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

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Experiment Validation", func() {
	ctx := context.Background()

	Context("When creating an experiment with an invalid spec.duration.maxIteration", func() {
		testName := "test-invalid-duration"
		testNamespace := "default"
		It("Should fail to create experiment", func() {
			experiment := v2alpha1.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha1.TestingPatternCanary).
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
			experiment := v2alpha1.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha1.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(10, 1, 1).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())
		})
	})

	Context("When creating a valid new Experiment", func() {
		It("Should successfully complete late initialization", func() {
			By("Providing a request-count metric")
			m := v2alpha1.NewMetric("request-count", "iter8").
				WithType(v2alpha1.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			Expect(k8sClient.Create(ctx, m)).Should(Succeed())
			By("creating a reward metric")
			reward := v2alpha1.NewMetric("reward", "default").
				WithType(v2alpha1.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			Expect(k8sClient.Create(ctx, reward)).Should(Succeed())
			By("creating an indicator")
			indicator := v2alpha1.NewMetric("indicataor", "default").
				WithType(v2alpha1.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			Expect(k8sClient.Create(ctx, indicator)).Should(Succeed())
			By("creating an objective")
			objective := v2alpha1.NewMetric("objective", "default").
				WithType(v2alpha1.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			Expect(k8sClient.Create(ctx, objective)).Should(Succeed())
			By("creating an objective that is not in the cluster")
			fake := v2alpha1.NewMetric("fake", "default").
				WithType(v2alpha1.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()

			By("Creating a new Experiment")
			testName := "late-initialization"
			testNamespace := "default"
			experiment := v2alpha1.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha1.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithRequestCount("request-count").
				WithReward(*reward, v2alpha1.PreferredDirectionHigher).
				WithIndicator(*indicator).
				WithObjective(*objective, nil, nil, false).
				WithObjective(*fake, nil, nil, true).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			By("Getting experiment after late initialization has run (spec.Duration !=- nil)")
			Eventually(func() bool {
				return hasValue(testName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.InitTime != nil &&
						exp.Status.LastUpdateTime != nil &&
						exp.Status.CompletedIterations != nil &&
						len(exp.Status.Conditions) == 3
				})
			}).Should(BeTrue())
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
			experiment := v2alpha1.NewExperiment(testName, testNamespace).
				WithTarget("early-reconcile-targets").
				WithTestingPattern(v2alpha1.TestingPatternCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(initialInterval, expectedIterations, 1).
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate", nil).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			By("Changing the interval before the reconcile event triggers")
			time.Sleep(2 * time.Second)
			createdExperiment := &v2alpha1.Experiment{}
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
