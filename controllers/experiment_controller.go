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
	"fmt"
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

	v2beta1 "github.com/iter8-tools/etc3/api/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Iter8Config   Iter8Config
	HTTP          HTTPTransport
	ReleaseEvents chan event.GenericEvent
	JobManager    JobManager
}

/* RBAC roles are handwritten in config/rbac-iter8 so that different roles can be assigned
//   to the controller and to the handlers
// +kubebuilder:rbac:groups=iter8.tools,resources=experiments,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=iter8.tools,resources=experiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=iter8.tools.resources=metrics,verbs=get;list;watch
*/

// Reconcile attempts to align the resource with the spec
func (r *ExperimentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("experiment", req.NamespacedName)
	ctx = context.WithValue(ctx, LoggerKey, log)

	log.Info("Reconcile called")
	defer log.Info("Reconcile completed")

	// Fetch instance on which started
	instance := &v2beta1.Experiment{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		// if object not found, it has been deleted
		if errors.IsNotFound(err) {
			log.Info("Experiment not found")
			// we make sure to have deleted all jobs and trigger any waiting experiment
			r.cleanupDeletedExperiments(ctx, instance)
			return ctrl.Result{}, nil
		}
		// other error reading instance; return
		log.Error(err, "Unable to read experiment object")
		return ctrl.Result{}, nil
	}

	log.Info("Reconcile", "instance", instance)
	ctx = context.WithValue(ctx, OriginalStatusKey, instance.Status.DeepCopy())

	// If instance has never been seen before, initialize status object
	if instance.Status.StartTime == nil {
		instance.InitializeStatus()
		if err := r.Status().Update(ctx, instance); err != nil {
			log.Error(err, "Failed to update Status after initialization.")
		}
		r.recordExperimentProgress(ctx, instance,
			v2beta1.ReasonExperimentInitialized, "Experiment status initialized")
		return r.endRequest(ctx, instance)
	}
	log.Info("Status initialized")

	// If experiment already completed, stop
	if instance.Status.GetCondition(v2beta1.ExperimentConditionExperimentCompleted).IsTrue() {
		log.Info("Experiment already completed.")
		return r.endRequest(ctx, instance)
	}
	log.Info("Experiment is active")

	// Check status of all handlers associated with this experiment.
	// If any have been launched we should wait for them to complete before continuing.
	// If they have completed (or failed), we should take appropriate action.
	// The checkHandlers() method is type aware; it tells us (via the first return value)
	// whether to stop or continue. If we stop, we return the second and third return values.
	if stop, result, err := r.checkHandlersStatus(ctx, instance, allHandlerTypes); stop {
		return result, err
	}

	// LATE INITIALIZATION of instance.Spec
	// Note: we don't actually modify instance; this is in memory only
	instance.Spec.InitializeSpec()

	// VALIDATE EXPERIMENT: basic validation of experiment object
	// See IsExperimentValid() for list of validations done
	// TODO move to validating web hook
	if !r.IsExperimentValid(ctx, instance) {
		return r.failExperiment(ctx, instance, nil)
	}

	// advance stage from Waiting to Initializing
	// when we advance for the first time, we exit to force update; will be retriggered
	if ok := r.advanceStage(ctx, instance, v2beta1.ExperimentStageInitializing); ok {
		log.Info("Update stage advance to: Initializing")
		return r.endRequest(ctx, instance)
	}

	// RUN START HANDLER if necessary
	// Note: We checked above if the start handler was running.
	// If we get here it either hasn't been launched or it has already completed.
	// We get here many times, but we want to execute the start handler only once.
	// Use a prerequisite checker to check that it has never been launched before.
	if stop, result, err := r.launchHandlerWrapper(ctx, instance, HandlerTypeStart,
		handlerLaunchModifier{prerequisiteCheck: func() bool {
			return HandlerStatusNotLaunched == r.GetHandlerStatus(ctx, instance, r.GetHandler(instance, HandlerTypeStart), nil)
		}}); stop {
		return result, err
	}
	log.Info("Start Handling Complete")

	// using spec.criteria, read the metrics objects into status.metrics
	if shouldReadMetrics(instance) {
		if ok := r.ReadMetrics(ctx, instance); !ok {
			return r.failExperiment(ctx, instance, nil)
		}
		// we updated status if we read metrics
		// endRequest writes the change the to cluster and retriggers reconcile
		return r.endRequest(ctx, instance)
	}

	// advance stage from Initializing to Running
	// when we advance for the first time, we've just finished the start handler (if there is one),
	// so we update Status.CurrentWeightDistribution
	// when we advance for the first time, we exit to force update; will be retriggered
	if ok := r.advanceStage(ctx, instance, v2beta1.ExperimentStageRunning); ok {
		log.Info("Updating stage advance to: Running")
		if err := updateObservedWeights(ctx, instance, r.RestConfig); err != nil {
			r.recordExperimentFailed(ctx, instance, v2beta1.ReasonInvalidExperiment, "Specification of version weightObjectRef invalid: %s", err.Error())
			return r.failExperiment(ctx, instance, nil)
		}
		return r.endRequest(ctx, instance)
	}

	// VERSION VALIDATION (versionInfo should be created by start handler)
	// See IsVersionInfoValid() for list of validations done
	if !r.IsVersionInfoValid(ctx, instance) {
		return r.failExperiment(ctx, instance, nil)
	}

	// If not set, set an initial status.versionRecommendedForPromotion
	instance.Status.SetVersionRecommendedForPromotion(instance.Spec.Versions[0])

	// INITIAL WEIGHT DISTRIBUTION (FixedSplit only)
	// if instance.Spec.GetAlgorithm() == v2beta1.AlgorithmTypeFixedSplit {
	// 	redistributeWeight (ctx, instance, instance.Spec.GetWeightDistribution())
	// }

	// EXECUTE ITERATION
	return r.doIteration(ctx, instance)
}

