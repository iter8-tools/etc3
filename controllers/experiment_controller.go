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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/util"
)

// experiment.controller.go - implements reconcile loop
//     - handles most of flow except for core of iterate loop which is in iterate.go

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Log            logr.Logger
	Scheme         *runtime.Scheme
	Config         *rest.Config
	SpecModified   bool
	StatusModified bool
}

// +kubebuilder:rbac:groups=iter8.tools,resources=experiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=iter8.tools,resources=experiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile attempts to align the resource with the spec
func (r *ExperimentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("experiment", req.NamespacedName)
	ctx = context.WithValue(ctx, util.LoggerKey, log)

	log.Info("Reconcile() called")
	defer log.Info("Reconcile() completed")

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
		return ctrl.Result{}, err
	}

	log.Info("found instance", "instance", instance, "updatedStatus", r.StatusModified) //, "spec", instance.Spec, "status", instance.Status)

	// ADD FINALIZER
	// If instance does not have a finalizer, add one here (if desired)
	// IF DELETION, RUN FINALIZER and REMOVE FINALIZER
	// If instance deleted and have a finalizer, run it now

	// If instance has never been seen before, initialize status object
	if instance.Status.InitTime == nil {
		instance.InitStatus()
		log.Info("updating instance status after status initialization")
		if err := r.Status().Update(ctx, instance); err != nil {
			log.Error(err, "Failed to update when initializing status.")
			return ctrl.Result{}, err
		}
		r.StatusModified = false
		log.Info("Updated status")
	}

	// If experiment already completed, stop
	if instance.Status.GetCondition(v2alpha1.ExperimentConditionExperimentCompleted).IsTrue() {
		log.Info("Experiment already completed.")
		return ctrl.Result{}, nil
	}

	// Check if we are in the process of terminating an experiment and take appropriate action:
	// If {finish, failure, rollback} handler running, just quit (wait until done)
	// If {} handler was running and is now completed, endExperiment
	if instance.HasFinishHandler() {
		if instance.IsFinishHandlerRunning() {
			return r.endRequest(ctx, instance)
		}
		if instance.IsFinishHandlerCompleted() {
			return r.endExperiment(ctx, instance)
		}
	}

	if instance.HasFailureHandler() {
		if instance.IsFailureHandlerRunning() {
			return r.endRequest(ctx, instance)
		}
		if instance.IsFailureHandlerCompleted() {
			return r.endExperiment(ctx, instance)
		}
	}

	if instance.HasRollbackHandler() {
		if instance.IsRollbackHandlerRunning() {
			return r.endRequest(ctx, instance)
		}
		if instance.IsRollbackHandlerCompleted() {
			return r.endExperiment(ctx, instance)
		}
	}

	// LATE INITIALIZATION of spec
	specChanged := instance.SpecLateInitialization()
	if !r.AlreadyReadMetrics(instance) {
		specChanged = r.ReadMetrics(ctx, instance) || specChanged
	}
	if specChanged {
		r.SpecModified = true
	}
	if err := r.updateIfNeeded(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("Late initialization complete.")

	// // VALIDATE EXPERIMENT
	// // Basic validation of experiment object

	// // If experiment is completed, jump to finish handler
	// // If experiment is paused, stop

	// TARGET ACQUISITION
	// ensure that the target is not involved in another experiment
	// TODO how to record experiment with annotation in target

	// // START HANDLER
	// // if !startHandlerCompleted() {
	// // 	// check ExperimentConditionStartHandlerFinished OR
	// // 	// start job completed
	// // 	runStartHanlder() - create job and run it
	// // 	update ExperimentConditionStartHandlerLaunched True
	// // 	endRequest()
	// // }

	// // VERSION VALIDATION
	// // verify that versionInfo is present
	// // verify that the number of versions is suitable to the spec.type
	// // verify things like: if Canary then exactly 2 versions in versionInfo

	// // INITIAL WEIGHT DISTRIBUTION (FixedSplit only)
	// // if instance.Spec.GetAlgorithm() == v2alpha1.AlgorithmTypeFixedSplit {
	// // 	redistributeWeight (ctx, instance, instance.Spec.GetWeightDistribution())
	// // }

	// EXECUTE ITERATION
	log.Info("Executing Iteration", "maxIterations", instance.Spec.GetMaxIterations(), "completed iterations", *instance.Status.CompletedIterations)
	if r.moreIterationsNeeded(instance) && r.sufficientTimePassedSincePreviousIteration(ctx, instance) {
		result, err := r.doIteration(ctx, instance)
		if err != nil {
			r.endRequest(ctx, instance)
		}
		return result, err
	}

	// // FUTURE PROMOTE LOGIC ?

	// // FINISH HANDLER
	// // if experimentFinished() {
	// // 	if !finishHandlerCalled() {
	// // 		runFinishHandler()
	// // 		update ExperimentConditionFinishHandlerLaunched True
	// // 	}
	// // 	endRequest()
	// // }

	// //

	return ctrl.Result{}, nil
}

// SetupWithManager ..
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v2alpha1.Experiment{}).
		Complete(r)
}

// endRequest writes any changes (if needed) in preparation for ending processing of this reconcile request
func (r *ExperimentReconciler) endRequest(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("endRequest() called")
	defer log.Info("endRequest() completed")

	r.updateIfNeeded(ctx, instance)
	return ctrl.Result{}, nil
}

// endExperiment is called to mark an experiment as completed and
// triggers next experiment object
func (r *ExperimentReconciler) endExperiment(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("endExperiment() called")
	defer log.Info("endExperiment() completed")

	r.markExperimentCompleted(ctx, instance, "")
	r.updateIfNeeded(ctx, instance)

	// trigger next experiment

	return ctrl.Result{}, nil
}

// finishExperiment calls the finish handler or (if none) ends the experiment
func (r *ExperimentReconciler) finishExperiment(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("finishExperiment() called")
	defer log.Info("finishExperiment() completed")

	// run finish handlers
	if instance.HasFinishHandler() {
		r.startFinishHandler(ctx, instance)
		r.startRollbackHandler(ctx, instance)
		return r.endRequest(ctx, instance)
	} else {
		return r.endExperiment(ctx, instance)
	}
}

func (r *ExperimentReconciler) startFinishHandler(ctx context.Context, instance *v2alpha1.Experiment) error {
	log := util.Logger(ctx)
	log.Info("startFinishHandler() called")
	defer log.Info("startFinishHandler() ended")
	return nil
}

func (r *ExperimentReconciler) failExperiment(ctx context.Context, instance *v2alpha1.Experiment) (ctrl.Result, error) {
	log := util.Logger(ctx)
	log.Info("failExperiment() called")
	// set ExperimentConditionExperimentCompleted True
	// set reommendedBaseline to spec.VersionInfo.baseline
	// set ExperimentConditionExperimentSucceeded False
	// call FAILURE handler
	// queue next experiment
	r.updateIfNeeded(ctx, instance)
	return ctrl.Result{}, nil
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

	if r.SpecModified {
		log.Info("updating spec", "spec", instance.Spec)
		if err := r.Update(ctx, instance); err != nil && !validUpdateErr(err) {
			log.Error(err, "Failed to update spec")
			return err
		}
		r.SpecModified = false
	}

	return nil
}
