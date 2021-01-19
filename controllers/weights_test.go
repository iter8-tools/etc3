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
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/util"
)

var _ = Describe("Weight Patching", func() {
	restCfg, _ := config.GetConfig()
	namespace := "default"

	ctx := context.Background()
	ctx = context.WithValue(ctx, util.LoggerKey, ctrl.Log)

	Context("When experimentType is Performance", func() {
		experiment := v2alpha1.NewExperiment("noVersionInfo", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypePerformance).
			Build()
		It("should succeed without error", func() {
			Expect(redistributeWeight(ctx, experiment, restCfg)).Should(Succeed())
		})
	})

	Context("When algorithm is FixedSplit", func() {
		experiment := v2alpha1.NewExperiment("noVersionInfo", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithAlgorithm(v2alpha1.AlgorithmTypeFixedSplit).
			Build()
		It("should succeed without error", func() {
			Expect(redistributeWeight(ctx, experiment, restCfg)).Should(Succeed())
		})
	})

	Context("When no versionInfo", func() {
		experiment := v2alpha1.NewExperiment("noVersionInfo", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			Build()
		It("Should fail with error", func() {
			err := redistributeWeight(ctx, experiment, restCfg)
			Expect(err).Should(MatchError("Cannot redistribute weight; no version information present"))
		})
	})

	Context("When WeightObjRef is not set", func() {
		experiment := v2alpha1.NewExperiment("noWeightObRef", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithDuration(10, 0).
			WithBaselineVersion("baseline", nil).
			Build()
		It("Should not add a patch", func() {
			patches := map[corev1.ObjectReference][]patchIntValue{}
			err := addPatch(ctx, experiment, experiment.Spec.VersionInfo.Baseline, &patches)
			Expect(err).Should(BeNil())
			Expect(patches).Should(BeEmpty())
		})
	})

	Context("When WeightObjRef set but no FieldPath", func() {
		experiment := v2alpha1.NewExperiment("noFieldPath", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithDuration(10, 0).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
			}).
			Build()
		It("Should not add a patch", func() {
			patches := map[corev1.ObjectReference][]patchIntValue{}
			err := addPatch(ctx, experiment, experiment.Spec.VersionInfo.Baseline, &patches)
			Expect(err).Should(BeNil())
			Expect(patches).Should(BeEmpty())
		})
	})

	Context("When full WeightObjRef set but no weight recommendation", func() {
		experiment := v2alpha1.NewExperiment("noWeightRecommendation", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithDuration(10, 0).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  "/path/to/weight",
			}).
			Build()
		It("Should not fail and not add a patch", func() {
			patches := map[corev1.ObjectReference][]patchIntValue{}
			err := addPatch(ctx, experiment, experiment.Spec.VersionInfo.Baseline, &patches)
			Expect(err).Should(MatchError("No weight recommendation provided"))
			Expect(patches).Should(BeEmpty())
		})
	})

	Context("When full WeightObjRef and weight recommendation matches current value", func() {
		experiment := v2alpha1.NewExperiment("recommendationIsCurrent", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithDuration(10, 0).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  "/path/to/weight",
			}).
			WithCurrentWeight("baseline", int32(25)).
			WithRecommendedWeight("baseline", int32(25)).
			Build()
		It("Should not fail and not add a patch", func() {
			patches := map[corev1.ObjectReference][]patchIntValue{}
			err := addPatch(ctx, experiment, experiment.Spec.VersionInfo.Baseline, &patches)
			Expect(err).Should(BeNil())
			Expect(patches).Should(BeEmpty())
		})
	})
	Context("When full WeightObjRef and weight recommendation does not match the current value", func() {
		experiment := v2alpha1.NewExperiment("recommendationNotCurrent", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithDuration(10, 0).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  "/path/to/weight",
			}).
			WithCurrentWeight("baseline", int32(25)).
			WithRecommendedWeight("baseline", int32(50)).
			Build()
		It("Should add a patch", func() {
			patches := map[corev1.ObjectReference][]patchIntValue{}
			err := addPatch(ctx, experiment, experiment.Spec.VersionInfo.Baseline, &patches)
			Expect(err).Should(BeNil())
			Expect(len(patches)).Should(Equal(1))
		})
	})
	Context("When multiple versions require updates to the same object", func() {
		experiment := v2alpha1.NewExperiment("recommendationNotCurrent", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithDuration(10, 0).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  "/path/to/weight/0",
			}).
			WithCandidateVersion("candidate", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  "/path/to/weight/1",
			}).
			WithCurrentWeight("baseline", int32(25)).
			WithCurrentWeight("candidate", int32(75)).
			WithRecommendedWeight("baseline", int32(35)).
			WithRecommendedWeight("candidate", int32(65)).
			Build()
		It("There are multiple patches for one object", func() {
			patches := map[corev1.ObjectReference][]patchIntValue{}
			err := addPatch(ctx, experiment, experiment.Spec.VersionInfo.Baseline, &patches)
			Expect(err).Should(BeNil())
			Expect(len(patches)).Should(Equal(1))
			for _, version := range experiment.Spec.VersionInfo.Candidates {
				Expect(addPatch(ctx, experiment, version, &patches)).Should(Succeed())
			}
			Expect(len(patches)).Should(Equal(1))
			key := getKey(*experiment.Spec.VersionInfo.Baseline.WeightObjRef)
			patchList, ok := (patches)[key]
			Expect(ok).Should(BeTrue())
			Expect(len(patchList)).Should(Equal(2))
		})
	})

	Context("When multiple versions require updates to different objects", func() {
		experiment := v2alpha1.NewExperiment("recommendationNotCurrent", namespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary).
			WithDuration(10, 0).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs0",
				Namespace:  namespace,
				FieldPath:  "/path/to/weight/0",
			}).
			WithCandidateVersion("candidate", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs1",
				Namespace:  namespace,
				FieldPath:  "/path/to/weight/1",
			}).
			WithCurrentWeight("baseline", int32(25)).
			WithCurrentWeight("candidate", int32(75)).
			WithRecommendedWeight("baseline", int32(35)).
			WithRecommendedWeight("candidate", int32(65)).
			Build()
		It("There is one patch for each object", func() {
			patches := map[corev1.ObjectReference][]patchIntValue{}
			err := addPatch(ctx, experiment, experiment.Spec.VersionInfo.Baseline, &patches)
			Expect(err).Should(BeNil())
			Expect(len(patches)).Should(Equal(1))
			for _, version := range experiment.Spec.VersionInfo.Candidates {
				Expect(addPatch(ctx, experiment, version, &patches)).Should(Succeed())
			}
			Expect(len(patches)).Should(Equal(2))
		})
	})

})
