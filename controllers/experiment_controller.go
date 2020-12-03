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

package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/configuration"
	"github.com/iter8-tools/etc3/util"
)

// experiment.controller.go - implements reconcile loop
//     - handles most of flow except for core of iterate loop which is in iterate.go

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Log            logr.Logger
	Scheme         *runtime.Scheme
	RestConfig     *rest.Config
	EventRecorder  record.EventRecorder
	Iter8Config    configuration.Iter8Config
	StatusModified bool
}

/* RBAC roles are handwritten in config/rbac-iter8 so that different roles can be assigned
//   to the controller and to the handlers
// +kubebuilder:rbac:groups=iter8.tools,resources=experiments,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=iter8.tools,resources=experiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=iter8.tools.resources=metrics,verbs=get;list;watch
*/

// Reconcile attempts to align the resource with the spec
func (r *ExperimentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("experiment", req.NamespacedName)
	ctx = context.WithValue(ctx, util.LoggerKey, log)

	log.Info("Reconcile called")
	defer log.Info("Reconcile completed")

	// Fetch instance on which started
	instance := &v2alpha1.Experiment{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		// if object not found, it has been deleted, we can ignore
		// (if it is being deleted and there is a finalizer, we would have found it)
		if errors.IsNotFound(err) {
			log.Info("Experiment not found.")
			return ctrl.Result{}, nil
		}
		// other error reading instance; return
		log.Error(err, "Unable to read experiment object.")
		return ctrl.Result{}, err // TODO
	}

	log.Info("found instance", "instance", instance, "updatedStatus", r.StatusModified) //, "spec", instance.Spec, "status", instance.Status)

	// ADD FINALIZER (check first...)
	// If instance does not have a finalizer, add one here (if desired)
	// IF DELETION, RUN FINALIZER and REMOVE FINALIZER
	// If instance deleted and have a finalizer, run it now

	// If instance has never been seen before, initialize status object
	if instance.Status.InitTime == nil {
		instance.InitializeStatus()
		if err := r.Status().Update(ctx, instance); err != nil {
			log.Error(err, "Failed to update Status after initialization.")
		}
		// r.recordExperimentInitialized(ctx, instance, "Experiment status initialized")
		r.recordExperimentProgress(ctx, instance,
			v2alpha1.ReasonExperimentInitialized, "Experiment status initialized")
		// r.recordEvent(ctx, instance,
		// 	v2alpha1.ExperimentConditionExperimentCompleted, v1.ConditionFalse,
		// 	v2alpha1.ReasonExperimentInitialized, "Experiment status initialized")
		return r.endRequest(ctx, instance)
	}
	log.Info("Status initialized")

	// If experiment already completed, stop
	if instance.Status.GetCondition(v2alpha1.ExperimentConditionExperimentCompleted).IsTrue() {
		log.Info("Experiment already completed.")
		return r.endExperiment(ctx, instance)
	}
	log.Info("Experiment active")

	// Check if we are in the process of terminating an experiment and take appropriate action:
	// If a terminal handler (finish, failure, or rollback) is running, just quit (wait until done)
	// If a terminal handler was running and is now completed (failed), endExperiment (failExperiment)
	// If there is no terminal handler or it has not been launched, proceed
	for _, handlerType := range []HandlerType{HandlerTypeFinish, HandlerTypeFailure, HandlerTypeRollback} {
		handler := r.GetHandler(instance, handlerType)
		switch hStatus := r.GetHandlerStatus(ctx, instance, handler); hStatus {
		case HandlerStatusRunning:
			return r.endRequest(ctx, instance)
		case HandlerStatusComplete:
			return r.endExperiment(ctx, instance)
		case HandlerStatusFailed:
			r.recordExperimentFailed(ctx, instance, v2alpha1.ReasonHandlerFailed, "%s handler '%s' failed", handlerType, *handler)
			// we don't call failExperiment here because we are already ending; we just end
			return r.endExperiment(ctx, instance)
		default: // case HandlerStatusNoHandler, HandlerStatusNotLaunched
			// do nothing
		}
	}
	log.Info("No terminal handlers running")

	// LATE INITIALIZATION of instance.Spec
	specChanged := r.LateInitialization(ctx, instance)
	if specChanged {
		log.Info("updating spec", "spec", instance.Spec)
		if err := r.Update(ctx, instance); err != nil && !validUpdateErr(err) {
			log.Error(err, "Failed to update Spec after late initialization.")
		}
		r.recordExperimentProgress(ctx, instance, v2alpha1.ReasonExperimentInitialized, "Late initialization complete")
		return r.endRequest(ctx, instance)
	}
	log.Info("Late initialization completed")

	// VALIDATE EXPERIMENT: basic validation of experiment object
	// See IsExperimentValid() for list of validations done
	if !r.IsExperimentValid(ctx, instance) {
		r.failExperiment(ctx, instance, nil)
	}

	// TARGET ACQUISITION
	// ensure that the target is not involved in another experiment
	// record experiment with annotation in target?
	// if !TargetAquired() {
	// 	if CanAquireTarget() {
	// 		AquireTarget()
	// 		   r.markExperimentProgress(ctx, instance, v2alpha1.ReasonTargetAcquired, "Target '%s' acquired", instance.Spec.Target)
	// 	}
	// 	r.endRequest()
	// }

	// RUN START HANDLER
	// Note: since we haven't already checked it may already have been started
	handler := r.GetHandler(instance, HandlerTypeStart)
	log.Info("Start handler", "handler", handler)
	switch r.GetHandlerStatus(ctx, instance, handler) {
	case HandlerStatusNotLaunched:
		if err := r.LaunchHandler(ctx, instance, *handler); err != nil {
			r.recordExperimentFailed(ctx, instance, v2alpha1.ReasonLaunchHandlerFailed, "failure launching %s handler '%s': %s", HandlerTypeStart, *handler, err.Error())
			return r.failExperiment(ctx, instance, err)
		}
		r.recordExperimentProgress(ctx, instance, v2alpha1.ReasonStartHandlerLaunched, "Start handler '%s' launched", *handler)
		return r.endRequest(ctx, instance)
	case HandlerStatusFailed:
		r.recordExperimentFailed(ctx, instance, v2alpha1.ReasonHandlerFailed, "%s handler '%s' failed", HandlerTypeStart, *handler)
		return r.failExperiment(ctx, instance, nil)
	case HandlerStatusRunning:
		return r.endRequest(ctx, instance)
	default: // case HandlerStatusNoHandler, HandlerStatusComplete:
		// do nothing; can proceed
	}
	log.Info("Start Handling Complete")

	// VERSION VALIDATION (versionInfo should be created by start handler)
	// See IsVersionInfoValid() for list of validations done
	if !r.IsVersionInfoValid(ctx, instance) {
		return r.failExperiment(ctx, instance, nil)
	}

	// If not set, set an initial status.recommendedBaseline
	instance.Status.SetRecommendedBaseline(instance.Spec.VersionInfo.Baseline.Name)

	// INITIAL WEIGHT DISTRIBUTION (FixedSplit only)
	// if instance.Spec.GetAlgorithm() == v2alpha1.AlgorithmTypeFixedSplit {
	// 	redistributeWeight (ctx, instance, instance.Spec.GetWeightDistribution())
	// }

	// EXECUTE ITERATION
	return r.doIteration(ctx, instance)
}

