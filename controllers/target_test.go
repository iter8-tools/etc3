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
	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Target Acquisition", func() {
	var (
		testNamespace string
	)
	BeforeEach(func() {
		testNamespace = "default"

		// Expect(k8sClient.DeleteAllOf(ctx(), &v2alpha1.Experiment{})).Should(Succeed())
	})
	AfterEach(func() {
		// Expect(k8sClient.DeleteAllOf(ctx(), &v2alpha1.Experiment{})).Should(Succeed())
	})

	// This is indirectly tested by the test case below; this is an explicit test
	Context("Experiment already has the target", func() {
		var experiment *v2alpha1.Experiment
		JustBeforeEach(func() {
			experiment = v2alpha1.NewExperiment("already-has-target", testNamespace).
				WithTarget("targettest1").
				WithStrategy(v2alpha1.StrategyTypeCanary).
				WithCondition(v2alpha1.ExperimentConditionTargetAcquired, corev1.ConditionTrue, v2alpha1.ReasonTargetAcquired, "").
				Build()
		})
		It("should know it has the target", func() {
			ok := reconciler.acquireTarget(ctx(), experiment)
			Expect(ok).Should(BeTrue())
		})
	})

	Context("Experiment wanting to acquire a target", func() {
		var (
			hasName, wantsName string
			has, wants         *v2alpha1.Experiment
		)
		JustBeforeEach(func() {
			hasName = "willget-target"
			has = v2alpha1.NewExperiment(hasName, testNamespace).
				WithTarget("unavailable-target").
				WithStrategy(v2alpha1.StrategyTypeConformance).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(3, 2).
				WithBaselineVersion("baseline", nil).
				Build()
			wantsName = "wants-target"
			wants = v2alpha1.NewExperiment(wantsName, testNamespace).
				WithTarget("unavailable-target").
				WithStrategy(v2alpha1.StrategyTypeConformance).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(1, 1).
				WithBaselineVersion("baseline", nil).
				Build()
		})
		It("will acquire the target only after a target holder is completed", func() {
			By("Creating an experiment with a unique target name")
			Expect(k8sClient.Create(ctx(), has)).Should(Succeed())
			// defer k8sClient.Delete(ctx, has)
			Eventually(func() bool { return hasTarget(hasName, testNamespace) }).Should(BeTrue())
			// Eventually hasName should be Running
			Eventually(func() bool {
				return hasValue(hasName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.Stage != nil && *exp.Status.Stage == v2alpha1.ExperimentStageRunning
				})
			}).Should(BeTrue())

			By("Creating experiment wanting the same target")
			Expect(k8sClient.Create(ctx(), wants)).Should(Succeed())
			// defer k8sClient.Delete(ctx(), has)
			Eventually(func() bool { return isDeployed(wantsName, testNamespace) }).Should(BeTrue())
			Eventually(func() bool {
				return hasValue(wantsName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					// status is initialized
					return exp.Status.InitTime != nil
				})
			}).Should(BeTrue())

			By("Waiting for the target")
			Expect(hasTarget(wantsName, testNamespace)).Should(BeFalse())
			// wantsName should be Waiting
			Eventually(func() bool {
				return hasValue(wantsName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.Stage != nil && *exp.Status.Stage == v2alpha1.ExperimentStageWaiting
				})
			}).Should(BeTrue())

			By("Eventually the first experiment completes")
			Eventually(func() bool { return completes(hasName, testNamespace) }, 8).Should(BeTrue())
			// hasName should be Completed
			Eventually(func() bool {
				return hasValue(hasName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return *exp.Status.Stage == v2alpha1.ExperimentStageCompleted
				})
			}).Should(BeTrue())

			By("Allowing the second to acquire the target and proceed")
			Eventually(func() bool { return hasTarget(wantsName, testNamespace) }).Should(BeTrue())
			// wantsName should be Waiting
			Eventually(func() bool {
				return hasValue(wantsName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.Stage != nil && (*exp.Status.Stage == v2alpha1.ExperimentStageRunning || *exp.Status.Stage == v2alpha1.ExperimentStageCompleted)
				})
			}).Should(BeTrue())
			Eventually(func() bool { return completes(wantsName, testNamespace) }).Should(BeTrue())
			Eventually(func() bool {
				return hasValue(wantsName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.Stage != nil && *exp.Status.Stage == v2alpha1.ExperimentStageCompleted
				})
			}).Should(BeTrue())

		})
	})

})

