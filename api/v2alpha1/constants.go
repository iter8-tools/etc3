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

package v2alpha1

// StrategyType identifies the type of experiment type
// +kubebuilder:validation:Enum=canary;A/B;A/B/N;performance;bluegreen
type StrategyType string

const (
	// StrategyTypeCanary indicates an experiment is a canary experiment
	StrategyTypeCanary StrategyType = "canary"

	// StrategyTypeAB indicates an experiment is a A/B experiment
	StrategyTypeAB StrategyType = "A/B"

	// StrategyTypeABN indicates an experiment is a A/B/n experiment
	StrategyTypeABN StrategyType = "A/B/N"

	// StrategyTypePerformance indicates an experiment is a performance experiment
	StrategyTypePerformance StrategyType = "performance"

	// StrategyTypeBlueGreen indicates an experiment is a blue-green experiment
	StrategyTypeBlueGreen StrategyType = "bluegreen"
)

// AlgorithmType identifies the algorithms that can be used
// +kubebuilder:validation:Enum=fixed_split;progressive
type AlgorithmType string

const (
	// AlgorithmTypeFixedSplit indicates the weight distribution algorithm is a fixed split
	AlgorithmTypeFixedSplit AlgorithmType = "fixed_split"

	// AlgorithmTypeProgressive indicates that the the weight distribution algorithm is progressive
	AlgorithmTypeProgressive AlgorithmType = "progressive"
)

// PreferredDirectionType defines the valid values for reward.PreferredDirection
// +kubebuilder:validation:Enum=higher;lower
type PreferredDirectionType string

const (
	// PreferredDirectionHigher indicates that a higher value is "better"
	PreferredDirectionHigher PreferredDirectionType = "higher"

	// PreferredDirectionLower indicates that a lower value is "better"
	PreferredDirectionLower PreferredDirectionType = "lower"
)

// ExperimentConditionType limits conditions can be set by controller
// +kubebuilder:validation:Enum:=ExperimentInitialized;StartHandlerLaunched;StartHandlerCompleted;FinishHandlerLaunched;FinishHandlerCompleted;MetricsSynced;AnalyticsServiceNormal;ExperimentCompleted
type ExperimentConditionType string

const (
	// ExperimentConditionExperimentInitialized ..
	// Unknown at start, set to False immediately; True when done
	ExperimentConditionExperimentInitialized ExperimentConditionType = "ExperimentInitialized"

	// ExperimentConditionStartHandlerLaunched ..
	// False until launched, True thereafter
	ExperimentConditionStartHandlerLaunched ExperimentConditionType = "StartHandlerLaunched"

	// ExperimentConditionStartHandlerCompleted ..
	// False until completed; True when done
	ExperimentConditionStartHandlerCompleted ExperimentConditionType = "StartHandlerCompleted"

	// ExperimentConditionFinishHandlerLaunched ..
	// False until launched; True thereafter
	ExperimentConditionFinishHandlerLaunched ExperimentConditionType = "FinishHandlerLaunched"

	// ExperimentConditionFinishHandlerCompleted ..
	// Unknown until called; False until completed; True when done
	ExperimentConditionFinishHandlerCompleted ExperimentConditionType = "FinishHandlerCompleted"

	// ExperimentConditionMetricsSynced ..
	// Unknown before reading metrics
	// True when done; False if any error
	// Future: go to paused state if can't find metric; resume when defined or experiment changed
	ExperimentConditionMetricsSynced ExperimentConditionType = "MetricsSynced"

	// ExperimentConditionAnalyticsServiceNormal ..
	// Unknown before any attemtps to call analytics service
	// True while calls successful
	// False if a call fails
	ExperimentConditionAnalyticsServiceNormal ExperimentConditionType = "AnalyticsServiceNormal"

	// ExperimentConditionExperimentCompleted has status True when the experiment is completed
	// Unknown initially, set to False during initialization
	ExperimentConditionExperimentCompleted ExperimentConditionType = "ExperimentCompleted"
)

// PhaseType has options for phases that an experiment can be at
type PhaseType string

const (
	// PhasePaused indicates experiment is paused; this occurs because a needed resource
	// is not available. For example, another experiment may already be in progress using
	// the same target.
	PhasePaused PhaseType = "Paused"

	// PhaseProgressing indicates experiment is progressing
	PhaseProgressing PhaseType = "Progressing"

	// PhaseCompleted indicates experiment has competed (successfully or not)
	PhaseCompleted PhaseType = "Completed"
)

// A set of reason setting the experiment condition status
// TBD
const (
	ReasonIterationUpdate     = "IterationUpdate"
	ReasonExperimentCompleted = "ExperimentCompleted"
)