// SetupWithManager ..
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v2alpha1.Experiment{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

// LateInitialization initializes any fields in e.Spec not already set
// Returns true if a change was made, false if not
func (r *ExperimentReconciler) LateInitialization(ctx context.Context, instance *v2alpha1.Experiment) bool {
	changed := instance.Spec.InitializeHandlers(r.Iter8Config)
	changed = instance.Spec.InitializeWeights() || changed
	changed = instance.Spec.InitializeDuration() || changed
	changed = instance.Spec.InitializeCriteria(r.Iter8Config) || changed
	changed = r.ReadMetrics(ctx, instance) || changed

	return changed
}

// endRequest writes any changes (if needed) in preparation for ending processing of this reconcile request
func (r *ExperimentReconciler) endRequest(ctx context.Context, instance *v2alpha1.Experiment, interval ...time.Duration) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("endRequest called")
	defer log.Info("endRequest completed")

	r.updateIfNeeded(ctx, instance)

	if len(interval) > 0 {
		log.Info("Requeue for next iteration", "interval", interval, "iterations", instance.Status.GetCompletedIterations())
		return ctrl.Result{RequeueAfter: interval[0]}, nil
	}
	return ctrl.Result{}, nil
}

// endExperiment is called to mark an experiment as completed and triggers next experiment object
func (r *ExperimentReconciler) endExperiment(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("endExperiment called")
	defer log.Info("endExperiment completed")

	// trigger next experiment
	return r.endRequest(ctx, instance)
}

