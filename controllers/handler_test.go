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
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Test that we start the appropriate handlers at the appropriate times.
// Test deleteHandlerJob at the end as well
var _ = Describe("Handlers Run", func() {
	var (
		namespace string
	)
	BeforeEach(func() {
		namespace = "default"

		k8sClient.DeleteAllOf(ctx(), &v2alpha1.Experiment{}, client.InNamespace(namespace))
	})

	Context("When an experiment with a start handler runs", func() {
		Specify("the start handler is run", func() {
			By("Defining an experiment with a start handler")
			name, target := "has-start-handler", "has-start-handler"
			handler := "test-handler"
			iterations, loops := int32(2), int32(1)
			experiment := v2alpha1.NewExperiment(name, namespace).
				WithTarget(target).
				WithTestingPattern(v2alpha1.TestingPatternConformance).
				WithHandlers(map[string]string{"start": handler}).
				WithDuration(1, iterations, loops).
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())

			By("Checking that the start handler jobs are created")
			Eventually(func() bool {
				handlerJob := &batchv1.Job{}
				jbNm := jobName(experiment, handler, nil)
				err := k8sClient.Get(ctx(), types.NamespacedName{Name: jbNm, Namespace: "iter8"}, handlerJob)
				return err == nil
			}, 3).Should(BeTrue())
		})
	})
	Context("When an experiment with a finish handler finishes", func() {
		Specify("the finish handler is run", func() {
			By("Defining an experiment with a finsih handler")
			// for simplicity, no start handler
			name, target := "has-finish-handler", "has-finish-handler"
			handler := "test-handler"
			iterations, loops := int32(2), int32(1)
			experiment := v2alpha1.NewExperiment(name, namespace).
				WithTarget(target).
				WithTestingPattern(v2alpha1.TestingPatternConformance).
				WithHandlers(map[string]string{"start": "none", "finish": handler}).
				WithDuration(1, iterations, loops).
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that the finish handler jobs are created")
			Eventually(func() bool {
				handlerJob := &batchv1.Job{}
				jbNm := jobName(experiment, handler, nil)
				err := k8sClient.Get(ctx(), types.NamespacedName{Name: jbNm, Namespace: "iter8"}, handlerJob)
				return err == nil
			}, 10).Should(BeTrue())
			By("Checking that the experiment has executed all iterations")
			Eventually(func() bool {
				return hasValue(name, namespace, func(exp *v2alpha1.Experiment) bool {
					return exp.Status.GetCompletedIterations() == iterations*loops
				})
			}).Should(BeTrue())
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
		Specify("the loop handler is started", func() {
			By("Defining an experiment with a loop handler")
			name, target := "has-loop-handler", "has-loop-handler"
			handler := "test-handler"
			iterations, loops := int32(1), int32(2)
			testLoop := 1
			experiment := v2alpha1.NewExperiment(name, namespace).
				WithTarget(target).
				WithTestingPattern(v2alpha1.TestingPatternConformance).
				WithHandlers(map[string]string{"start": "none", "loop": handler}).
				WithDuration(1, iterations, loops).
				WithBaselineVersion("baseline", nil).
				Build()
			Expect(k8sClient.Create(ctx(), experiment)).Should(Succeed())
			By("Checking that the loop handler jobs are created")
			Eventually(func() bool {
				handlerJob := &batchv1.Job{}
				jbNm := jobName(experiment, handler, &testLoop)
				err := k8sClient.Get(ctx(), types.NamespacedName{Name: jbNm, Namespace: "iter8"}, handlerJob)
				return err == nil
			}, 10).Should(BeTrue())
		})
	})

})
