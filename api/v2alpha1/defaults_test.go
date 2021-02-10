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

package v2alpha1_test

import (
	"reflect"

	"github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Stages", func() {
	Context("When stages are compared", func() {
		It("Evaluates the order correctly", func() {
			Expect(v2alpha1.ExperimentStageCompleted.After(v2alpha1.ExperimentStageRunning)).Should(BeTrue())
			Expect(v2alpha1.ExperimentStageRunning.After(v2alpha1.ExperimentStageInitializing)).Should(BeTrue())
			Expect(v2alpha1.ExperimentStageInitializing.After(v2alpha1.ExperimentStageRunning)).Should(BeFalse())
		})
	})
})

var _ = Describe("Initialization", func() {
	Context("When initialize", func() {
		experiment := v2alpha1.NewExperiment("experiment", "namespace").
			WithTarget("target").
			WithTestingPattern(v2alpha1.TestingPatternCanary).
			WithHandlers(map[string]string{"start": "none", "finish": "none"}).
			WithRequestCount("request-count").
			Build()
		It("is initialized", func() {
			By("Initializing Status")
			experiment.InitializeStatus()
			By("Inspecting Status")
			Expect(experiment.Status.InitTime).ShouldNot(BeNil())
			Expect(experiment.Status.LastUpdateTime).ShouldNot(BeNil())
			Expect(experiment.Status.CompletedIterations).ShouldNot(BeNil())
			Expect(len(experiment.Status.Conditions)).Should(Equal(3))

			By("Initializing Spec")
			c := configuration.NewIter8Config().
				WithTestingPattern(string(v2alpha1.TestingPatternCanary), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
				WithTestingPattern(string(v2alpha1.TestingPatternAB), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
				WithTestingPattern(string(v2alpha1.TestingPatternConformance), map[string]string{"start": "start"}).
				WithRequestCount("request-count").
				WithEndpoint("http://iter8-analytics:8080").
				Build()
			experiment.Spec.InitializeSpec(c)
			Expect(experiment.Spec.GetMaxIterations()).Should(Equal(v2alpha1.DefaultMaxIterations))
			Expect(experiment.Spec.GetIntervalSeconds()).Should(Equal(int32(v2alpha1.DefaultIntervalSeconds)))
			Expect(experiment.Spec.GetMaxCandidateWeight()).Should(Equal(v2alpha1.DefaultMaxCandidateWeight))
			Expect(experiment.Spec.GetMaxCandidateWeightIncrement()).Should(Equal(v2alpha1.DefaultMaxCandidateWeightIncrement))
			Expect(experiment.Spec.GetDeploymentPattern()).Should(Equal(v2alpha1.DefaultDeploymentPattern))
			// Expect(len(experiment.Spec.Metrics)).Should(Equal(1))
			Expect(*experiment.Spec.GetRequestCount(configuration.Iter8Config{})).Should(Equal("request-count"))
		})
	})
})

var _ = Describe("Handler Initialization", func() {
	var (
		experiment *v2alpha1.Experiment
		cfg        configuration.Iter8Config
	)
	Context("When hanlders are set in an experiment", func() {
		BeforeEach(func() {
			experiment = v2alpha1.NewExperiment("test", "default").
				WithTestingPattern(v2alpha1.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "expStart", "finish": "expFinish"}).
				Build()
		})

		It("the value of the start handler should match the value in experiment when a configuration is present", func() {
			cfg = configuration.NewIter8Config().
				WithTestingPattern(string(v2alpha1.TestingPatternCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()

			Expect(experiment.Spec.Strategy.Handlers).ShouldNot(BeNil())
			Expect(experiment.Spec.Strategy.Handlers.Start).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("expStart"))
			Expect(experiment.Spec.Strategy.Handlers.Finish).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetFinishHandler(cfg)).Should(Equal("expFinish"))
			Expect(experiment.Spec.GetRollbackHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFailureHandler(cfg)).Should(BeNil())
		})
		It("the value of GetStartHandler should should be nil when set to none even if a configuration is present", func() {
			cfg = configuration.NewIter8Config().
				WithTestingPattern(string(v2alpha1.TestingPatternCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()
			none := v2alpha1.NoneHandler
			experiment.Spec.Strategy.Handlers.Start = &none

			Expect(experiment.Spec.Strategy.Handlers).ShouldNot(BeNil())
			Expect(experiment.Spec.Strategy.Handlers.Start).ShouldNot(BeNil())
			Expect(experiment.Spec.GetStartHandler(cfg)).Should(BeNil())
		})
		It("the value of the start handler should match the value in experiment when ca onfiguration is not present", func() {
			cfg = configuration.NewIter8Config().
				Build()
			Expect(experiment.Spec.Strategy.Handlers).ShouldNot(BeNil())
			Expect(experiment.Spec.Strategy.Handlers.Start).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("expStart"))
			Expect(experiment.Spec.Strategy.Handlers.Finish).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetFinishHandler(cfg)).Should(Equal("expFinish"))
			Expect(experiment.Spec.GetRollbackHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFailureHandler(cfg)).Should(BeNil())
		})
	})

	Context("When handlers are not defined in an experiment", func() {
		experiment := v2alpha1.NewExperiment("test", "default").
			WithTestingPattern(v2alpha1.TestingPatternCanary).
			Build()

		It("the value is set by late initialization from the configuration", func() {
			experiment := v2alpha1.NewExperiment("test", "default").
				WithTestingPattern(v2alpha1.TestingPatternCanary).
				Build()
			cfg := configuration.NewIter8Config().
				WithTestingPattern(string(v2alpha1.TestingPatternCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()
			experiment.Spec.InitializeHandlers(cfg)
			Expect(experiment.Spec.Strategy.Handlers).ShouldNot(BeNil())
			Expect(experiment.Spec.Strategy.Handlers.Start).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("cfgStart"))
		})

		It("the value of the start handler shoud match the value im the configuration when it is present", func() {
			cfg := configuration.NewIter8Config().
				WithTestingPattern(string(v2alpha1.TestingPatternCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()
			Expect(experiment.Spec.Strategy.Handlers).Should(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("cfgStart"))
			Expect(*experiment.Spec.GetFinishHandler(cfg)).Should(Equal("cfgFinish"))
			Expect(experiment.Spec.GetRollbackHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFailureHandler(cfg)).Should(BeNil())
		})

		It("the value of the start handler shoud be nil when when no configuration is present", func() {
			cfg := configuration.NewIter8Config().
				Build()
			Expect(experiment.Spec.Strategy.Handlers).Should(BeNil())
			Expect(experiment.Spec.GetStartHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFinishHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetRollbackHandler(cfg)).Should(BeNil())
			Expect(experiment.Spec.GetFailureHandler(cfg)).Should(BeNil())
		})
	})
})

var _ = Describe("Generated Code", func() {
	Context("When copy object", func() {
		It("the copy should be the same", func() {
			experiment := v2alpha1.NewExperiment("test", "default").
				WithTarget("copy").
				WithTestingPattern(v2alpha1.TestingPatternCanary).
				WithDeploymentPattern(v2alpha1.DeploymentPatternFixedSplit).
				WithHandlers(map[string]string{"start": "st", "finish": "fin", "failure": "fail", "rollback": "back"}).
				WithDuration(3, 2).
				WithRequestCount("request-count").
				WithBaselineVersion("baseline", &corev1.ObjectReference{
					Kind:       "kind",
					Namespace:  "namespace",
					Name:       "name",
					APIVersion: "apiVersion",
					FieldPath:  "path",
				}).
				WithCandidateVersion("candidate", nil).
				WithCurrentWeight("baseline", 25).WithCurrentWeight("candidate", 75).
				WithRecommendedWeight("baseline", 0).WithRecommendedWeight("candidate", 100).
				WithCondition(v2alpha1.ExperimentConditionExperimentFailed, corev1.ConditionTrue, v2alpha1.ReasonHandlerFailed, "foo %s", "bar").
				Build()
			now := metav1.Now()
			message := "message"
			winner := "winner"
			q := resource.Quantity{}
			ss := int32(1)
			experiment.Status.Analysis = &v2alpha1.Analysis{
				AggregatedMetrics: &v2alpha1.AggregatedMetricsAnalysis{
					AnalysisMetaData: v2alpha1.AnalysisMetaData{
						Provenance: "provenance",
						Timestamp:  now,
						Message:    &message,
					},
					Data: map[string]v2alpha1.AggregatedMetricsData{
						"metric1": {
							Max: &q,
							Min: &q,
							Data: map[string]v2alpha1.AggregatedMetricsVersionData{
								"metric": {
									Min:        &q,
									Max:        &q,
									Value:      &q,
									SampleSize: &ss,
								},
							},
						},
					},
				},
				WinnerAssessment: &v2alpha1.WinnerAssessmentAnalysis{
					AnalysisMetaData: v2alpha1.AnalysisMetaData{},
					Data: v2alpha1.WinnerAssessmentData{
						WinnerFound: true,
						Winner:      &winner,
					},
				},
				VersionAssessments: &v2alpha1.VersionAssessmentAnalysis{
					AnalysisMetaData: v2alpha1.AnalysisMetaData{},
					Data: map[string]v2alpha1.BooleanList{
						"baseline":  []bool{false},
						"candidate": []bool{false},
					},
				},
				Weights: &v2alpha1.WeightsAnalysis{
					AnalysisMetaData: v2alpha1.AnalysisMetaData{},
					Data: []v2alpha1.WeightData{
						{Name: "baseline", Value: 25},
						{Name: "candidate", Value: 75},
					},
				},
			}
			Expect(reflect.DeepEqual(experiment, experiment.DeepCopyObject())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment, experiment.DeepCopy())).Should(BeTrue())
			// Expect(reflect.DeepEqual(experiment.Status, experiment.Status.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis, experiment.Status.Analysis.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis.WinnerAssessment, experiment.Status.Analysis.WinnerAssessment.DeepCopy())).Should(BeTrue())
			// Expect(reflect.DeepEqual(experiment.Status.Analysis.WinnerAssessment.Data, experiment.Status.Analysis.WinnerAssessment.Data.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis.VersionAssessments, experiment.Status.Analysis.VersionAssessments.DeepCopy())).Should(BeTrue())
			//Expect(reflect.DeepEqual(experiment.Status.Analysis.VersionAssessments.Data, experiment.Status.Analysis.VersionAssessments.Data.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis.Weights, experiment.Status.Analysis.Weights.DeepCopy())).Should(BeTrue())
			//Expect(reflect.DeepEqual(experiment.Status.Analysis.Weights.Data, experiment.Status.Analysis.Weights.Data.DeepCopy())).Should(BeTrue())
		})
	})
})
