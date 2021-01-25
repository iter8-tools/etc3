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
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validation of VersionInfo", func() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, util.LoggerKey, ctrl.Log)
	testNamespace := "default"

	// The first test validates when no VersionInfo is present (all strategies)
	// The way to have no versions is to have no VersionInfo

	Context("Experiment is a Conformance test", func() {
		bldr := v2alpha1.NewExperiment("conformance-test", testNamespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeConformance)
		It("should be valid only when 1 version is identified", func() {
			By("returning false when no versions are specified")
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning true when 1 version is identified")
			bldr = bldr.WithBaselineVersion("baseline", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeTrue())
			By("returning false when more than 1 version is identified")
			bldr = bldr.WithCandidateVersion("candidate-1", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
		})
	})

	Context("Experiment is an AB test", func() {
		bldr := v2alpha1.NewExperiment("ab-test", testNamespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeAB)
		It("should be valid only when 2 versions are identified", func() {
			By("returning false when no versions are specified")
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning false when 1 version is identified")
			bldr = bldr.WithBaselineVersion("baseline", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning true when exactly than 2 versions are identified")
			bldr = bldr.WithCandidateVersion("candidate-1", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeTrue())
			By("returning false when more than 2 version are identified")
			bldr = bldr.WithCandidateVersion("candidate-2", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
		})
	})

	Context("Experiment is an Canary test", func() {
		bldr := v2alpha1.NewExperiment("canary-test", testNamespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeCanary)
		It("should be valid only when 2 versions are identified", func() {
			By("returning false when no versions are specified")
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning false when 1 version is identified")
			bldr = bldr.WithBaselineVersion("baseline", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning true when exactly than 2 versions are identified")
			bldr = bldr.WithCandidateVersion("candidate-1", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeTrue())
			By("returning false when more than 2 version are identified")
			bldr = bldr.WithCandidateVersion("candidate-2", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
		})
	})

	Context("Experiment is an BlueGreen test", func() {
		bldr := v2alpha1.NewExperiment("bluegreen-test", testNamespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeBlueGreen)
		It("should be valid only when 2 versions are identified", func() {
			By("returning false when no versions are specified")
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning false when 1 version is identified")
			bldr = bldr.WithBaselineVersion("baseline", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning true when exactly than 2 versions are identified")
			bldr = bldr.WithCandidateVersion("candidate-1", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeTrue())
			By("returning false when more than 2 version are identified")
			bldr = bldr.WithCandidateVersion("candidate-2", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
		})
	})

	Context("Experiment is an ABN test", func() {
		bldr := v2alpha1.NewExperiment("abn-test", testNamespace).
			WithTarget("target").
			WithStrategy(v2alpha1.StrategyTypeABN)
		It("should be valid only when more than 2 or more versions are identified", func() {
			By("returning false when no versions are specified")
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning false when 1 version is identified")
			bldr = bldr.WithBaselineVersion("baseline", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeFalse())
			By("returning false when 2 versions are identified")
			bldr = bldr.WithCandidateVersion("candidate-1", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeTrue())
			By("returning true when more than 2 version are identified")
			bldr = bldr.WithCandidateVersion("candidate-2", nil)
			Expect(reconciler.IsVersionInfoValid(ctx, bldr.Build())).Should(BeTrue())
		})
	})

})
