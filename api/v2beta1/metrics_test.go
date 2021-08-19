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

package v2beta1

import (
	"context"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Experiment", func() {
	ctx := context.Background()

	Context("When experiment has backend and metrics", func() {
		It("should create the experiment", func() {
			By("reading experiment " + "withbackend.yaml")
			s := Experiment{}
			Expect(readExperimentFromFile(path.Join("..", "..", "test", "data", "withbackend.yaml"), &s)).To(Succeed())

			By("creating the experiment")
			Expect(k8sClient.Create(ctx, &s)).Should(Succeed())

			By("reading the experiment")
			exp := Experiment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "experiment", Namespace: "default"}, &exp)).Should(Succeed())

			By("verifying the experiment")
			Expect(len(exp.Spec.Backends) > 0).To(BeTrue())
			Expect(exp.Spec.Backends[0].Name).To(Equal("backend-1"))
			Expect(*exp.Spec.Backends[0].AuthType).To(Equal(BasicAuthType))
			Expect(*exp.Spec.Backends[0].Method).To(Equal(POSTMethodType))
			Expect(*exp.Spec.Backends[0].Provider).To(Equal("provider"))
			Expect(*exp.Spec.Backends[0].JQExpression).To(Equal("jq"))
			Expect(*exp.Spec.Backends[0].Secret).To(Equal("default/my-secret"))
			Expect(exp.Spec.Backends[0].Headers[0].Name).To(Equal("header"))
			Expect(exp.Spec.Backends[0].Headers[0].Value).To(Equal("{{.variable-1}}::{{.variable-2}}"))
			Expect(*exp.Spec.Backends[0].URL).To(Equal("https://provider.url"))
			Expect(len(exp.Spec.Backends[0].VersionInfo)).To(Equal(2))

			Expect(len(exp.Spec.Backends[0].Metrics)).Should(Equal(2))
			Expect(exp.Spec.Backends[0].Metrics[0].Name).Should(Equal("backend-1/reward-metric"))
			Expect(*exp.Spec.Backends[0].Metrics[0].Description).Should(Equal("reward-metric description"))
			Expect(len(exp.Spec.Backends[0].Metrics[0].Params)).Should(Equal(1))
			Expect(*exp.Spec.Backends[0].Metrics[0].Units).Should(Equal("ms"))
			Expect(*exp.Spec.Backends[0].Metrics[0].Type).Should(Equal(GaugeMetricType))
			Expect(*exp.Spec.Backends[0].Metrics[0].Body).Should(Equal("maybe empty"))

			Expect(exp.Spec.Backends[0].Metrics[1].Name).Should(Equal("backend-1/objective-metric"))
			Expect(*exp.Spec.Backends[0].Metrics[1].Description).Should(Equal("objective-metric description"))
			Expect(len(exp.Spec.Backends[0].Metrics[1].Params)).Should(Equal(1))
			Expect(*exp.Spec.Backends[0].Metrics[1].Units).Should(Equal("ms"))
			Expect(*exp.Spec.Backends[0].Metrics[1].Type).Should(Equal(GaugeMetricType))
			Expect(*exp.Spec.Backends[0].Metrics[1].Body).Should(Equal("maybe empty"))

			By("deleting the experiment")
			Expect(k8sClient.Delete(ctx, &s)).Should(Succeed())
		})
	})

	Context("When experiment has no metrics", func() {
		It("should create the experiment", func() {
			By("reading experiment " + "nometrics.yaml")
			s := Experiment{}
			Expect(readExperimentFromFile(path.Join("..", "..", "test", "data", "nometrics.yaml"), &s)).To(Succeed())

			By("creating the experiment")
			Expect(k8sClient.Create(ctx, &s)).Should(Succeed())

			By("reading the experiment")
			exp := Experiment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "experiment", Namespace: "default"}, &exp)).Should(Succeed())

			By("verifying the experiment")
			Expect(len(exp.Spec.Backends)).Should(Equal(0))

			By("deleting the experiment")
			Expect(k8sClient.Delete(ctx, &s)).Should(Succeed())
		})
	})

})
