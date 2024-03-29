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
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestRemoveString(t *testing.T) {
	sl := []string{"hello", "world", "goodbye", "everyone"}
	res := removeString(sl, "world")
	assert.Equal(t, []string{"hello", "goodbye", "everyone"}, res)
}

var _ = Describe("Experiment Validation", func() {
	var jqe string = "expr"
	var url string = "url"

	ctx := context.Background()

	Context("When creating an experiment with an invalid spec.duration.maxIteration", func() {
		testName := "test-invalid-duration"
		testNamespace := "default"
		It("Should fail to create experiment", func() {
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDuration(10, 0, 1).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).ShouldNot(Succeed())
		})
	})

	Context("When creating an experiment with a valid spec.duration.maxIteration", func() {
		testName := "test-valid-duration"
		testNamespace := "default"
		It("Should succeed in creating experiment", func() {
			ctx := context.Background()
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDuration(10, 1, 1).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())
		})
	})

	Context("When creating a valid new Experiment", func() {
		It("Should successfully complete late initialization", func() {
			By("Providing a request-count metric")
			m := v2alpha2.NewMetric("request-count", "iter8").
				WithType(v2alpha2.CounterMetricType).
				WithParams([]v2alpha2.NamedValue{{
					Name:  "param",
					Value: "value",
				}}).
				WithProvider("prometheus").
				WithJQExpression(&jqe).
				WithURLTemplate(&url).
				Build()
			// ns := &corev1.Namespace{
			// 	ObjectMeta: metav1.ObjectMeta{Name: "iter8"},
			// }
			// Expect(k8sClient.Create(ctx, ns)).Should(Succeed())
			Expect(k8sClient.Create(ctx, m)).Should(Succeed())
			// createdMetric := &v2alpha2.Metric{}
			// Eventually(func() bool {
			// 	err := k8sClient.Get(ctx, types.NamespacedName{Name: "request-count", Namespace: "iter8"}, createdMetric)
			// 	if err != nil {
			// 		return false
			// 	}
			// 	return true
			// }).Should(BeTrue())
			By("creating a reward metric")
			reward := v2alpha2.NewMetric("reward", "default").
				WithType(v2alpha2.CounterMetricType).
				WithParams([]v2alpha2.NamedValue{{
					Name:  "param",
					Value: "value",
				}}).
				WithProvider("prometheus").
				WithJQExpression(&jqe).
				WithURLTemplate(&url).
				Build()
			Expect(k8sClient.Create(ctx, reward)).Should(Succeed())
			By("creating an indicator")
			indicator := v2alpha2.NewMetric("indicataor", "default").
				WithType(v2alpha2.CounterMetricType).
				WithParams([]v2alpha2.NamedValue{{
					Name:  "param",
					Value: "value",
				}}).
				WithProvider("prometheus").
				WithJQExpression(&jqe).
				WithURLTemplate(&url).
				Build()
			Expect(k8sClient.Create(ctx, indicator)).Should(Succeed())
			By("creating an objective")
			objective := v2alpha2.NewMetric("objective", "default").
				WithType(v2alpha2.CounterMetricType).
				WithParams([]v2alpha2.NamedValue{{
					Name:  "param",
					Value: "value",
				}}).
				WithProvider("prometheus").
				WithJQExpression(&jqe).
				WithURLTemplate(&url).
				Build()
			Expect(k8sClient.Create(ctx, objective)).Should(Succeed())
			By("creating an objective that is not in the cluster")
			fake := v2alpha2.NewMetric("fake", "default").
				WithType(v2alpha2.CounterMetricType).
				WithParams([]v2alpha2.NamedValue{{
					Name:  "param",
					Value: "value",
				}}).
				WithProvider("prometheus").
				WithJQExpression(&jqe).
				WithURLTemplate(&url).
				Build()

			By("Creating a new Experiment")
			testName := "late-initialization"
			testNamespace := "default"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithRequestCount("request-count").
				WithReward(*reward, v2alpha2.PreferredDirectionHigher).
				WithIndicator(*indicator).
				WithObjective(*objective, nil, nil, false).
				WithObjective(*fake, nil, nil, true).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			By("Getting experiment after late initialization has run (spec.Duration !=- nil)")
			Eventually(func() bool {
				return hasValue(testName, testNamespace, func(exp *v2alpha2.Experiment) bool {
					return exp.Status.InitTime != nil &&
						exp.Status.LastUpdateTime != nil &&
						exp.Status.CompletedIterations != nil &&
						len(exp.Status.Conditions) == 3
				})
			}).Should(BeTrue())
		})
	})
})