var _ = Describe("Finalizer", func() {
	var (
		testNamespace string
	)
	BeforeEach(func() {
		testNamespace = "default"

		// Expect(k8sClient.DeleteAllOf(ctx(), &v2alpha1.Experiment{})).Should(Succeed())
	})
	AfterEach(func() {
		// Expect(k8sClient.DeleteAllOf(ctx(), &v2alpha1.Experiment{})).Should(Succeed())
	})

	Context("Experiment wanting to acquire a target", func() {
		var (
			hasName, wantsName string
			has, wants         *v2alpha1.Experiment
		)
		JustBeforeEach(func() {
			hasName = "willget-target-finalizer"
			has = v2alpha1.NewExperiment(hasName, testNamespace).
				WithTarget("unavailable-target-finalizer").
				WithStrategy(v2alpha1.StrategyTypeConformance).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(3, 2).
				WithBaselineVersion("baseline", nil).
				Build()
			wantsName = "wants-target-finalizer"
			wants = v2alpha1.NewExperiment(wantsName, testNamespace).
				WithTarget("unavailable-target-finalizer").
				WithStrategy(v2alpha1.StrategyTypeConformance).
				WithHandlers(map[string]string{"start": "none", "finish": "none"}).
				WithDuration(1, 1).
				WithBaselineVersion("baseline", nil).
				Build()
		})
		It("will acquire the target when a holder is deleted", func() {
			By("Creating an experiment with a unique target name")
			Expect(k8sClient.Create(ctx(), has)).Should(Succeed())
			// defer k8sClient.Delete(ctx, has)
			Eventually(func() bool { return hasTarget(hasName, testNamespace) }).Should(BeTrue())
			// Eventually hasName should be Running
			Eventually(func() bool {
				return hasValue(hasName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.Stage != nil && *exp.Status.Stage == v2alpha1.ExperimentStageRunning
				})
			}).Should(BeTrue())

			By("Creating experiment wanting the same target")
			Expect(k8sClient.Create(ctx(), wants)).Should(Succeed())
			// defer k8sClient.Delete(ctx(), wants)
			Eventually(func() bool { return isDeployed(wantsName, testNamespace) }).Should(BeTrue())
			Eventually(func() bool {
				return hasValue(wantsName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					// status is initialized
					return exp.Status.InitTime != nil
				})
			}).Should(BeTrue())

			By("Waiting for the target")
			Expect(hasTarget(wantsName, testNamespace)).Should(BeFalse())
			// wantsName should be Waiting
			Eventually(func() bool {
				return hasValue(wantsName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.Stage != nil && *exp.Status.Stage == v2alpha1.ExperimentStageWaiting
				})
			}).Should(BeTrue())

			By("Deleting the first experiment")
			exp := &v2alpha1.Experiment{}
			Expect(k8sClient.Get(ctx(), types.NamespacedName{Name: hasName, Namespace: testNamespace}, exp)).Should(Succeed())
			Expect(k8sClient.Delete(ctx(), exp, client.PropagationPolicy(metav1.DeletePropagationBackground))).Should(Succeed())
			Eventually(func() bool { return isDeleted(hasName, testNamespace) }, 8).Should(BeTrue())

			By("Allowing the second to acquire the target and proceed")
			Eventually(func() bool { return hasTarget(wantsName, testNamespace) }).Should(BeTrue())
			// wantsName should be Running
			Eventually(func() bool {
				return hasValue(wantsName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.Stage != nil && *exp.Status.Stage == v2alpha1.ExperimentStageRunning
				})
			}).Should(BeTrue())
			Eventually(func() bool { return completes(wantsName, testNamespace) }).Should(BeTrue())
			Eventually(func() bool {
				return hasValue(wantsName, testNamespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.Stage != nil && *exp.Status.Stage == v2alpha1.ExperimentStageCompleted
				})
			}).Should(BeTrue())
		})
	})

})
