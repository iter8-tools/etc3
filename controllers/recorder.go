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

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Methods here share a change simultaneously in many ways:
//   - change in status condition
//   - log it
//   - issue a Kubernetes event
//   - issue a notification
//   - call a webhook
// This can also be used to take a specfic action. This should probably be avoided.

func (r *ExperimentReconciler) markAnalyticsServiceError(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	if updated, reason := instance.Status.MarkAnalyticsServiceError(messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason + ", " + fmt.Sprintf(messageFormat, messageA...))
		// record event
		// send notifications
		r.StatusModified = true
	}
}

func (r *ExperimentReconciler) markAnalyticsServiceRunning(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	if updated, reason := instance.Status.MarkAnalyticsServiceRunning(messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason)
		// record event
		// send notifications
		r.StatusModified = true
	}
}

func (r *ExperimentReconciler) markMetricUnavailable(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	if updated, reason := instance.Status.MarkMetricUnavailable(messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason)
		// record event
		// send notifications
		r.StatusModified = true
	}
}

func (r *ExperimentReconciler) markMetricsSynced(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	log := util.Logger(ctx)
	log.Info("markMetricsSynced() called")
	defer log.Info("markMetricsSynced() completed")

	if updated, reason := instance.Status.MarkMetricsSynced(log, messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason)
		// record event
		// send notifications
		r.StatusModified = true
	}
}

func (r *ExperimentReconciler) markIterationUpdate(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	if updated, reason := instance.Status.MarkIterationUpdate(messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason + ", " + fmt.Sprintf(messageFormat, messageA...))
		// record event
		// send notifications
		r.StatusModified = true
	}
}

func (r *ExperimentReconciler) markExperimentCompleted(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	if updated, reason := instance.Status.MarkExperimentCompleted(messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason + ", " + fmt.Sprintf(messageFormat, messageA...))
		// record event
		// send notifications

		now := metav1.Now()
		instance.Status.EndTime = &now
		r.StatusModified = true
	}
}
