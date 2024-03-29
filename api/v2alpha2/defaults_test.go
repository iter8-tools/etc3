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
	"time"

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
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
	Context("Before Initialization", func() {
		experiment := v2alpha2.NewExperiment("experiment", "namespace").
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			Build()
		Specify("status values should be unset", func() {
			Expect(experiment.Status.InitTime).Should(BeNil())
			Expect(experiment.Status.LastUpdateTime).Should(BeNil())
			Expect(experiment.Status.CompletedIterations).Should(BeNil())
			Expect(len(experiment.Status.Conditions)).Should(Equal(0))
		})
		Specify("methods on spec should handle nil gracefully", func() {
			Expect(experiment.Spec.GetIterationsPerLoop()).Should(Equal(v2alpha2.DefaultIterationsPerLoop))
			Expect(experiment.Spec.GetMaxLoops()).Should(Equal(v2alpha2.DefaultMaxLoops))
			Expect(experiment.Spec.GetIntervalSeconds()).Should(Equal(int32(v2alpha2.DefaultIntervalSeconds)))
			Expect(experiment.Spec.GetIntervalAsDuration()).Should(Equal(time.Second * time.Duration(experiment.Spec.GetIntervalSeconds())))
			Expect(experiment.Spec.GetMaxCandidateWeight()).Should(Equal(v2alpha2.DefaultMaxCandidateWeight))
			Expect(experiment.Spec.GetMaxCandidateWeightIncrement()).Should(Equal(v2alpha2.DefaultMaxCandidateWeightIncrement))
			Expect(experiment.Spec.GetDeploymentPattern()).Should(Equal(v2alpha2.DefaultDeploymentPattern))
			Expect(experiment.Spec.GetRequestCount()).Should(BeNil())
			Expect(*experiment.Spec.GetStartHandler()).Should(Equal(v2alpha2.DefaultStartHandler))
			Expect(*experiment.Spec.GetFinishHandler()).Should(Equal(v2alpha2.DefaultFinishHandler))
			Expect(*experiment.Spec.GetRollbackHandler()).Should(Equal(v2alpha2.DefaultRollbackHandler))
			Expect(*experiment.Spec.GetFailureHandler()).Should(Equal(v2alpha2.DefaultFailureHandler))
			Expect(*experiment.Spec.GetLoopHandler()).Should(Equal(v2alpha2.DefaultLoopHandler))
		})
	})

	Context("After Initialization", func() {
		experiment := v2alpha2.NewExperiment("experiment", "namespace").
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
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
			Expect(experiment.Status.GetCondition(v2alpha2.ExperimentConditionExperimentCompleted).IsTrue()).Should(Equal(false))
			Expect(experiment.Status.GetCondition(v2alpha2.ExperimentConditionExperimentCompleted).IsFalse()).Should(Equal(true))
			Expect(experiment.Status.GetCondition(v2alpha2.ExperimentConditionExperimentCompleted).IsUnknown()).Should(Equal(false))

			By("Initializing Spec")
			experiment.Spec.InitializeSpec()
			Expect(experiment.Spec.GetIterationsPerLoop()).Should(Equal(v2alpha2.DefaultIterationsPerLoop))
			Expect(experiment.Spec.GetMaxLoops()).Should(Equal(v2alpha2.DefaultMaxLoops))
			Expect(experiment.Spec.GetIntervalSeconds()).Should(Equal(int32(v2alpha2.DefaultIntervalSeconds)))
			Expect(experiment.Spec.GetIntervalAsDuration()).Should(Equal(time.Second * time.Duration(experiment.Spec.GetIntervalSeconds())))
			Expect(experiment.Spec.GetMaxCandidateWeight()).Should(Equal(v2alpha2.DefaultMaxCandidateWeight))
			Expect(experiment.Spec.GetMaxCandidateWeightIncrement()).Should(Equal(v2alpha2.DefaultMaxCandidateWeightIncrement))
			Expect(experiment.Spec.GetDeploymentPattern()).Should(Equal(v2alpha2.DefaultDeploymentPattern))
			Expect(*experiment.Spec.GetRequestCount()).Should(Equal("request-count"))
			Expect(*experiment.Spec.GetStartHandler()).Should(Equal(v2alpha2.DefaultStartHandler))
			Expect(*experiment.Spec.GetFinishHandler()).Should(Equal(v2alpha2.DefaultFinishHandler))
			Expect(*experiment.Spec.GetRollbackHandler()).Should(Equal(v2alpha2.DefaultRollbackHandler))
			Expect(*experiment.Spec.GetFailureHandler()).Should(Equal(v2alpha2.DefaultFailureHandler))
			Expect(*experiment.Spec.GetLoopHandler()).Should(Equal(v2alpha2.DefaultLoopHandler))
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
	var jqe string = "expr"

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
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(experiment.Spec.Criteria).ShouldNot(BeNil())
			Expect(experiment.Spec.Criteria.Rewards).ShouldNot(BeEmpty())
		})
	})
})

