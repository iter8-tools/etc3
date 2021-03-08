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

// target.go implements code to lock a target for an experiment

package controllers

import (
	"context"

	v2alpha2 "github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *ExperimentReconciler) acquireTarget(ctx context.Context, instance *v2alpha2.Experiment) bool {
	log := util.Logger(ctx)
	log.Info("acquireTarget called")
	defer log.Info("acquireTarget completed")

	// do we already have the target?
	log.Info("acquireTarget", "Acquired", instance.Status.GetCondition(v2alpha2.ExperimentConditionTargetAcquired))
	if instance.Status.GetCondition(v2alpha2.ExperimentConditionTargetAcquired).IsTrue() {
		return true
	}

	// get the set of experiments (across all namespaces) that share the target and which are not completed
	shareTarget := r.activeContendersForTarget(ctx, instance.Spec.Target)

	// If another experiment has acquired the target, we cannot
	// While checking, keep track of the highest priority (earliest init time) among the set of competitors
	// If no one has acquired the target, we will compare priorities
	earliest := instance.Status.InitTime
	for _, e := range shareTarget {
		if !sameInstance(instance, e) {
			if e.Status.GetCondition(v2alpha2.ExperimentConditionTargetAcquired).IsTrue() {
				log.Info("acquireTarget", "target already owned by", e.Name)
				return false
			}
			// keep track of the competitor with the highest priority (earliest init time)
			if e.Status.InitTime.Before(earliest) {
				earliest = e.Status.InitTime
			}
		}
	}

	// we didn't find a competeitor who has already acquired the target
	// we can if we have the highest priority (started first)
	log.Info("acquireTarget", "instance InitTime", instance.Status.InitTime, "earliest", earliest.Time)
	if !earliest.Before(instance.Status.InitTime) {
		log.Info("acquireTarget target available; acquiring")
		r.recordTargetAcquired(ctx, instance, "")
	}

	// otherwise, return we cannot acquire target: there is another experiment with priority
	return false
}

func (r *ExperimentReconciler) activeContendersForTarget(ctx context.Context, target string) []*v2alpha2.Experiment {
	log := util.Logger(ctx)
	log.Info("activeContendersForTarget called")
	defer log.Info("activeContendersForTarget completed")

	result := []*v2alpha2.Experiment{}

	experiments := &v2alpha2.ExperimentList{}
	if err := r.List(ctx, experiments); err != nil {
		log.Error(err, "activeContendersForTarget Unable to list experiments")
		return result
	}

	for i := range experiments.Items {
		if experiments.Items[i].Spec.Target == target &&
			experiments.Items[i].Status.GetCondition(v2alpha2.ExperimentConditionExperimentCompleted).IsFalse() {
			result = append(result, &experiments.Items[i])
		}
	}

	log.Info("activeContendersForTarget", "result", result)
	return result
}

func sameInstance(instance1 *v2alpha2.Experiment, instance2 *v2alpha2.Experiment) bool {
	return instance1.Name == instance2.Name && instance1.Namespace == instance2.Namespace
}

// triggerWaitingExperiments looks at all targets (it finds them by looking at all experiments)
// and triggers the next experiment waiting for the target if:
//    (a) there isn't already an active experiment for the target, and
//    (b) the next active experiment for the target is this instance (because acquireTarget will
//        be called soon -- when endExperiment() is called)
func (r *ExperimentReconciler) triggerWaitingExperiments(ctx context.Context, instance *v2alpha2.Experiment) {
	log := util.Logger(ctx)
	log.Info("triggerWaitingExperiments called")
	defer log.Info("triggerWaitingExperiments completed")

	targetsAlreadyChecked := []string{}

	experiments := &v2alpha2.ExperimentList{}
	if err := r.List(ctx, experiments); err != nil {
		log.Error(err, "triggerWaitingExperiments: Unable to list experiments")
		return
	}

	for _, experiment := range experiments.Items {
		target := experiment.Spec.Target
		if containsString(targetsAlreadyChecked, target) {
			continue
		}
		targetsAlreadyChecked = append(targetsAlreadyChecked, target)
		r.triggerNextExperiment(ctx, target, instance)
	}
}

// nextWaitingExperiment identifies the next experiment waiting for a given target.
// If there are none (either because there are no waiting experiments for the given target
// or because the target is already in use), nil is returned.
// If instance is specified (blacklisted), it will not be returned.
func (r *ExperimentReconciler) nextWaitingExperiment(ctx context.Context, target string, instance *v2alpha2.Experiment) *v2alpha2.Experiment {
	log := util.Logger(ctx)
	log.Info("nextWaitingExperiment called")
	defer log.Info("nextWaitingExperiment completed")

	shareTarget := r.activeContendersForTarget(ctx, target)

	earliest := metav1.Now()
	next := (*v2alpha2.Experiment)(nil)

	for _, e := range shareTarget {
		// not interested in instance if it is provided
		if instance != nil && sameInstance(e, instance) {
			continue
		}

		// Note that we've already filtered out the completed ones so if there is another
		// experiment that has acquired the target, we can't/shouldn't suggest another
		if e.Status.GetCondition(v2alpha2.ExperimentConditionTargetAcquired).IsTrue() {
			log.Info("nextWaitingExperiment", "target already owned by", e.Name)
			return nil
		}
		// keep track of the competitor with the highest priority (earliest init time)
		if e.Status.InitTime.Before(&earliest) {
			earliest = *e.Status.InitTime
			next = e
		}
	}
	if next != nil {
		log.Info("nextWaitingExperiment", "name", next.Name, "namespace", next.Namespace)
	}
	return next
}

// triggerNextExperiment finds the next experiment to trigger. If there is one, it is triggered.
func (r *ExperimentReconciler) triggerNextExperiment(ctx context.Context, target string, instance *v2alpha2.Experiment) {
	log := util.Logger(ctx)
	log.Info("triggerNextExperiment called", "target", target)
	defer log.Info("triggerNextExperiment completed")

	next := r.nextWaitingExperiment(ctx, target, instance)
	if nil != next {
		log.Info("triggerNextExperiment", "name", next.Name, "namespace", next.Namespace)
		// found one
		r.ReleaseEvents <- event.GenericEvent{
			Meta:   next,
			Object: next,
		}
	}
}
