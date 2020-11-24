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
	"github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ExperimentReconciler) moreIterationsNeeded(instance *v2alpha1.Experiment) bool {
	// Are there more iterations to execute
	return *instance.Status.CompletedIterations < instance.Spec.GetMaxIterations()
}

func (r *ExperimentReconciler) sufficientTimePassedSincePreviousIteration(ctx context.Context, instance *v2alpha1.Experiment) bool {
	log := util.Logger(ctx)

	// Is this the first iteration or has enough time passed since last iteration?
	if *instance.Status.CompletedIterations == 0 {
		return true
	}
	now := time.Now()
	interval := instance.Spec.GetIntervalAsDuration()
	log.Info("sufficientTimePassedSincePreviousIteration", "lastUpdateTime", instance.Status.LastUpdateTime, "interval", interval, "sum", instance.Status.LastUpdateTime.Add(interval), "now", now)
	return instance.Status.LastUpdateTime == nil || now.After(instance.Status.LastUpdateTime.Add(interval))
}

func (r *ExperimentReconciler) doIteration(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("doIterate() called")
	defer log.Info("doIterate() completed")

	// record start time of experiment if not already set
	if err := r.setStartTimeIfNotSet(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	analyticsEndpoint := util.GetAnalyticsService()
	analysis, err := analytics.Invoke(log, analyticsEndpoint, *instance)
	log.Info("Invoke returned", "analysis", analysis)
	if err != nil {
		r.markAnalyticsServiceError(ctx, instance, "Unable to contact analytics engine %s", analyticsEndpoint)
		return r.failExperiment(ctx, instance)
	}

	// VALIDATE analysis object:
	//   - has 4 entries: aggregatedMetrics, winnerAssessment, versionAssessments, weights
	//   - versionAssessments have entry for each version, objective
	//   - weights has entry for each version
	// If not valid: failExperiment(context, instance)

	// update analytics in instance.status
	instance.Status.Analysis = analysis
	r.StatusModified = true
	log.Info("Updated status with analysis", "status.Analysis", instance.Status.Analysis)
	log.Info("Updated status with analysis", "status", instance.Status)

	// TODO -- encapsulate in rollbackExperiment()
	// Handle failure of objective (possibly rollback)
	if r.mustRollback(ctx, instance) {
		// do we consider this a completed iteration?
		if instance.HasRollbackHandler() {
			r.startRollbackHandler(ctx, instance)

			return r.endIteration(ctx, instance)
		}
		// We don't need to check if any handlers are running
		// Recall that we do this at the start of the reconciler
		return r.finishExperiment(ctx, instance)
	}

	// update weight distribution
	// if we failed some versions, how do we distribute weight?
	err = r.redistributeWeight(ctx, instance)
	if err != nil {
		return r.failExperiment(ctx, instance)
	}

	// update completedIterations counter
	*instance.Status.CompletedIterations++
	r.StatusModified = true

	// if there are no more iterations to execute, run finish handler(s) if present
	if !r.moreIterationsNeeded(instance) {
		return r.finishExperiment(ctx, instance)
	}

	return r.endIteration(ctx, instance)
}

func (r *ExperimentReconciler) setStartTimeIfNotSet(ctx context.Context, instance *v2alpha1.Experiment) error {
	if instance.Status.StartTime == nil {
		now := metav1.Now()
		instance.Status.StartTime = &now

		if err := r.Status().Update(ctx, instance); err != nil {
			util.Logger(ctx).Error(err, "Failed to update when initializing status")
			return err
		}
		r.StatusModified = false
	}
	return nil
}

func hasCriteria(instance *v2alpha1.Experiment) bool {
	return instance.Spec.Criteria != nil
}

func (r *ExperimentReconciler) redistributeWeight(ctx context.Context, instance *v2alpha1.Experiment) error {
	log := util.Logger(ctx)
	log.Info("redistributeWeight() called")
	defer log.Info("redistributeWeight() ended")

	if versionInfo := instance.Spec.VersionInfo; versionInfo == nil {
		// should never get here; should have been validated before this
		log.Info("Cannot redistribute weight; no version information present.")
		return nil
	}

	// TODO collect multiple changes for each object
	if err := r.updateWeightForVersion(ctx, instance, instance.Spec.VersionInfo.Baseline); err != nil {
		return err
	}
	for _, version := range instance.Spec.VersionInfo.Candidates {
		if err := r.updateWeightForVersion(ctx, instance, version); err != nil {
			return err
		}
	}

	// set status.currentWeightDistribution to match set weights
	// for now copy from status.analysis.weights
	instance.Status.CurrentWeightDistribution = make([]v2alpha1.WeightData, len(instance.Status.Analysis.Weights.Data))
	for i, w := range instance.Status.Analysis.Weights.Data {
		instance.Status.CurrentWeightDistribution[i] = w
	}
	return nil
}

func (r *ExperimentReconciler) updateWeightForVersion(ctx context.Context, instance *v2alpha1.Experiment, version v2alpha1.VersionDetail) error {
	log := util.Logger(ctx)

	// n-1 versions should have a weightObjRef.fieldPath; should be caught by validation
	if version.WeightObjRef == nil {
		log.Info("Unable to update weight; no weightObjectReference", "version", version)
		return nil
	}
	if version.WeightObjRef.FieldPath == "" {
		log.Info("Unable to update weight; no field specified", "version", version)
		return nil
	}

	weight := getWeightRecommendation(version.Name, instance.Status.Analysis.Weights.Data)
	if weight == nil {
		log.Info("Unable to find weight recommendation.", "version", version, "weights", instance.Status.Analysis.Weights)
		// fatal error; expected a weight recommendation for all versions
		return errors.New("Unable to find weight recommendation")
	}
	_, err := r.patchWeight(ctx, version.WeightObjRef, *weight)
	if err != nil {
		log.Error(err, "Failed to update weight", "version", version)
		return err
	}

	return nil
}

func getWeightRecommendation(version string, weights []v2alpha1.WeightData) *int32 {
	for _, w := range weights {
		if w.Name == version {
			weight := w.Value
			return &weight
		}
	}
	return nil
}

func (r *ExperimentReconciler) endIteration(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("endIteration() called", "instance", instance)

	// update lastUpdateTime (any anything else that might have changed)
	now := metav1.Now()
	instance.Status.LastUpdateTime = &now
	r.StatusModified = true
	r.updateIfNeeded(ctx, instance)

	interval := instance.Spec.GetIntervalAsDuration()
	log.Info("Requeue for next iteration", "interval", interval, "iterations", instance.Status.GetCompletedIterations())
	return ctrl.Result{RequeueAfter: interval}, nil
}

// mustRollback determines if the experiment should be rolled back.
func (r *ExperimentReconciler) mustRollback(ctx context.Context, instance *v2alpha1.Experiment) bool {
	return len(r.versionsMustRollback(ctx, instance)) > 0
}

// versionsMustRollback identifies any versions that to be rollbacked:
//   - Is there an objective for which rollbackOnFailure set to true AND
//   -    there is a version for which this objective is failing
// Returns list of versions that failed an objective
// Assumes the analysis has already been added to status.analysis (avoids another input parameter)
func (r *ExperimentReconciler) versionsMustRollback(ctx context.Context, instance *v2alpha1.Experiment) []string {
	log := util.Logger(ctx)
	log.Info("mustRollbackVersions() called")
	defer log.Info("mustRollbackVersions() ended")

	strategy := instance.Spec.Strategy.Type
	failedVersions := make([]string, 0)
	for index, o := range instance.Spec.Criteria.Objectives {
		if o.GetRollbackOnFailure(strategy) {
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

func (r *ExperimentReconciler) rollbackExperiment(ctx context.Context, instance *v2alpha1.Experiment, failedVersions []string) {
	log := util.Logger(ctx)
	log.Info("rollbackExperiment() called")
	defer log.Info("rollbackExperiment() ended")
}

func (r *ExperimentReconciler) startRollbackHandler(ctx context.Context, instance *v2alpha1.Experiment) error {
	log := util.Logger(ctx)
	log.Info("startRollbackHandler() called")
	defer log.Info("startRollbackHandler() ended")
	return nil
}
