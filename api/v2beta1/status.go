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

package v2beta1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	//DefaultCompletedIterations is the number of iterations that have completed; ie, 0
	DefaultCompletedIterations = 0
)

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

	stage := ExperimentStageWaiting
	e.Status.Stage = &stage

	e.TestingPattern()

	completedIterations := int32(0)
	e.Status.CompletedIterations = &completedIterations
}

// TestingPattern assigns a "testing pattern" to the experiment. Note that if the
// experiment is defined incorrectly, the pattern assigned may be incorrect.
func (e *Experiment) TestingPattern() TestingPatternType {
	if e.Status.TestingPattern == nil {
		// set e.Status.TestingPattern
		numVersions := len(e.Spec.Versions)
		hasReward := e.Spec.Criteria != nil && len(e.Spec.Criteria.Rewards) > 0
		hasObjectives := e.Spec.Criteria != nil && len(e.Spec.Criteria.Objectives) > 0

		testingPattern := TestingPatternSLOValidation

		if hasReward {
			if !hasObjectives {
				if numVersions == 2 {
					testingPattern = TestingPatternAB
				} else if numVersions > 2 {
					testingPattern = TestingPatternABN
				}
			} else {
				if numVersions == 2 {
					testingPattern = TestingPatternHybridAB
				} else if numVersions > 2 {
					testingPattern = TestingPatternHybridABN
				}
			}
		}

		// Note that numVersions >= 0 (crd validation ensures this)
		// If numVersions == 1 and hasReward --> validation error (cf. validNumberOfRewards())

		e.Status.TestingPattern = &testingPattern
	}

	// return TestingPattern
	return *e.Status.TestingPattern
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

// SetVersionRecommendedForPromotion sets a version recommended for promotion to either:
// the recommended winner or the current baseline
func (s *ExperimentStatus) SetVersionRecommendedForPromotion(currentBaseline string) {
	recommendation := identfiedWinner(s.Analysis)
	if recommendation == nil {
		recommendation = &currentBaseline
	}
	if s.VersionRecommendedForPromotion == nil {
		s.VersionRecommendedForPromotion = recommendation
	}
	if *s.VersionRecommendedForPromotion != *recommendation {
		s.VersionRecommendedForPromotion = recommendation
	}
}

func identfiedWinner(analysis *Analysis) *string {
	if analysis == nil || analysis.WinnerAssessment == nil {
		return nil
	}
	if !analysis.WinnerAssessment.Data.WinnerFound {
		return nil
	}
	if analysis.WinnerAssessment.Data.Winner == nil {
		return nil
	}
	return analysis.WinnerAssessment.Data.Winner
}

// MarkCondition sets a condition with a status, reason and message.
// The reason and method are also combined to set status.Message
// Note that we compare all fields to determine if we are actually changing anything.
// We do this because we want to also expose the message externally (via Kubernetes events and
// notifications) but want to do so only once -- the first time it is set.
func (s *ExperimentStatus) MarkCondition(condition ExperimentConditionType, status corev1.ConditionStatus, reason string, messageFormat string, messageA ...interface{}) bool {
	conditionMessage := fmt.Sprintf(messageFormat, messageA...)

	statusMessage := reason
	if len(conditionMessage) > 0 {
		statusMessage += ": " + conditionMessage
	}
	s.Message = &statusMessage

	c := s.GetCondition(condition)
	updated := status != c.Status || c.Reason == nil || c.Message == nil || reason != *c.Reason || conditionMessage != *c.Message
	c.Status = status
	c.Reason = &reason
	c.Message = &conditionMessage
	now := metav1.Now()
	c.LastTransitionTime = &now
	return updated
}
