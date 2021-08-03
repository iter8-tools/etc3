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

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validation of VersionInfo", func() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, LoggerKey, ctrl.Log)
	testNamespace := "default"
	var jqe string = "expr"

	// The first test validates when no VersionInfo is present (all strategies)
	// The way to have no versions is to have no VersionInfo

	Context("Experiment is a Conformance test", func() {
		var bldr *v2alpha2.ExperimentBuilder
		BeforeEach(func() {
			bldr = v2alpha2.NewExperiment("conformance-test", testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternConformance)
		})
		It("should be invalid when no versions are specified", func() {
			experiment := bldr.
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when exactly 1 version is specified", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when candidate versions are specified", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when there is no reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when there is a reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

	Context("Experiment is an AB test", func() {
		var bldr *v2alpha2.ExperimentBuilder
		BeforeEach(func() {
			bldr = v2alpha2.NewExperiment("ab-test", testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternAB)
		})
		It("should be invalid when no versions are specified", func() {
			experiment := bldr.
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be invalid when exactly 1 version is specified", func() {
			experiment := bldr.
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when exactly 2 versions are specified", func() {
			experiment := bldr.
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when more than 2 version are specified", func() {
			experiment := bldr.
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithCandidateVersion("candidate-2", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be invalid when there is no reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when there is a single reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})
		It("should be invalid when there is are multiple rewards", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithReward(*v2alpha2.NewMetric("metric-1", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithReward(*v2alpha2.NewMetric("metric-2", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

	Context("Experiment is an Canary test", func() {
		var bldr *v2alpha2.ExperimentBuilder
		BeforeEach(func() {
			bldr = v2alpha2.NewExperiment("canary-test", testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary)
		})

		It("should be invalid when no versions are specified", func() {
			experiment := bldr.
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be invalid when exactly 1 version is specified", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when exactly 2 versions are specified", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when more than 2 version are specified", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithCandidateVersion("candidate-2", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when there is no reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when there is a reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

	Context("Experiment is an ABN test", func() {
		var bldr *v2alpha2.ExperimentBuilder
		BeforeEach(func() {
			bldr = v2alpha2.NewExperiment("abn-test", testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternABN)
		})

		It("should be invalid when no versions are specified", func() {
			experiment := bldr.
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be invalid when exactly 1 version is specified", func() {
			experiment := bldr.
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when exactly 2 versions are specified", func() {
			experiment := bldr.
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be valid when more than 2 version are specified", func() {
			experiment := bldr.
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithCandidateVersion("candidate-2", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when no reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithCandidateVersion("candidate-2", nil).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when 1 reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithCandidateVersion("candidate-2", nil).
				WithReward(*v2alpha2.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when more than 1 reward", func() {
			experiment := bldr.
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate-1", nil).
				WithCandidateVersion("candidate-2", nil).
				WithReward(*v2alpha2.NewMetric("metric-1", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				WithReward(*v2alpha2.NewMetric("metric-2", "default").WithJQExpression(&jqe).Build(), v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

	Context("Experiment has common names", func() {
		experiment := v2alpha2.NewExperiment("abn-test", testNamespace).
			WithTarget("target").
			WithBaselineVersion("baseline", nil).
			WithCandidateVersion("candidate", nil).
			WithTestingPattern(v2alpha2.TestingPatternABN).
			Build()
		It("should fail", func() {
			By("adding another canidate with the same name")
			experiment.Spec.VersionInfo.Candidates = append(experiment.Spec.VersionInfo.Candidates,
				v2alpha2.VersionDetail{Name: "candidate"})
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

	Context("Spec.VersionInfo.*.WeightObjRef.FieldPath validity", func() {
		bldr := v2alpha2.NewExperiment("invalid-fieldpath", testNamespace).
			WithTarget("target").
			WithTestingPattern(v2alpha2.TestingPatternConformance)
		It("Should reject the experiment if fieldpath starts without '.'", func() {
			experiment := bldr.WithBaselineVersion("baseline", &v1.ObjectReference{
				Name:      "object",
				Namespace: "ns",
				FieldPath: "foo",
			}).Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
			Expect(containsSubString(events, "Fieldpaths must start with '.'")).Should(BeTrue())
		})
		It("Should accept the experiment if fieldpath starts with '.'", func() {
			experiment := bldr.WithBaselineVersion("baseline", &v1.ObjectReference{
				Name:      "object",
				Namespace: "ns",
				FieldPath: ".foo",
			}).Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})
	})

})
