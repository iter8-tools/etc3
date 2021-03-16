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

package v2alpha2_test

import (
	"reflect"

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Stages", func() {
	Context("When stages are compared", func() {
		It("Evaluates the order correctly", func() {
			Expect(v2alpha2.ExperimentStageCompleted.After(v2alpha2.ExperimentStageRunning)).Should(BeTrue())
			Expect(v2alpha2.ExperimentStageRunning.After(v2alpha2.ExperimentStageInitializing)).Should(BeTrue())
			Expect(v2alpha2.ExperimentStageInitializing.After(v2alpha2.ExperimentStageRunning)).Should(BeFalse())
		})
	})
})

var _ = Describe("Initialization", func() {
	Context("When initialize", func() {
		experiment := v2alpha2.NewExperiment("experiment", "namespace").
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
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
				WithTestingPattern(string(v2alpha2.TestingPatternCanary), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
				WithTestingPattern(string(v2alpha2.TestingPatternAB), map[string]string{"start": "start", "finish": "finish", "rollback": "finish", "failure": "finish"}).
				WithTestingPattern(string(v2alpha2.TestingPatternConformance), map[string]string{"start": "start"}).
				WithRequestCount("request-count").
				WithEndpoint("http://iter8-analytics:8080").
				Build()
			experiment.Spec.InitializeSpec(c)
			Expect(experiment.Spec.GetIterationsPerLoop()).Should(Equal(v2alpha2.DefaultIterationsPerLoop))
			Expect(experiment.Spec.GetMaxLoops()).Should(Equal(v2alpha2.DefaultMaxLoops))
			Expect(experiment.Spec.GetIntervalSeconds()).Should(Equal(int32(v2alpha2.DefaultIntervalSeconds)))
			Expect(experiment.Spec.GetMaxCandidateWeight()).Should(Equal(v2alpha2.DefaultMaxCandidateWeight))
			Expect(experiment.Spec.GetMaxCandidateWeightIncrement()).Should(Equal(v2alpha2.DefaultMaxCandidateWeightIncrement))
			Expect(experiment.Spec.GetDeploymentPattern()).Should(Equal(v2alpha2.DefaultDeploymentPattern))
			// Expect(len(experiment.Spec.Metrics)).Should(Equal(1))
			Expect(*experiment.Spec.GetRequestCount(configuration.Iter8Config{})).Should(Equal("request-count"))
		})
	})
})

