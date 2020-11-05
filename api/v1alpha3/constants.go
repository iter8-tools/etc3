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

// ExperimentTypeType identifies the type of experiment type
// +kubebuilder:validation:Enum=canary;A/B;A/B/N;performance;bluegreen
type ExperimentTypeType string

const (
	// Canary indicates an experiment is a canary experiment
	Canary ExperimentTypeType = "canary"

	// AB indicates an experiment is a A/B experiment
	AB ExperimentTypeType = "A/B"

	// ABN indicates an experiment is a A/B/n experiment
	ABN ExperimentTypeType = "A/B/N"

	// Performance indicates an experiment is a performance experiment
	Performance ExperimentTypeType = "performance"

	// BlueGreen indicates an experiment is a blue-green experiment
	BlueGreen ExperimentTypeType = "bluegreen"
)

// PreferredDirectionType defines the valid values for reward.PreferredDirection
// +kubebuilder:validation:Enum=higher;lower
type PreferredDirectionType string

const (
	// Higher indicates that a higher value is "better"
	Higher PreferredDirectionType = "higher"

	// Lower indicates that a lower value is "better"
	Lower PreferredDirectionType = "lower"
)

// ExperimentConditionType limits conditions can be set by controller
type ExperimentConditionType string

const (
	// ExperimentCreated
	// MetricsRead
	// StartHandlerInvoked (in-progress, finished, error)
	// FinishHandlerInvoked (in-progress, finished, error)
	// ExperimentCompleted
	// AnalyticsActive

	// ExperimentConditionTargetsProvided has status True when the Experiment detects all elements specified in targetService
	ExperimentConditionTargetsProvided ExperimentConditionType = "TargetsProvided"

	// ExperimentConditionAnalyticsServiceNormal has status True when the analytics service is operating normally
	ExperimentConditionAnalyticsServiceNormal ExperimentConditionType = "AnalyticsServiceNormal"

	// ExperimentConditionMetricsSynced has status True when metrics are successfully synced with config map
	ExperimentConditionMetricsSynced ExperimentConditionType = "MetricsSynced"

	// ExperimentConditionExperimentCompleted has status True when the experiment is completed
	ExperimentConditionExperimentCompleted ExperimentConditionType = "ExperimentCompleted"

	// ExperimentConditionRoutingRulesReady has status True when routing rules are ready
	ExperimentConditionRoutingRulesReady ExperimentConditionType = "RoutingRulesReady"
)

// PhaseType has options for phases that an experiment can be at
type PhaseType string

const (
	// PhasePause indicates experiment is paused
	PhasePause PhaseType = "Pause"

	// PhaseProgressing indicates experiment is progressing
	PhaseProgressing PhaseType = "Progressing"

	// PhaseCompleted indicates experiment has competed (successfully or not)
	PhaseCompleted PhaseType = "Completed"
)

// A set of reason setting the experiment condition status
const (
	ReasonTargetsFound            = "TargetsFound"
	ReasonTargetsError            = "TargetsError"
	ReasonAnalyticsServiceError   = "AnalyticsServiceError"
	ReasonAnalyticsServiceRunning = "AnalyticsServiceRunning"
	ReasonIterationUpdate         = "IterationUpdate"
	ReasonAssessmentUpdate        = "AssessmentUpdate"
	ReasonTrafficUpdate           = "TrafficUpdate"
	ReasonExperimentCompleted     = "ExperimentCompleted"
	ReasonSyncMetricsError        = "SyncMetricsError"
	ReasonSyncMetricsSucceeded    = "SyncMetricsSucceeded"
	ReasonRoutingRulesError       = "RoutingRulesError"
	ReasonRoutingRulesReady       = "RoutingRulesReady"
	ReasonActionPause             = "ActionPause"
	ReasonActionResume            = "ActionResume"
)
