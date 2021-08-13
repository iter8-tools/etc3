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
// +kubebuilder:printcolumn:name="completed iterations",type="string",JSONPath=".status.completedIterations"
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

	// Versions is list of version labels
	// +kubebuilder:validation:MinItems:=1
	Versions []string `json:"versions" yaml:"versions"`

	// VersionInfo is information about versions that is typically provided by the domain start handler
	// +optional
	VersionInfo *VersionInfo `json:"versionInfo,omitempty" yaml:"versionInfo,omitempty"`

	// Strategy identifies the type of experiment and its properties
	Strategy Strategy `json:"strategy" yaml:"strategy"`

	// Criteria contains a list of Criterion for assessing the candidates
	// Note that the number of rewards that can be/must be specified depends on the testing pattern
	// +optional
	Criteria *Criteria `json:"criteria,omitempty" yaml:"criteria,omitempty"`

	// Duration describes how long the experiment will last.
	// +optional
	Duration *Duration `json:"duration,omitempty" yaml:"duration,omitempty"`
}

// MetricInfo is name/value pair; entry for list of metrics
type MetricInfo struct {
	// Name is identifier for metric.  Can be of the form "name" or "namespace/name"
	Name string `json:"name" yaml:"name"`

	// MetricObj is the referenced metric
	// +kubebuilder:validation:EmbeddedResource
	MetricObj Metric `json:"metricObj" yaml:"metricObj"`
}

// VersionInfo is information about versions that is typically provided by the domain start handler.
type VersionInfo struct {
	// Baseline is baseline version
	Baseline VersionDetail `json:"baseline" yaml:"baseline"`

	// Candidates is list candidate versions
	// +optional
	Candidates []VersionDetail `json:"candidates,omitempty" yaml:"candidates,omitempty"`
}

// VersionDetail is detail about a single version
type VersionDetail struct {

	// Name is a name for the version
	Name string `json:"name" yaml:"name"`

	// Variables is a list of variables that can be used by handlers and in metrics queries
	// +optional
	Variables []NamedValue `json:"variables,omitempty" yaml:"variables,omitempty"`

	// WeightObjRef is a reference to another kubernetes object
	// +optional
	WeightObjRef *corev1.ObjectReference `json:"weightObjRef,omitempty" yaml:"weightObjRef,omitempty"`
}

