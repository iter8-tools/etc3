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

// validation.go - methods to validate an experiment resource

package controllers

import (
	"context"

	"github.com/iter8-tools/etc3/api/v2beta1"
)

// IsExperimentValid verifies that instance.Spec is valid; this should be done after late initialization
// DONE 1. Validate task specification
func (r *ExperimentReconciler) IsExperimentValid(ctx context.Context, instance *v2beta1.Experiment) bool {
	return r.AreTasksValid(ctx, instance)
}

// IsVersionInfoValid verifies that Spec.versionInfo is valid
// DONE Verify at least one version (this is mostly to ensure tests are valid)
// DONE Verify that the names of the versions are all unique
// DONE Verify that the number of rewards (spec.criteria.rewards) is acceptable
func (r *ExperimentReconciler) IsVersionInfoValid(ctx context.Context, instance *v2beta1.Experiment) bool {
	// Verify at least one version
	if len(instance.Spec.VersionInfo) < 1 {
		r.recordExperimentFailed(ctx, instance, v2beta1.ReasonInvalidExperiment, "There must be at least one version")
		return false
	}

	// Verify that the names of the versionns are all unique
	if !versionsUnique(instance.Spec) {
		r.recordExperimentFailed(ctx, instance, v2beta1.ReasonInvalidExperiment, "Version names are not unique")
		return false
	}

	// Verify that the number of rewards (spec.criteria.rewards) is acceptable
	if !validNumberOfRewards(instance.Spec) {
		r.recordExperimentFailed(ctx, instance, v2beta1.ReasonInvalidExperiment, "Invalid number of rewards (at most 1 allowed)")
		return false
	}

	return true
}

func versionsUnique(s v2beta1.ExperimentSpec) bool {
	versions := []string{}
	for _, v := range s.VersionInfo {
		if containsString(versions, v) {
			return false
		}
		versions = append(versions, v)
	}
	return true
}

// AreTasksValid ensures that each task either has a valid task string or a valid run string but not both
func (r *ExperimentReconciler) AreTasksValid(ctx context.Context, instance *v2beta1.Experiment) bool {
	for _, a := range instance.Spec.Actions {
		num := 0
		for _, t := range a {
			if t.Task != nil && len(*t.Task) > 0 {
				num++
			}
			if t.Run != nil && len(*t.Run) > 0 {
				num++
			}
		}
		if num != 1 {
			return false
		}
	}
	return true
}

// verify that there is at most 1 reward
// verify that there are no rewards if just 1 version
func validNumberOfRewards(s v2beta1.ExperimentSpec) bool {
	numRewards := 0
	if s.Criteria != nil {
		numRewards = len(s.Criteria.Rewards)
	}
	numVersions := len(s.VersionInfo)

	// if numVersions is 1 then should be no reward
	// if numVersions is 2 then can be 1 reward or not
	// if numVersions is 3 or more then must be 1 reward
	return numVersions == 1 && numRewards == 0 ||
		numVersions == 2 && numRewards <= 1 ||
		numVersions > 2 && numRewards == 1
}
