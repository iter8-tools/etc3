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
	"path"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/types"

	"github.com/iter8-tools/etc3/api/v2alpha2"
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
		"invalid3.yaml": "spec.jqExpression is not set",
		"invalid4.yaml": "spec.jqExpression is \"\"",
		"invalid5.yaml": "spec has an extra field",
		"invalid6.yaml": "spec.method is not GET/POST",
		"invalid7.yaml": "spec.authType is not Basic/Bearer/APIKey",
	} {
		Context("When "+feature, func() {
			metric := v2alpha2.Metric{}
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
		metric := v2alpha2.NewMetric("test", "default").
			WithDescription("valid metric").
			WithParams([]v2alpha2.NamedValue{{
				Name:  "foo",
				Value: "bar",
			}}).
			WithType(v2alpha2.GaugeMetricType).
			WithSampleSize("namespace/name").
			WithProvider("provider").
			WithJQExpression("expr").
			WithURLTemplate("url").
			Build()

		It("should be created", func() {
			Expect(k8sClient.Create(ctx, metric)).Should(Succeed())
			k8sClient.Delete(ctx, metric) // cleanup
		})
	})
})

var _ = Describe("Metrics with method", func() {
	ctx := context.Background()

	Context("When a metric is created with the method field", func() {
		metric := v2alpha2.NewMetric("test", "default").
			WithDescription("valid metric").
			WithParams([]v2alpha2.NamedValue{{
				Name:  "foo",
				Value: "bar",
			}}).
			WithType(v2alpha2.GaugeMetricType).
			WithMethod(v2alpha2.POSTMethodType).
			WithSampleSize("namespace/name").
			WithProvider("provider").
			WithJQExpression("expr").
			WithURLTemplate("url").
			Build()

		It("the method field is preserved", func() {
			Expect(k8sClient.Create(ctx, metric)).Should(Succeed())
			fetchedMetric := v2alpha2.NewMetric("test", "default").Build()
			k8sClient.Get(ctx, types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			}, fetchedMetric)
			Expect(*fetchedMetric.Spec.Method).Should(Equal(v2alpha2.POSTMethodType))
			k8sClient.Delete(ctx, metric) // cleanup
		})
	})

	Context("When a metric is created without the method field", func() {
		metric := v2alpha2.NewMetric("test", "default").
			WithDescription("valid metric").
			WithParams([]v2alpha2.NamedValue{{
				Name:  "foo",
				Value: "bar",
			}}).
			WithType(v2alpha2.GaugeMetricType).
			WithSampleSize("namespace/name").
			WithProvider("provider").
			WithJQExpression("expr").
			WithURLTemplate("url").
			Build()

		It("the method field is defaulted to GET", func() {
			Expect(k8sClient.Create(ctx, metric)).Should(Succeed())
			fetchedMetric := v2alpha2.NewMetric("test", "default").Build()
			k8sClient.Get(ctx, types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			}, fetchedMetric)
			Expect(*fetchedMetric.Spec.Method).Should(Equal(v2alpha2.GETMethodType))
			k8sClient.Delete(ctx, metric) // cleanup
		})
	})
})

var _ = Describe("Metrics with authtype", func() {
	ctx := context.Background()

	Context("When a metric is created with the authType field", func() {
		metric := v2alpha2.NewMetric("test", "default").
			WithDescription("valid metric").
			WithParams([]v2alpha2.NamedValue{{
				Name:  "foo",
				Value: "bar",
			}}).
			WithType(v2alpha2.GaugeMetricType).
			WithMethod(v2alpha2.POSTMethodType).
			WithAuthType(v2alpha2.BasicAuthType).
			WithSampleSize("namespace/name").
			WithProvider("provider").
			WithJQExpression("expr").
			WithURLTemplate("url").
			Build()

		It("the authtype field is preserved", func() {
			Expect(k8sClient.Create(ctx, metric)).Should(Succeed())
			fetchedMetric := v2alpha2.NewMetric("test", "default").Build()
			k8sClient.Get(ctx, types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			}, fetchedMetric)
			Expect(*fetchedMetric.Spec.AuthType).Should(Equal(v2alpha2.BasicAuthType))
			k8sClient.Delete(ctx, metric) // cleanup
		})
	})

	Context("When a metric is created without AuthType field", func() {
		metric := v2alpha2.NewMetric("test", "default").
			WithDescription("valid metric").
			WithParams([]v2alpha2.NamedValue{{
				Name:  "foo",
				Value: "bar",
			}}).
			WithType(v2alpha2.GaugeMetricType).
			WithSampleSize("namespace/name").
			WithProvider("provider").
			WithJQExpression("expr").
			WithURLTemplate("url").
			Build()

		It("the AuthType field is not defaulted", func() {
			Expect(k8sClient.Create(ctx, metric)).Should(Succeed())
			fetchedMetric := v2alpha2.NewMetric("test", "default").Build()
			k8sClient.Get(ctx, types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			}, fetchedMetric)
			Expect(fetchedMetric.Spec.AuthType).Should(BeNil())
			k8sClient.Delete(ctx, metric) // cleanup
		})
	})
})

func readMetricFromFile(templateFile string, job *v2alpha2.Metric) error {
	yamlFile, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, job); err == nil {
		return err
	}

	return nil
}
