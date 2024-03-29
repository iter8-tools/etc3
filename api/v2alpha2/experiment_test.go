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

package v2alpha2_test

import (
	"context"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/controllers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Experiment", func() {
	ctx := context.Background()

	type testcase struct {
		file      string
		feature   string
		fieldPath []string
	}

	testcases := []testcase{
		{
			file:      "expspec.yaml",
			feature:   "with strategy containing actions",
			fieldPath: []string{"spec", "strategy", "actions", "start"},
		},
		{
			file:      "expwithbuiltins.yaml",
			feature:   "with status containing builtin data",
			fieldPath: []string{"status", "analysis", "aggregatedBuiltinHists", "data", "DurationHistogram", "Avg"},
		},
	}

	for _, tc := range testcases {
		tc := tc
		Context(tc.feature, func() {
			It("should deal "+tc.feature, func() {
				By("reading experiment")
				s := v2alpha2.Experiment{}
				Expect(readExperimentFromFile(controllers.CompletePath("../../test/data", tc.file), &s)).To(Succeed())

				By("creating the experiment")
				Expect(k8sClient.Create(ctx, &s)).Should(Succeed())

				By("updating status...")
				Expect(k8sClient.Status().Update(ctx, &s)).Should(Succeed())

				By("fetching the experiment with json fields")
				exp2 := &unstructured.Unstructured{}
				exp2.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v2alpha2.GroupVersion.Group,
					Version: v2alpha2.GroupVersion.Version,
					Kind:    "Experiment",
				})
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Namespace: s.Namespace,
					Name:      s.Name}, exp2)).Should(Succeed())

				_, found, err := unstructured.NestedFieldCopy(exp2.Object, tc.fieldPath...)

				Expect(found).To(BeTrue())
				Expect(err).To(BeNil())

				By("deleting the experiment")
				Expect(k8sClient.Delete(ctx, &s)).Should(Succeed())
			})
		})
	}

})

func readExperimentFromFile(templateFile string, exp *v2alpha2.Experiment) error {
	yamlFile, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, exp); err == nil {
		return err
	}

	return nil
}
