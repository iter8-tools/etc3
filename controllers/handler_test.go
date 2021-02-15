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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Start Handler", func() {
	var (
		testNamespace string
	)
	BeforeEach(func() {
		testNamespace = "default"

		k8sClient.DeleteAllOf(ctx(), &v2alpha1.Experiment{})
	})
	AfterEach(func() {
		k8sClient.DeleteAllOf(ctx(), &v2alpha1.Experiment{})
	})

	Context("When an experiment with a start handler runs", func() {
		Specify("the start handler is run", func() {
			By("Defining an experiment with a start handler")
			By("Checking that the start handler jobs are created")
		})
	})
	Context("When an experiment with a finish handler finishes", func() {
		Specify("the finish handler is run", func() {
			By("Defining an experiment with a finsih handler")
			// for simplicity, no start handler
			By("Checking that the finish handler jobs are created")
		})
	})
	Context("When an experiment with a failure handler fails", func() {
		Specify("the failure handler runs", func() {
			By("Defining an experiment with a finish handler")
			// for simplicity, no start handler
			By("Checking that the finish handler jobs are created")
		})
	})
	Context("When an experiment with a loop handler passes loop boundry", func() {
		var testName string = "has-loop-handler"
		It("the loop handler is started", func() {
			By("Defining an experiment with a loop handler")
			experiment := v2alpha1.NewExperiment(testName, testNamespace).
				WithTarget("unavailable-target-finalizer").
				WithTestingPattern(v2alpha1.TestingPatternConformance).
				WithHandlers(map[string]string{"start": "none", "loop": "none"}).
				WithDuration(2, 2, 2).
				WithBaselineVersion("baseline", nil).
				Build()
				// for simplicity, no start handler
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that the loop handler jobs are created")
			By("Checkong that no loop handler jobs are created for the last loop")
		})
	})

})
