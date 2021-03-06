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

var _ = Describe("Metrics", func() {
	var jqe string = "expr"
	var url string = "url"

	var testName string
	var testNamespace, metricsNamespace string
	var goodObjective, goodObjective2, badObjective, reward *v2alpha2.Metric
	BeforeEach(func() {
		testNamespace = "default"
		metricsNamespace = "metric-namespace"

		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Experiment{}, client.InNamespace(testNamespace))
		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Metric{}, client.InNamespace(testNamespace))
		k8sClient.DeleteAllOf(ctx(), &v2alpha2.Metric{}, client.InNamespace(metricsNamespace))

		By("Providing a request-count metric")
		m := v2alpha2.NewMetric("request-count", metricsNamespace).
			WithType(v2alpha2.CounterMetricType).
			WithParams([]v2alpha2.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			Build()
		Expect(k8sClient.Create(ctx(), m)).Should(Succeed())
		goodObjective2 = v2alpha2.NewMetric("objective-with-good-reference-2", metricsNamespace).
			WithType(v2alpha2.CounterMetricType).
			WithParams([]v2alpha2.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			WithSampleSize("request-count").
			Build()
		Expect(k8sClient.Create(ctx(), goodObjective2)).Should(Succeed())
		By("creating an objective that does not reference the request-count")
		goodObjective = v2alpha2.NewMetric("objective-with-good-reference", "default").
			WithType(v2alpha2.CounterMetricType).
			WithParams([]v2alpha2.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			WithSampleSize(metricsNamespace + "/request-count").
			Build()
		Expect(k8sClient.Create(ctx(), goodObjective)).Should(Succeed())
		By("creating an objective that references request-count")
		badObjective = v2alpha2.NewMetric("objective-with-bad-reference", "default").
			WithType(v2alpha2.CounterMetricType).
			WithParams([]v2alpha2.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			WithSampleSize("request-count").
			Build()
		Expect(k8sClient.Create(ctx(), badObjective)).Should(Succeed())
		reward = v2alpha2.NewMetric("rwrd", "default").
			WithType(v2alpha2.CounterMetricType).
			WithParams([]v2alpha2.NamedValue{{
				Name:  "param",
				Value: "value",
			}}).
			WithProvider("prometheus").
			WithJQExpression(&jqe).
			WithURLTemplate(&url).
			Build()
		Expect(k8sClient.Create(ctx(), reward)).Should(Succeed())
	})

	Context("When creating an experiment referencing valid metrics", func() {
		// experiment (in default namespace) refers to metric "objective-with-good-reference"
		// which has a sampleSize "metricNamespace/request-count" which is correct
		It("Should successfully read the metrics and proceed", func() {
			By("Creating experiment")
			testName = "valid-reference"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithRequestCount(metricsNamespace+"/request-count").
				WithObjective(*goodObjective, nil, nil, false).
				WithReward(*reward, v2alpha2.PreferredDirectionHigher).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it starts Running")
			// this assumes that it runs for a while
			Eventually(func() bool {
				return containsSubString(events, "Advanced to Running") //v2alpha2.ReasonStageAdvanced)
			}, 5).Should(BeTrue())
		})
	})

	Context("failed start handler", func() {
		Specify("experiment teminated in a failed state", func() {
			By("Creating an experiment with a start handler")
			name, target := "has-failing-handler", "has-failing-handler"
			iterations, loops := int32(2), int32(1)
			handler := "start"
			experiment := v2alpha2.NewExperiment(name, testNamespace).
				WithTarget(target).
				WithTestingPattern(v2alpha2.TestingPatternConformance).
				WithAction(handler, []v2alpha2.TaskSpec{}).
				WithRequestCount(metricsNamespace+"/request-count").
				WithDuration(30, iterations, loops).
				WithBaselineVersion("baseline", nil).
				Build()

			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			Eventually(func() bool { return fails(name, testNamespace) }, 5).Should(BeTrue())
		})
	})

	Context("When creating an experiment which refers to a non-existing metric", func() {
		// experiment (in default ns) refers to metric "request-count" (not in default namespace)
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "invalid-metric"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithRequestCount("request-count").
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonMetricUnavailable) &&
					containsSubString(events, "default/request-count")
			}, 5).Should(BeTrue())
			Eventually(func() bool { return fails(testName, testNamespace) }, 5).Should(BeTrue())
		})
	})
	Context("When creating another experiment which refers to a non-existing metric", func() {
		// experiment (in default ns) refers to metric "iter8/request-count" (not in iter8 namespace)
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "invalid-metric"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithRequestCount("iter8/request-count").
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonMetricUnavailable)
			}, 5).Should(BeTrue())
			Eventually(func() bool { return fails(testName, testNamespace) }, 5).Should(BeTrue())
		})
	})

	Context("When creating an experiment referencing a metric with a bad reference", func() {
		// experiment (in default namespace) refers to metric "objective-with-bad-reference"
		// which has a sampleSize "request-count" (not in same ns as the referring metric (default))
		It("Should fail to read metrics", func() {
			By("Creating experiment")
			testName = "invalid-reference"
			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithRequestCount(metricsNamespace+"/request-count").
				WithObjective(*badObjective, nil, nil, false).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it fails")
			// this depends on an experiment that should run for a while
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonMetricUnavailable)
			}, 5).Should(BeTrue())
			// Eventually(func() bool { return fails(testName, testNamespace) }, 5).Should(BeTrue())
		})
	})

	Context("When creating an experiment referencing a metric with a bad reference", func() {
		// experiment (in default namespace) refers to metric "objective-with-bad-reference"
		// which has a sampleSize "request-count" (not in same ns as the referring metric (default))
		It("Should successfully read metrics", func() {
			By("Creating experiment")
			testName = "good-reference-2"

			experiment := v2alpha2.NewExperiment(testName, testNamespace).
				WithTarget("target").
				WithTestingPattern(v2alpha2.TestingPatternCanary).
				WithRequestCount(metricsNamespace+"/objective-with-good-reference-2").
				WithObjective(*goodObjective2, nil, nil, false).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it starts Running")
			// this assumes that it runs for a while
			Eventually(func() bool {
				return containsSubString(events, v2alpha2.ReasonStageAdvanced)
			}, 5).Should(BeTrue())
		})
	})
	Context("When converting a namespacedname", func() {
		var ns *string
		var nm string
		It("Should return nil, '' on input ''", func() {
			ns, nm = namespaceName("")
			Expect(ns).To(BeNil())
			Expect(nm).To(Equal(""))
		})
		It("Should return nil, 'name' on input 'name'", func() {
			ns, nm = namespaceName("name")
			Expect(ns).To(BeNil())
			Expect(nm).To(Equal("name"))
		})
		It("Should return 'namespace', 'name' on input 'namespace/name", func() {
			ns, nm = namespaceName("namespace/name")
			Expect(ns).ToNot(BeNil())
			Expect(*ns).To(Equal("namespace"))
			Expect(nm).To(Equal("name"))
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
