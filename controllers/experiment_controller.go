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
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/iter8-tools/etc3/analytics"
	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/configuration"
	"github.com/iter8-tools/etc3/util"
)

// experiment.controller.go - implements reconcile loop
//     - handles most of flow except for core of iterate loop which is in iterate.go

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	RestConfig    *rest.Config
	EventRecorder record.EventRecorder
	Iter8Config   configuration.Iter8Config
	HTTP          analytics.HTTP
	ReleaseEvents chan event.GenericEvent
}

const (
	iter8FinalizerName = "experiments.iter8.tools.finalizer"
)

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
			log.Info("Experiment not found")
			return ctrl.Result{}, nil
		}
		// other error reading instance; return
		log.Error(err, "Unable to read experiment object")
		return ctrl.Result{}, nil
	}

	log.Info("Reconcile", "instance", instance)
	ctx = context.WithValue(ctx, util.OriginalStatusKey, instance.Status.DeepCopy())

	// Add FINALIZER if not present; run finalizer if deleting experiment
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The experiment is not being deleted, so if it doesn't have a finalizer we add one
		// and return; update will retrigger reconcile
		if !containsString(instance.ObjectMeta.Finalizers, iter8FinalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, iter8FinalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// the experiment is being deleted, look for a finalizer and run it
		if containsString(instance.ObjectMeta.Finalizers, iter8FinalizerName) {
			if err := r.finalizeExperiment(ctx, instance); err != nil {
				// if failed, return error so can retry
				return ctrl.Result{}, err
			}

			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, iter8FinalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
			// on success, remove finalizer so that deletion can proceed
			return ctrl.Result{}, nil
		}
		// is being deleted and there was no finalizer; just exit
		return ctrl.Result{}, nil
	}

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
		return r.endRequest(ctx, instance)
	}
	log.Info("Experiment is active")

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
			return r.endExperiment(ctx, instance, "Experiment completed successfully")
		case HandlerStatusFailed:
			r.recordExperimentFailed(ctx, instance, v2alpha1.ReasonHandlerFailed, "%s handler '%s' failed", handlerType, *handler)
			// we don't call failExperiment here because we are already ending; we just end
			return r.endExperiment(ctx, instance, "Experiment failed")
		default: // case HandlerStatusNoHandler, HandlerStatusNotLaunched
			// do nothing
		}
	}

	// LATE INITIALIZATION of instance.Spec
	// TODO move to mutating webhook
	originalSpec := instance.Spec.DeepCopy()
	if ok := r.LateInitialization(ctx, instance); !ok {
		return r.failExperiment(ctx, instance, nil)
	}

	if !reflect.DeepEqual(originalSpec, &instance.Spec) {
		if err := r.Update(ctx, instance); err != nil && !validUpdateErr(err) {
			log.Error(err, "Failed to update Spec after late initialization.")
		}
		r.recordExperimentProgress(ctx, instance, v2alpha1.ReasonExperimentInitialized, "Late initialization complete")
		return r.endRequest(ctx, instance)
	}
	log.Info("Late initialization completed")

	// VALIDATE EXPERIMENT: basic validation of experiment object
	// See IsExperimentValid() for list of validations done
	// TODO move to validating web hook
	if !r.IsExperimentValid(ctx, instance) {
		return r.failExperiment(ctx, instance, nil)
	}

	// TARGET ACQUISITION
	// Ensure that we are the only experiment proceding with the same target
	// If we find another, end request and wait to be triggered again
	if !r.acquireTarget(ctx, instance) {
		// do not have the target, quit
		return r.endRequest(ctx, instance)
	}

	// advance stage from Waiting to Initializing
	// when we advance for the first time, we exit to force update; will be retriggered
	if ok := r.advanceStage(ctx, instance, v2alpha1.ExperimentStageInitializing); ok {
		log.Info("Reconcile ending after advance to Initializing")
		return r.endRequest(ctx, instance)
	}

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

	// advance stage from Initializing to Running
	// when we advance for the first time, we exit to force update; will be retriggered
	if ok := r.advanceStage(ctx, instance, v2alpha1.ExperimentStageRunning); ok {
		log.Info("Reconcile ending after advance to Running")
		return r.endRequest(ctx, instance)
	}

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

	jobPredicateFuncs := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			namespace := e.MetaNew.GetNamespace()
			return namespace == r.Iter8Config.Namespace
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	jobToExperiment := handler.ToRequestsFunc(
		func(a handler.MapObject) []ctrl.Request {
			lbls := a.Meta.GetLabels()
			experimentName, ok := lbls["iter8/experimentName"]
			if !ok {
				return nil
			}
			experimentNamespace, ok := lbls["iter8/experimentNamespace"]
			if !ok {
				return nil
			}
			return []ctrl.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      experimentName,
						Namespace: experimentNamespace,
					},
				},
			}
		},
	)

	return ctrl.NewControllerManagedBy(mgr).
		For(&v2alpha1.Experiment{}).
		Watches(&source.Kind{Type: &batchv1.Job{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: jobToExperiment},
			builder.WithPredicates(jobPredicateFuncs)).
		Watches(&source.Channel{Source: r.ReleaseEvents}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

// Helper functions for FINALIZERS

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// Helper function for LATE INITIALIZATION

// LateInitialization initializes any fields in e.Spec not already set
// Returns false if something went wrong
func (r *ExperimentReconciler) LateInitialization(ctx context.Context, instance *v2alpha1.Experiment) bool {
	instance.Spec.InitializeSpec(r.Iter8Config)
	return r.ReadMetrics(ctx, instance)
}

// Helper functions for maintaining stages
func (r *ExperimentReconciler) advanceStage(ctx context.Context, instance *v2alpha1.Experiment, to v2alpha1.ExperimentStageType) bool {
	log := util.Logger(ctx)
	log.Info("advanceStage called", "current stage", *instance.Status.Stage, "to", to)
	defer log.Info("advanceStage completed")

	stage := *instance.Status.Stage
	if to.After(stage) {
		stage = to
		instance.Status.Stage = &stage
		log.Info("advanceStage advanced", "to", to)
		return true
	}
	return false
}

// Helper functions for TERMINATION

// endRequest writes any changes (if needed) in preparation for ending processing of this reconcile request
func (r *ExperimentReconciler) endRequest(ctx context.Context, instance *v2alpha1.Experiment, interval ...time.Duration) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("endRequest called")
	defer log.Info("endRequest completed")

	err := r.updateStatus(ctx, instance)

	if len(interval) > 0 {
		log.Info("Requeue for next iteration", "interval", interval, "iterations", instance.Status.GetCompletedIterations())
		return ctrl.Result{RequeueAfter: interval[0]}, err
	}
	return ctrl.Result{}, err
}

// endExperiment is called to mark an experiment as completed and triggers next experiment object
func (r *ExperimentReconciler) endExperiment(ctx context.Context, instance *v2alpha1.Experiment, msg string) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("endExperiment called")
	defer log.Info("endExperiment completed")

	// advance stage from Finishing to Completed
	// when we do so for the first time, record the completion event and trigger the next experiment
	if ok := r.advanceStage(ctx, instance, v2alpha1.ExperimentStageCompleted); ok {
		r.recordExperimentCompleted(ctx, instance, msg)
		r.updateStatus(ctx, instance)
		r.triggerNextExperiment(ctx, instance)
	}

	return r.endRequest(ctx, instance)
}

