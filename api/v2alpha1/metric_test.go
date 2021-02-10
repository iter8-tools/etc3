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
	"path"

	"github.com/ghodss/yaml"

	"github.com/iter8-tools/etc3/api/v2alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	testMetricsDir string = "../test/data"
)

var _ = Describe("Metrics Not Created When Invalid", func() {
	ctx := context.Background()

	for file, feature := range map[string]string{
		"invalid1.yaml": "spec.params is {}",
		"invalid2.yaml": "spec.type is invalid",
		"invalid3.yaml": "spec.provider is not set",
		"invalid4.yaml": "spec.provider is \"\"",
		"invalid5.yaml": "spec.sampleSize.name is \"\"",
		"invalid6.yaml": "spec.sampleSize.name is not set",
		"invalid7.yaml": "spec has an extra field",
	} {
		Context("When "+feature, func() {
			metric := v2alpha1.Metric{}
			readMetricFromFile(path.Join(testMetricsDir, file), &metric)

			It("should fail to create", func() {
				Expect(k8sClient.Create(ctx, &metric)).ShouldNot(Succeed())
			})
		})
	}

})

var _ = Describe("Metrics Are Created When Valid", func() {
	ctx := context.Background()

	Context("When metric is valid", func() {
		metric := v2alpha1.NewMetric("test", "default").
			WithDescription("valid metric").
			WithParams(map[string]string{"foo": "bar"}).
			WithType(v2alpha1.GaugeMetricType).
			WithSampleSize("namespace", "name").
			WithProvider("provider").
			Build()

		It("should be created", func() {
			Expect(k8sClient.Create(ctx, metric)).Should(Succeed())
		})
	})
})

func readMetricFromFile(templateFile string, job *v2alpha1.Metric) error {
	yamlFile, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, job); err == nil {
		return err
	}

	return nil
}
