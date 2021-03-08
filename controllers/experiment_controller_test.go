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
	"time"

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Experiment Validation", func() {
	ctx := context.Background()

	Context("When creating an experiment with an invalid spec.duration.maxIteration", func() {
		testName := "test-invalid-duration"
		testNamespace := "default"
		It("Should fail to create experiment", func() {
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(10, 0, 1).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).ShouldNot(Succeed())
		})
	})

	Context("When creating an experiment with a valid spec.duration.maxIteration", func() {
		testName := "test-valid-duration"
		testNamespace := "default"
		It("Should succeed in creating experiment", func() {
			ctx := context.Background()
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(10, 1, 1).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())
		})
	})

	Context("When creating a valid new Experiment", func() {
		It("Should successfully complete late initialization", func() {
			By("Providing a request-count metric")
			m := v2alpha2.NewMetric("request-count", "iter8").
				WithType(v2alpha2.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			// ns := &corev1.Namespace{
			// 	ObjectMeta: metav1.ObjectMeta{Name: "iter8"},
			// }
			// Expect(k8sClient.Create(ctx, ns)).Should(Succeed())
			Expect(k8sClient.Create(ctx, m)).Should(Succeed())
			// createdMetric := &v2alpha2.Metric{}
			// Eventually(func() bool {
			// 	err := k8sClient.Get(ctx, types.NamespacedName{Name: "request-count", Namespace: "iter8"}, createdMetric)
			// 	if err != nil {
			// 		return false
			// 	}
			// 	return true
			// }).Should(BeTrue())
			By("creating a reward metric")
			reward := v2alpha2.NewMetric("reward", "default").
				WithType(v2alpha2.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			Expect(k8sClient.Create(ctx, reward)).Should(Succeed())
			By("creating an indicator")
			indicator := v2alpha2.NewMetric("indicataor", "default").
				WithType(v2alpha2.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			Expect(k8sClient.Create(ctx, indicator)).Should(Succeed())
			By("creating an objective")
			objective := v2alpha2.NewMetric("objective", "default").
				WithType(v2alpha2.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			Expect(k8sClient.Create(ctx, objective)).Should(Succeed())
			By("creating an objective that is not in the cluster")
			fake := v2alpha2.NewMetric("fake", "default").
				WithType(v2alpha2.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()

			By("Creating a new Experiment")
			testName := "late-initialization"
			testNamespace := "default"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithRequestCount("request-count").
				WithReward(*reward, v2alpha2.PreferredDirectionHigher).
				WithIndicator(*indicator).
				WithObjective(*objective, nil, nil, false).
				WithObjective(*fake, nil, nil, true).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			By("Getting experiment after late initialization has run (spec.Duration !=- nil)")
			Eventually(func() bool {
				return hasValue(testName, testNamespace, func(exp *v2alpha2.Experiment) bool {
					return exp.Status.InitTime != nil &&
						exp.Status.LastUpdateTime != nil &&
						exp.Status.CompletedIterations != nil &&
						len(exp.Status.Conditions) == 3
				})
			}).Should(BeTrue())
		})
	})
})

var _ = Describe("Metrics", func() {
	var testName string
	var testNamespace, metricsNamespace string
	var goodObjective, badObjective *v2alpha2.Metric
	BeforeEach(func() {
		testNamespace = "default"
		metricsNamespace = "metric-namespace"

		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Experiment{}, client.InNamespace(testNamespace))
		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Metric{}, client.InNamespace(testNamespace))
		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Metric{}, client.InNamespace(metricsNamespace))

		By("Providing a request-count metric")
		m := v2alpha2.NewMetric("request-count", metricsNamespace).
			WithType(v2alpha2.CounterMetricType).
			WithParams(map[string]string{"param": "value"}).
			WithProvider("prometheus").
			Build()
		Expect(k8sClient.Create(ctx(), m)).Should(Succeed())
		By("creating an objective that does not reference the request-count")
		goodObjective = v2alpha2.NewMetric("objective-with-good-reference", "default").
			WithType(v2alpha2.CounterMetricType).
			WithParams(map[string]string{"param": "value"}).
			WithProvider("prometheus").
			WithSampleSize(metricsNamespace + "/request-count").
			Build()
		Expect(k8sClient.Create(ctx(), goodObjective)).Should(Succeed())
		By("creating an objective that references request-count")
		badObjective = v2alpha2.NewMetric("objective-with-bad-reference", "default").
			WithType(v2alpha2.CounterMetricType).
			WithParams(map[string]string{"param": "value"}).
			WithProvider("prometheus").
			WithSampleSize("request-count").
			Build()
		Expect(k8sClient.Create(ctx(), badObjective)).Should(Succeed())
	})

	Context("When creating an experiment referencing valid metrics", func() {
		// experiment (in default namespace) refers to metric "objective-with-good-reference"
		// which has a sampleSize "metricNamespace/request-count" which is correct
		It("Should successfully read the metrics and proceed", func() {
			By("Creating experiment")
			testName = "valid-reference"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithRequestCount(metricsNamespace+"/request-count").
				WithObjective(*goodObjective, nil, nil, false).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it starts Running")
			// this assumes that it runs for a while
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonStageAdvanced)
			}, 5).Should(BeTrue())
		})
	})

	Context("When creating an experiment which refers to a non-existing metric", func() {
		// experiment (in default ns) refers to metric "request-count" (not in default namespace)
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "invalid-metric"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithRequestCount("request-count").
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonMetricUnavailable)
			}, 5).Should(BeTrue())
			// Eventually(func() bool { return fails(testName, testNamespace) }, 5).Should(BeTrue())
		})
	})
	Context("When creating another experiment which refers to a non-existing metric", func() {
		// experiment (in default ns) refers to metric "iter8/request-count" (not in iter8 namespace)
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "invalid-metric"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithRequestCount("iter8/request-count").
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonMetricUnavailable)
			}, 5).Should(BeTrue())
			// Eventually(func() bool { return fails(testName, testNamespace) }, 5).Should(BeTrue())
		})
	})

	Context("When creating an experiment referencing a metric with a bad reference", func() {
		// experiment (in default namespace) refers to metric "objective-with-bad-reference"
		// which has a sampleSize "request-count" (not in same ns as the referring metric (default))
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "invalid-reference"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithRequestCount(metricsNamespace+"/request-count").
				WithObjective(*badObjective, nil, nil, false).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonMetricUnavailable)
			}, 5).Should(BeTrue())
			// Eventually(func() bool { return fails(testName, testNamespace) }, 5).Should(BeTrue())
		})
	})

})

var _ = Describe("Experiment proceeds", func() {
	ctx := context.Background()

	Context("Early event trigger", func() {
		testName := "early-reconcile"
		testNamespace := "default"
		It("Experiment should complete", func() {
			By("Creating Experiment with 10s interval")
			expectedIterations := int32(2)
			initialInterval := int32(5)
			modifiedInterval := int32(10)
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("early-reconcile-targets").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(initialInterval, expectedIterations, 1).
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate", nil).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			By("Changing the interval before the reconcile event triggers")
			time.Sleep(2 * time.Second)
			createdExperiment := &v2alpha2.Experiment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: testName, Namespace: testNamespace}, createdExperiment)).Should(Succeed())
			createdExperiment.Spec.Duration.IntervalSeconds = &modifiedInterval
			Expect(k8sClient.Update(ctx, createdExperiment)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: testName, Namespace: testNamespace}, createdExperiment)
				if err != nil {
					return false
				}
				return createdExperiment.Status.GetCompletedIterations() == expectedIterations
				// return true
			}, 2*modifiedInterval*expectedIterations).Should(BeTrue())

		})
	})
})
