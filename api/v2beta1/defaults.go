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

// defaults.go - methods to get values for optional spec fields that return a default value when none set
//             - methods to initialize spec fields with default or derived values

package v2beta1

import (
	"time"
)

const (
	// DefaultStartHandler is the prefix of the default start handler
	DefaultStartHandler string = "start"

	// DefaultFinishHandler is the prefix of the default finish handler
	DefaultFinishHandler string = "finish"

	// DefaultFailureHandler is the prefix of the default failure handler
	DefaultFailureHandler string = "finish"

	// DefaultRollbackHandler is the prefix of the default rollback handler
	DefaultRollbackHandler string = "finish"

	// DefaultLoopHandler is the prefix of the default loop handler
	DefaultLoopHandler string = "loop"

	// DefaultMaxCandidateWeight is the default traffic percentage used in experiment, which is 100
	DefaultMaxCandidateWeight int32 = 100

	// DefaultMaxCandidateWeightIncrement is the default maxIncrement for traffic update, which is 10
	DefaultMaxCandidateWeightIncrement int32 = 10

	// DefaultDeploymentPattern is the default deployment pattern for experiments
	// It takes effect when the testing pattern is canary, A/B or A/B/n
	DefaultDeploymentPattern DeploymentPatternType = DeploymentPatternProgressive

	// DefaultIntervalSeconds is default interval duration as a string
	DefaultIntervalSeconds = 20

	// DefaultIterationsPerLoop is the default number of iterations, 15
	DefaultIterationsPerLoop int32 = 15

	// DefaultMaxLoops is the default maximum number of loops, 1
	// reserved for future use
	DefaultMaxLoops int32 = 1
)

// DefaultBlueGreenSplit is the default split to be used for bluegreen experiment
var DefaultBlueGreenSplit = []int32{0, 100}

// GetNumberOfCandidates returns the number of candidates in VersionInfo
func (s *ExperimentSpec) GetNumberOfCandidates() int {
	if s.VersionInfo == nil {
		return 0
	}
	return len((*s.VersionInfo).Candidates)
}

// HasBaseline determines if a baseline has been identified in a s.VersionInfo
func (s *ExperimentSpec) HasBaseline() bool {
	return !(s.VersionInfo == nil)
}

// GetNumberOfBaseline returns the number of baselines in VersionInfo (1 if present, 0 otherwise)
func (s *ExperimentSpec) GetNumberOfBaseline() int {
	if s.HasBaseline() {
		return 1
	}
	return 0
}

//////////////////////////////////////////////////////////////////////
// spec.strategy.handlers
//////////////////////////////////////////////////////////////////////

// GetStartHandler returns the name of the handler to be called when an experiment starts
func (s *ExperimentSpec) GetStartHandler() *string {
	handler := DefaultStartHandler
	return &handler
}

// GetFinishHandler returns the handler that should be called when an experiment ha completed.
func (s *ExperimentSpec) GetFinishHandler() *string {
	handler := DefaultFinishHandler
	return &handler
}

// GetRollbackHandler returns the handler to be called if a candidate fails its objective(s)
func (s *ExperimentSpec) GetRollbackHandler() *string {
	handler := DefaultRollbackHandler
	return &handler
}

// GetFailureHandler returns the handler to be called if there is a failure during experiment execution
func (s *ExperimentSpec) GetFailureHandler() *string {
	handler := DefaultFailureHandler
	return &handler
}

// GetLoopHandler returns the handler to be called at the end of each loop (except the last)
func (s *ExperimentSpec) GetLoopHandler() *string {
	handler := DefaultLoopHandler
	return &handler
}

//////////////////////////////////////////////////////////////////////
// spec.strategy.weights
//////////////////////////////////////////////////////////////////////

// GetMaxCandidateWeight return spec.strategy.weights.maxCandidateWeight if set
// Otherwise it returns DefaultMaxCandidateWeight (100)
func (s *ExperimentSpec) GetMaxCandidateWeight() int32 {
	if s.Strategy.Weights == nil || s.Strategy.Weights.MaxCandidateWeight == nil {
		return DefaultMaxCandidateWeight
	}
	return *s.Strategy.Weights.MaxCandidateWeight
}

// InitializeMaxCandidateWeight initializes spec.strategy.weights.maxCandiateWeight if not already set
func (s *ExperimentSpec) InitializeMaxCandidateWeight() {
	if s.Strategy.Weights == nil {
		s.Strategy.Weights = &Weights{}
	}
	if s.Strategy.Weights.MaxCandidateWeight == nil {
		weight := s.GetMaxCandidateWeight()
		s.Strategy.Weights.MaxCandidateWeight = &weight
	}
}

// GetMaxCandidateWeightIncrement return spec.strategy.weights.maxCandidateWeightIncrement if set
// Otherwise it returns DefaultMaxCandidateWeightIncrement (10)
func (s *ExperimentSpec) GetMaxCandidateWeightIncrement() int32 {
	if s.Strategy.Weights == nil || s.Strategy.Weights.MaxCandidateWeightIncrement == nil {
		return DefaultMaxCandidateWeightIncrement
	}
	return *s.Strategy.Weights.MaxCandidateWeightIncrement
}

// InitializeMaxCandidateWeightIncrement initializes spec.strategy.weights.maxCandidateWeightIncrement if not already set
func (s *ExperimentSpec) InitializeMaxCandidateWeightIncrement() {
	if s.Strategy.Weights == nil {
		s.Strategy.Weights = &Weights{}
	}
	if s.Strategy.Weights.MaxCandidateWeightIncrement == nil {
		increment := s.GetMaxCandidateWeightIncrement()
		s.Strategy.Weights.MaxCandidateWeightIncrement = &increment
	}
}

