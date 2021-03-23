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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/util"
)

var _ = Describe("Reading Weights Using internal method observeWeight", func() {
	var namespace string
	BeforeEach(func() {
		namespace = "default"
		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Experiment{}, client.InNamespace(namespace))
	})
	Context("When try to read field from object", func() {
		name := "read"
		var experiment *v2alpha2.Experiment
		var objRef *corev1.ObjectReference
		JustBeforeEach(func() {
			experiment = v2alpha2.NewExperiment(name, namespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDuration(10, 5, 3).
				Build()
			objRef = &corev1.ObjectReference{
				APIVersion: "iter8.tools/v2alpha2",
				Kind:       "Experiment",
				Name:       name,
				Namespace:  namespace,
			}
		})
		It("A FieldPath into an array returns a valid value", func() {
			objRef.FieldPath = ".status.currentWeightDistribution[2].value"
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			time.Sleep(3 * time.Second)
			exp := v2alpha2.Experiment{}
			Expect(k8sClient.Get(ctx(), types.NamespacedName{Namespace: namespace, Name: name}, &exp)).Should(Succeed())
			exp.Status.CurrentWeightDistribution = []v2alpha2.WeightData{
				{Name: "v1", Value: 10},
				{Name: "v2", Value: 20},
				{Name: "v3", Value: 30},
				{Name: "v4", Value: 40},
			}
			Expect(k8sClient.Status().Update(ctx(), &exp)).Should(Succeed())
			value, _ := observeWeight(ctx(), objRef, namespace, cfg)
			Expect(*value).To(Equal(int32(30)))
		})
		It("A FieldPath returns a valid value", func() {
			objRef.FieldPath = ".spec.duration.maxLoops"
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			value, _ := observeWeight(ctx(), objRef, namespace, cfg)
			Expect(*value).To(Equal(int32(3)))
		})
		It("No FieldPath returns an error", func() {
			experiment.Name = "no-fieldpath"
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			_, err := observeWeight(ctx(), objRef, namespace, cfg)
			Expect(err).To(HaveOccurred())
		})
		It("Invalid FieldPath returns an error", func() {
			experiment.Name = "invalid-fieldpath"
			objRef.FieldPath = ".invalid.path"
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			_, err := observeWeight(ctx(), objRef, namespace, cfg)
			Expect(err).To(HaveOccurred())
		})
		It("Valid path to non int returns an error", func() {
			experiment.Name = "non-int-fieldpath"
			objRef.FieldPath = ".spec.target"
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			_, err := observeWeight(ctx(), objRef, namespace, cfg)
			Expect(err).To(HaveOccurred())
		})
		It("Reference to invalid object returns an error", func() {
			experiment.Name = "invalid-ref"
			objRef.Name = "no-such-object"
			objRef.FieldPath = ".spec.duration.maxLoops"
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			_, err := observeWeight(ctx(), objRef, namespace, cfg)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Updating weights from reconcile", func() {
	var namespace string
	BeforeEach(func() {
		namespace = "default"
		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Experiment{}, client.InNamespace(namespace))
	})

	Context("When weightObjectRef has errors preventing read", func() {
		Specify("When one weightObjectRef is invalid, the experiment fails", func() {
			By("Defining with one invalid weightObjectRef")
			name := "badweightref"
			objRefb := &corev1.ObjectReference{
				APIVersion: "iter8.tools/v2alpha2",
				Kind:       "Experiment",
				Name:       name,
				Namespace:  namespace,
				FieldPath:  ".spec.Duration.intervalSeconds",
			}
			objRef1 := &corev1.ObjectReference{
				APIVersion: "iter8.tools/v2alpha2",
				Kind:       "Experiment",
				Name:       name,
				Namespace:  namespace,
				FieldPath:  ".spec.Duration.bad",
			}
			experiment := v2alpha2.NewExperiment(name, namespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithBaselineVersion("baseline", objRefb).
				WithCandidateVersion("candidate-1", objRef1).
				WithDuration(10, 5, 3).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that the experiment failed and the expected reason is recorded")
			Eventually(func() bool { return fails(name, namespace) }, 5).Should(BeTrue())
			Eventually(func() bool { return issuedEvent("Specification weightObjectRef invalid") }).Should(BeTrue())
		})
	})

	Context("When create an experiment where all versions have a weightRefObj", func() {
		name := "observe-weights-all"
		It("should read all the weights", func() {
			objRef := &corev1.ObjectReference{
				APIVersion: "iter8.tools/v2alpha2",
				Kind:       "Experiment",
				Name:       name,
				Namespace:  namespace,
				FieldPath:  ".spec.duration.maxLoops",
			}
			experiment := v2alpha2.NewExperiment(name, namespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDuration(10, 5, 3).
				WithBaselineVersion("baseline", objRef).
				WithCandidateVersion("candidate", objRef).
				Build()

			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			Eventually(func() bool {
				return hasValue(name, namespace, func(exp *v2alpha2.Experiment) bool {
					return len(exp.Status.CurrentWeightDistribution) == 2 &&
						exp.Status.CurrentWeightDistribution[0].Name == "baseline" &&
						exp.Status.CurrentWeightDistribution[0].Value == 3 &&
						exp.Status.CurrentWeightDistribution[1].Name == "candidate" &&
						exp.Status.CurrentWeightDistribution[1].Value == 3
				})
			}, 5).Should(BeTrue())
		})
	})

	Context("When create an experiment where 1 version does not have a weightRefObj", func() {
		name := "observe-weights-1"
		It("should compute the missing weight", func() {
			objRef := &corev1.ObjectReference{
				APIVersion: "iter8.tools/v2alpha2",
				Kind:       "Experiment",
				Name:       name,
				Namespace:  namespace,
				FieldPath:  ".spec.duration.maxLoops",
			}
			experiment := v2alpha2.NewExperiment(name, namespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDuration(10, 5, 3).
				WithBaselineVersion("baseline", objRef).
				WithCandidateVersion("candidate", nil).
				Build()

			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			Eventually(func() bool {
				return hasValue(name, namespace, func(exp *v2alpha2.Experiment) bool {
					return len(exp.Status.CurrentWeightDistribution) == 2 &&
						exp.Status.CurrentWeightDistribution[0].Name == "baseline" &&
						exp.Status.CurrentWeightDistribution[0].Value == 3 &&
						exp.Status.CurrentWeightDistribution[1].Name == "candidate" &&
						exp.Status.CurrentWeightDistribution[1].Value == 97
				})
			}, 5).Should(BeTrue())
		})
	})

	Context("When create an experiment where more than one version does not have a weightRefObj", func() {
		name := "observe-weights-2"
		It("should not compute the missing weights; it should fail", func() {
			objRef := &corev1.ObjectReference{
				APIVersion: "iter8.tools/v2alpha2",
				Kind:       "Experiment",
				Name:       name,
				Namespace:  namespace,
				FieldPath:  ".spec.duration.maxLoops",
			}
			experiment := v2alpha2.NewExperiment(name, namespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDuration(10, 5, 3).
				WithBaselineVersion("baseline", objRef).
				WithCandidateVersion("candidate", nil).
				WithCandidateVersion("candidate2", nil).
				Build()

			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			Eventually(func() bool { return fails(name, namespace) }, 5).Should(BeTrue())
			Eventually(func() bool { return issuedEvent("Specification weightObjectRef invalid") }).Should(BeTrue())
		})
	})

})

var _ = Describe("patch", func() {
	var namespace string
	BeforeEach(func() {
		namespace = "default"
		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Experiment{}, client.InNamespace(namespace))
	})
	Context("path gets updated", func() {
		name := "write"
		var bldr *v2alpha2.ExperimentBuilder
		var objRef *corev1.ObjectReference
		JustBeforeEach(func() {
			bldr = v2alpha2.NewExperiment(name, namespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDuration(10, 5, 3)

			objRef = &corev1.ObjectReference{
				APIVersion: "iter8.tools/v2alpha2",
				Kind:       "Experiment",
				Name:       name,
				Namespace:  namespace,
			}
		})
		It("Should patch the desired field", func() {
			By("Create experiment")
			objRef.FieldPath = ".spec.duration.maxLoops"
			experiment := bldr.WithBaselineVersion("v1", nil).
				WithCandidateVersion("v2", objRef).
				Build()

			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			time.Sleep(3 * time.Second)
			exp := v2alpha2.Experiment{}
			Expect(k8sClient.Get(ctx(), types.NamespacedName{Namespace: namespace, Name: name}, &exp)).Should(Succeed())
			By("Updating experiment with recommended weights")
			exp.Status.Analysis = &v2alpha2.Analysis{
				Weights: &v2alpha2.WeightsAnalysis{
					AnalysisMetaData: v2alpha2.AnalysisMetaData{
						Provenance: "provenance",
						Timestamp:  metav1.Now(),
					},
					Data: []v2alpha2.WeightData{
						{Name: "v1", Value: 14},
						{Name: "v2", Value: 16},
					},
				},
			}
			Expect(k8sClient.Status().Update(ctx(), &exp)).Should(Succeed())
			By("calling redistributeWeight")
			Expect(shouldRedistribute(&exp)).Should(BeTrue())
			Expect(redistributeWeight(ctx(), &exp, reconciler.RestConfig)).Should(Succeed())
			By("verifying that the weight was changed")
			value, _ := observeWeight(ctx(), objRef, namespace, cfg)
			Expect(*value).To(Equal(int32(16)))
		})
	})
})

var _ = Describe("Weight Patching", func() {
	restCfg, _ := config.GetConfig()
	namespace := "default"

	ctx := context.Background()
	ctx = context.WithValue(ctx, util.LoggerKey, ctrl.Log)

	Context("When experimentType is Conformance", func() {
		experiment := v2alpha2.NewExperiment("noVersionInfo", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternConformance).
			Build()
		It("should succeed without error", func() {
			Expect(redistributeWeight(ctx, experiment, restCfg)).Should(Succeed())
		})
	})

	Context("When algorithm is FixedSplit", func() {
		experiment := v2alpha2.NewExperiment("noVersionInfo", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			WithDeploymentPattern(v2alpha2.DeploymentPatternFixedSplit).
			Build()
		It("should succeed without error", func() {
			Expect(redistributeWeight(ctx, experiment, restCfg)).Should(Succeed())
		})
	})

	Context("When no versionInfo", func() {
		experiment := v2alpha2.NewExperiment("noVersionInfo", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			Build()
		It("Should fail with error", func() {
			err := redistributeWeight(ctx, experiment, restCfg)
			Expect(err).Should(MatchError("Cannot redistribute weight; no version information present"))
		})
	})

	Context("When WeightObjRef is not set", func() {
		experiment := v2alpha2.NewExperiment("noWeightObRef", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			WithDuration(10, 0, 1).
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
		experiment := v2alpha2.NewExperiment("noFieldPath", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			WithDuration(10, 0, 1).
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
		experiment := v2alpha2.NewExperiment("noWeightRecommendation", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			WithDuration(10, 0, 1).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  ".path.to.weight",
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
		experiment := v2alpha2.NewExperiment("recommendationIsCurrent", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			WithDuration(10, 0, 1).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  ".path.to.weight",
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
		experiment := v2alpha2.NewExperiment("recommendationNotCurrent", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			WithDuration(10, 0, 1).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  ".path.to.weight",
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
		experiment := v2alpha2.NewExperiment("recommendationNotCurrent", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			WithDuration(10, 0, 1).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  ".path.to.weight[0]",
			}).
			WithCandidateVersion("candidate", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs",
				Namespace:  namespace,
				FieldPath:  ".path.to.weight[1]",
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
		experiment := v2alpha2.NewExperiment("recommendationNotCurrent", namespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternCanary).
			WithDuration(10, 0, 1).
			WithBaselineVersion("baseline", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs0",
				Namespace:  namespace,
				FieldPath:  ".path.to.weight[0]",
			}).
			WithCandidateVersion("candidate", &corev1.ObjectReference{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
				Name:       "vs1",
				Namespace:  namespace,
				FieldPath:  ".path.to.weight[1]",
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
