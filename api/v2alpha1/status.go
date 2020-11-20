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

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	//DefaultCompletedIterations is the number of iterations that have completed; ie, 0
	DefaultCompletedIterations = 0
)

var experimentCondSet = []ExperimentConditionType{
	ExperimentConditionExperimentInitialized,
	ExperimentConditionStartHandlerLaunched,
	ExperimentConditionStartHandlerCompleted,
	ExperimentConditionFinishHandlerLaunched,
	ExperimentConditionFinishHandlerCompleted,
	ExperimentConditionRollbackHandlerLaunched,
	ExperimentConditionRollbackHandlerCompleted,
	ExperimentConditionFailureHandlerLaunched,
	ExperimentConditionFailureHandlerCompleted,
	ExperimentConditionMetricsSynced,
	ExperimentConditionAnalyticsServiceNormal,
}

func (s *ExperimentStatus) addCondition(conditionType ExperimentConditionType) *ExperimentCondition {
	condition := &ExperimentCondition{
		Type:   conditionType,
		Status: corev1.ConditionUnknown,
	}
	now := metav1.Now()
	condition.LastTransitionTime = &now
	s.Conditions = append(s.Conditions, condition)
	return condition
}

// GetCondition returns condition of given conditionType
func (s *ExperimentStatus) GetCondition(condition ExperimentConditionType) *ExperimentCondition {
	for _, c := range s.Conditions {
		if c.Type == condition {
			return c
		}
	}

	return s.addCondition(condition)
}

// IsTrue tells whether the experiment condition is true or not
func (c *ExperimentCondition) IsTrue() bool {
	return c.Status == corev1.ConditionTrue
}

// IsFalse tells whether the experiment condition is false or not
func (c *ExperimentCondition) IsFalse() bool {
	return c.Status == corev1.ConditionFalse
}

// IsUnknown tells whether the experiment condition is false or not
func (c *ExperimentCondition) IsUnknown() bool {
	return c.Status == corev1.ConditionUnknown
}

// InitStatus initialize status value of an experiment
func (e *Experiment) InitStatus() {
	// sets relevant unset conditions to Unknown state.
	for _, c := range experimentCondSet {
		e.Status.addCondition(c)
	}

	// TODO be explicit about the condition value

	now := metav1.Now()
	e.Status.InitTime = &now // metav1.Now()

	e.Status.LastUpdateTime = &now // metav1.Now()

	// e.Status.Phase = PhaseProgressing

	completedIterations := int32(0)
	e.Status.CompletedIterations = &completedIterations
}

// IsFinishHandlerRunning ..
func (e *Experiment) IsFinishHandlerRunning() bool {
	if !e.HasFinishHandler() {
		return false
	}
	return e.Status.GetCondition(ExperimentConditionFinishHandlerLaunched).IsTrue() &&
		e.Status.GetCondition(ExperimentConditionFinishHandlerCompleted).IsFalse()
}

// IsFinishHandlerCompleted ..
func (e *Experiment) IsFinishHandlerCompleted() bool {
	if !e.HasFinishHandler() {
		return false
	}
	return e.Status.GetCondition(ExperimentConditionFinishHandlerLaunched).IsTrue() &&
		e.Status.GetCondition(ExperimentConditionFinishHandlerCompleted).IsTrue()

}

// IsRollbackHandlerRunning ..
func (e *Experiment) IsRollbackHandlerRunning() bool {
	if !e.HasRollbackHandler() {
		return false
	}
	return e.Status.GetCondition(ExperimentConditionRollbackHandlerLaunched).IsTrue() &&
		e.Status.GetCondition(ExperimentConditionRollbackHandlerCompleted).IsFalse()
}

// IsRollbackHandlerCompleted ..
func (e *Experiment) IsRollbackHandlerCompleted() bool {
	if !e.HasFinishHandler() {
		return false
	}
	return e.Status.GetCondition(ExperimentConditionRollbackHandlerLaunched).IsTrue() &&
		e.Status.GetCondition(ExperimentConditionRollbackHandlerCompleted).IsTrue()
}

