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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *ExperimentReconciler) markAnalyticsServiceError(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentFailed(ctx, instance, v2alpha1.ReasonAnalyticsServiceError, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markLaunchHandlerFailed(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentFailed(ctx, instance, v2alpha1.ReasonLaunchHandlerFailed, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markHandlerFailedError(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentFailed(ctx, instance, v2alpha1.ReasonHandlerFailed, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markWeightRedistributionFailed(ctx context.Context, instance *v2alpha1.Experiment,
	reason string, messageFormat string, messageA ...interface{}) {
	r.markExperimentFailed(ctx, instance, v2alpha1.ReasonWeightRedistributionFailed, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markMetricUnavailable(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentFailed(ctx, instance, v2alpha1.ReasonHandlerFailed, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markInvalidExperiment(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentFailed(ctx, instance, v2alpha1.ReasonInvalidExperiment, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markExperimentFailed(ctx context.Context, instance *v2alpha1.Experiment,
	reason string, messageFormat string, messageA ...interface{}) {
	if updated, reason := instance.Status.MarkExperimentFailed(reason, messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason + ", " + fmt.Sprintf(messageFormat, messageA...))
		r.EventRecorder.Eventf(instance, corev1.EventTypeWarning, reason, messageFormat, messageA...)
		// send notificastions
		r.StatusModified = true
	}
}

func (r *ExperimentReconciler) markExperimentCompleted(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	if updated, reason := instance.Status.MarkExperimentCompleted(messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason + ", " + fmt.Sprintf(messageFormat, messageA...))
		r.EventRecorder.Eventf(instance, corev1.EventTypeNormal, reason, messageFormat, messageA...)
		// send notifications

		now := metav1.Now()
		instance.Status.EndTime = &now
		r.StatusModified = true
	}
}

func (r *ExperimentReconciler) markExperimentInitialized(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentProgress(ctx, instance, v2alpha1.ReasonExperimentInitialized, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markStartHandlerLaunched(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentProgress(ctx, instance, v2alpha1.ReasonStartHandlerLaunched, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markStartHandlerCompleted(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentProgress(ctx, instance, v2alpha1.ReasonStartHandlerCompleted, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markTargetAcquired(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentProgress(ctx, instance, v2alpha1.ReasonTargetAcquired, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markIterationCompleted(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentProgress(ctx, instance, v2alpha1.ReasonIterationCompleted, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markTerminalHandlerLaunched(ctx context.Context, instance *v2alpha1.Experiment,
	messageFormat string, messageA ...interface{}) {
	r.markExperimentProgress(ctx, instance, v2alpha1.ReasonTerminalHandlerLaunched, messageFormat, messageA...)
}

func (r *ExperimentReconciler) markExperimentProgress(ctx context.Context, instance *v2alpha1.Experiment,
	reason string, messageFormat string, messageA ...interface{}) {
	if updated, reason := instance.Status.MarkExperimentProgressing(reason, messageFormat, messageA...); updated {
		util.Logger(ctx).Info(reason + ", " + fmt.Sprintf(messageFormat, messageA...))
		r.EventRecorder.Eventf(instance, corev1.EventTypeNormal, reason, messageFormat, messageA...)
		// send notifications

		now := metav1.Now()
		instance.Status.EndTime = &now
		r.StatusModified = true
	}
}
