/*
Copyright 2021.

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

// metric_types.go - go type definitions for Iter8 metrics API.
// Iter8 uses HTTP requests to query metric databases during experiments. An Iter8 metric contains the data needed by Iter8 to construct the HTTP request and extract the metric value from the (JSON) response provided by the metric database.
// Iter8 metrics are intended to enable metric queries to *any* REST API.
// Metric type definitions in this file strictly adhere to K8s API conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
// In particular, enumeration fields follow https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#constants;
// optional fields follow https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#optional-vs-required

package v2alpha3

import (
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetricType identifies the type of the metric.
// +kubebuilder:validation:Enum=Counter;Gauge
type MetricType string

const (
	// CounterMetricType corresponds to Prometheus Counter metric type
	CounterMetricType MetricType = "Counter"

	// GaugeMetricType is an enhancement of Prometheus Gauge metric type
	GaugeMetricType MetricType = "Gauge"
)

// AuthType identifies the type of authentication used in the HTTP request
// +kubebuilder:validation:Enum=Basic;Bearer;APIKey
type AuthType string

const (
	// BasicAuthType corresponds to authentication with basic auth
	BasicAuthType AuthType = "Basic"

	// BearerAuthType corresponds to authentication with bearer token
	BearerAuthType AuthType = "Bearer"

	// APIKeyAuthType corresponds to authentication with API keys
	APIKeyAuthType AuthType = "APIKey"
)

// MethodType identifies the HTTP request method (aka verb) used in the HTTP request
// +kubebuilder:validation:Enum=GET;POST
type MethodType string

const (
	// GETMethodType corresponds to HTTP GET method
	GETMethodType MethodType = "GET"

	// POSTMethodType corresponds to HTTP POST method
	POSTMethodType MethodType = "POST"
)

// NamedLevel contains the name of a version and the level of the version to be used in mock metric generation.
// The semantics of level are the following:
// If the metric is a counter, if level is x, and time elapsed since the start of the experiment is y, then x*y is the metric value.
// Note: this will keep increasing over time as counters do.
// If the metric is gauge, if level is x, the metric value is a random value with mean x.
// Note: due to randomness, this stay around x but can go up or down as a gauges do.
type NamedLevel struct {
	// Name of the version
	Name string `json:"name" yaml:"name"`

	// Level of the version
	Level resource.Quantity `json:"level" yaml:"level"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MetricSpec defines the desired state of Metric
type MetricSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Params are key/value pairs corresponding to HTTP request parameters
	// Value may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using version information.
	// +optional
	Params []NamedValue `json:"params,omitempty" yaml:"params,omitempty"`

	// Text description of the metric
	// +optional
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

	// Units of the metric. Used for informational purposes.
	// +optional
	Units *string `json:"units,omitempty" yaml:"units,omitempty"`

	// Type of the metric
	// +kubebuilder:default:="Gauge"
	// +optional
	Type *MetricType `json:"type,omitempty" yaml:"type,omitempty"`

	// SampleSize is a reference to a counter metric resource.
	// The value of the SampleSize metric denotes the number of data points over which this metric is computed.
	// This field is relevant only when Type == Gauge
	// +optional
	SampleSize *string `json:"sampleSize,omitempty" yaml:"sampleSize,omitempty"`

	// AuthType is the type of authentication used in the HTTP request
	// +optional
	AuthType *AuthType `json:"authType,omitempty" yaml:"authType,omitempty"`

	// Method is the HTTP method used in the HTTP request
	// +kubebuilder:default:="GET"
	// +optional
	Method *MethodType `json:"method,omitempty" yaml:"method,omitempty"`

	// Body is the string used to construct the (json) body of the HTTP request
	// Body may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using version information.
	// +optional
	Body *string `json:"body,omitempty" yaml:"body,omitempty"`

	// Provider identifies the type of metric database. Used for informational purposes.
	// +optional
	Provider *string `json:"provider,omitempty" yaml:"provider,omitempty"`

	// JQExpression defines the jq expression used by Iter8 to extract the metric value from the (JSON) response returned by the HTTP URL queried by Iter8.
	// An empty string is a valid jq expression.
	// +optional
	JQExpression *string `json:"jqExpression,omitempty" yaml:"jqExpression,omitempty"`

	// Secret is a reference to the Kubernetes secret.
	// Secret contains data used for HTTP authentication.
	// Secret may also contain data used for placeholder substitution in HeaderTemplates and URLTemplate.
	// +optional
	Secret *string `json:"secret,omitempty" yaml:"secret,omitempty"`

	// HeaderTemplates are key/value pairs corresponding to HTTP request headers and their values.
	// Value may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using Secret.
	// Placeholder substitution will be attempted only when Secret != nil.
	// +optional
	HeaderTemplates []NamedValue `json:"headerTemplates,omitempty" yaml:"headerTemplates,omitempty"`

	// URLTemplate is a template for the URL queried during the HTTP request.
	// Typically, URLTemplate is expected to be the actual URL without any placeholders.
	// However, as indicated by its name, URLTemplate may be templated.
	// In this case, Iter8 will attempt to substitute placeholders in the URLTemplate at query time using Secret.
	// Placeholder substitution will be attempted only when Secret != nil.
	// +optional
	URLTemplate *string `json:"urlTemplate,omitempty" yaml:"urlTemplate,omitempty"`

	// Mock enables mocking of metric values, which is useful in tests and tutorial/documentation.
	// Iter8 metrics can be either counter (which keep increasing over time) or gauge (which can increase or decrease over time).
	// Mock enables mocking of both.
	// +optional
	Mock []NamedLevel `json:"mock,omitempty" yaml:"mock,omitempty"`
}

// Metric is the Schema for the metrics API
//+kubebuilder:object:root=true
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

// MetricList contains a list of Metric
//+kubebuilder:object:root=true
type MetricList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Metric `json:"items" yaml:"items"`
}

func init() {
	SchemeBuilder.Register(&Metric{}, &MetricList{})
}
