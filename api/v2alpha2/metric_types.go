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

// metric_types.go - go model for metric CRD

package v2alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetricType identifies the type of the metric
// +kubebuilder:validation:Enum=counter;gauge
type MetricType string

const (
	// CounterMetricType corresponds to Prometheus counter metric type
	CounterMetricType MetricType = "counter"

	// GaugeMetricType is an enhancement of Prometheus gauge metric type
	GaugeMetricType MetricType = "gauge"
)

// MetricSpec defines the desired state of Metric
type MetricSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Params are key/value pairs used to construct a REST query to the metrics backend
	// +optional
	Params *[]NamedValue `json:"params,omitempty" yaml:"params,omitempty"`

	// Text description of the metric
	// +optional
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

	// Units used for display only
	// +optional
	Units *string `json:"units,omitempty" yaml:"units,omitempty"`

	// Type of the metric
	// +kubebuilder:default:="gauge"
	// +optional
	Type MetricType `json:"type" yaml:"type"`

	// SampleSize is a reference to a counter metric resource.
	// It needs to indicte the number of data points over which this metric is computed.
	// +kubebuilder:validation:MinLength:=1
	// +optional
	SampleSize *string `json:"sampleSize,omitempty" yaml:"sampleSize,omitempty"`

	// Provider identifies the metric backend including its authentication properties and its unmarshaller
	// +kubebuilder:validation:MinLength:=1
	Provider string `json:"provider" yaml:"provider"`

	// SecretRef the name of a kubernetes Secret containing authentication details for the metrics backend
	// +kubebuilder:validation:MinLength:=1
	// +optional
	SecretRef *string `json:"secret,omitempty" yaml:"secret,omitempty"`

	// HeaderTemplates are templates for headers that should be passed to the metrics backend.
	// Typically these are authentication headers. Any fields of the form '$name' are treated as
	// variables whose value is looked up in the secret referred to by spec.secretRef. If the
	// secret does not contain the variable as a key, the full string "$name" is used.
	// +optional
	HeaderTemplates *[]NamedValue `json:"headers,omitempty" yaml:"headers,omitempty"`

	// URLTemplate is a template of the url of metrics backend. Any fields pf the form '$name' are
	// treated as variables whose value is looked up in the secret referred to by spec.secretRef.
	// If the secret does not contain the variable as a key, the full string "$name" is used.
	URLTemplate string `json:"urlTemplate" yaml:"urlTemplate"`
}

// NamedValue name/value to be used in constructing a REST query to backend metrics server
type NamedValue struct {
	// Name of parameter
	Name string `json:"name" yaml:"name"`

	// Value of parameter
	Value string `json:"value" yaml:"value"`
}

// MetricStatus defines the observed state of Metric
type MetricStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Metric is the Schema for the metrics API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="description",type="string",JSONPath=".spec.description"
type Metric struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec MetricSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
	// metrics are fixed; there is no need for a status
	// cf. https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#resources
	// See section: Objects > Spec and Status
	// Status MetricStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MetricList contains a list of Metric
type MetricList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Metric `json:"items" yaml:"items"`
}

func init() {
	SchemeBuilder.Register(&Metric{}, &MetricList{})
}