var _ = Describe("Experiment proceeds", func() {
	ctx := context.Background()

	Context("Early event trigger", func() {
		testName := "early-reconcile"
		testNamespace := "default"
		It("Experiment should complete", func() {
			By("Creating Experiment with 10s interval")
			expectedIterations := int32(2)
			initialInterval := int32(5)
			modifiedInterval := int32(10)
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("early-reconcile-targets").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithDuration(initialInterval, expectedIterations, 1).
				WithDeploymentPattern(v2alpha2.DeploymentPatternFixedSplit).
				WithBaselineVersion("baseline", nil).
				WithCandidateVersion("candidate", nil).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			By("Changing the interval before the reconcile event triggers")
			time.Sleep(2 * time.Second)
			createdExperiment := &v2alpha2.Experiment{}
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

var _ = Describe("Empty Criteria section", func() {
	var dataDir string = "../test/data"

	Context("When the Criteria section has empty lists", func() {
		var testName string = "norealcriteria.yaml"
		experiment := v2alpha2.Experiment{}
		readExperimentFromFile(path.Join(dataDir, testName), &experiment)

		Specify("The experiment should read the (non-existent) metrics", func() {
			Expect(k8sClient.Create(ctx(), &experiment)).Should(Succeed())
			// will fail after this point because there is no versionInfo is present
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonInvalidExperiment)
			}, 5).Should(BeTrue())
		})
	})

})

var _ = Describe("Missing criteria.requestCount", func() {
	var jqe string = "expr"
	var url string = "url"

	var testNamespace string = "default"
	Context("When there is no criteria.requestCount", func() {
		Specify("The controller should read the other metrics", func() {
			var testName string = "norequestcount"
			By("Defining a Gauge metric that references requestcount")
			metric := v2alpha2.NewMetric("referencesrequestcount", "default").
				WithType("Gauge").
				WithProvider("provider").
				WithJQExpression(&jqe).
				WithURLTemplate(&url).
				WithSampleSize("requestcount").
				Build()
			Expect(k8sClient.Create(ctx(), metric)).Should(Succeed())
			By("Defining an experiment with no request count")
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternType(v2alpha2.TestingPatternConformance)).
				WithIndicator(*metric).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			// will fail because samplesize reference is not available
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonMetricUnavailable)
			}, 5).Should(BeTrue())
		})
	})
})

var _ = Describe("Loop Execution", func() {
	var testName string
	var testNamespace string = "default"
	BeforeEach(func() {
		testNamespace = "default"

		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Experiment{}, client.InNamespace(testNamespace))
	})
	AfterEach(func() {
		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Experiment{}, client.InNamespace(testNamespace))
	})
	Context("When creating an experiment with 3 loops", func() {
		// experiment (in default namespace) refers to metric "objective-with-good-reference"
		// which has a sampleSize "metricNamespace/request-count" which is correct
		It("Should successfully execute three times", func() {
			By("Creating experiment")
			testName = "loops"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternConformance).
				WithBaselineVersion("baseline", nil).
				WithDuration(1, 1, 3).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it loops exactly 3 times")
			Eventually(func() bool {
				return containsSubString(events, "Completed Loop 3")
			}, 5).Should(BeTrue())
			Eventually(func() bool {
				return containsSubString(events, "Completed Loop 4")
			}, 1).Should(BeFalse())

		})
	})
})