func (r *ExperimentReconciler) finishExperiment(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("finishExperiment called")
	defer log.Info("finishExperiment completed")

	result, err := r.terminate(ctx, instance, HandlerTypeFinish)
	r.recordExperimentCompleted(ctx, instance, "Experiment completed successfully")
	return result, err
}

func (r *ExperimentReconciler) rollbackExperiment(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("rollbackExperiment called")
	defer log.Info("rollbackExperiment ended")

	result, err := r.terminate(ctx, instance, HandlerTypeRollback)
	r.recordExperimentCompleted(ctx, instance, "Experiment rolled back")
	return result, err
}

func (r *ExperimentReconciler) failExperiment(ctx context.Context, instance *v2alpha1.Experiment, err error) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("failExperiment called")
	defer log.Info("failExperiment completed")

	if err != nil {
		log.Error(err, err.Error())
	}
	result, err := r.terminate(ctx, instance, HandlerTypeFailure)
	r.recordExperimentCompleted(ctx, instance, "Experiment failed")
	return result, err
}

// terminate calls the specified terminal handler (finish, rollback or fail) and ends the request
// Checking on completion of any terminal handlers takes place earlier, so can just launch
// If no handler exists, we end the experiment instead
func (r *ExperimentReconciler) terminate(ctx context.Context, instance *v2alpha1.Experiment, handlerType HandlerType) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("terminate called", "handlerType", handlerType)
	defer log.Info("terminate completed", "handlerType", handlerType)

	// run handler
	handler := r.GetHandler(instance, handlerType)
	if handler != nil {
		if err := r.LaunchHandler(ctx, instance, *handler); err != nil {
			r.recordExperimentFailed(ctx, instance, v2alpha1.ReasonLaunchHandlerFailed, "failure launching %s handler '%s': %s", handlerType, *handler, err.Error())
			if handlerType != HandlerTypeFailure {
				// can't call failExperiment if we are already in failExperiment
				return r.failExperiment(ctx, instance, err)
			}
			return r.endExperiment(ctx, instance)
		}
		r.recordExperimentProgress(ctx, instance, v2alpha1.ReasonTerminalHandlerLaunched, "%s handler '%s' launched", handlerType, *handler)
		return r.endRequest(ctx, instance)
	}
	return r.endExperiment(ctx, instance)
}

func validUpdateErr(err error) bool {
	if err == nil {
		return true
	}
	benignMsg := "the object has been modified"
	return strings.Contains(err.Error(), benignMsg)
}

func (r *ExperimentReconciler) updateIfNeeded(ctx context.Context, instance *v2alpha1.Experiment) error {
	log := util.Logger(ctx)
	if r.StatusModified {
		log.Info("updating status", "status", instance.Status)
		if err := r.Status().Update(ctx, instance); err != nil && !validUpdateErr(err) {
			log.Error(err, "Failed to update status")
			return err
		}
		r.StatusModified = false
	}

	return nil
}