// IsFailureHandlerRunning ..
func (e *Experiment) IsFailureHandlerRunning() bool {
	if !e.HasFailureHandler() {
		return false
	}
	return e.Status.GetCondition(ExperimentConditionFailureHandlerLaunched).IsTrue() &&
		e.Status.GetCondition(ExperimentConditionFailureHandlerCompleted).IsFalse()
}

// IsFailureHandlerCompleted ..
func (e *Experiment) IsFailureHandlerCompleted() bool {
	if !e.HasFinishHandler() {
		return false
	}
	return e.Status.GetCondition(ExperimentConditionFailureHandlerLaunched).IsTrue() &&
		e.Status.GetCondition(ExperimentConditionFailureHandlerCompleted).IsTrue()
}

// GetCompletedIterations ..
func (s *ExperimentStatus) GetCompletedIterations() int32 {
	if s.CompletedIterations == nil {
		return 0
	}
	return *s.CompletedIterations
}

// IncrementCompletedIterations ..
func (s *ExperimentStatus) IncrementCompletedIterations() int32 {
	if s.CompletedIterations == nil {
		iteration := int32(DefaultCompletedIterations)
		s.CompletedIterations = &iteration
	}
	*s.CompletedIterations++
	return *s.CompletedIterations
}

// MarkExperimentCompleted sets the condition that the experiemnt is completed
func (s *ExperimentStatus) MarkExperimentCompleted(messageFormat string, messageA ...interface{}) (bool, string) {
	reason := ReasonExperimentCompleted
	message := composeMessage(reason, messageFormat, messageA...)
	// s.Phase = PhaseCompleted
	s.Message = &message
	return s.GetCondition(ExperimentConditionExperimentCompleted).
		markCondition(corev1.ConditionTrue, reason, messageFormat, messageA...), reason
}

// MarkAnalyticsServiceError sets the condition that the analytics service breaks down
// Return true if it's converted from true or unknown
func (s *ExperimentStatus) MarkAnalyticsServiceError(messageFormat string, messageA ...interface{}) (bool, string) {
	reason := ReasonAnalyticsServiceError
	message := composeMessage(reason, messageFormat, messageA...)
	s.Message = &message
	// s.Phase = PhasePause
	return s.GetCondition(ExperimentConditionAnalyticsServiceNormal).
		markCondition(corev1.ConditionFalse, reason, messageFormat, messageA...), reason
}

// MarkAnalyticsServiceRunning sets the condition that the analytics service is operating normally
// Return true if it's converted from false or unknown
func (s *ExperimentStatus) MarkAnalyticsServiceRunning(messageFormat string, messageA ...interface{}) (bool, string) {
	reason := ReasonAnalyticsServiceRunning
	return s.GetCondition(ExperimentConditionAnalyticsServiceNormal).
		markCondition(corev1.ConditionTrue, reason, messageFormat, messageA...), reason
}

// MarkIterationUpdate sets the condition that the iteration updated
func (s *ExperimentStatus) MarkIterationUpdate(messageFormat string, messageA ...interface{}) (bool, string) {
	reason := ReasonIterationUpdate
	message := composeMessage(reason, messageFormat, messageA...)
	// s.Phase = PhaseProgressing
	s.Message = &message
	now := metav1.Now()
	s.LastUpdateTime = &now
	return s.GetCondition(ExperimentConditionExperimentCompleted).
		markCondition(corev1.ConditionFalse, reason, messageFormat, messageA...), reason
}

func (c *ExperimentCondition) markCondition(status corev1.ConditionStatus, reason, messageFormat string, messageA ...interface{}) bool {
	message := fmt.Sprintf(messageFormat, messageA...)
	updated := status != c.Status || reason != *c.Reason || message != *c.Message
	c.Status = status
	c.Reason = &reason
	c.Message = &message
	now := metav1.Now()
	c.LastTransitionTime = &now
	return updated
}

func composeMessage(reason, messageFormat string, messageA ...interface{}) string {
	out := reason
	msg := fmt.Sprintf(messageFormat, messageA...)
	if len(msg) > 0 {
		out += ": " + msg
	}
	return out
}
