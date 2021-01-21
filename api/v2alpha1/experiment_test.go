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

package v2alpha1_test

import (
	"context"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Experiment's handler field", func() {
	ctx := context.Background()

	for file, feature := range map[string]string{
		"expspec.yaml": "containing unknown fields",
	} {
		Context("when "+feature, func() {
			us := &unstructured.Unstructured{}
			us.Object = map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "exp",
					"namespace": "default",
				},
				"spec": map[string]interface{}{},
			}

			It("should read experiment successfully", func() {
				s := map[string]interface{}{}
				Expect(readExperimentFromFile(util.CompletePath("../../test/data", file), &s)).To(Succeed())
				us.Object["spec"] = s["spec"]
			})

			It("should create the experiment", func() {
				us.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v2alpha1.GroupVersion.Group,
					Version: v2alpha1.GroupVersion.Version,
					Kind:    "Experiment",
				})
				testLogger.Info("unstructured object", "us", us)
				Expect(k8sClient.Create(ctx, us)).Should(Succeed())
			})

			exp2 := &unstructured.Unstructured{}
			exp2.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   v2alpha1.GroupVersion.Group,
				Version: v2alpha1.GroupVersion.Version,
				Kind:    "Experiment",
			})
			It("should fetch the experiment with the unknown fields", func() {
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Namespace: "default",
					Name:      "exp"}, exp2)).Should(Succeed())
				testLogger.Info("fetched", "experiment", exp2)
				_, found, err := unstructured.NestedFieldCopy(exp2.Object, "spec", "strategy", "handlers", "startTasks")
				Expect(found).To(BeTrue())
				Expect(err).To(BeNil())
			})
		})
	}

})

func readExperimentFromFile(templateFile string, m *map[string]interface{}) error {
	yamlFile, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, m); err == nil {
		return err
	}

	return nil
}
