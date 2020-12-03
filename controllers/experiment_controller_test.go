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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Experiment controller", func() {
	Context("When creating a new Experiment", func() {
		It("Should set status.InitTime", func() {
			By("Creating a new Experiment")
			ctx := context.Background()
			experiment := &v2alpha1.Experiment{
				TypeMeta: v1.TypeMeta{
					APIVersion: "iter8.tools/v2alpha1",
					Kind:       "Experiment",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v2alpha1.ExperimentSpec{
					Target: "target",
					Strategy: v2alpha1.Strategy{
						Type: v2alpha1.StrategyTypeCanary,
					},
				},
			}
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			createdExperiment := &v2alpha1.Experiment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: "test", Namespace: "default"}, createdExperiment)
				if err != nil {
					return false
				}
				return createdExperiment.Spec.Duration != nil
				// return true
			}).Should(BeTrue())
			Expect(createdExperiment.Status.InitTime).ShouldNot(BeNil())
			Expect(createdExperiment.Spec.Duration.MaxIterations).ShouldNot(BeNil())
			Expect(*createdExperiment.Spec.Duration.MaxIterations).Should(Equal(int32(15)))
		})
	})
})
