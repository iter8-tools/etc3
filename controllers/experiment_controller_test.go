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

	v2beta1 "github.com/iter8-tools/etc3/api/v2beta1"
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
	ctx := context.Background()

	Context("When creating an experiment with an invalid spec.duration.maxLoops", func() {
		testName := "test-invalid-duration"
		testNamespace := "default"
		It("Should fail to create experiment", func() {
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithDuration(10, 0).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).ShouldNot(Succeed())
		})
	})

	Context("When creating an experiment with a valid spec.duration.maxLoops", func() {
		testName := "test-valid-duration"
		testNamespace := "default"
		It("Should succeed in creating experiment", func() {
			ctx := context.Background()
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithDuration(10, 1).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())
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
			expectedLoops := int32(2)
			initialInterval := int32(5)
			modifiedInterval := int32(10)
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").WithVersion("candidate").
				WithDuration(initialInterval, expectedLoops).
				Build()
			Expect(k8sClient.Create(ctx, experiment)).Should(Succeed())

			By("Changing the interval before the reconcile event triggers")
			time.Sleep(2 * time.Second)
			createdExperiment := &v2beta1.Experiment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: testName, Namespace: testNamespace}, createdExperiment)).Should(Succeed())
			createdExperiment.Spec.Duration.MinIntervalBetweenLoops = &modifiedInterval
			Expect(k8sClient.Update(ctx, createdExperiment)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: testName, Namespace: testNamespace}, createdExperiment)
				if err != nil {
					return false
				}
				return createdExperiment.Status.GetCompletedLoops() == expectedLoops
				// return true
			}, 2*modifiedInterval*expectedLoops).Should(BeTrue())

		})
	})
})

var _ = Describe("Empty Criteria section", func() {
	var dataDir string = "../test/data"

	Context("When the Criteria section has empty lists", func() {
		var testName string = "norealcriteria.yaml"
		experiment := v2beta1.Experiment{}
		readExperimentFromFile(path.Join(dataDir, testName), &experiment)

		Specify("The experiment should read the (non-existent) metrics", func() {
			Expect(k8sClient.Create(ctx(), &experiment)).Should(Succeed())
			Eventually(func() bool {
				return issuedEvent(v2beta1.ReasonExperimentCompleted)
			}, 5).Should(BeTrue())
		})
	})

})

var _ = Describe("Loop Execution", func() {
	var testName string
	var testNamespace string = "default"
	BeforeEach(func() {
		testNamespace = "default"

		k8sClient.DeleteAllOf(ctx(), &v2beta1.Experiment{}, client.InNamespace(testNamespace))
	})
	AfterEach(func() {
		k8sClient.DeleteAllOf(ctx(), &v2beta1.Experiment{}, client.InNamespace(testNamespace))
	})
	Context("When creating an experiment with 3 loops", func() {
		// experiment (in default namespace) refers to metric "objective-with-good-reference"
		// which has a sampleSize "metricNamespace/request-count" which is correct
		It("Should successfully execute three times", func() {
			By("Creating experiment")
			testName = "loops"
			experiment := v2beta1.NewExperiment(testName, testNamespace).
				WithVersion("baseline").
				WithDuration(1, 3).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that it loops exactly 3 times")
			Eventually(func() bool {
				return issuedEvent("Completed Loop 3")
			}, 5).Should(BeTrue())
			Eventually(func() bool {
				return issuedEvent("Completed Loop 4")
			}, 1).Should(BeFalse())

		})
	})
})
