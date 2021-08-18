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

// constants.go - values of constants used in experiment model

package v2beta1

// TestingPatternType identifies the type of experiment type
// +kubebuilder:validation:Enum=SLOValidation;A/B;A/B/N;Hybrid-A/B;Hybrid-A/B/N
type TestingPatternType string

const (
	// TestingPatternSLOValidation indicates an experiment tests for SLO validation
	TestingPatternSLOValidation TestingPatternType = "SLOValidation"

	// TestingPatternAB indicates an experiment is a A/B experiment
	TestingPatternAB TestingPatternType = "A/B"

	// TestingPatternABN indicates an experiment is a A/B/n experiment
	TestingPatternABN TestingPatternType = "A/B/N"

	// TestingPatternHybridAB indicates an experiment is a Hybrid-A/B experiment
	TestingPatternHybridAB TestingPatternType = "Hybrid-A/B"

	// TestingPatternHybridABN indicates an experiment is a Hybrid-A/B/n experiment
	TestingPatternHybridABN TestingPatternType = "Hybrid-A/B/N"
)

// PreferredDirectionType defines the valid values for reward.PreferredDirection
// +kubebuilder:validation:Enum=High;Low
type PreferredDirectionType string

const (
	// PreferredDirectionHigher indicates that a higher value is "better"
	PreferredDirectionHigher PreferredDirectionType = "High"

	// PreferredDirectionLower indicates that a lower value is "better"
	PreferredDirectionLower PreferredDirectionType = "Low"
)

// ExperimentConditionType limits conditions can be set by controller
// +kubebuilder:validation:Enum:=Completed;Failed
type ExperimentConditionType string

const (
	// ExperimentConditionExperimentCompleted has status True when the experiment is completed
	// Unknown initially, set to False during initialization
	ExperimentConditionExperimentCompleted ExperimentConditionType = "Completed"

	// ExperimentConditionExperimentFailed has status True when the experiment has failed
	// False until failure occurs
	ExperimentConditionExperimentFailed ExperimentConditionType = "Failed"
)

// A set of reason setting the experiment condition status
const (
	ReasonExperimentInitialized      = "ExperimentInitialized"
	ReasonLoopCompleted              = "LoopUpdate"
	ReasonExperimentCompleted        = "ExperimentCompleted"
	ReasonAnalyticsServiceError      = "AnalyticsServiceError"
	ReasonMetricUnavailable          = "MetricUnavailable"
	ReasonMetricsUnreadable          = "MetricsUnreadable"
	ReasonHandlerLaunched            = "HandlerLaunched"
	ReasonHandlerCompleted           = "HandlerCompleted"
	ReasonHandlerFailed              = "HandlerFailed"
	ReasonLaunchHandlerFailed        = "LaunchHandlerFailed"
	ReasonWeightRedistributionFailed = "WeightRedistributionFailed"
	ReasonInvalidExperiment          = "InvalidExperiment"
	ReasonStageAdvanced              = "StageAdvanced"
)

// ExperimentStageType identifies valid stages of an experiment
// +kubebuilder:validation:Enum:=Initializing;Running;Finishing;Completed
type ExperimentStageType string

const (
	// ExperimentStageInitializing indicates an experiment has acquired access to the target
	// and a start handler, if  any, is running
	ExperimentStageInitializing ExperimentStageType = "Initializing"

	// ExperimentStageRunning indicates an experiment is running
	ExperimentStageRunning ExperimentStageType = "Running"

	// ExperimentStageFinishing indicates an experiment has completed its loops and is
	// running any termination handler (either success or  failure)
	ExperimentStageFinishing ExperimentStageType = "Finishing"

	// ExperimentStageCompleted indicates an experiment has completed
	ExperimentStageCompleted ExperimentStageType = "Completed"
)

// After Determines if a stage is after another
func (stage ExperimentStageType) After(otherStage ExperimentStageType) bool {
	orderedStages := []ExperimentStageType{
		ExperimentStageInitializing,
		ExperimentStageRunning,
		ExperimentStageFinishing,
		ExperimentStageCompleted,
	}

	return stageIndex(stage, orderedStages) > stageIndex(otherStage, orderedStages)
}

func stageIndex(value ExperimentStageType, stages []ExperimentStageType) int {
	for pos, val := range stages {
		if val == value {
			return pos
		}
	}
	return -1
}

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
