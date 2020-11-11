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

package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Experiment is the Schema for the experiments API
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Experiment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExperimentSpec   `json:"spec,omitempty"`
	Status ExperimentStatus `json:"status,omitempty"`
}

// ExperimentList contains a list of Experiment
// +kubebuilder:object:root=true
type ExperimentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Experiment `json:"items"`
}

// ExperimentSpec defines the desired state of Experiment
type ExperimentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Target is used to enable concurrent experimentation
	// Two experiments cannot be running concurrently for the same target.
	// +kubebuilder:validation:MinLength:=1
	Target *string `json:"target"`

	// VersionInfo is information about versions that is typically provided by the domain start handler
	// +optional
	VersionInfo *VersionInfo `json:"versionInfo,omitempty"`

	// Strategy identifies the type of experiment and its properties
	Strategy Strategy `json:"strategy"`

	// Criteria contains a list of Criterion for assessing the candidates
	// Note that at most one reward metric is allowed
	// If more than one reward criterion is included, the first will be used while others would be omitted
	// +optional
	Criteria *Criteria `json:"criteria,omitempty"`

	// Duration describes how long the experiment will last.
	// +optional
	Duration *Duration `json:"duration,omitempty"`

	// Metrics is a map of all the metrics used in the experiment
	// It is inserted by the controller from the references in spec.criteria
	// +optional
	Metrics *map[string]Metric `json:"metrics,omitempty"`
}

// VersionInfo is information about versions that is typically provided by the domain start handler.
type VersionInfo struct {
	// Baseline is baseline version
	Baseline DomainVersion `json:"baseline"`

	// Candidates is list candidate versions
	// +optional
	Candidates []DomainVersion `json:"candidates,omitempty"`
}

// DomainVersion ..
type DomainVersion struct {

	// Name is a name for the version
	Name string `json:"name"`

	// Tags is map of tags that can be used for metrics queries
	// +optional
	Tags *map[string]string `json:"tags,omitempty"`

	// WeightObjRef is a reference to another kubernetes object
	// +optional
	WeightObjRef *corev1.ObjectReference `json:"weightObjRef,omitempty"`
}

// Strategy identifies the type of experiment and its properties
// The behavior of the experiment can be modified by setting advanced properties.
type Strategy struct {
	// ExperimentType is the type of the experiment, one of several predefined types
	Type ExperimentTypeType `json:"type"`

	// Handlers define domain specific behavior and are called at well defined points in the lifecycle of an experiment.
	// Specifically at the start (start handler), at the end (finish handler).
	// A special handler can be specified to handle error cases.
	// +optional
	Handlers *Handlers `json:"handlers,omitempty"`

	// Weights modify the behavior of the traffic split algorithm.
	// Defaults depend on the experiment type.
	// +optional
	Weights *Weights `json:"weights,omitempty"`
}

// Handlers define domain specific behavior and are called at well defined points in the lifecycle of an experiment.
// Specifically at the start (start handler), at the end (finish handler).
// A special handler can be specified to handle error cases.
type Handlers struct {
	// Start handler implmenents any domain specific set up for an experiment.
	// It should ensure that any needed resources are available and in an appropriate state.
	// It must update the spec.versionInfo field of the experiment resource.
	// +optional
	Start *string `json:"start,omitempty"`

	// Finish handler implements any domain specific actions that should take place at the end of an experiment.
	// For now, this includes any promotion logic that is needed for a winning version.
	// In the future, this function might be migrated into the controller itself.
	// +optional
	Finish *string `json:"finish,omitempty"`

	// Rollback handler should implement any domain specific actions that should take place when an objective is violated.
	// For now, this includes any rollback logic thast is needed.
	// In the future, this function might be migrated into the controller itself.
	// +optional
	Rollback *string `json:"rollback"`
}

// Weights modify the behavior of the traffic split algorithm.
type Weights struct {
	// MaxCandidateWeight is the maximum percent of traffic that should be sent to the
	// candidate versions during an experiment
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=100
	// +optional
	MaxCandidateWeight *int32 `json:"maxCandidateWeight,omitempty"`

	// MaxCandidateWeightIncrement the maximum permissible increase in traffic to a candidate in one iteration
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:validation:Maximum:=100
	// +optional
	MaxCandidateWeightIncrement *int32 `json:"maxCandidateWeightIncrement,omitempty"`

	// Algorithm is the traffic split algorithm
	// Default will be None for performance experiments,
	// "fixed_split" for bluegreen experiments, and
	// "progressive" otherwise
	// +optional
	Algorithm *AlgorithmType `json:"algorithm,omitempty"`

	// Split used only by the fixed_split algorithm.
	// For bluegreen experiments, it will default to [0, 100]
	// Otherwise, it will default to a uniform split among baseline and candidates.
	// Will be ignored by all other algorithms (warning if possible!)
	// + optional
	Split []int32 `json:"split,omitempty"`
}

