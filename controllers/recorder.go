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

// recorder.go - methods to modify status.conditions. Each method allows for a single place to:
//     - change status.condition
//     - logs change
//     - issue kubernetes event (not currently implemented)
//     - send notification (not currently implemented)

package controllers

import (
	"context"
	"fmt"

	v2beta1 "github.com/iter8-tools/etc3/api/v2beta1"
	corev1 "k8s.io/api/core/v1"
)

func (r *ExperimentReconciler) recordExperimentFailed(ctx context.Context, instance *v2beta1.Experiment,
	reason string, messageFormat string, messageA ...interface{}) {
	r.recordEvent(ctx, instance,
		v2beta1.ExperimentConditionExperimentFailed, corev1.ConditionTrue,
		reason, messageFormat, messageA...)
}

func (r *ExperimentReconciler) recordExperimentCompleted(ctx context.Context, instance *v2beta1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.recordEvent(ctx, instance,
		v2beta1.ExperimentConditionExperimentCompleted, corev1.ConditionTrue,
		v2beta1.ReasonExperimentCompleted, messageFormat, messageA...)
}

func (r *ExperimentReconciler) recordExperimentProgress(ctx context.Context, instance *v2beta1.Experiment,
	reason string, messageFormat string, messageA ...interface{}) {
	r.recordEvent(ctx, instance,
		v2beta1.ExperimentConditionExperimentCompleted, corev1.ConditionFalse,
		reason, messageFormat, messageA...)
}

// record the event in a variety of ways. Note that we do not want to report an event more than once
// in a log message, kubernetes event or notification. Consequently, we must pay attention to whether
// or not we are recording an event for the first time or repeating it. We do this by first updating
// a condition on instance.Status. If the condition changes, we report the event externally.
func (r *ExperimentReconciler) recordEvent(ctx context.Context, instance *v2beta1.Experiment,
	condition v2beta1.ExperimentConditionType, status corev1.ConditionStatus,
	reason string, messageFormat string, messageA ...interface{}) {
	ok := instance.Status.MarkCondition(condition, status, reason, messageFormat, messageA...)
	if ok {
		Logger(ctx).Info(reason + ", " + fmt.Sprintf(messageFormat, messageA...))
		r.EventRecorder.Eventf(instance, corev1.EventTypeNormal, reason, messageFormat, messageA...)
		// FUTURE: send notifications
	}
}
