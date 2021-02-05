/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/configuration"
	"github.com/iter8-tools/etc3/util"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var lg logr.Logger = ctrl.Log.WithName("etc3").WithName("test")
var recorder record.EventRecorder
var reconciler *ExperimentReconciler

type testHTTP struct {
	analysis *v2alpha1.Analysis
}

func (t *testHTTP) Post(url, contentType string, body []byte) ([]byte, int, error) {
	statuscode := 200
	b, err := json.Marshal(t.analysis)
	if err != nil {
		statuscode = 500
	}
	return b, statuscode, err
}

type testRecorder struct{}

func (r testRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	fmt.Printf("%s (%s): %s\n", eventtype, reason, message)
}
func (r testRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	fmt.Printf("%s (%s): %s\n", eventtype, reason, fmt.Sprintf(messageFmt, args...))
}
func (r testRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	fmt.Printf("%s (%s): %s\n", eventtype, reason, fmt.Sprintf(messageFmt, args...))
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	// logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
	ctrl.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = v2alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	iter8config := configuration.NewIter8Config().
		WithStrategy(string(v2alpha1.StrategyTypeCanary), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
		WithStrategy(string(v2alpha1.StrategyTypeAB), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
		WithStrategy(string(v2alpha1.StrategyTypeConformance), map[string]string{"start": "start"}).
		WithStrategy(string(v2alpha1.StrategyTypeBlueGreen), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
		WithRequestCount("request-count").
		WithEndpoint("http://iter8-analytics:8080").
		Build()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	testTransport := &testHTTP{
		analysis: &v2alpha1.Analysis{
			AggregatedMetrics: &v2alpha1.AggregatedMetricsAnalysis{
				AnalysisMetaData: v2alpha1.AnalysisMetaData{},
				Data:             map[string]v2alpha1.AggregatedMetricsData{},
			},
			WinnerAssessment: &v2alpha1.WinnerAssessmentAnalysis{
				AnalysisMetaData: v2alpha1.AnalysisMetaData{},
				Data:             v2alpha1.WinnerAssessmentData{},
			},
			VersionAssessments: &v2alpha1.VersionAssessmentAnalysis{
				AnalysisMetaData: v2alpha1.AnalysisMetaData{},
				Data:             map[string]v2alpha1.BooleanList{},
			},
			Weights: &v2alpha1.WeightsAnalysis{
				AnalysisMetaData: v2alpha1.AnalysisMetaData{},
				Data:             []v2alpha1.WeightData{},
			},
		},
	}

	recorder = testRecorder{}

	reconciler = &ExperimentReconciler{
		Client:        k8sClient,
		Log:           lg,
		Scheme:        k8sManager.GetScheme(),
		RestConfig:    cfg,
		EventRecorder: recorder,
		Iter8Config:   iter8config,
		HTTP:          testTransport,
		ReleaseEvents: make(chan event.GenericEvent),
	}

	Expect(reconciler.SetupWithManager(k8sManager)).Should(Succeed())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func isDeployed(name string, ns string) bool {
	exp := &v2alpha1.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}
	return true
}

func hasTarget(name string, ns string) bool {
	exp := &v2alpha1.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}

	return exp.Status.GetCondition(v2alpha1.ExperimentConditionTargetAcquired).IsTrue()
}

func completes(name string, ns string) bool {
	exp := &v2alpha1.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}

	return exp.Status.GetCondition(v2alpha1.ExperimentConditionExperimentCompleted).IsTrue()
}

func completesSuccessfully(name string, ns string) bool {
	exp := &v2alpha1.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}
	completed := exp.Status.GetCondition(v2alpha1.ExperimentConditionExperimentCompleted).IsTrue()
	successful := exp.Status.GetCondition(v2alpha1.ExperimentConditionExperimentFailed).IsFalse()

	return completed && successful
}

func isDeleted(name string, ns string) bool {
	exp := &v2alpha1.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	return err != nil &&
		(errors.IsNotFound(err) || errors.IsGone(err))
}

type check func(*v2alpha1.Experiment) bool

func hasValue(name string, ns string, check check) bool {
	exp := &v2alpha1.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}

	return check(exp)
}

func ctx() context.Context {
	return context.WithValue(context.Background(), util.LoggerKey, ctrl.Log)
}
