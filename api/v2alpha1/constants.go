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

package v2alpha1

// StrategyType identifies the type of experiment type
// +kubebuilder:validation:Enum=Canary;A/B;A/B/N;Performance;BlueGreen
type StrategyType string

const (
	// StrategyTypeCanary indicates an experiment is a canary experiment
	StrategyTypeCanary StrategyType = "Canary"

	// StrategyTypeAB indicates an experiment is a A/B experiment
	StrategyTypeAB StrategyType = "A/B"

	// StrategyTypeABN indicates an experiment is a A/B/n experiment
	StrategyTypeABN StrategyType = "A/B/N"

	// StrategyTypePerformance indicates an experiment is a performance experiment
	StrategyTypePerformance StrategyType = "Performance"

	// StrategyTypeBlueGreen indicates an experiment is a blue-green experiment
	StrategyTypeBlueGreen StrategyType = "BlueGreen"
)

// AlgorithmType identifies the algorithms that can be used
// +kubebuilder:validation:Enum=FixedSplit;Progressive
type AlgorithmType string

const (
	// AlgorithmTypeFixedSplit indicates the weight distribution algorithm is a fixed split
	AlgorithmTypeFixedSplit AlgorithmType = "FixedSplit"

	// AlgorithmTypeProgressive indicates that the the weight distribution algorithm is progressive
	AlgorithmTypeProgressive AlgorithmType = "Progressive"
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
// +kubebuilder:validation:Enum:=ExperimentCompleted;ExperimentFailed
type ExperimentConditionType string

const (
	// ExperimentConditionExperimentCompleted has status True when the experiment is completed
	// Unknown initially, set to False during initialization
	ExperimentConditionExperimentCompleted ExperimentConditionType = "ExperimentCompleted"

	// ExperimentConditionExperimentFailed has status True when the experiment has failed
	// False until failure occurs
	ExperimentConditionExperimentFailed ExperimentConditionType = "ExperimentFailed"
)

// A set of reason setting the experiment condition status
const (
	ReasonExperimentInitialized      = "ExperimentInitialized"
	ReasonStartHandlerLaunched       = "StartHandlerLaunched"
	ReasonStartHandlerCompleted      = "StartHandlerCompleted"
	ReasonTargetAcquired             = "TargetAcquired"
	ReasonIterationCompleted         = "IterationUpdate"
	ReasonTerminalHandlerLaunched    = "TerminalHandlerLaunched"
	ReasonExperimentCompleted        = "ExperimentCompleted"
	ReasonAnalyticsServiceError      = "AnalyticsServiceError"
	ReasonMetricUnavailable          = "MetricUnavailable"
	ReasonMetricsUnreadable          = "MetricsUnreadable"
	ReasonHandlerFailed              = "HandlerFailed"
	ReasonLaunchHandlerFailed        = "LaunchHandlerFailed"
	ReasonWeightRedistributionFailed = "WeightRedistributionFailed"
	ReasonInvalidExperiment          = "InvalidExperiment"
)