// GetDeploymentPattern returns spec.strategy.deploymentPattern if set
func (s *ExperimentSpec) GetDeploymentPattern() DeploymentPatternType {
	if s.Strategy.DeploymentPattern == nil {
		return DefaultDeploymentPattern
	}
	return *s.Strategy.DeploymentPattern
}

// InitializeDeploymentPattern initializes spec.strategy.deploymentPattern if not already set
func (s *ExperimentSpec) InitializeDeploymentPattern() {
	if s.Strategy.DeploymentPattern == nil {
		deploymentPattern := s.GetDeploymentPattern()
		s.Strategy.DeploymentPattern = &deploymentPattern
	}
}

// InitializeWeights initializes weights if not already set
func (s *ExperimentSpec) InitializeWeights() {
	s.InitializeMaxCandidateWeight()
	s.InitializeMaxCandidateWeightIncrement()
	s.InitializeDeploymentPattern()
	// Must wait until versionInfo has been defined by start handler before
	// initializing weight distribution because need to know the candidates
	// s.InitializeWeightDistribution()

}

//////////////////////////////////////////////////////////////////////
// spec.duration
//////////////////////////////////////////////////////////////////////

// GetIntervalSeconds returns specified(or default) interval for each duration
func (s *ExperimentSpec) GetIntervalSeconds() int32 {
	if s.Duration == nil || s.Duration.IntervalSeconds == nil {
		return DefaultIntervalSeconds
	}
	return *s.Duration.IntervalSeconds
}

// GetIntervalAsDuration returns spec.duration.intervalSeconds as a time.Duration (in ns)
func (s *ExperimentSpec) GetIntervalAsDuration() time.Duration {
	return time.Second * time.Duration(s.GetIntervalSeconds())
}

// InitializeInterval sets duration.interval if not already set using the default value
func (s *ExperimentSpec) InitializeInterval() {
	if s.Duration == nil {
		s.Duration = &Duration{}
	}
	if s.Duration.IntervalSeconds == nil {
		interval := int32(DefaultIntervalSeconds)
		s.Duration.IntervalSeconds = &interval
	}
}

// GetIterationsPerLoop returns the specified (or default) iterations
func (s *ExperimentSpec) GetIterationsPerLoop() int32 {
	if s.Duration == nil || s.Duration.IterationsPerLoop == nil {
		return DefaultIterationsPerLoop
	}
	return *s.Duration.IterationsPerLoop
}

// GetMaxLoops returns specified (or default) max mumber of loops
func (s *ExperimentSpec) GetMaxLoops() int32 {
	if s.Duration == nil || s.Duration.MaxLoops == nil {
		return DefaultMaxLoops
	}
	return *s.Duration.MaxLoops
}

// InitializeIterationsPerLoop sets duration.iterationsPerLoop to the default if not already set
func (s *ExperimentSpec) InitializeIterationsPerLoop() {
	if s.Duration == nil {
		s.Duration = &Duration{}
	}
	if s.Duration.IterationsPerLoop == nil {
		iterations := s.GetIterationsPerLoop()
		s.Duration.IterationsPerLoop = &iterations
	}
}

// InitializeMaxLoops sets duration.iterationsPerLoop to the default if not already set
func (s *ExperimentSpec) InitializeMaxLoops() {
	if s.Duration == nil {
		s.Duration = &Duration{}
	}
	if s.Duration.MaxLoops == nil {
		loops := s.GetMaxLoops()
		s.Duration.MaxLoops = &loops
	}
}

// InitializeDuration initializes spec.durations if not already set
func (s *ExperimentSpec) InitializeDuration() {
	s.InitializeInterval()
	s.InitializeIterationsPerLoop()
	s.InitializeMaxLoops()
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////
// criteria
//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

// GetRequestCount returns the requst count metric
// If there are no criteria specified or no request count specified, this is nil
func (s *ExperimentSpec) GetRequestCount() *string {
	if s.Criteria == nil {
		return nil
	}
	// if s.Criteria.RequestCount == nil {
	// 	rc := cfg.RequestCount
	// 	return &rc
	// }
	return s.Criteria.RequestCount
}

// InitializeRequestCount sets the request count metric to the default value if not already set
func (s *ExperimentSpec) InitializeRequestCount() {
	if s.Criteria == nil {
		return
	}
	if s.Criteria.RequestCount == nil {
		s.Criteria.RequestCount = s.GetRequestCount()
	}
}

//////////////////////////////////////////////////////////////////////
// objective
//////////////////////////////////////////////////////////////////////

// GetRollbackOnFailure identifies if the experiment should be rolledback on failure of an objective
func (o *Objective) GetRollbackOnFailure(deploymentPattern DeploymentPatternType) bool {
	if o.RollbackOnFailure == nil {
		return deploymentPattern == DeploymentPatternBlueGreen
	}
	return *o.RollbackOnFailure
}

// InitializeObjectives initializes the rollbackOnFailure field of all objectives if
// the strategy type is "bluegreen"
func (s *ExperimentSpec) InitializeObjectives() {
	if s.Criteria == nil {
		return
	}

	for _, o := range s.Criteria.Objectives {
		if s.GetDeploymentPattern() == DeploymentPatternBlueGreen && o.RollbackOnFailure == nil {
			rollback := true
			o.RollbackOnFailure = &rollback
		}
	}
}

// InitializeCriteria initializes any criteria details not already set
func (s *ExperimentSpec) InitializeCriteria() {
	if s.Criteria != nil {
		s.InitializeRequestCount()
		s.InitializeObjectives()
	}
}

// InitializeSpec initializes values in Spec to default values if not already set
func (s *ExperimentSpec) InitializeSpec() {
	s.InitializeWeights()
	s.InitializeDuration()
	s.InitializeCriteria()
}
