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

// status.go - methods to get and update status fields

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
	ExperimentConditionExperimentFailed,
	ExperimentConditionExperimentCompleted,
}

func (s *ExperimentStatus) addCondition(conditionType ExperimentConditionType, status corev1.ConditionStatus) *ExperimentCondition {
	condition := &ExperimentCondition{
		Type:   conditionType,
		Status: status,
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

	return s.addCondition(condition, corev1.ConditionUnknown)
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

// InitializeStatus initialize status value of an experiment
func (e *Experiment) InitializeStatus() {
	// sets relevant unset conditions to Unknown state.
	e.Status.addCondition(ExperimentConditionExperimentCompleted, corev1.ConditionFalse)
	e.Status.addCondition(ExperimentConditionExperimentFailed, corev1.ConditionFalse)

	now := metav1.Now()
	e.Status.InitTime = &now // metav1.Now()

	e.Status.LastUpdateTime = &now // metav1.Now()

	completedIterations := int32(0)
	e.Status.CompletedIterations = &completedIterations
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
	s.Message = &message
	return s.GetCondition(ExperimentConditionExperimentCompleted).
		markCondition(corev1.ConditionTrue, reason, messageFormat, messageA...), reason
}

// MarkExperimentProgressing sets the condition that the experiemnt is progressing
func (s *ExperimentStatus) MarkExperimentProgressing(reason string, messageFormat string, messageA ...interface{}) (bool, string) {
	message := composeMessage(reason, messageFormat, messageA...)
	s.Message = &message
	return s.GetCondition(ExperimentConditionExperimentCompleted).
		markCondition(corev1.ConditionFalse, reason, messageFormat, messageA...), reason
}

// MarkExperimentFailed sets the condition that the experiment failed
// Return true if it's converted from true or unknown
func (s *ExperimentStatus) MarkExperimentFailed(reason string, messageFormat string, messageA ...interface{}) (bool, string) {
	message := composeMessage(reason, messageFormat, messageA...)
	s.Message = &message
	return s.GetCondition(ExperimentConditionExperimentFailed).
		markCondition(corev1.ConditionTrue, reason, messageFormat, messageA...), reason
}

func (c *ExperimentCondition) markCondition(status corev1.ConditionStatus, reason string, messageFormat string, messageA ...interface{}) bool {
	fmt.Printf("status = %v\n", status)
	fmt.Printf("resason = %s\n", reason)
	fmt.Printf("messageFormat = %s\n", messageFormat)
	fmt.Printf("condition = %v\n", c)
	message := fmt.Sprintf(messageFormat, messageA...)
	updated := status != c.Status || c.Reason == nil || c.Message == nil || reason != *c.Reason || message != *c.Message
	c.Status = status
	c.Reason = &reason
	c.Message = &message
	now := metav1.Now()
	c.LastTransitionTime = &now
	return updated
}

func composeMessage(reason string, messageFormat string, messageA ...interface{}) string {
	out := reason
	msg := fmt.Sprintf(messageFormat, messageA...)
	if len(msg) > 0 {
		out += ": " + msg
	}
	return out
}
