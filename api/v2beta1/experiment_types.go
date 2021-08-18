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

// experiment_types.go - go model for experiment CRD

package v2beta1

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Experiment is the Schema for the experiments API
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="type",type="string",JSONPath=".status.testingPattern"
// +kubebuilder:printcolumn:name="stage",type="string",JSONPath=".status.stage"
// +kubebuilder:printcolumn:name="completed loops",type="string",JSONPath=".status.completedLoops"
// +kubebuilder:printcolumn:name="message",type="string",JSONPath=".status.message"
type Experiment struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   ExperimentSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status ExperimentStatus `json:"status,omitempty" yaml:"spec,omitempty"`
}

// ExperimentList contains a list of Experiment
//+kubebuilder:object:root=true
type ExperimentList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Experiment `json:"items"`
}

// ExperimentSpec defines the desired state of Experiment
type ExperimentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// VersionInfo is list of version labels
	// +kubebuilder:validation:MinItems:=1
	VersionInfo []string `json:"versionInfo" yaml:"versionInfo"`

	// Actions define the collections of tasks that are executed by handlers.
	// Specifically, start and finish actions are invoked by start and finish handlers respectively.
	// +optional
	Actions ActionMap `json:"actions,omitempty" yaml:"actions,omitempty"`

	// Criteria contains a list of Criterion for assessing the candidates
	// Note that the number of rewards that can be/must be specified depends on the testing pattern
	// +optional
	Criteria *Criteria `json:"criteria,omitempty" yaml:"criteria,omitempty"`

	// Duration describes how long the experiment will last.
	// +optional
	Duration *Duration `json:"duration,omitempty" yaml:"duration,omitempty"`

	// Backend is details of metrics backend or overrides to a backend defined in a configmap
	Backends []Backend `json:"backends,omitempty" yaml:"backends,omitempty"`

	// Metrics are metrics definitions used by Criteria; may override predefined metrics
	Metrics []Metric `json:"metrics,omitempty" yaml:"metrics,omitempty"`
}

// ActionMap type for containing a collection of actions.
type ActionMap map[string]Action

// Action is a slice of task specifications.
type Action []TaskSpec

// TaskSpec contains the specification of a task.
type TaskSpec struct {
	// Task uniquely identifies the task to be executed.
	// Examples include 'common/bash', etc.
	// +optional
	Task *string `json:"task,omitempty" yaml:"task,omitempty"`
	// Run is identifies the bash script to be run.
	// TaskSpec must include exactly one of the two fields, run or task.
	// +optional
	Run *string `json:"run,omitempty" yaml:"run,omitempty"`
	// If specifies if this task should be executed.
	// Task will be evaluated if condition specified by if evaluates to true, and not otherwise.
	// +optional
	If *string `json:"if,omitempty" yaml:"if,omitempty"`
	// With holds inputs to this task.
	// Different task require different types of inputs. Hence, this data is held as json.RawMessage to be decoded by individual task libraries.
	// +optional
	With map[string]apiextensionsv1.JSON `json:"with,omitempty" yaml:"with,omitempty"`
}

// Criteria is list of criteria to be evaluated throughout the experiment
type Criteria struct {
	// Rewards is a list of metrics that should be used to evaluate the reward for a version in the experiment.
	// +optional
	Rewards []Reward `json:"rewards,omitempty" yaml:"rewards,omitempty"`

	// Objectives is a list of conditions on metrics that must be tested on each loop of the experiment.
	// Failure of an objective might reduces the likelihood that a version will be selected as the winning version.
	// +optional
	Objectives []Objective `json:"objectives,omitempty" yaml:"objectives,omitempty"`
}

// Reward ..
type Reward struct {
	// Metric ..
	Metric string `json:"metric" yaml:"metric"`

	// PreferredDirection identifies whether higher or lower values of the reward metric are preferred
	// valid values are "higher" and "lower"
	PreferredDirection PreferredDirectionType `json:"preferredDirection" yaml:"preferredDirection"`
}

// Objective is a service level objective
type Objective struct {
	// Metric is the name of the metric resource that defines the metric to be measured.
	// If the value contains a "/", the prefix will be considered to be a namespace name.
	// If the value does not contain a "/", the metric should be defined either in the same namespace
	// or in the default domain namespace (defined as a property of iter8 when installed).
	// The experiment namespace takes precedence.
	Metric string `json:"metric" yaml:"metric"`

	// UpperLimit is the maximum acceptable value of the metric.
	// +optional
	UpperLimit *resource.Quantity `json:"upperLimit,omitempty" yaml:"upperLimit,omitempty"`

	// LowerLimit is the minimum acceptable value of the metric.
	// +optional
	LowerLimit *resource.Quantity `json:"lowerLimit,omitempty" yaml:"lowerLimit,omitempty"`
}

// Duration of an experiment
type Duration struct {
	// IntervalSeconds is the length of an interval of the experiment in seconds
	// Default is 20 (seconds)
	// +kubebuilder:validation:Minimum:=1
	// +optional
	MinIntervalBetweenLoops *int32 `json:"minIntervalBetweenLoops,omitempty" yaml:"minIntervalBetweenLoops,omitempty"`

	// MaxLoops is the maximum number of loops
	// Default is 15
	// Reserved for future use
	// +kubebuilder:validation:Minimum:=1
	// +optional
	MaxLoops *int32 `json:"maxLoops,omitempty" yaml:"maxLoops,omitempty"`
}

