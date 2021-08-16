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
	"time"

	v2beta1 "github.com/iter8-tools/etc3/api/v2beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Metrics", func() {
	var jqe string = "expr"
	var url string = "url"

	var testName string
	var testNamespace, metricsNamespace string
	var goodObjectiveMetric, goodObjective2Metric, badObjectiveMetric, rewardMetric *v2beta1.Metric
	BeforeEach(func() {
		testNamespace = "default"
		metricsNamespace = "metric-namespace"

		k8sClient.DeleteAllOf(ctx(), &v2beta1.Experiment{}, client.InNamespace(testNamespace))
		k8sClient.DeleteAllOf(ctx(), &v2beta1.Metric{}, client.InNamespace(testNamespace))
		k8sClient.DeleteAllOf(ctx(), &v2beta1.Metric{}, client.InNamespace(metricsNamespace))

		By("Providing a request-count metric")
		m := v2beta1.NewMetric("request-count", metricsNamespace).
			WithType(v2beta1.CounterMetricType).
			WithParams([]v2beta1.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			Build()
		Expect(k8sClient.Create(ctx(), m)).Should(Succeed())
		goodObjective2Metric = v2beta1.NewMetric("objective-with-good-reference-2", metricsNamespace).
			WithType(v2beta1.CounterMetricType).
			WithParams([]v2beta1.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			WithSampleSize("request-count").
			Build()
		Expect(k8sClient.Create(ctx(), goodObjective2Metric)).Should(Succeed())
		By("creating an objective that does not reference the request-count")
		goodObjectiveMetric = v2beta1.NewMetric("objective-with-good-reference", "default").
			WithType(v2beta1.CounterMetricType).
			WithParams([]v2beta1.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			WithSampleSize(metricsNamespace + "/request-count").
			Build()
		Expect(k8sClient.Create(ctx(), goodObjectiveMetric)).Should(Succeed())
		By("creating an objective that references request-count")
		badObjectiveMetric = v2beta1.NewMetric("objective-with-bad-reference", "default").
			WithType(v2beta1.CounterMetricType).
			WithParams([]v2beta1.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			WithSampleSize("request-count").
			Build()
		Expect(k8sClient.Create(ctx(), badObjectiveMetric)).Should(Succeed())
		rewardMetric = v2beta1.NewMetric("rwrd", "default").
			WithType(v2beta1.CounterMetricType).
			WithParams([]v2beta1.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			Build()
		Expect(k8sClient.Create(ctx(), rewardMetric)).Should(Succeed())
	})

	Context("When creating an experiment referencing valid metrics", func() {
		// experiment (in default namespace) refers to metric "objective-with-good-reference"
		// which has a sampleSize "metricNamespace/request-count" which is correct
		It("Should successfully read the metrics and proceed", func() {
			By("Creating experiment")
			testName = "valid-reference"
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithRequestCount(metricsNamespace+"/request-count").
				WithObjective(*goodObjectiveMetric, nil, nil, false).
				WithReward(*rewardMetric, v2beta1.PreferredDirectionHigher).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it starts Running")
			// this assumes that it runs for a while
			Eventually(func() bool {
				return issuedEvent("Advanced to Running") //v2beta1.ReasonStageAdvanced)
			}, 5).Should(BeTrue())
		})
	})

	Context("failed start handler", func() {
		Specify("experiment teminated in a failed state", func() {
			By("Creating an experiment with a start handler")
			name := "has-failing-handler"
			iterations, loops := int32(2), int32(1)
			handler := "start"
			experiment := v2beta1.NewExperiment(name, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithAction(handler, []v2beta1.TaskSpec{}).
				WithRequestCount(metricsNamespace+"/request-count").
				WithDuration(30, iterations, loops).
				Build()

			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			Eventually(func() bool { return fails(name, testNamespace) }, 5).Should(BeTrue())
		})
	})

	Context("When creating an experiment which refers to a non-existing metric", func() {
		// experiment (in default ns) refers to metric "request-count" (not in default namespace)
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "non-existing-metric-1"
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithRequestCount("request-count").
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return issuedEvent(v2beta1.ReasonMetricUnavailable) &&
					issuedEvent("default/request-count")
			}, 5).Should(BeTrue())
			Eventually(func() bool { return fails(testName, testNamespace) }, 5).Should(BeTrue())
		})
	})
	Context("When creating another experiment which refers to a non-existing metric", func() {
		// experiment (in default ns) refers to metric "iter8/request-count" (not in iter8 namespace)
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "non-existing-metric-2"
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithRequestCount("iter8/request-count").
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return issuedEvent(v2beta1.ReasonMetricUnavailable)
			}, 5).Should(BeTrue())
		})
	})

	Context("When creating an experiment referencing a metric with a bad reference", func() {
		// experiment (in default namespace) refers to metric "objective-with-bad-reference"
		// which has a sampleSize "request-count" (not in same ns as the referring metric (default))
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "invalid-reference"
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithRequestCount(metricsNamespace+"/request-count").
				WithObjective(*badObjectiveMetric, nil, nil, false).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return issuedEvent(v2beta1.ReasonMetricUnavailable)
			}, 5).Should(BeTrue())
			// Eventually(func() bool { return fails(testName, testNamespace) }, 5).Should(BeTrue())
		})
	})

	Context("When creating an experiment referencing a metric with a bad reference", func() {
		// experiment (in default namespace) refers to metric "objective-with-bad-reference"
		// which has a sampleSize "request-count" (not in same ns as the referring metric (default))
		It("Should successfully read metrics", func() {
			By("Creating experiment")
			testName = "good-reference-2"

			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithRequestCount(metricsNamespace+"/objective-with-good-reference-2").
				WithObjective(*goodObjective2Metric, nil, nil, false).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it starts Running")
			// this assumes that it runs for a while
			Eventually(func() bool {
				return issuedEvent(v2beta1.ReasonStageAdvanced)
			}, 5).Should(BeTrue())
		})
	})
	Context("When creating an experiment with a metric that references a local request", func() {
		// Test that "metrics-namesapce/request-count" and "request-count" are not both added as metrics
		It("Should successfully read metrics", func() {
			By("Creating experiment")
			testName = "singlerefcount"

			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithRequestCount(metricsNamespace+"/request-count").
				WithObjective(*goodObjective2Metric, nil, nil, false).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("that the metrics are read")
			time.Sleep(10 * time.Second)
			exp := v2beta1.Experiment{}
			Expect(k8sClient.Get(ctx(), types.NamespacedName{Namespace: testNamespace, Name: testName}, &exp)).Should(Succeed())
			Expect(len(exp.Status.Metrics)).To(Equal(2))
		})
	})

	Context("When request count specified with and without namespace in same experiment", func() {
		Specify("It should be read only once", func() {
			By("Defining experiment with requestcount specified with and without namespace")
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithRequestCount(metricsNamespace+"/request-count").   // specified with namespace
				WithObjective(*goodObjective2Metric, nil, nil, false). // specifid without namespace
				Build()
			By("reading the metrics")
			Expect(reconciler.ReadMetrics(ctx(), experiment)).Should(BeTrue())
			By("counting the number of metrics read")
			Expect(len(experiment.Status.Metrics)).To(Equal(2))
		})
	})
})
