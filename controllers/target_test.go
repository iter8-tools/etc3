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

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Target Acquisition", func() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, util.LoggerKey, ctrl.Log)
	testNamespace := "default"

	// This is indirectly tested by the test case below; this is an explicit test
	Context("Experiment already has the target", func() {
		experiment := v2alpha1.NewExperiment("already-has-target", testNamespace).
			WithTarget("targettest1").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithCondition(v2alpha1.ExperimentConditionTargetAcquired, corev1.ConditionTrue, v2alpha1.ReasonTargetAcquired, "").
			Build()
		It("should know it has the target", func() {
			ok := reconciler.acquireTarget(ctx, experiment)
			Expect(ok).Should(BeTrue())
		})
	})

	Context("Experiment wanting to acquire a target", func() {
		hasName := "willget-target"
		wantsName := "wants-target"
		has := v2alpha1.NewExperiment(hasName, testNamespace).
			WithTarget("unavailable-target").
			WithStrategy(v2alpha1.StrategyTypePerformance).
			WithHandlers(map[string]string{"start": "none", "finish": "none"}).
			WithDuration(3, 2).
			WithBaselineVersion("baseline", nil).
			Build()
		wants := v2alpha1.NewExperiment(wantsName, testNamespace).
			WithTarget("unavailable-target").
			WithStrategy(v2alpha1.StrategyTypePerformance).
			WithHandlers(map[string]string{"start": "none", "finish": "none"}).
			WithDuration(1, 1).
			WithBaselineVersion("baseline", nil).
			Build()
		It("will acquire the target if no other experiments using it", func() {
			By("Creating an experiment with a unique target name")
			Expect(k8sClient.Create(ctx, has)).Should(Succeed())
			Eventually(func() bool { return hasTarget(ctx, hasName, testNamespace) }).Should(BeTrue())
		})

		It("will not acquire the target when another experiment already has it", func() {
			By("Creating experiment wanting the same target")
			Expect(k8sClient.Create(ctx, wants)).Should(Succeed())
			Eventually(func() bool { return isDeployed(ctx, wantsName, testNamespace) }).Should(BeTrue())

			By("Waiting for the target")
			Expect(hasTarget(ctx, wantsName, testNamespace)).Should(BeFalse())
		})

		It("will get the target when the original holder finishes", func() {
			By("Eventually the first experiment completes")
			Eventually(func() bool { return completes(ctx, hasName, testNamespace) }, 8).Should(BeTrue())

			By("Allowing the second to acquire the target and proceed")
			Eventually(func() bool { return hasTarget(ctx, wantsName, testNamespace) }).Should(BeTrue())
			Eventually(func() bool { return completes(ctx, wantsName, testNamespace) }).Should(BeTrue())
		})
	})

})

var _ = Describe("Finalizer", func() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, util.LoggerKey, ctrl.Log)
	testNamespace := "default"

	Context("Experiment wanting to acquire a target", func() {
		hasName := "willget-target-finalizer"
		wantsName := "wants-target-finalizer"
		has := v2alpha1.NewExperiment(hasName, testNamespace).
			WithTarget("unavailable-target").
			WithStrategy(v2alpha1.StrategyTypePerformance).
			WithHandlers(map[string]string{"start": "none", "finish": "none"}).
			WithDuration(3, 2).
			WithBaselineVersion("baseline", nil).
			Build()
		wants := v2alpha1.NewExperiment(wantsName, testNamespace).
			WithTarget("unavailable-target").
			WithStrategy(v2alpha1.StrategyTypePerformance).
			WithHandlers(map[string]string{"start": "none", "finish": "none"}).
			WithDuration(1, 1).
			WithBaselineVersion("baseline", nil).
			Build()
		It("will acquire the target if no other experiments using it", func() {
			By("Creating an experiment with a unique target name")
			Expect(k8sClient.Create(ctx, has)).Should(Succeed())
			Eventually(func() bool { return hasTarget(ctx, hasName, testNamespace) }).Should(BeTrue())
		})

		It("will not acquire the target when another experiment already has it", func() {
			By("Creating experiment wanting the same target")
			Expect(k8sClient.Create(ctx, wants)).Should(Succeed())
			Eventually(func() bool { return isDeployed(ctx, wantsName, testNamespace) }).Should(BeTrue())

			By("Waiting for the target")
			Expect(hasTarget(ctx, wantsName, testNamespace)).Should(BeFalse())
		})

		It("will get the target when the original holder is deleted", func() {
			By("Deleting the first experiment")
			exp := &v2alpha1.Experiment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: hasName, Namespace: testNamespace}, exp)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, exp, client.PropagationPolicy(metav1.DeletePropagationBackground))).Should(Succeed())
			Eventually(func() bool { return isDeleted(ctx, hasName, testNamespace) }, 8).Should(BeTrue())

			By("Allowing the second to acquire the target and proceed")
			Eventually(func() bool { return hasTarget(ctx, wantsName, testNamespace) }).Should(BeTrue())
			Eventually(func() bool { return completes(ctx, wantsName, testNamespace) }).Should(BeTrue())
		})
	})

})
