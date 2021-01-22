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

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *ExperimentReconciler) acquireTarget(ctx context.Context, instance *v2alpha1.Experiment) bool {
	log := util.Logger(ctx)
	log.Info("acquireTarget called")
	defer log.Info("acquireTarget completed")

	// do we already have the target?
	log.Info("acquireTarget", "Acquired", instance.Status.GetCondition(v2alpha1.ExperimentConditionTargetAcquired))
	if instance.Status.GetCondition(v2alpha1.ExperimentConditionTargetAcquired).IsTrue() {
		return true
	}

	// get the set of experiments (across all namespaces) that share the target and which are not completed
	// the members of this set are our competetors for the target
	shareTarget := r.otherActiveContendersForTarget(ctx, instance)

	// If another experiment has aquired the target, we cannot
	// While checking, keep track of the highest priority (earliest init time) among the set of competitors
	// If no one has acquired the target, we will compare priorities
	earliest := instance.Status.InitTime
	for _, e := range shareTarget {
		if !sameInstance(instance, e) {
			if e.Status.GetCondition(v2alpha1.ExperimentConditionTargetAcquired).IsTrue() {
				log.Info("acquiredTarget", "target owned by", e.Name)
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
	log.Info("acquiredTarget", "instance InitTime", instance.Status.InitTime, "earliest", earliest.Time)
	if !earliest.Before(instance.Status.InitTime) {
		log.Info("acquiredTarget acquiring")
		r.recordTargetAcquired(ctx, instance, "")
	}

	// otherwise, return we cannot aquire target: there is another experiment with priority
	return false
}

func (r *ExperimentReconciler) otherActiveContendersForTarget(ctx context.Context, instance *v2alpha1.Experiment) []*v2alpha1.Experiment {
	log := util.Logger(ctx)
	log.Info("otherContendersForTarget called")
	defer log.Info("otherContendersForTarget completed")

	result := []*v2alpha1.Experiment{}

	experiments := &v2alpha1.ExperimentList{}
	if err := r.List(ctx, experiments); err != nil {
		log.Error(err, "Unable to list experiments")
		return result
	}

	for _, exp := range experiments.Items {
		if exp.Spec.Target == instance.Spec.Target &&
			exp.Status.GetCondition(v2alpha1.ExperimentConditionExperimentCompleted).IsFalse() {
			result = append(result, &exp)
		}
	}

	log.Info("otherContendersForTarget", "result", result)
	return result
}

func sameInstance(instance1 *v2alpha1.Experiment, instance2 *v2alpha1.Experiment) bool {
	return instance1.Name == instance2.Name && instance1.Namespace == instance2.Namespace
}

// nextExperimentToRun should be called by triggerNextExperiment when we are releasing the target
func (r *ExperimentReconciler) nextExperimentToRun(ctx context.Context, instance *v2alpha1.Experiment) *v2alpha1.Experiment {
	log := util.Logger(ctx)
	log.Info("nextExperimentToRun called")
	defer log.Info("nextExperimentToRun completed")

	shareTarget := r.otherActiveContendersForTarget(ctx, instance)

	earliest := metav1.Now()
	next := (*v2alpha1.Experiment)(nil)

	for _, e := range shareTarget {
		// not interested in ourself
		if sameInstance(e, instance) {
			continue
		}

		// Note that we've already filtered out the completed ones so if there is another
		// experiment that thas acquired the target, we can't/shouldn't suggest another
		if e.Status.GetCondition(v2alpha1.ExperimentConditionTargetAcquired).IsTrue() {
			log.Info("acquiredTarget", "target owned by", e.Name)
			return nil
		}
		// keep track of the competitor with the highest priority (earliest init time)
		if e.Status.InitTime.Before(&earliest) {
			earliest = *e.Status.InitTime
			next = e
		}
	}
	return next
}

func (r *ExperimentReconciler) triggerNextExperiment(ctx context.Context, instance *v2alpha1.Experiment) {
	next := r.nextExperimentToRun(ctx, instance)
	if nil != next {
		// found one
		r.ReleaseEvents <- event.GenericEvent{
			Meta:   nil,
			Object: nil,
		}
	}
}
