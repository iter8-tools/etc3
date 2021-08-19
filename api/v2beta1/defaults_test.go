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

package v2beta1

import (
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Stages", func() {
	Context("When stages are compared", func() {
		It("Evaluates the order correctly", func() {
			Expect(ExperimentStageCompleted.After(ExperimentStageRunning)).Should(BeTrue())
			Expect(ExperimentStageRunning.After(ExperimentStageInitializing)).Should(BeTrue())
			Expect(ExperimentStageInitializing.After(ExperimentStageRunning)).Should(BeFalse())
		})
	})
})

var _ = Describe("Initialization", func() {
	Context("Before Initialization", func() {
		experiment := NewExperiment("experiment", "namespace").
			WithVersion("baseline").WithVersion("candidate").
			Build()
		Specify("status values should be unset", func() {
			Expect(experiment.Status.StartTime).Should(BeNil())
			Expect(experiment.Status.LastUpdateTime).Should(BeNil())
			Expect(experiment.Status.CompletedLoops).Should(BeNil())
			Expect(len(experiment.Status.Conditions)).Should(Equal(0))
			Expect(experiment.Status.TestingPattern).Should(BeNil())
		})
		Specify("methods on spec should handle nil gracefully", func() {
			Expect(experiment.Spec.GetMaxLoops()).Should(Equal(DefaultMaxLoops))
			Expect(experiment.Spec.GetIntervalSeconds()).Should(Equal(int32(DefaultMinIntervalBetweenLoops)))
			Expect(experiment.Spec.GetIntervalAsDuration()).Should(Equal(time.Second * time.Duration(experiment.Spec.GetIntervalSeconds())))
			Expect(*experiment.Spec.GetStartHandler()).Should(Equal(DefaultStartHandler))
			Expect(*experiment.Spec.GetFinishHandler()).Should(Equal(DefaultFinishHandler))
			Expect(*experiment.Spec.GetFailureHandler()).Should(Equal(DefaultFailureHandler))
			Expect(*experiment.Spec.GetLoopHandler()).Should(Equal(DefaultLoopHandler))
		})
	})

	Context("After Initialization", func() {
		experiment := NewExperiment("experiment", "namespace").
			WithVersion("baseline").WithVersion("candidate").
			Build()
		It("is initialized", func() {
			By("Initializing Status")
			experiment.InitializeStatus()
			By("Inspecting Status")
			Expect(experiment.Status.StartTime).ShouldNot(BeNil())
			Expect(experiment.Status.LastUpdateTime).ShouldNot(BeNil())
			Expect(experiment.Status.CompletedLoops).ShouldNot(BeNil())
			Expect(len(experiment.Status.Conditions)).Should(Equal(2))
			Expect(experiment.Status.GetCondition(ExperimentConditionExperimentCompleted).IsTrue()).Should(Equal(false))
			Expect(experiment.Status.GetCondition(ExperimentConditionExperimentCompleted).IsFalse()).Should(Equal(true))
			Expect(experiment.Status.GetCondition(ExperimentConditionExperimentCompleted).IsUnknown()).Should(Equal(false))
			Expect(*experiment.Status.TestingPattern).Should(Equal(TestingPatternSLOValidation))

			By("Initializing Spec")
			experiment.Spec.InitializeSpec()
			Expect(experiment.Spec.GetMaxLoops()).Should(Equal(DefaultMaxLoops))
			Expect(experiment.Spec.GetIntervalSeconds()).Should(Equal(int32(DefaultMinIntervalBetweenLoops)))
			Expect(experiment.Spec.GetIntervalAsDuration()).Should(Equal(time.Second * time.Duration(experiment.Spec.GetIntervalSeconds())))
			Expect(*experiment.Spec.GetStartHandler()).Should(Equal(DefaultStartHandler))
			Expect(*experiment.Spec.GetFinishHandler()).Should(Equal(DefaultFinishHandler))
			Expect(*experiment.Spec.GetFailureHandler()).Should(Equal(DefaultFailureHandler))
			Expect(*experiment.Spec.GetLoopHandler()).Should(Equal(DefaultLoopHandler))
		})
	})
})

var _ = Describe("Criteria", func() {
	Context("Criteria", func() {
		builder := NewExperiment("test", "default").
			WithVersion("baseline").WithVersion("candidate")
		It("", func() {
			experiment := builder.DeepCopy().Build()
			Expect(experiment.Spec.Criteria).Should(BeNil())

			experiment = builder.DeepCopy().
				WithReward("default/metric", PreferredDirectionHigher).
				Build()
			Expect(experiment.Spec.Criteria).ShouldNot(BeNil())
			Expect(experiment.Spec.Criteria.Rewards).ShouldNot(BeEmpty())
		})
	})
})

