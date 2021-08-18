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

	// DefaultLoopHandler is the prefix of the default loop handler
	DefaultLoopHandler string = "loop"

	// DefaultMinIntervalBetweenLoops is default interval duration as a string
	DefaultMinIntervalBetweenLoops = 20

	// DefaultMaxLoops is the default maximum number of loops, 1
	// reserved for future use
	DefaultMaxLoops int32 = 15
)

// DefaultBlueGreenSplit is the default split to be used for bluegreen experiment
var DefaultBlueGreenSplit = []int32{0, 100}

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

/// GetFailureHandler returns the handler to be called if there is a failure during experiment execution
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
// spec.duration
//////////////////////////////////////////////////////////////////////

// GetIntervalSeconds returns specified(or default) interval for each duration
func (s *ExperimentSpec) GetIntervalSeconds() int32 {
	if s.Duration == nil || s.Duration.MinIntervalBetweenLoops == nil {
		return DefaultMinIntervalBetweenLoops
	}
	return *s.Duration.MinIntervalBetweenLoops
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
	if s.Duration.MinIntervalBetweenLoops == nil {
		interval := int32(DefaultMinIntervalBetweenLoops)
		s.Duration.MinIntervalBetweenLoops = &interval
	}
}

// GetMaxLoops returns specified (or default) max mumber of loops
func (s *ExperimentSpec) GetMaxLoops() int32 {
	if s.Duration == nil || s.Duration.MaxLoops == nil {
		return DefaultMaxLoops
	}
	return *s.Duration.MaxLoops
}

// InitializeMaxLoops sets duration.maxLoops to the default if not already set
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
	s.InitializeMaxLoops()
}

// InitializeSpec initializes values in Spec to default values if not already set
func (s *ExperimentSpec) InitializeSpec() {
	s.InitializeDuration()
}