// ExperimentStatus defines the observed state of Experiment
type ExperimentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// List of conditions
	// +optional
	Conditions []*ExperimentCondition `json:"conditions,omitempty" yaml:"conditions,omitempty"`

	// StartTime is the time when the experiment is created. It is set by the controller
	// when the experiment is initialized.
	// +optional
	// matches
	StartTime *metav1.Time `json:"startTime,omitempty" yaml:"startTime,omitempty"`

	// LastUpdateTime is the last time a loop has been updated
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty" yaml:"lastUpdateTime,omitempty"`

	// Stage indicates where the experiment is in its process of execution
	// +optional
	Stage *ExperimentStageType `json:"stage,omitempty" yaml:"stage,omitempty"`

	// TestingPattern identifies the type of experiment being executed
	TestingPattern *TestingPatternType `json:"testingPattern,omitempty" yaml:"testingPattern,omitempty"`

	// CurrentLoops is the number of loops that have completed
	// It is undefined until the experiment starts.
	// +optional
	CompletedLoops *int32 `json:"completedLoops,omitempty" yaml:"completedLoops,omitempty"`

	// CurrentWeightDistribution is currently applied traffic weights
	// +optional
	CurrentWeightDistribution []int32 `json:"currentWeightDistribution,omitempty" yaml:"currentWeightDistribution,omitempty"`

	// Analysis returned by the last analyis
	// +optional
	Analysis *Analysis `json:"analysis,omitempty" yaml:"analysis,omitempty"`

	// Message specifies message to show in the kubectl printer
	// +optional
	Message *string `json:"message,omitempty" yaml:"message,omitempty"`
}

// ExperimentCondition describes a condition of an experiment
type ExperimentCondition struct {
	// Type of the condition
	Type ExperimentConditionType `json:"type" yaml:"type"`

	// Status of the condition
	Status corev1.ConditionStatus `json:"status" yaml:"status"`

	// LastTransitionTime is the time when this condition is last updated
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`

	// Reason for the last update
	// +optional
	Reason *string `json:"reason,omitempty" yaml:"reason,omitempty"`

	// Detailed explanation on the update
	// +optional
	Message *string `json:"message,omitempty" yaml:"message,omitempty"`
}

// Analysis is data from an analytics provider
type Analysis struct {
	// Metrics
	Metrics []map[string]QuantityList `json:"metrics,omitempty" yaml:"metrics,omitempty"`

	// Winner
	Winner *Winner `json:"winner,omitempty" yaml:"winner,omitempty"`

	// Objectives
	// if not empty, the length of the outer slice must match the length of Spec.Versions
	// if not empty, the length of an inner slice must match the number of Spec.Criteria.Objectives
	Objectives []BooleanList `json:"objectives,omitempty" yaml:"objectives,omitempty"`

	// Weights
	// if not empty, the length of the slice must match the length of Spec.Versions
	Weights []int32 `json:"weights,omitempty" yaml:"weights,omitempty"`
}

// BooleanList ..
type BooleanList []bool

// QuantityList ..
type QuantityList []resource.Quantity

// Winner ..
type Winner struct {
	// WinnerFound whether or not a winning version has been identified
	WinnerFound bool `json:"winnerFound" yaml:"winnerFound"`

	// Winner if found
	// +optional
	Winner *string `json:"winner,omitempty" yaml:"winner,omitempty"`
}

// Backend ..
type Backend struct {
	// Name is label of backend
	Name string `json:"name" yaml:"name"`

	// Text description of the backend
	// +optional
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

	// BackendDetail
	BackendDetail `json:",inline" yaml:",inline"`

	// Metrics is list of metrics available from this backend
	// +optional
	Metrics []Metric `json:"metrics,omitempty" yaml:"metrics,omitempty"`
}

// Metric ..
type Metric struct {
	// Name is label of metric
	Name string `json:"name" yaml:"name"`

	// Text description of the metric
	// +optional
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

	// Params are key/value pairs corresponding to HTTP request parameters
	// Value may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using version information.
	// +optional
	Params []NamedValue `json:"params,omitempty" yaml:"params,omitempty"`

	// Units of the metric. Used for informational purposes.
	// +optional
	Units *string `json:"units,omitempty" yaml:"units,omitempty"`

	// Type of the metric
	// +optional
	Type *MetricType `json:"type,omitempty" yaml:"type,omitempty"`

	// Body is the string used to construct the (json) body of the HTTP request
	// Body may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using version information.
	// +optional
	Body *string `json:"body,omitempty" yaml:"body,omitempty"`
}

type BackendDetail struct {
	// AuthType is the type of authentication used in the HTTP request
	// +optional
	AuthType *AuthType `json:"authType,omitempty" yaml:"authType,omitempty"`

	// Method is the HTTP method used in the HTTP request
	// +optional
	Method *MethodType `json:"method,omitempty" yaml:"method,omitempty"`

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

	// Headers are key/value pairs corresponding to HTTP request headers and their values.
	// Value may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using Secret.
	// Placeholder substitution will be attempted only when Secret != nil.
	// +optional
	Headers []NamedValue `json:"headers,omitempty" yaml:"headers,omitempty"`

	// URL is a template for the URL queried during the HTTP request.
	// Typically, URL is expected to be the actual URL without any placeholders.
	// However, as indicated by its name, URL may be templated.
	// In this case, Iter8 will attempt to substitute placeholders in the URL at query time using Secret.
	// Placeholder substitution will be attempted only when Secret != nil.
	// +optional
	URL *string `json:"url,omitempty" yaml:"url,omitempty"`

	// VersionInfo is map of name/value pairs for substitution into Body and Params
	// +optional
	VersionInfo []VersionDetail `json:"versionInfo,omitempty" yaml:"versionInfo,omitempty"`
}

type VersionDetail map[string]string

func init() {
	SchemeBuilder.Register(&Experiment{}, &ExperimentList{})
}
