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

	v2beta1 "github.com/iter8-tools/etc3/api/v2beta1"
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

	Context("Experiment is a SLOValidation test", func() {
		var bldr *v2beta1.ExperimentBuilder
		BeforeEach(func() {
			bldr = v2beta1.NewExperiment("conformance-test", testNamespace)
		})
		It("should be invalid when no versions are specified", func() {
			experiment := bldr.
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when exactly 1 version is specified", func() {
			experiment := bldr.
				WithVersion("baseline").
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be valid when 2 versions are specified", func() {
			experiment := bldr.
				WithVersion("baseline").WithVersion("candidate-1").
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when there is a reward (1 version)", func() {
			experiment := bldr.
				WithVersion("baseline").
				WithReward(*v2beta1.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2beta1.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

	Context("Experiment is an AB test", func() {
		var bldr *v2beta1.ExperimentBuilder
		BeforeEach(func() {
			bldr = v2beta1.NewExperiment("ab-test", testNamespace).
				WithVersion("v1").WithVersion("v2")
		})

		It("should be valid when there is a single reward", func() {
			experiment := bldr.
				WithReward(*v2beta1.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2beta1.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when there is are multiple rewards", func() {
			experiment := bldr.
				WithReward(*v2beta1.NewMetric("metric-1", "default").WithJQExpression(&jqe).Build(), v2beta1.PreferredDirectionHigher).
				WithReward(*v2beta1.NewMetric("metric-2", "default").WithJQExpression(&jqe).Build(), v2beta1.PreferredDirectionHigher).
				Build()
			Expect(len(experiment.Spec.VersionInfo)).Should(Equal(2))
			Expect(experiment.Spec.Criteria).ShouldNot(BeNil())
			Expect(len(experiment.Spec.Criteria.Rewards)).Should(Equal(2))
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

	Context("Experiment is an ABN test", func() {
		var bldr *v2beta1.ExperimentBuilder
		BeforeEach(func() {
			bldr = v2beta1.NewExperiment("abn-test", testNamespace).
				WithVersion("v1").WithVersion("v2").WithVersion("v3")
		})

		It("should be invalid when no reward", func() {
			experiment := bldr.
				WithVersion("baseline").
				WithVersion("candidate-1").
				WithVersion("candidate-2").
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})

		It("should be valid when 1 reward", func() {
			experiment := bldr.
				WithVersion("baseline").
				WithVersion("candidate-1").
				WithVersion("candidate-2").
				WithReward(*v2beta1.NewMetric("metric", "default").WithJQExpression(&jqe).Build(), v2beta1.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeTrue())
		})

		It("should be invalid when more than 1 reward", func() {
			experiment := bldr.
				WithVersion("baseline").
				WithVersion("candidate-1").
				WithVersion("candidate-2").
				WithReward(*v2beta1.NewMetric("metric-1", "default").WithJQExpression(&jqe).Build(), v2beta1.PreferredDirectionHigher).
				WithReward(*v2beta1.NewMetric("metric-2", "default").WithJQExpression(&jqe).Build(), v2beta1.PreferredDirectionHigher).
				Build()
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

	Context("Experiment has common names", func() {
		experiment := v2beta1.NewExperiment("abn-test", testNamespace).
			WithVersion("baseline").WithVersion("candidate").WithVersion("candidate").
			Build()
		It("should fail", func() {
			Expect(reconciler.IsVersionInfoValid(ctx, experiment)).Should(BeFalse())
		})
	})

})