// Criteria is list of criteria to be evaluated throughout the experiment
type Criteria struct {

	// RequestCount identifies metric to be used to count how many requests a version has seen
	// Typically set by the controller (based on setup configuration) but can be overridden by the user
	// + optional
	RequestCount *string `json:"requestCount,omitempty"`

	// Reward is the metric that should be used to evaluate the reward for a version in the experiment.
	// +optional
	Reward *Reward `json:"reward,omitempty"`

	// Indicators is a list of metrics to be measured and reported on each iteration of the experiment.
	// +optional
	Indicators []string `json:"indicators,omitempty"`

	// Objectives is a list of conditions on metrics that must be tested on each iteration of the experiment.
	// Failure of an objective might reduces the likelihood that a version will be selected as the winning version.
	// Failure of an objective might also trigger an experiment rollback.
	// +optional
	Objectives []Objective `json:"objectives,omitempty"`
}

// Reward ..
type Reward struct {
	// Metric ..
	Metric string `json:"metric"`

	// PreferredDirection identifies whether higher or lower values of the reward metric are preferred
	// valid values are "higher" and "lower"
	PreferredDirection PreferredDirectionType `json:"preferredDirection"`
}

// Objective is a service level objective
type Objective struct {
	// Metric is the name of the metric resource that defines the metric to be measured.
	// If the value contains a "/", the prefix will be considered to be a namespace name.
	// If the value does not contain a "/", the metric should be defined either in the same namespace
	// or in the default domain namespace (defined as a property of iter8 when installed).
	// The experiment namespace takes precedence.
	Metric string `json:"metric"`

	// UpperLimit is the maximum acceptable value of the metric.
	// +optional
	UpperLimit *resource.Quantity `json:"upperLimit,omitempty"`

	// UpperLimit is the minimum acceptable value of the metric.
	// +optional
	LowerLimit *resource.Quantity `json:"lowerLimit,omitempty"`

	// RollbackOnFailure indicates that if the criterion is not met, the experiment should be ended
	// default is false
	// +optional
	RollbackOnFailure *bool `json:"rollback_on_failure,omitempty"`
}

// Duration of an experiment
type Duration struct {
	// Interval is the length of an interval in the experiment
	// Default is 20s
	// +optional
	Interval *string `json:"interval,omitempty"`

	// MaxIterations is the maximum number of iterations
	// Default is 15
	// +optional
	MaxIterations *int32 `json:"maxIterations,omitempty"`
}

// ExperimentStatus defines the observed state of Experiment
type ExperimentStatus struct {
	// List of conditions
	// +optional
	Conditions []*ExperimentCondition `json:"conditions,omitempty"`

	// InitTimestamp is the timestamp when the experiment is initialized
	// +optional
	// matches example
	InitTimestamp *metav1.Time `json:"initTimestamp,omitempty"`

	// StartTimestamp is the timestamp when the experiment starts
	// +optional
	// matches
	StartTimestamp *metav1.Time `json:"startTimestamp,omitempty"`

	// EndTimestamp is the timestamp when experiment completes
	// +optional
	EndTimestamp *metav1.Time `json:"endTimestamp,omitempty"`

	// LastUpdateTime is the last time iteration has been updated
	// +optional
	LastUpdateTimestamp *metav1.Time `json:"lastUpdateTimestamp,omitempty"`

	// CurrentIteration is the current iteration number.
	// It is undefined until the experiment starts.
	// +optional
	CurrentIteration *int32 `json:"currentIteration,omitempty"`

	// Phase marks the phase the experiment is at
	Phase PhaseType `json:"phase"`

	// CurrentWeights is currently applied traffic weights
	// +optional
	CurrentWeights []WeightData `json:"currentWeights,omitempty"`

	// Analysis returned by the last analyis
	// +optional
	Analysis *Analysis `json:"analysis,omitempty"`

	// RecommendedBaseline is the version recommended as the baseline after the experiment completes.
	// Will be set to the winner (status.analysis[].data.winner)
	// or to the current baseline in the case of a rollback.
	// +optional
	RecommendedBaseline *string `json:"recommendedBaseline,omitempty"`

	// Message specifies message to show in the kubectl printer
	// +optional
	Message *string `json:"message,omitempty"`
}