// Strategy identifies the type of experiment and its properties
// The behavior of the experiment can be modified by setting advanced properties.
type Strategy struct {
	// DeploymentPattern is the deployment pattern of an experiment.
	// It takes effect when the testing pattern is one of Canary, A/B or A/B/n.
	// It defaults to Progressive.
	// +optional
	DeploymentPattern *DeploymentPatternType `json:"deploymentPattern,omitempty" yaml:"deploymentPattern,omitempty"`

	// Actions define the collections of tasks that are executed by handlers.
	// Specifically, start and finish actions are invoked by start and finish handlers respectively.
	// +optional
	Actions ActionMap `json:"actions,omitempty" yaml:"actions,omitempty"`

	// Weights modify the behavior of the traffic split algorithm.
	// Defaults depend on the experiment type.
	// +optional
	Weights *Weights `json:"weights,omitempty" yaml:"weights,omitempty"`
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

// Weights modify the behavior of the traffic split algorithm.
type Weights struct {
	// MaxCandidateWeight is the maximum percent of traffic that should be sent to the
	// candidate versions during an experiment
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:validation:Maximum:=100
	// +optional
	MaxCandidateWeight *int32 `json:"maxCandidateWeight,omitempty" yaml:"maxCandidateWeight,omitempty"`

	// MaxCandidateWeightIncrement the maximum permissible increase in traffic to a candidate in one iteration
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:validation:Maximum:=100
	// +optional
	MaxCandidateWeightIncrement *int32 `json:"maxCandidateWeightIncrement,omitempty" yaml:"maxCandidateWeightIncrement,omitempty"`
}

// Criteria is list of criteria to be evaluated throughout the experiment
type Criteria struct {

	// RequestCount identifies metric to be used to count how many requests a version has seen
	// Typically set by the controller (based on setup configuration) but can be overridden by the user
	// + optional
	RequestCount *string `json:"requestCount,omitempty" yaml:"requestCount,omitempty"`

	// Rewards is a list of metrics that should be used to evaluate the reward for a version in the experiment.
	// +optional
	Rewards []Reward `json:"rewards,omitempty" yaml:"rewards,omitempty"`

	// Indicators is a list of metrics to be measured and reported on each iteration of the experiment.
	// +optional
	Indicators []string `json:"indicators,omitempty" yaml:"indicators,omitempty"`

	// Objectives is a list of conditions on metrics that must be tested on each iteration of the experiment.
	// Failure of an objective might reduces the likelihood that a version will be selected as the winning version.
	// Failure of an objective might also trigger an experiment rollback.
	// +optional
	Objectives []Objective `json:"objectives,omitempty" yaml:"objectives,omitempty"`

	// Strength identifies the required degree of support the analytics must provide before it will
	// assert success for an objective.
	// +optional
	Strength apiextensionsv1.JSON `json:"strength,omitempty" yaml:"strength,omitempty"`
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

	// UpperLimit is the minimum acceptable value of the metric.
	// +optional
	LowerLimit *resource.Quantity `json:"lowerLimit,omitempty" yaml:"lowerLimit,omitempty"`

	// RollbackOnFailure indicates that if the criterion is not met, the experiment should be ended
	// default is false
	// +optional
	RollbackOnFailure *bool `json:"rollback_on_failure,omitempty" yaml:"rollback_on_failure,omitempty"`
}

// Duration of an experiment
type Duration struct {
	// IntervalSeconds is the length of an interval of the experiment in seconds
	// Default is 20 (seconds)
	// +kubebuilder:validation:Minimum:=1
	// +optional
	IntervalSeconds *int32 `json:"intervalSeconds,omitempty" yaml:"intervalSeconds,omitempty"`

	// IterationsPerLoop is the maximum number of iterations
	// Default is 15
	// +kubebuilder:validation:Minimum:=1
	// +optional
	IterationsPerLoop *int32 `json:"iterationsPerLoop,omitempty" yaml:"iterationsPerLoop,omitempty"`

	// MaxLoops is the maximum number of loops
	// Default is 1
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

	// LastUpdateTime is the last time iteration has been updated
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty" yaml:"lastUpdateTime,omitempty"`

	// Stage indicates where the experiment is in its process of execution
	// +optional
	Stage *ExperimentStageType `json:"stage,omitempty" yaml:"stage,omitempty"`

	// TestingPattern identifies the type of experiment being executed
	TestingPattern *TestingPatternType `json:"testingPattern,omitempty" yaml:"testingPattern,omitempty"`

	// CurrentIteration is the current iteration number.
	// It is undefined until the experiment starts.
	// +optional
	CompletedIterations *int32 `json:"completedIterations,omitempty" yaml:"completedIterations,omitempty"`

	// CurrentWeightDistribution is currently applied traffic weights
	// +optional
	CurrentWeightDistribution []WeightData `json:"currentWeightDistribution,omitempty" yaml:"currentWeightDistribution,omitempty"`

	// Analysis returned by the last analyis
	// +optional
	Analysis *Analysis `json:"analysis,omitempty" yaml:"analysis,omitempty"`

	// VersionRecommendedForPromotion is the version recommended as the baseline after the experiment completes.
	// Will be set to the winner (status.analysis[].data.winner)
	// or to the current baseline in the case of a rollback.
	// +optional
	VersionRecommendedForPromotion *string `json:"versionRecommendedForPromotion,omitempty" yaml:"versionRecommendedForPromotion,omitempty"`

	// Message specifies message to show in the kubectl printer
	// +optional
	Message *string `json:"message,omitempty" yaml:"message,omitempty"`

	// Metrics is a list of all the metrics used in the experiment
	// It is inserted by the controller from the references in spec.criteria
	// Key is the name as referenced in spec.criteria
	// +optional
	Metrics []MetricInfo `json:"metrics,omitempty" yaml:"metrics,omitempty"`
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
	// AggregatedBuiltinHistograms -- aggregated builtin metrics will be derived from this data structure
	AggregatedBuiltinHists *AggregatedBuiltinHists `json:"aggregatedBuiltinHists,omitempty" yaml:"aggregatedBuiltinHists,omitempty"`

	// AggregatedMetrics
	AggregatedMetrics *AggregatedMetricsAnalysis `json:"aggregatedMetrics,omitempty" yaml:"aggregatedMetrics,omitempty"`

	// WinnerAssessment
	WinnerAssessment *WinnerAssessmentAnalysis `json:"winnerAssessment,omitempty" yaml:"winnerAssessment,omitempty"`

	// VersionAssessments
	VersionAssessments *VersionAssessmentAnalysis `json:"versionAssessments,omitempty" yaml:"versionAssessments,omitempty"`

	// Weights
	Weights *WeightsAnalysis `json:"weights,omitempty" yaml:"weights,omitempty"`
}

// AnalysisMetaData ..
type AnalysisMetaData struct {
	// Provenance is source of data
	Provenance string `json:"provenance" yaml:"provenance"`

	// Timestamp is the timestamp when the controller got its data from an analytics engine
	Timestamp metav1.Time `json:"timestamp" yaml:"timestamp"`

	// Message optional messsage for user
	// +optional
	Message *string `json:"message,omitempty" yaml:"message,omitempty"`
}

// AggregatedBuiltinHists ..
type AggregatedBuiltinHists struct {
	AnalysisMetaData `json:",inline" yaml:",inline"`
	// This field needs leeway to evolve. At the moment, it would look like DurationHists from fortio output, but further experimentation is needed. Hence, `apiextensionsv1.JSON` is a safe starting point.
	Data apiextensionsv1.JSON `json:"data" yaml:"data"`
}

// WinnerAssessmentAnalysis ..
type WinnerAssessmentAnalysis struct {
	AnalysisMetaData `json:",inline" yaml:",inline"`

	// Data
	Data WinnerAssessmentData `json:"data" yaml:"data"`
}

// VersionAssessmentAnalysis ..
type VersionAssessmentAnalysis struct {
	AnalysisMetaData `json:",inline" yaml:",inline"`

	// Data is a map from version name to an array of indicators as to whether or not the objectives are satisfied
	// The order of the array entries is the same as the order of objectives in spec.criteria.objectives
	// There must be an entry for each objective
	Data map[string]BooleanList `json:"data" yaml:"data"`
}

// BooleanList ..
type BooleanList []bool

// WeightsAnalysis ..
type WeightsAnalysis struct {
	AnalysisMetaData `json:",inline" yaml:",inline"`

	// Data
	Data []WeightData `json:"data" yaml:"data"`
}

// AggregatedMetricsAnalysis ..
type AggregatedMetricsAnalysis struct {
	AnalysisMetaData `json:",inline" yaml:",inline"`

	// Data is a map from metric name to most recent metric data
	Data map[string]AggregatedMetricsData `json:"data" yaml:"data"`
}

// WinnerAssessmentData ..
type WinnerAssessmentData struct {
	// WinnerFound whether or not a winning version has been identified
	WinnerFound bool `json:"winnerFound" yaml:"winnerFound"`

	// Winner if found
	// +optional
	Winner *string `json:"winner,omitempty" yaml:"winner,omitempty"`
}

// AggregatedMetricsData ..
type AggregatedMetricsData struct {
	// Max value observed for this metric across all versions
	// +optional
	Max *resource.Quantity `json:"max,omitempty" yaml:"max,omitempty"`

	// Min value observed for this metric across all versions
	// +optional
	Min *resource.Quantity `json:"min,omitempty" yaml:"min,omitempty"`

	// Data is a map from version name to the most recent aggregated metrics data for that version
	Data map[string]AggregatedMetricsVersionData `json:"data" yaml:"data"`
}

// WeightData is the weight for a version
type WeightData struct {
	// Name the name of a version
	Name string `json:"name" yaml:"name"`

	// Value is the weight assigned to name
	Value int32 `json:"value" yaml:"value"`
}

// AggregatedMetricsVersionData ..
type AggregatedMetricsVersionData struct {
	// Max value observed for this metric for this version
	// +optional
	Max *resource.Quantity `json:"max,omitempty" yaml:"max,omitempty"`

	// Min value observed for this metric for this version
	// +optional
	Min *resource.Quantity `json:"min,omitempty" yaml:"min,omitempty"`

	// Value of the metric observed for this version
	// +optional
	Value *resource.Quantity `json:"value,omitempty" yaml:"value,omitempty"`

	// SampleSize is the size of the sample used for computing this metric.
	// This field is applicable only to Gauge metrics
	// +kubebuilder:validation:Minimum:=0
	SampleSize *int32 `json:"sampleSize,omitempty" yaml:"sampleSize,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Experiment{}, &ExperimentList{})
}