var _ = Describe("Generated Code", func() {
	Context("When an Experiment object is copied", func() {
		Specify("the copy should be the same as the original", func() {
			testStr := "test"
			ifStr := "conditional-expression"
			experimentBuilder := NewExperiment("test", "default").
				WithVersion("baseline").WithVersion("candidate").
				WithDuration(3, 2).
				WithCurrentWeight("baseline", 25).WithCurrentWeight("candidate", 75).
				WithRecommendedWeight("baseline", 0).WithRecommendedWeight("candidate", 100).
				WithCurrentWeight("baseline", 30).WithRecommendedWeight("baseline", 10).
				WithCondition(ExperimentConditionExperimentFailed, corev1.ConditionTrue, ReasonHandlerFailed, "foo %s", "bar").
				WithAction("start", []TaskSpec{
					{Task: &testStr},
					{Run: &testStr},
					{If: &ifStr},
				}).
				WithReward("default/reward", PreferredDirectionHigher).
				WithObjective("default/reward", nil, nil)
			experiment := experimentBuilder.Build()
			experiment.InitializeStatus()
			winner := "winner"
			q := resource.Quantity{}
			analysis := &Analysis{
				Metrics: []map[string]QuantityList{
					{
						"metric1": []resource.Quantity{q},
					},
					{
						"metric1": []resource.Quantity{q},
					},
				},
				Winner: &Winner{
					WinnerFound: true,
					Winner:      &winner,
				},
				Objectives: []BooleanList{
					[]bool{false},
					[]bool{false},
				},
				Weights: []int32{25, 74},
			}
			experiment.Status.Analysis = analysis
			experimentList := ExperimentList{
				Items: []Experiment{*experiment},
			}
			backendDescription := "backend description"
			backendAuthType := BasicAuthType
			backendMethod := POSTMethodType
			backendProvider := "backend provider"
			backendJQExpresssion := "expression"
			backendSecret := "namespace/secret"
			backendURL := "url"
			backend := Backend{
				Name:        "backend",
				Description: &backendDescription,
				BackendDetail: BackendDetail{
					AuthType:     &backendAuthType,
					Method:       &backendMethod,
					Provider:     &backendProvider,
					JQExpression: &backendJQExpresssion,
					Secret:       &backendSecret,
					URL:          &backendURL,
					VersionInfo:  []VersionDetail{},
				},
				Metrics: []Metric{},
			}
			experiment.Spec.Backends = []Backend{backend}
			metricDesription := "metric description"
			metricUnits := "ms"
			metricType := GaugeMetricType
			metricBody := "body"
			metric := Metric{
				Name:        "metric name",
				Description: &metricDesription,
				Params:      []NamedValue{},
				Units:       &metricUnits,
				Type:        &metricType,
				Body:        &metricBody,
			}
			experiment.Spec.Backends[0].Metrics = []Metric{metric}

			Expect(reflect.DeepEqual(experimentBuilder, experimentBuilder.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment, experiment.DeepCopyObject())).Should(BeTrue())
			// Expect(reflect.DeepEqual(experimentList, experimentList.DeepCopyObject())).Should(BeTrue())
			Expect(len(experimentList.Items)).Should(Equal(len(experimentList.DeepCopy().Items)))

			Expect(reflect.DeepEqual(experiment.Spec, *experiment.Spec.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Criteria, experiment.Spec.Criteria.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Criteria.Rewards[0], *experiment.Spec.Criteria.Rewards[0].DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Duration, experiment.Spec.Duration.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Actions, experiment.Spec.Actions.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Actions["start"], experiment.Spec.Actions["start"].DeepCopy())).Should(BeTrue())

			Expect(reflect.DeepEqual(experiment.Spec.Backends[0], *experiment.Spec.Backends[0].DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Spec.Backends[0].Metrics[0], *experiment.Spec.Backends[0].Metrics[0].DeepCopy())).Should(BeTrue())

			Expect(reflect.DeepEqual(experiment.Status, *experiment.Status.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis, experiment.Status.Analysis.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Analysis.Winner, experiment.Status.Analysis.Winner.DeepCopy())).Should(BeTrue())
			Expect(reflect.DeepEqual(experiment.Status.Conditions[0], experiment.Status.Conditions[0].DeepCopy())).Should(BeTrue())
		})
	})
})