// SetupWithManager is the method called when setting up the experiment reconciler with the controller manager.
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {

	jobPredicateFuncs := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			namespace := e.ObjectNew.GetNamespace()
			return namespace == r.Iter8Config.Namespace
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	// Enque requests from map func, which we will use for mapping jobs to reconcile requests, requires a func with this signature
	jobToExperiment := func(a client.Object) []ctrl.Request {
		// jobToExperiment := handler.ToRequestsFunc(
		// 	func(a handler.MapObject) []ctrl.Request {
		lbls := a.GetLabels()
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
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v2beta1.Experiment{}).
		Watches(&source.Kind{Type: &batchv1.Job{}},
			handler.EnqueueRequestsFromMapFunc(jobToExperiment),
			builder.WithPredicates(jobPredicateFuncs)).
		Watches(&source.Channel{Source: r.ReleaseEvents}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

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

// Helper functions for maintaining stages

func (r *ExperimentReconciler) advanceStage(ctx context.Context, instance *v2beta1.Experiment, to v2beta1.ExperimentStageType) bool {
	log := Logger(ctx)
	log.Info("advanceStage called", "current stage", *instance.Status.Stage, "to", to)
	defer log.Info("advanceStage completed")

	stage := *instance.Status.Stage
	if to.After(stage) {
		stage = to
		instance.Status.Stage = &stage
		r.recordExperimentProgress(ctx, instance, v2beta1.ReasonStageAdvanced, "Advanced to %s", to)
		return true
	}
	return false
}

// Helper functions for TERMINATION

// endRequest writes any changes (if needed) in preparation for ending processing of this reconcile request
func (r *ExperimentReconciler) endRequest(ctx context.Context, instance *v2beta1.Experiment, interval ...time.Duration) (ctrl.Result, error) {
	log := Logger(ctx)
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
func (r *ExperimentReconciler) endExperiment(ctx context.Context, instance *v2beta1.Experiment, msg string) (ctrl.Result, error) {
	log := Logger(ctx)
	log.Info("endExperiment called")
	defer log.Info("endExperiment completed")

	// advance stage from Finishing to Completed
	// when we advance to Completed for the first time, any terminal handler has completed. We update
	// Status.CurrentWeightDistribution to reflect any possible change to distributiom.
	// when we do so for the first time, record the completion event and trigger the next experiment
	if ok := r.advanceStage(ctx, instance, v2beta1.ExperimentStageCompleted); ok {
		log.Info("Updating stage advance to: Completed")
		r.recordExperimentCompleted(ctx, instance, msg)
		r.updateStatus(ctx, instance)
	}

	return r.endRequest(ctx, instance)
}

func (r *ExperimentReconciler) finishExperiment(ctx context.Context, instance *v2beta1.Experiment) (ctrl.Result, error) {
	log := Logger(ctx)
	log.Info("finishExperiment called")
	defer log.Info("finishExperiment completed")

	if stop, result, err := r.launchHandlerWrapper(ctx, instance, HandlerTypeFinish,
		handlerLaunchModifier{onSuccessfulLaunch: func() { r.advanceStage(ctx, instance, v2beta1.ExperimentStageFinishing) }},
	); stop {
		return result, err
	}

	return r.endExperiment(ctx, instance, "Experiment completed successfully")
}

func (r *ExperimentReconciler) rollbackExperiment(ctx context.Context, instance *v2beta1.Experiment) (ctrl.Result, error) {
	log := Logger(ctx)
	log.Info("rollbackExperiment called")
	defer log.Info("rollbackExperiment ended")

	if stop, result, err := r.launchHandlerWrapper(ctx, instance, HandlerTypeRollback,
		handlerLaunchModifier{onSuccessfulLaunch: func() { r.advanceStage(ctx, instance, v2beta1.ExperimentStageFinishing) }},
	); stop {
		return result, err
	}

	return r.endExperiment(ctx, instance, "Experiment rolled back")
}

func (r *ExperimentReconciler) failExperiment(ctx context.Context, instance *v2beta1.Experiment, err error) (ctrl.Result, error) {
	log := Logger(ctx)
	log.Info("failExperiment called")
	defer log.Info("failExperiment completed")

	if err != nil {
		log.Error(err, err.Error())
	}

	// For context see: https://github.com/iter8-tools/etc3/issues/201
	// if stop, result, err := r.launchHandlerWrapper(ctx, instance, HandlerTypeFailure,
	// 	handlerLaunchModifier{onSuccessfulLaunch: func() { r.advanceStage(ctx, instance, v2beta1.ExperimentStageFinishing) }},
	// ); stop {
	// 	return result, err
	// }

	return r.endExperiment(ctx, instance, "Experiment failed")
}

func validUpdateErr(err error) bool {
	if err == nil {
		return true
	}
	benignMsg := "the object has been modified"
	return strings.Contains(err.Error(), benignMsg)
}

func (r *ExperimentReconciler) updateStatus(ctx context.Context, instance *v2beta1.Experiment) error {
	log := Logger(ctx)
	originalStatus := OriginalStatus(ctx)

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

// checkHandlersStatus checks the status of a set of handlers and takes appropriate action:
// If running, tell caller to stop (to wait for completion)
// If failed, call failExperiment and tell caller to stop
// If completed successfully, tell caller to continue
func (r *ExperimentReconciler) checkHandlersStatus(ctx context.Context, instance *v2beta1.Experiment,
	handlerTypes []HandlerType) (bool, ctrl.Result, error) {

	log := Logger(ctx)
	log.Info("checkHandlersStatus called", "handlerTypes", handlerTypes)
	defer log.Info("checkHandlersStatus completed")

	dummyResult := ctrl.Result{}
	stop := true

	for _, handlerType := range handlerTypes {
		handler := r.GetHandler(instance, handlerType)
		if handler == nil {
			continue
		}

		if handlerType == HandlerTypeLoop {
			for loop := 1; loop <= int(instance.Spec.GetMaxLoops()); loop++ {
				if stop, result, err := r.checkHandlerStatus(ctx, instance, handlerType, handler, &loop); stop {
					return stop, result, err
				}
			}
		} else {
			if stop, result, err := r.checkHandlerStatus(ctx, instance, handlerType, handler, nil); stop {
				return stop, result, err
			}
		}
	}

	return !stop, dummyResult, nil
}

func (r *ExperimentReconciler) checkHandlerStatus(ctx context.Context, instance *v2beta1.Experiment,
	handlerType HandlerType, handler *string, handlerInstance *int) (bool, ctrl.Result, error) {

	log := Logger(ctx)
	log.Info("checkHandlerStatus called", "handlerType", handlerType, "handler", handler)
	defer log.Info("checkHandlerStatus completed")

	dummyResult := ctrl.Result{}
	stop := true

	switch r.GetHandlerStatus(ctx, instance, handler, handlerInstance) {
	case HandlerStatusRunning:
		// exit; keep waiting for handler to complete
		result, err := r.endRequest(ctx, instance)
		return stop, result, err
	case HandlerStatusComplete:
		switch handlerType {
		case HandlerTypeFinish, HandlerTypeFailure, HandlerTypeRollback:
			// terminal handler completed; we end the experiment
			result, err := r.endExperiment(ctx, instance, "Experiment Completed")
			return stop, result, err
		case HandlerTypeLoop:
			// we update Status.CurrentWeightDistribution then allow reconcile to continue
			if err := updateObservedWeights(ctx, instance, r.RestConfig); err != nil {
				r.recordExperimentFailed(ctx, instance, v2beta1.ReasonInvalidExperiment, "Specification of version weightObjectRef invalid: %s", err.Error())
				return stop, dummyResult, err
			}
			return !stop, dummyResult, nil
		default: // HandlerTypeStart
			// allow reconcile to continue
			return !stop, dummyResult, nil
		}
	case HandlerStatusFailed:
		// a handler failed; don't call a failure handler; just stop
		msg := fmt.Sprintf("%s actions failed", handlerType)
		r.recordExperimentFailed(ctx, instance, v2beta1.ReasonHandlerFailed, msg)
		result, err := r.endExperiment(ctx, instance, msg)
		return stop, result, err
	default: // HandlerStatusNotLaunched, HandlerStatusNoHandler:
		return !stop, dummyResult, nil
	}

}

type handlerLaunchPrerequisiteChecker func() bool
type handlerLaunchOnSuccess func()
type handlerLaunchModifier struct {
	// A prerequisite check is an extra check before launching an handler
	// Used, for example, when launching the start handler to verify that the handler wasn't run in the past
	prerequisiteCheck handlerLaunchPrerequisiteChecker
	// the current loop; used by the loop handler to generate unique job names
	loop *int
	// Any steps to be called on a successful launch.
	// A method is used instead of relying on the caller because the launch method recommends whether the caller
	// should proceed or teminate. The launch method handles the behavior before returning.
	// Used, for example, when a finish handler launches to advance the stage from Running to Finishing
	onSuccessfulLaunch handlerLaunchOnSuccess
}

// launchHandlerWrapper wraps launchHandler with the following additional behavior:
// Determine if a handler actually exists
// Run an prerequisite check if one was provided
// If the handler successfully launches, run any provided modifier
// If the handler fails to launch, try launching the failure handler
// Record any succesful launch
// Return (a) whether or not the caller should stop processing the current Reconcile()
//        (b) a ctrl.Result to be used by the caller when it should stop
//        (c) an error if one occured
// The caller should continue only if there is no handler or the prerequisite check failed
// If a handler was launched the caller should wait for completion
// If an error occurred, failExperiment was called and the caller should stop
func (r *ExperimentReconciler) launchHandlerWrapper(
	ctx context.Context, instance *v2beta1.Experiment, handlerType HandlerType,
	modifier handlerLaunchModifier) (bool, ctrl.Result, error) {

	log := Logger(ctx)
	log.Info("launchHandlerWrapper called", "handlerType", handlerType)
	defer log.Info("launchHandlerWrapper completed", "handlerType", handlerType)

	dummyResult := ctrl.Result{}
	stop := true

	// run handler
	handler := r.GetHandler(instance, handlerType)
	if handler == nil {
		log.Info("launchHandlerWrapper no handler", "handlerType", handlerType)
		return !stop, dummyResult, nil
	}

	// verify any prerequisites; if not met, don't launch
	// For example, to check that the handler hasn't been run in the past
	if modifier.prerequisiteCheck != nil && !modifier.prerequisiteCheck() {
		log.Info("launchHandlerWrapper prerequisite check rejected launch", "handlerType", handlerType)
		return !stop, dummyResult, nil
	}

	if err := r.LaunchHandler(ctx, instance, *handler, modifier.loop); err != nil {
		// An error occurred trying to launch a handler; recommend immediate termination
		result, err := r.endExperiment(ctx, instance, "failure executing failure handler")
		return stop, result, err
	}

	// successfully launched the handler; run any modifier
	// an example is to advance the stage after successfully launching a finishHandler
	if modifier.onSuccessfulLaunch != nil {
		modifier.onSuccessfulLaunch()
	}

	// record launch
	r.recordExperimentProgress(ctx, instance, v2beta1.ReasonHandlerLaunched, "%s handler '%s' launched", handlerType, *handler)

	// tell caller to stop (to wait for handler to complete)
	result, err := r.endRequest(ctx, instance)
	return stop, result, err
}

func (r *ExperimentReconciler) cleanupDeletedExperiments(ctx context.Context, instance *v2beta1.Experiment) {
	log := Logger(ctx)
	log.Info("cleanupDeletedExperiments called")
	defer log.Info("cleanupDeletedExperiments completed")

	// Identify any experiments that have been deleted
	// Delete any remaining jobs associated with them

	// to identify experiments get list of jobs and map to experiments via labels
	experiments := r.identifyDeletedExperiments(ctx, r.identifyExperimentsFromHandlers(ctx))
	for key, jobs := range experiments {
		for _, job := range jobs {
			log.Info("Deleting handler job", "experiment", key, "jobNamespace", job.ObjectMeta.Namespace, "jobName", job.ObjectMeta.Name)
			r.Delete(ctx, &job, client.PropagationPolicy(metav1.DeletePropagationBackground))
		}
	}
}

func (r *ExperimentReconciler) identifyExperimentsFromHandlers(ctx context.Context) map[string][]batchv1.Job {
	log := Logger(ctx)
	log.Info("identifyExperimentsFromHandlers called")
	defer log.Info("identifyExperimentsFromHandlers completed")

	experimentToJobs := map[string][]batchv1.Job{}

	jobs := &batchv1.JobList{}
	if err := r.List(ctx, jobs, client.InNamespace(r.Iter8Config.Namespace)); err != nil {
		log.Error(err, "identifyExperimentsFromHandlers Unable to list experiments")
		return experimentToJobs
	}
	for _, job := range jobs.Items {
		nm, ok := job.ObjectMeta.GetLabels()[LabelExperimentName]
		if !ok {
			// no label LabelExperimentName; ignore
			continue
		}
		ns, ok := job.ObjectMeta.GetLabels()[LabelExperimentNamespace]
		if !ok {
			// no label LabelExperimentNamespace; ignore
			continue
		}
		key := ns + "/" + nm
		log.Info("identifyExperimentsFromHandlers", "key", key)
		if experimentToJobs[key] == nil {
			experimentToJobs[key] = []batchv1.Job{}
		}
		experimentToJobs[key] = append(experimentToJobs[key], job)
	}

	return experimentToJobs
}

func (r *ExperimentReconciler) identifyDeletedExperiments(ctx context.Context, experimentToJobs map[string][]batchv1.Job) map[string][]batchv1.Job {
	log := Logger(ctx)
	log.Info("identifyDeletedExperiments called")
	defer log.Info("identifyDeletedExperiments completed")

	deletedExperimentToJobs := map[string][]batchv1.Job{}

	for key, value := range experimentToJobs {
		splitKey := strings.Split(key, "/")
		experiment := v2beta1.Experiment{}
		err := r.Get(ctx, types.NamespacedName{Name: splitKey[1], Namespace: splitKey[0]}, &experiment)
		if err == nil {
			// we found the experiment, it is not deleted
			continue
		}
		if !errors.IsNotFound(err) {
			// some other error; ignore
			continue
		}
		// the experiment is not present
		log.Info("identifyDeletedExperiments", "key", key)
		deletedExperimentToJobs[key] = value
	}
	return deletedExperimentToJobs
}