func (r *ExperimentReconciler) finishExperiment(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("finishExperiment called")
	defer log.Info("finishExperiment completed")

	if launched := r.launchTerminalHandler(ctx, instance, HandlerTypeFinish); launched {
		// a handler was launched so we end request; Reconcile() will be triggered again when it finishes/fails
		return r.endRequest(ctx, instance)
	}
	// no handler was launched; we don't expect Reconcile() to be triggered again so we take final actions
	return r.endExperiment(ctx, instance, "Experiment completed successfully")
}

func (r *ExperimentReconciler) rollbackExperiment(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("rollbackExperiment called")
	defer log.Info("rollbackExperiment ended")

	if launched := r.launchTerminalHandler(ctx, instance, HandlerTypeRollback); launched {
		// a handler was launched so we end request; Reconcile() will be triggered again when it finishes/fails
		return r.endRequest(ctx, instance)
	}
	// no handler was launched; we don't expect Reconcile() to be triggered again so we take final actions
	return r.endExperiment(ctx, instance, "Experiment rolled back")
}

func (r *ExperimentReconciler) failExperiment(ctx context.Context, instance *v2alpha1.Experiment, err error) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("failExperiment called")
	defer log.Info("failExperiment completed")

	if err != nil {
		log.Error(err, err.Error())
	}
	if launched := r.launchTerminalHandler(ctx, instance, HandlerTypeFailure); launched {
		// a handler was launched so we end request; Reconcile() will be triggered again when it finishes/fails
		return r.endRequest(ctx, instance)
	}
	// no handler was launched; we don't expect Reconcile() to be triggered again so we take final actions
	return r.endExperiment(ctx, instance, "Experiment failed")
}

