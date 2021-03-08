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
	"encoding/json"
	"io/ioutil"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Experiment with actions", func() {
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
	}

	for _, tc := range testcases {
		Context(tc.feature, func() {
			It("should deal with actions with tasks with inputs", func() {
				By("reading experiment")
				s := v2alpha2.Experiment{}
				Expect(readExperimentFromFile(util.CompletePath("../../test/data", tc.file), &s)).To(Succeed())

				testMap := make(map[string]interface{})
				withBytes, err := json.Marshal(s.Spec.Strategy.Actions["finish"][0].With)
				Expect(err).ToNot(HaveOccurred())
				err = json.Unmarshal(withBytes, &testMap)
				Expect(err).ToNot(HaveOccurred())
				Expect(testMap["args"]).To(Equal([]interface{}{"apply", "-k", "https://github.com/my-org/my-repo/path/to/overlays/{{ Status.RecommendedBaseline }}"}))
				Expect(testMap["cmd"]).To(Equal("kubectl"))

				By("creating the experiment")
				Expect(k8sClient.Create(ctx, &s)).Should(Succeed())

				By("fetching the experiment with unknown fields")
				exp2 := &unstructured.Unstructured{}
				exp2.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v2alpha2.GroupVersion.Group,
					Version: v2alpha2.GroupVersion.Version,
					Kind:    "Experiment",
				})
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Namespace: "default",
					Name:      "exp"}, exp2)).Should(Succeed())
				log.Log.Info("fetched", "experiment", exp2)
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
