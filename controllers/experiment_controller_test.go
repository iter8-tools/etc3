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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/configuration"
)

var _ = Describe("Experiment validation", func() {
	Context("When creating an experiment with an invalid spec.duration.maxIteration", func() {
		testName := "test-invalid-duration"
		testNamespace := "default"
		It("Should fail to create experiment", func() {
			ctx := context.Background()
			experiment := v2alpha1.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(10, 0).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).ShouldNot(Succeed())
		})
	})
})

var _ = Describe("Experiment validation", func() {
	Context("When creating an experiment with a valid spec.duration.maxIteration", func() {
		testName := "test-valid-duration"
		testNamespace := "default"
		It("Should succeed in creating experiment", func() {
			ctx := context.Background()
			experiment := v2alpha1.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(10, 1).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())
		})
	})
})

var _ = Describe("Late Initialization", func() {
	var ctx context.Context
	Context("When creating a valid new Experiment", func() {
		It("Should successfully complete late initialization", func() {
			By("Providing a request-count metric")
			ctx = context.Background()
			m := v2alpha1.NewMetric("request-count", "default").
				WithType(v2alpha1.CounterMetricType).
				WithParams(map[string]string{"param": "value"}).
				WithProvider("prometheus").
				Build()
			Expect(k8sClient.Create(ctx, m)).Should(Succeed())
			createdMetric := &v2alpha1.Metric{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: "request-count", Namespace: "default"}, createdMetric)
				if err != nil {
					return false
				}
				return true
			}).Should(BeTrue())

			By("Creating a new Experiment")
			// ctx := context.Background()
			experiment := v2alpha1.NewExperiment("test", "default").
				WithTarget("target").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithRequestCount("request-count").
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			By("Getting experiment after late initialization has run (spec.Duration !=- nil)")
			createdExperiment := &v2alpha1.Experiment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: "test", Namespace: "default"}, createdExperiment)
				if err != nil {
					return false
				}
				return createdExperiment.Spec.Duration != nil
				// return true
			}).Should(BeTrue())
			//
			By("Inspecting status")
			Expect(createdExperiment.Status.InitTime).ShouldNot(BeNil())
			Expect(createdExperiment.Status.LastUpdateTime).ShouldNot(BeNil())
			Expect(createdExperiment.Status.CompletedIterations).ShouldNot(BeNil())
			Expect(len(createdExperiment.Status.Conditions)).Should(Equal(2))
			By("Inspecting spec")
			Expect(createdExperiment.Spec.GetMaxIterations()).Should(Equal(v2alpha1.DefaultMaxIterations))
			Expect(createdExperiment.Spec.GetIntervalSeconds()).Should(Equal(int32(v2alpha1.DefaultIntervalSeconds)))
			Expect(createdExperiment.Spec.GetMaxCandidateWeight()).Should(Equal(v2alpha1.DefaultMaxCandidateWeight))
			Expect(createdExperiment.Spec.GetMaxCandidateWeightIncrement()).Should(Equal(v2alpha1.DefaultMaxCandidateWeightIncrement))
			Expect(createdExperiment.Spec.GetAlgorithm()).Should(Equal(v2alpha1.DefaultAlgorithm))
			Expect(len(createdExperiment.Spec.Metrics)).Should(Equal(1))
			Expect(*createdExperiment.Spec.GetRequestCount(configuration.Iter8Config{})).Should(Equal("request-count"))
		})
	})
})

var _ = Describe("Experiment proceeds", func() {
	Context("Early event trigger", func() {
		ctx := context.Background()
		testName := "early-reconcile"
		testNamespace := "default"
		It("Experiment should complete", func() {
			By("Creating Experiment with 10s interval")
			expectedIterations := int32(2)
			initialInterval := int32(5)
			modifiedInterval := int32(10)
			experiment := v2alpha1.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(initialInterval, expectedIterations).
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