var _ = Describe("Handler Initialization", func() {
	var (
		experiment *v2alpha2.Experiment
		cfg        configuration.Iter8Config
	)
	Context("When hanlders are set in an experiment", func() {
		BeforeEach(func() {
			experiment = v2alpha2.NewExperiment("test", "default").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "expStart", "finish": "expFinish"}).
				Build()
		})

		It("the value of the start handler should match the value in experiment when a configuration is present", func() {
			cfg = configuration.NewIter8Config().
				WithTestingPattern(string(v2alpha2.TestingPatternCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
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
				WithTestingPattern(string(v2alpha2.TestingPatternCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()
			none := v2alpha2.NoneHandler
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
		experiment := v2alpha2.NewExperiment("test", "default").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			Build()

		It("the value is set by late initialization from the configuration", func() {
			experiment := v2alpha2.NewExperiment("test", "default").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				Build()
			cfg := configuration.NewIter8Config().
				WithTestingPattern(string(v2alpha2.TestingPatternCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
				Build()
			experiment.Spec.InitializeHandlers(cfg)
			Expect(experiment.Spec.Strategy.Handlers).ShouldNot(BeNil())
			Expect(experiment.Spec.Strategy.Handlers.Start).ShouldNot(BeNil())
			Expect(*experiment.Spec.GetStartHandler(cfg)).Should(Equal("cfgStart"))
		})

		It("the value of the start handler shoud match the value im the configuration when it is present", func() {
			cfg := configuration.NewIter8Config().
				WithTestingPattern(string(v2alpha2.TestingPatternCanary), map[string]string{"start": "cfgStart", "finish": "cfgFinish"}).
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

var _ = Describe("VersionInfo", func() {
	Context("When count versions", func() {
		builder := v2alpha2.NewExperiment("test", "default").WithTarget("target")
		It("should count correctly", func() {
			experiment := builder.DeepCopy().Build()
			Expect(experiment.Spec.GetNumberOfBaseline()).Should(Equal(0))
			Expect(experiment.Spec.GetNumberOfCandidates()).Should(Equal(0))

			experiment = builder.DeepCopy().
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(experiment.Spec.GetNumberOfBaseline()).Should(Equal(1))
			Expect(experiment.Spec.GetNumberOfCandidates()).Should(Equal(0))

			experiment = builder.DeepCopy().
				WithCandidateVersion("candidate", nil).
				Build()
				//			Expect(experiment.Spec.GetNumberOfBaseline()).Should(Equal(0))
			Expect(experiment.Spec.GetNumberOfCandidates()).Should(Equal(1))

			experiment = builder.DeepCopy().
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate", nil).
				Build()
			Expect(experiment.Spec.GetNumberOfBaseline()).Should(Equal(1))
			Expect(experiment.Spec.GetNumberOfCandidates()).Should(Equal(1))

			experiment = builder.DeepCopy().
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate", nil).
				WithCandidateVersion("c", nil).
				Build()
			Expect(experiment.Spec.GetNumberOfBaseline()).Should(Equal(1))
			Expect(experiment.Spec.GetNumberOfCandidates()).Should(Equal(2))
		})
	})
})

var _ = Describe("Criteria", func() {
	Context("Criteria", func() {
		builder := v2alpha2.NewExperiment("test", "default").WithTarget("target")
		It("", func() {
			experiment := builder.DeepCopy().Build()
			Expect(experiment.Spec.Criteria).Should(BeNil())

			experiment = builder.DeepCopy().
				WithIndicator(*v2alpha2.NewMetric("metric", "default").Build()).
				Build()
			Expect(experiment.Spec.Criteria).ShouldNot(BeNil())
			Expect(experiment.Spec.Criteria.Rewards).Should(BeEmpty())

			experiment = builder.DeepCopy().
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression("expr").Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(experiment.Spec.Criteria).ShouldNot(BeNil())
			Expect(experiment.Spec.Criteria.Rewards).ShouldNot(BeEmpty())
		})
	})
})

var _ = Describe("Generated Code", func() {
	Context("When a Metric object is copied", func() {
		Specify("the copy should be the same as the original", func() {
			metricBuilder := v2alpha2.NewMetric("reward", "default").
				WithDescription("reward metric").
				WithParams(map[string]string{"query": "query"}).
				WithProvider("prometheus").
				WithJQExpression("expr").
				WithType(v2alpha2.CounterMetricType).
				WithUnits("ms").
				WithSampleSize("sample/default")
			metric := metricBuilder.Build()
			metricList := *&v2alpha2.MetricList{
				Items: []v2alpha2.Metric{*metric},
			}

			Expect(reflect.DeepEqual(metricBuilder, metricBuilder.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(metric, metric.DeepCopyObject())).Should(BeTrue())
			Expect(len(metricList.Items)).Should(Equal(len(metricList.DeepCopy().Items)))
		})
	})

	Context("When an Experiment object is copied", func() {
		Specify("the copy should be the same as the original", func() {
			experimentBuilder := v2alpha2.NewExperiment("test", "default").
				WithTarget("copy").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDeploymentPattern(v2alpha2.DeploymentPatternFixedSplit).
				WithHandlers(map[string]string{"start": "st", "finish": "fin", "failure": "fail", "rollback": "back", "loop": "lo"}).
				WithDuration(3, 2, 1).
				WithBaselineVersion("baseline", nil).
				WithBaselineVersion("baseline", &corev1.ObjectReference{
					Kind:       "kind",
					Namespace:  "namespace",
					Name:       "name",
					APIVersion: "apiVersion",
					FieldPath:  "path",
				}).
				WithCandidateVersion("candidate", nil).WithCandidateVersion("candidate", nil).
				WithCurrentWeight("baseline", 25).WithCurrentWeight("candidate", 75).
				WithRecommendedWeight("baseline", 0).WithRecommendedWeight("candidate", 100).
				WithCurrentWeight("baseline", 30).WithRecommendedWeight("baseline", 10).
				WithCondition(v2alpha2.ExperimentConditionExperimentFailed, corev1.ConditionTrue, v2alpha2.ReasonHandlerFailed, "foo %s", "bar").
				WithAction("start", []v2alpha2.TaskSpec{{Task: "task"}}).
				WithRequestCount("request-count").
				WithReward(*v2alpha2.NewMetric("reward", "default").WithJQExpression("expr").Build(), v2alpha2.PreferredDirectionHigher).
				WithIndicator(*v2alpha2.NewMetric("indicator", "default").WithJQExpression("expr").Build()).
				WithObjective(*v2alpha2.NewMetric("reward", "default").WithJQExpression("expr").Build(), nil, nil, false)
			experiment := experimentBuilder.Build()
			now := metav1.Now()
			message := "message"
			winner := "winner"
			q := resource.Quantity{}
			ss := int32(1)
			experiment.Status.Analysis = &v2alpha2.Analysis{
				AggregatedMetrics: &v2alpha2.AggregatedMetricsAnalysis{
					AnalysisMetaData: v2alpha2.AnalysisMetaData{
						Provenance: "provenance",
						Timestamp:  now,
						Message:    &message,
					},
					Data: map[string]v2alpha2.AggregatedMetricsData{
						"metric1": {
							Max: &q,
							Min: &q,
							Data: map[string]v2alpha2.AggregatedMetricsVersionData{
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
				WinnerAssessment: &v2alpha2.WinnerAssessmentAnalysis{
					AnalysisMetaData: v2alpha2.AnalysisMetaData{},
					Data: v2alpha2.WinnerAssessmentData{
						WinnerFound: true,
						Winner:      &winner,
					},
				},
				VersionAssessments: &v2alpha2.VersionAssessmentAnalysis{
					AnalysisMetaData: v2alpha2.AnalysisMetaData{},
					Data: map[string]v2alpha2.BooleanList{
						"baseline":  []bool{false},
						"candidate": []bool{false},
					},
				},
				Weights: &v2alpha2.WeightsAnalysis{
					AnalysisMetaData: v2alpha2.AnalysisMetaData{},
					Data: []v2alpha2.WeightData{
						{Name: "baseline", Value: 25},
						{Name: "candidate", Value: 75},
					},
				},
			}
			experimentList := *&v2alpha2.ExperimentList{
				Items: []v2alpha2.Experiment{*experiment},
			}

			Expect(reflect.DeepEqual(experimentBuilder, experimentBuilder.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment, experiment.DeepCopyObject())).Should(BeTrue())
			Expect(len(experimentList.Items)).Should(Equal(len(experimentList.DeepCopy().Items)))
		})
	})
})
