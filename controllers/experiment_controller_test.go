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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Experiment validation", func() {
	Context("When creating an experiment with an invalid spec.duration.maxIteration", func() {
		It("Should fail to create experiment", func() {
			ctx := context.Background()
			experiment := v2alpha1.NewExperiment("test-invalid", "default").
				WithTarget("target").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				WithDuration(10, 0).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).ShouldNot(Succeed())
		})
	})
})

var _ = Describe("Experiment validation", func() {
	Context("When creating an experiment with a valid spec.duration.maxIteration", func() {
		It("Should succeed in creating experiment", func() {
			ctx := context.Background()
			experiment := v2alpha1.NewExperiment("test-valid", "default").
				WithTarget("target").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				WithDuration(10, 1).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())
		})
	})
})

var _ = Describe("Late Initialization", func() {
	Context("When creating a new Experiment", func() {
		It("Should do late initialization", func() {
			By("Creating a new Experiment")
			ctx := context.Background()
			experiment := v2alpha1.NewExperiment("test", "default").
				WithTarget("target").
				WithStrategy(v2alpha1.StrategyTypeCanary).
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
		})
	})
})
