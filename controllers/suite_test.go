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
	"strings"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
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
var events []string

type testHTTP struct {
	analysis *v2alpha2.Analysis
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
	events = append(events, fmt.Sprintf("%s (%s): %s\n", eventtype, reason, message))
}
func (r testRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	events = append(events, fmt.Sprintf("%s (%s): %s\n", eventtype, reason, fmt.Sprintf(messageFmt, args...)))
}
func (r testRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	events = append(events, fmt.Sprintf("%s (%s): %s\n", eventtype, reason, fmt.Sprintf(messageFmt, args...)))
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

	err = v2alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	iter8config := configuration.NewIter8Config().
		WithTestingPattern(string(v2alpha2.TestingPatternCanary), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
		WithTestingPattern(string(v2alpha2.TestingPatternAB), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
		WithTestingPattern(string(v2alpha2.TestingPatternConformance), map[string]string{"start": "start"}).
		WithRequestCount("request-count").
		WithEndpoint("http://iter8-analytics:8080").
		WithHandlersDir("../test/handlers").
		WithNamespace("iter8").
		Build()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	// Create iter8 namespace for use by some tests
	Expect(k8sClient.Create(ctx(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "iter8"},
	})).Should(Succeed())
	Expect(k8sClient.Create(ctx(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "metric-namespace"},
	})).Should(Succeed())

	testTransport := &testHTTP{
		analysis: &v2alpha2.Analysis{
			AggregatedMetrics: &v2alpha2.AggregatedMetricsAnalysis{
				AnalysisMetaData: v2alpha2.AnalysisMetaData{},
				Data:             map[string]v2alpha2.AggregatedMetricsData{},
			},
			WinnerAssessment: &v2alpha2.WinnerAssessmentAnalysis{
				AnalysisMetaData: v2alpha2.AnalysisMetaData{},
				Data:             v2alpha2.WinnerAssessmentData{},
			},
			VersionAssessments: &v2alpha2.VersionAssessmentAnalysis{
				AnalysisMetaData: v2alpha2.AnalysisMetaData{},
				Data:             map[string]v2alpha2.BooleanList{},
			},
			Weights: &v2alpha2.WeightsAnalysis{
				AnalysisMetaData: v2alpha2.AnalysisMetaData{},
				Data:             []v2alpha2.WeightData{},
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
	exp := &v2alpha2.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}
	return true
}

func hasTarget(name string, ns string) bool {
	exp := &v2alpha2.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}

	return exp.Status.GetCondition(v2alpha2.ExperimentConditionTargetAcquired).IsTrue()
}

func completes(name string, ns string) bool {
	exp := &v2alpha2.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}
	return exp.Status.GetCondition(v2alpha2.ExperimentConditionExperimentCompleted).IsTrue()
}

func completesSuccessfully(name string, ns string) bool {
	exp := &v2alpha2.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}
	completed := exp.Status.GetCondition(v2alpha2.ExperimentConditionExperimentCompleted).IsTrue()
	successful := exp.Status.GetCondition(v2alpha2.ExperimentConditionExperimentFailed).IsFalse()

	return completed && successful
}

func fails(name string, ns string) bool {
	exp := &v2alpha2.Experiment{}
	err := k8sClient.Get(ctx(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}
	completed := exp.Status.GetCondition(v2alpha2.ExperimentConditionExperimentCompleted).IsTrue()
	failed := exp.Status.GetCondition(v2alpha2.ExperimentConditionExperimentFailed).IsTrue()

	return completed && failed
}

func issuedEvent(message string) bool {
	return containsSubString(events, message)
}

func isDeleted(name string, ns string) bool {
	exp := &v2alpha2.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	return err != nil &&
		(errors.IsNotFound(err) || errors.IsGone(err))
}

type check func(*v2alpha2.Experiment) bool

func hasValue(name string, ns string, check check) bool {
	exp := &v2alpha2.Experiment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: ns}, exp)
	if err != nil {
		return false
	}
	return check(exp)
}

func isRunning(name string, ns string) bool {
	return hasValue(name, ns, func(exp *v2alpha2.Experiment) bool {
		return exp.Status.Stage != nil && *exp.Status.Stage == v2alpha2.ExperimentStageRunning
	})
}

func ctx() context.Context {
	return context.WithValue(context.Background(), util.LoggerKey, ctrl.Log)
}

// Helper functions to check and remove string from a slice of strings.
func containsSubString(slice []string, s string) bool {
	for _, item := range slice {
		if strings.Contains(item, s) {
			return true
		}
	}
	return false
}