// launchTerminalHandler calls the specified terminal handler (finish, rollback or fail) and ends the request
// Checking on completion of any terminal handlers takes place earlier, so can just launch
// If no handler exists, we end the experiment instead
func (r *ExperimentReconciler) launchTerminalHandler(ctx context.Context, instance *v2alpha1.Experiment, handlerType HandlerType) bool {
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
				return r.launchTerminalHandler(ctx, instance, HandlerTypeFailure)
			}
			// we did not successfully launch a failure handler
			return false
		}
		// advance stage from Running to Finishing
		// we will ever get called once
		log.Info("launchTerminalHandler ending after advance to Finishing", "handler", *handler)
		r.advanceStage(ctx, instance, v2alpha1.ExperimentStageFinishing)

		r.recordExperimentProgress(ctx, instance, v2alpha1.ReasonTerminalHandlerLaunched, "%s handler '%s' launched", handlerType, *handler)
		return true
	}
	// there was no handler, so none launched
	return false
}

func validUpdateErr(err error) bool {
	if err == nil {
		return true
	}
	benignMsg := "the object has been modified"
	return strings.Contains(err.Error(), benignMsg)
}

func (r *ExperimentReconciler) updateStatus(ctx context.Context, instance *v2alpha1.Experiment) error {
	log := util.Logger(ctx)
	originalStatus := util.OriginalStatus(ctx)

	// log.Info("updateStatus", "original status", *originalStatus)
	log.Info("updateStatus", "status", instance.Status)
	if !reflect.DeepEqual(originalStatus, &instance.Status) {
		if err := r.Status().Update(ctx, instance); err != nil && !validUpdateErr(err) {
			log.Error(err, "Failed to update status")
			return err
		}
	}
	return nil
}

func (r *ExperimentReconciler) finalizeExperiment(ctx context.Context, instance *v2alpha1.Experiment) error {
	log := util.Logger(ctx)
	log.Info("finalizeExperiment called")
	defer log.Info("finalizeExperiment completed")

	// The experiment finalizer does the following:
	//     1. Delete any handler jobs
	//     2. Trigger any waiting experiments

	//     1. Delete any handler jobs
	for _, handlerType := range []HandlerType{HandlerTypeStart, HandlerTypeFinish, HandlerTypeFailure, HandlerTypeRollback} {
		handler := r.GetHandler(instance, handlerType)
		if handler != nil {
			log.Info("finalizeExperiment deleting job", "handler", handler)
			if err := r.deleteHandlerJob(ctx, instance, handler); err != nil {
				return err
			}
		}
	}

	//     2. Trigger any waiting experiments
	// endExperiment() triggers any waiting experiment
	log.Info("finalizeExperiment triggering next experiment")
	// to avoid a possible race condition, we mark the experiment completed
	// and update its status before triggering the next experiment
	r.recordExperimentCompleted(ctx, instance, "Experiment deleted")
	r.updateStatus(ctx, instance)
	r.triggerNextExperiment(ctx, instance)

	return nil
}