var _ = Describe("Generated Code", func() {
	var jqe string = "expr"

	Context("When a Metric object is copied", func() {
		Specify("the copy should be the same as the original", func() {
			metricBuilder := v2alpha2.NewMetric("reward", "default").
				WithDescription("reward metric").
				WithParams([]v2alpha2.NamedValue{{
					Name:  "query",
					Value: "query",
				}}).
				WithProvider("prometheus").
				WithJQExpression(&jqe).
				WithType(v2alpha2.CounterMetricType).
				WithUnits("ms").
				WithSampleSize("sample/default")
			metric := metricBuilder.Build()
			metricList := v2alpha2.MetricList{
				Items: []v2alpha2.Metric{*metric},
			}

			Expect(reflect.DeepEqual(metricBuilder, metricBuilder.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(metric, metric.DeepCopyObject())).Should(BeTrue())
			Expect(len(metricList.Items)).Should(Equal(len(metricList.DeepCopy().Items)))
		})
	})

	Context("When an Experiment object is copied", func() {
		Specify("the copy should be the same as the original", func() {
			testStr := "test"
			experimentBuilder := v2alpha2.NewExperiment("test", "default").
				WithTarget("copy").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDeploymentPattern(v2alpha2.DeploymentPatternFixedSplit).
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
				WithAction("start", []v2alpha2.TaskSpec{
					{Task: &testStr},
					{Run: &testStr},
				}).
				WithRequestCount("request-count").
				WithReward(*v2alpha2.NewMetric("reward", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithIndicator(*v2alpha2.NewMetric("indicator", "default").WithJQExpression(&jqe).Build()).
				WithObjective(*v2alpha2.NewMetric("reward", "default").WithJQExpression(&jqe).Build(), nil, nil, false)
			experiment := experimentBuilder.Build()
			experiment.InitializeStatus()
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
			experimentList := v2alpha2.ExperimentList{
				Items: []v2alpha2.Experiment{*experiment},
			}

			Expect(reflect.DeepEqual(experimentBuilder, experimentBuilder.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment, experiment.DeepCopyObject())).Should(BeTrue())
			// Expect(reflect.DeepEqual(experimentList, experimentList.DeepCopyObject())).Should(BeTrue())
			Expect(len(experimentList.Items)).Should(Equal(len(experimentList.DeepCopy().Items)))

			// Expect(reflect.DeepEqual(experiment.Spec, experiment.Spec.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Criteria, experiment.Spec.Criteria.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Duration, experiment.Spec.Duration.DeepCopy())).Should(BeTrue())
			// Expect(reflect.DeepEqual(experiment.Spec.Strategy, experiment.Spec.Strategy.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Strategy.Weights, experiment.Spec.Strategy.Weights.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Strategy.Actions, experiment.Spec.Strategy.Actions.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Strategy.Actions["start"], experiment.Spec.Strategy.Actions["start"].DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.VersionInfo, experiment.Spec.VersionInfo.DeepCopy())).Should(BeTrue())
			// Expect(reflect.DeepEqual(experiment.Spec.VersionInfo.Baseline, experiment.Spec.VersionInfo.Baseline.DeepCopy())).Should(BeTrue())

			// Expect(reflect.DeepEqual(experiment.Status, experiment.Status.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis, experiment.Status.Analysis.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis.AggregatedBuiltinHists, experiment.Status.Analysis.AggregatedBuiltinHists.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis.AggregatedMetrics, experiment.Status.Analysis.AggregatedMetrics.DeepCopy())).Should(BeTrue())
			// Expect(reflect.DeepEqual(experiment.Status.Analysis.AggregatedMetrics.AnalysisMetaData, experiment.Status.Analysis.AggregatedMetrics.AnalysisMetaData.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis.VersionAssessments, experiment.Status.Analysis.VersionAssessments.DeepCopy())).Should(BeTrue())
			// Expect(reflect.DeepEqual(experiment.Status.Analysis.VersionAssessments, experiment.Status.Analysis.Weights.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis.WinnerAssessment, experiment.Status.Analysis.WinnerAssessment.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Conditions[0], experiment.Status.Conditions[0].DeepCopy())).Should(BeTrue())
		})
	})
})