// ExperimentCondition describes a condition of an experiment
type ExperimentCondition struct {
	// Type of the condition
	Type ExperimentConditionType `json:"type"`

	// Status of the condition
	Status corev1.ConditionStatus `json:"status"`

	// LastTransitionTime is the time when this condition is last updated
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason for the last update
	// +optional
	Reason *string `json:"reason,omitempty"`

	// Detailed explanation on the update
	// +optional
	Message *string `json:"message,omitempty"`
}

// Analysis is data from an analytics provider
type Analysis struct {
	// AggregatedMetrics
	AggregatedMetrics *AggregatedMetricsAnalysis `json:"aggregatedMetrics,omitempty"`

	// WinnerAssessment
	WinnerAssessment *WinnerAssessmentAnalysis `json:"winnerAssessment,omitempty"`

	// VersionAssessments
	VersionAssessments *VersionAssessmentAnalysis `json:"versionAssessments,omitempty"`

	// Weights
	Weights *WeightsAnalysis `json:"weights,omitempty"`
}

// AnalysisMetaData ..
type AnalysisMetaData struct {
	// Provenance is source of data
	Provenance string `json:"provenance"`

	// Timestamp is the timestamp when the controller got its data from an analytics engine
	Timestamp metav1.Time `json:"timestamp"`

	// Message optional messsage for user
	// +optional
	Message *string `json:"message,omitempty"`
}

// WinnerAssessmentAnalysis ..
type WinnerAssessmentAnalysis struct {
	AnalysisMetaData `json:",inline"`

	// Data
	Data WinnerAssessmentData `json:"data"`
}

// VersionAssessmentAnalysis ..
type VersionAssessmentAnalysis struct {
	AnalysisMetaData `json:",inline"`

	// Data
	Data []VersionAssessmentData `json:"data"`
}

// WeightsAnalysis ..
type WeightsAnalysis struct {
	AnalysisMetaData `json:",inline"`

	// Data
	Data []WeightData `json:"data"`
}

// AggregatedMetricsAnalysis ..
type AggregatedMetricsAnalysis struct {
	AnalysisMetaData `json:",inline"`

	// Data
	Data []AggregatedMetricsData `json:"data"`
}

// WinnerAssessmentData ..
type WinnerAssessmentData struct {
	// WinnerFound whether or not a winning version has been identified
	WinnerFound bool `json:"winnerFound"`

	// Winner if found
	// +optional
	Winner *string `json:"winner,omitempty"`
}

// VersionAssessmentData indicates if the objectives are satisfied for a given version
type VersionAssessmentData struct {
	// Name of version
	Name string `json:"name"`

	// SatisfiesObjectives whether the objectives (in spec.criteria.objectives) are satisfied
	// There should have one entry for each objective in spec.criteria.objectives
	// The order is the same as expressed in spec.critieria.objectives
	SatisfiesObjectives []bool `json:"satisfiesObjectives"`
}

// AggregatedMetricsData ..
type AggregatedMetricsData struct {
	// Name of metric
	Name string `json:"name"`

	// Max value observed for this metric across all versions
	// +optional
	Max *resource.Quantity `json:"max,omitempty"`

	// Min value observed for this metric across all versions
	// +optional
	Min *resource.Quantity `json:"min,omitempty"`

	// Versions is aggregated metrics data for each version; there should be an entry for each
	// version (baseline and all candidates)
	Versions []AggregatedMetricsVersionData `json:"versions"`
}

// WeightData is the weight for a version
type WeightData struct {
	// Name the name of a version
	Name string `json:"name"`

	// Value is the weight assigned to name
	Value int32 `json:"value"`
}

// AggregatedMetricsVersionData ..
type AggregatedMetricsVersionData struct {
	// Name of version
	Name string `json:"name"`

	// Max value observed for this metric for this version
	// +optional
	Max *resource.Quantity `json:"max,omitempty"`

	// Min value observed for this metric for this version
	// +optional
	Min *resource.Quantity `json:"min,omitempty"`

	// Value of the metric observed for this version
	// +optional
	Value *resource.Quantity `json:"value,omitempty"`

	// SampleSize is the number of requests observed for this version
	// +kubebuilder:validation:Minimum:=0
	SampleSize *int32 `json:"sampleSize,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Experiment{}, &ExperimentList{})
}
