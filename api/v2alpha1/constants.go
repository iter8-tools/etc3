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

// TestingPatternType identifies the type of experiment type
// +kubebuilder:validation:Enum=Canary;A/B;A/B/N;Conformance
type TestingPatternType string

const (
	// TestingPatternCanary indicates an experiment is a canary experiment
	TestingPatternCanary TestingPatternType = "Canary"

	// TestingPatternAB indicates an experiment is a A/B experiment
	TestingPatternAB TestingPatternType = "A/B"

	// TestingPatternABN indicates an experiment is a A/B/n experiment
	TestingPatternABN TestingPatternType = "A/B/N"

	// TestingPatternConformance indicates an experiment is a conformance experiment
	TestingPatternConformance TestingPatternType = "Conformance"
)

// ValidTestingPatternTypes are legal strategy types iter8 is aware of
// Should match list in github.com/iter8-tools/etc3/api/v2alpha1 (cf. constants.go)
var ValidTestingPatternTypes []TestingPatternType = []TestingPatternType{
	TestingPatternCanary,
	TestingPatternAB,
	TestingPatternABN,
	TestingPatternConformance,
}

// DeploymentPatternType identifies the deployment patterns that can be used
// +kubebuilder:validation:Enum=FixedSplit;Progressive;BlueGreen
type DeploymentPatternType string

const (
	// DeploymentPatternFixedSplit indicates the deployment pattern is fixed split
	DeploymentPatternFixedSplit DeploymentPatternType = "FixedSplit"

	// DeploymentPatternProgressive indicates that the deployment pattern progressive
	DeploymentPatternProgressive DeploymentPatternType = "Progressive"

	// DeploymentPatternBlueGreen indicates that the deployment pattern is progressive
	DeploymentPatternBlueGreen DeploymentPatternType = "BlueGreen"
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
// +kubebuilder:validation:Enum:=Completed;Failed;TargetAcquired
type ExperimentConditionType string

const (
	// ExperimentConditionExperimentCompleted has status True when the experiment is completed
	// Unknown initially, set to False during initialization
	ExperimentConditionExperimentCompleted ExperimentConditionType = "Completed"

	// ExperimentConditionExperimentFailed has status True when the experiment has failed
	// False until failure occurs
	ExperimentConditionExperimentFailed ExperimentConditionType = "Failed"

	// ExperimentConditionTargetAcquired has status True when an experiment has a lock on the target
	// False until can lock the target
	ExperimentConditionTargetAcquired ExperimentConditionType = "TargetAcquired"
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

// ExperimentStageType identifies valid stages of an experiment
// +kubebuilder:validation:Enum:=Waiting;Initializing;Running;Finishing;Completed
type ExperimentStageType string

const (
	// ExperimentStageWaiting indicates the experiment is not yet scheduled to run because it
	// does not yet have exclusive experiment access to the target
	ExperimentStageWaiting ExperimentStageType = "Waiting"

	// ExperimentStageInitializing indicates an experiment has acquired access to the target
	// and a start handler, if  any, is running
	ExperimentStageInitializing ExperimentStageType = "Initializing"

	// ExperimentStageRunning indicates an experiment is running
	ExperimentStageRunning ExperimentStageType = "Running"

	// ExperimentStageFinishing indicates an experiment has completed its iterations and is
	// running any termination handler (either success or  failure)
	ExperimentStageFinishing ExperimentStageType = "Finishing"

	// ExperimentStageCompleted indicates an experiment has completed
	ExperimentStageCompleted ExperimentStageType = "Completed"
)

// After Determines if a stage is after another
func (stage ExperimentStageType) After(otherStage ExperimentStageType) bool {
	orderedStages := []ExperimentStageType{
		ExperimentStageWaiting,
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
