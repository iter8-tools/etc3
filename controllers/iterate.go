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

// interate.go implments behavior of the iter8 control loop:
//    - query analytics service for updated statistics and recommendations
//    - redistribute weights

package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/iter8-tools/etc3/api/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ExperimentReconciler) sufficientTimePassedSincePreviousLoop(ctx context.Context, instance *v2beta1.Experiment) bool {
	log := Logger(ctx)

	// Is this the first loop or has enough time passed since last loop?
	if instance.Status.GetCompletedLoops() == 0 || instance.Status.LastUpdateTime == nil {
		return true
	}

	now := time.Now()
	interval := instance.Spec.GetIntervalAsDuration()
	expectedTime := instance.Status.LastUpdateTime.Add(interval)
	log.Info("sufficientTimePassedSincePreviousLoop", "lastUpdateTime", instance.Status.LastUpdateTime, "interval", interval, "sum", expectedTime, "now", now)

	if now.Before(expectedTime) {
		// is it close enough?
		difference := expectedTime.Sub(now)
		return difference < 100*time.Millisecond
	}
	// now is after expectedTime
	return true
}

func (r *ExperimentReconciler) doLoop(ctx context.Context, instance *v2beta1.Experiment) (ctrl.Result, error) {
	log := Logger(ctx)
	log.Info("doLoop called")
	defer log.Info("doLoop completed")

	// If we've already executed as many loops as requested, we  should finish the experiment
	// Check here since may have executed a loop handler
	if instance.Spec.GetMaxLoops() <= instance.Status.GetCompletedLoops() {
		return r.finishExperiment(ctx, instance)
	}

	if !r.sufficientTimePassedSincePreviousLoop(ctx, instance) {
		// not enough time has passed since the last loop, wait
		return ctrl.Result{}, errors.New("insufficient time has passed since previous loop")
	}

	// TODO  GET CURRENT WEIGHTS (from cluster)

	analyticsEndpoint := r.Iter8Config.Endpoint //r.GetAnalyticsService()
	analysis, err := Invoke(log, analyticsEndpoint, *instance, r.HTTP)
	log.Info("Invoke returned", "analysis", analysis)
	if err != nil {
		r.recordExperimentFailed(ctx, instance, v2beta1.ReasonAnalyticsServiceError, "Call to analytics engine failed")
		return r.failExperiment(ctx, instance, err)
	}

	// VALIDATE analysis object:
	// 1. has 4 entries: aggregatedMetrics, winnerAssessment, versionAssessments, weights
	// 2. versionAssessments have entry for each version, objective
	// 3. weights has entry for each version
	// If not valid: return r.failExperiment(context, instance)

	instance.Status.Analysis = analysis

	// // update weight distribution
	// if err := redistributeWeight(ctx, instance, r.RestConfig); err != nil {
	// 	r.recordExperimentFailed(ctx, instance, v2beta1.ReasonWeightRedistributionFailed, "Failure redistributing weights: %s", err.Error())
	// 	return r.failExperiment(ctx, instance, err)
	// }

	// // after weights have been redistributed, update Status.CurrentWeightDistribution
	// if err := updateObservedWeights(ctx, instance, r.RestConfig); err != nil {
	// 	r.recordExperimentFailed(ctx, instance, v2beta1.ReasonInvalidExperiment, "Specification of version weightObjectRef invalid: %s", err.Error())
	// 	return r.failExperiment(ctx, instance, nil)
	// }

	// END of loop processing: update counter and call loop handler

	// update completedLoops counter and udpate time
	loops := int(instance.Status.IncrementCompletedLoops())
	now := metav1.Now()
	instance.Status.LastUpdateTime = &now
	r.recordExperimentProgress(ctx, instance, v2beta1.ReasonLoopCompleted, "Completed Loop %d", instance.Status.GetCompletedLoops())

	// Call a loop handler if one is defined.
	// Note that we on the last loop, we will not execute this code; we called returned just above.
	if quit, result, err := r.launchHandlerWrapper(
		ctx, instance, HandlerTypeLoop, handlerLaunchModifier{loop: &loops}); quit {
		return result, err
	}

	// Not of loop or there is no loop handler --> schedule next loop
	return r.endRequest(ctx, instance, instance.Spec.GetIntervalAsDuration())
}
