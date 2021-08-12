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

	"github.com/iter8-tools/etc3/api/v2alpha2"
)

// IsExperimentValid verifies that instance.Spec is valid; this should be done after late initialization
// TODO 1. If fixed_split, we have an initial split (or are we just assuming start handler does it?)
// TODO 2. Warning if no criteria?
// TODO 3. For ab and abn there is a reward
// TODO 4. If rollbackOnFailure there is a rollback handler?
func (r *ExperimentReconciler) IsExperimentValid(ctx context.Context, instance *v2alpha2.Experiment) bool {
	return r.AreTasksValid(ctx, instance)
}

// IsVersionInfoValid verifies that Spec.versionInfo is valid
// DONE Verify that versionInfo is present
// DONE Verify that the number of versions (spec.versionInfo) is suitable to the spec.strategy.testingPattern
// DONE Verify that the names of the versions are all unique
// DONE Verify that the number of rewards (spec.criteria.rewards) is suitable to spec.strategy.testingPattern
// DONE Verify that each fieldpath starts with '.'
// TODO Verify any ObjectReferences are existing objects in the cluster
func (r *ExperimentReconciler) IsVersionInfoValid(ctx context.Context, instance *v2alpha2.Experiment) bool {
	// Verify that versionInfo is present
	if instance.Spec.VersionInfo == nil {
		r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonInvalidExperiment, "No versionInfo in experiment")
		return false
	}
	// Verify that the number of versions in Spec.versionInfo is suitable to the Spec.Strategy.Type
	if !candidatesMatchStrategy(instance.Spec) {
		r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonInvalidExperiment, "Invalid number of candidates for %s experiment", instance.Spec.Strategy.TestingPattern)
		return false
	}
	// Verify that the names of the versionns are all unique
	if !versionsUnique(instance.Spec) {
		r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonInvalidExperiment, "Version names are not unique")
		return false
	}

	// Verify that the number of rewards (spec.criteria.rewards) is suitable to spec.strategy.testingPattern
	if !validNumberOfRewards(instance.Spec) {
		r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonInvalidExperiment, "Invalid number of rewards for %s experiment", instance.Spec.Strategy.TestingPattern)
		return false
	}

	// Verify that any specified fieldpath starts with a '.'
	b := instance.Spec.VersionInfo.Baseline.WeightObjRef
	if b != nil && b.FieldPath != "" && b.FieldPath[0] != '.' {
		r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonInvalidExperiment, "Fieldpaths must start with '.'")
		return false
	}
	for _, c := range instance.Spec.VersionInfo.Candidates {
		if c.WeightObjRef != nil && len(c.WeightObjRef.FieldPath) != 0 && c.WeightObjRef.FieldPath[0] != '.' {
			r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonInvalidExperiment, "Fieldpaths must start with '.'")
			return false
		}
	}

	return true
}

func candidatesMatchStrategy(s v2alpha2.ExperimentSpec) bool {
	switch s.Strategy.TestingPattern {
	case v2alpha2.TestingPatternConformance:
		return len(s.VersionInfo.Candidates) == 0
	case v2alpha2.TestingPatternAB, v2alpha2.TestingPatternCanary:
		return len(s.VersionInfo.Candidates) == 1
	case v2alpha2.TestingPatternABN:
		return len(s.VersionInfo.Candidates) > 0
	}
	return true
}

func versionsUnique(s v2alpha2.ExperimentSpec) bool {
	versions := []string{s.VersionInfo.Baseline.Name}
	for _, candidate := range s.VersionInfo.Candidates {
		if containsString(versions, candidate.Name) {
			return false
		}
		versions = append(versions, candidate.Name)
	}
	return true
}

// AreTasksValid ensures that each task either has a valid task string or a valid run string but not both
func (r *ExperimentReconciler) AreTasksValid(ctx context.Context, instance *v2alpha2.Experiment) bool {
	for _, a := range instance.Spec.Strategy.Actions {
		for _, t := range a {
			num := 0
			if t.Task != nil && len(*t.Task) > 0 {
				num++
			}
			if t.Run != nil && len(*t.Run) > 0 {
				num++
			}
			if num != 1 {
				return false
			}
		}
	}
	return true
}

func validNumberOfRewards(s v2alpha2.ExperimentSpec) bool {
	switch s.Strategy.TestingPattern {
	case v2alpha2.TestingPatternConformance:
		return s.Criteria == nil || len(s.Criteria.Rewards) == 0
	case v2alpha2.TestingPatternCanary:
		return s.Criteria == nil || len(s.Criteria.Rewards) == 0
	case v2alpha2.TestingPatternAB:
		return s.Criteria != nil && len(s.Criteria.Rewards) == 1
	case v2alpha2.TestingPatternABN:
		return s.Criteria != nil && len(s.Criteria.Rewards) == 1
	}
	return true
}
