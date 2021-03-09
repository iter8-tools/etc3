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

	"github.com/iter8-tools/etc3/analytics"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func moreIterationsNeeded(instance *v2alpha2.Experiment) bool {
	// Are there more iterations to execute
	return *instance.Status.CompletedIterations < instance.Spec.GetIterationsPerLoop()*instance.Spec.GetMaxLoops()
}

// determine if a loop is completed by determing if the number of iterations executed
// is a multiple of duration.iterationsPerLoop
func completedLoop(instance *v2alpha2.Experiment) (int, bool) {
	if 0 == instance.Status.GetCompletedIterations()%instance.Spec.GetIterationsPerLoop() {
		return int(instance.Status.GetCompletedIterations() / instance.Spec.GetIterationsPerLoop()), true
	}
	return -1, false
}

func (r *ExperimentReconciler) sufficientTimePassedSincePreviousIteration(ctx context.Context, instance *v2alpha2.Experiment) bool {
	log := util.Logger(ctx)

	// Is this the first iteration or has enough time passed since last iteration?
	if *instance.Status.CompletedIterations == 0 || instance.Status.LastUpdateTime == nil {
		return true
	}

	now := time.Now()
	interval := instance.Spec.GetIntervalAsDuration()
	expectedTime := instance.Status.LastUpdateTime.Add(interval)
	log.Info("sufficientTimePassedSincePreviousIteration", "lastUpdateTime", instance.Status.LastUpdateTime, "interval", interval, "sum", expectedTime, "now", now)

	if now.Before(expectedTime) {
		// is it close enough?
		difference := expectedTime.Sub(now)
		return difference < 100*time.Millisecond
	}
	// now is after expectedTime
	return true
}

func (r *ExperimentReconciler) doIteration(ctx context.Context, instance *v2alpha2.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("doIteration called")
	defer log.Info("doIteration completed")

	// record start time of experiment if not already set
	if err := r.setStartTimeIfNotSet(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	if !r.sufficientTimePassedSincePreviousIteration(ctx, instance) {
		// not enough time has passed since the last iteration, wait
		return ctrl.Result{}, errors.New("Insufficient time has passed since previous iteration")
	}

	// TODO  GET CURRENT WEIGHTS (from cluster)

	analyticsEndpoint := r.Iter8Config.Endpoint //r.GetAnalyticsService()
	analysis, err := analytics.Invoke(log, analyticsEndpoint, *instance, r.HTTP)
	log.Info("Invoke returned", "analysis", analysis)
	if err != nil {
		r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonAnalyticsServiceError, "Unable to contact analytics engine %s", analyticsEndpoint)
		return r.failExperiment(ctx, instance, err)
	}

	// VALIDATE analysis object:
	// 1. has 4 entries: aggregatedMetrics, winnerAssessment, versionAssessments, weights
	// 2. versionAssessments have entry for each version, objective
	// 3. weights has entry for each version
	// If not valid: return r.failExperiment(context, instance)

	// update analytics in instance.status
	instance.Status.Analysis = analysis

	// Handle failure of objective (possibly rollback)
	if r.mustRollback(ctx, instance) {
		return r.rollbackExperiment(ctx, instance)
	}

	// update weight distribution
	if err := redistributeWeight(ctx, instance, r.RestConfig); err != nil {
		r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonWeightRedistributionFailed, "Failure redistributing weights: %s", err.Error())
		return r.failExperiment(ctx, instance, err)
	}

	// after weights have been redistributed, update Status.CurrentWeightDistribution
	updateObservedWeights(ctx, instance, r.RestConfig)

	// update status.versionRecommendedForPromotion if a new winner identified
	instance.Status.SetVersionRecommendedForPromotion(instance.Spec.VersionInfo.Baseline.Name)

	// update completedIterations counter and record completion
	r.completeIteration(ctx, instance)
	r.recordExperimentProgress(ctx, instance, v2alpha2.ReasonIterationCompleted, "Completed Iteration %d", *instance.Status.CompletedIterations)

	// if there are no more iterations to execute, finishExperiment
	// otherwise, just endRequest (and requeue for later)
	if !moreIterationsNeeded(instance) {
		return r.finishExperiment(ctx, instance)
	}

	// if we are at the end of a loop (we've executed Duration.IterationsPerLoop iterations)
	// then call a loop handler if one is defined.
	// Note that we on the last loop, we will not execute this code; we called returned just above.
	if loop, ok := completedLoop(instance); ok {
		r.recordExperimentProgress(ctx, instance, v2alpha2.ReasonIterationCompleted, "Completed Loop %d", loop)

		if quit, result, err := r.launchHandlerWrapper(
			ctx, instance, HandlerTypeLoop, handlerLaunchModifier{loop: &loop}); quit {
			return result, err
		}
	}

	// Not of loop or there is no loop handler --> schedule next iteration
	return r.endRequest(ctx, instance, instance.Spec.GetIntervalAsDuration())
}

func (r *ExperimentReconciler) setStartTimeIfNotSet(ctx context.Context, instance *v2alpha2.Experiment) error {
	if instance.Status.StartTime == nil {
		now := metav1.Now()
		instance.Status.StartTime = &now

		if err := r.Status().Update(ctx, instance); err != nil {
			util.Logger(ctx).Info("Failed to update when initializing status: " + err.Error())
			return err
		}
	}
	return nil
}

func hasCriteria(instance *v2alpha2.Experiment) bool {
	return instance.Spec.Criteria != nil
}

func (r *ExperimentReconciler) completeIteration(ctx context.Context, instance *v2alpha2.Experiment) {
	// update completedIterations counter
	*instance.Status.CompletedIterations++
	now := metav1.Now()
	instance.Status.LastUpdateTime = &now
}

// mustRollback determines if the experiment should be rolled back.
func (r *ExperimentReconciler) mustRollback(ctx context.Context, instance *v2alpha2.Experiment) bool {
	return len(r.versionsMustRollback(ctx, instance)) > 0
}

// versionsMustRollback identifies any versions that to be rollbacked:
//   - Is there an objective for which rollbackOnFailure set to true AND
//   -    there is a version for which this objective is failing
// Returns list of versions that failed an objective
// Assumes the analysis has already been added to status.analysis (avoids another input parameter)
func (r *ExperimentReconciler) versionsMustRollback(ctx context.Context, instance *v2alpha2.Experiment) []string {
	log := util.Logger(ctx)
	log.Info("mustRollbackVersions() called")
	defer log.Info("mustRollbackVersions() ended")

	deploymentPattern := instance.Spec.GetDeploymentPattern()
	failedVersions := make([]string, 0)
	if instance.Spec.Criteria == nil {
		// there are no criteria
		return failedVersions
	}
	for index, o := range instance.Spec.Criteria.Objectives {
		if o.GetRollbackOnFailure(deploymentPattern) {
			// need to rollback on failure; did some version fail for this objective?
			for version, satisfiesObjectives := range instance.Status.Analysis.VersionAssessments.Data {
				if !satisfiesObjectives[index] {
					failedVersions = append(failedVersions, version)
				}
			}
		}
	}
	return failedVersions
}
