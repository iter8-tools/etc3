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

	"github.com/iter8-tools/etc3/api/v2alpha1"
)

// IsExperimentValid verifies that instance.Spec is valid; this should be done after late initialization
// TODO 1. If fixed_split, we have an initial split (or are we just assuming start handler does it?)
// TODO 2. Warning if no criteria?
// TODO 3. For ab and abn there is a reward
// TODO 4. If rollbackOnFailure there is a rollback handler?
func (r *ExperimentReconciler) IsExperimentValid(ctx context.Context, instance *v2alpha1.Experiment) bool {
	return true
}

// IsVersionInfoValid verifies that Spec.versionInfo is valid
// DONE 1. verify that versionInfo is present
// DONE 2. verify that the number of versions in Spec.versionInfo is suitable to the Spec.Strategy.Type
// TODO 3. verify any ObjectReferences are existing objects in the cluster
func (r *ExperimentReconciler) IsVersionInfoValid(ctx context.Context, instance *v2alpha1.Experiment) bool {
	// 1. verify that versionInfo is present
	if instance.Spec.VersionInfo == nil {
		r.recordExperimentFailed(ctx, instance, v2alpha1.ReasonInvalidExperiment, "No versionInfo in experiment")
		return false
	}
	// 2. verify that the number of versions in Spec.versionInfo is suitable to the Spec.Strategy.Type
	if !candidatesMatchStrategy(instance.Spec) {
		r.recordExperimentFailed(ctx, instance, v2alpha1.ReasonInvalidExperiment, "Invlid number of candidates for %s experiment", instance.Spec.Strategy.Type)
		return false
	}

	return true
}

func candidatesMatchStrategy(s v2alpha1.ExperimentSpec) bool {
	switch s.Strategy.Type {
	case v2alpha1.StrategyTypePerformance:
		return len(s.VersionInfo.Candidates) == 0
	case v2alpha1.StrategyTypeAB, v2alpha1.StrategyTypeCanary, v2alpha1.StrategyTypeBlueGreen:
		return len(s.VersionInfo.Candidates) == 1
	case v2alpha1.StrategyTypeABN:
		return len(s.VersionInfo.Candidates) > 0
	}
	return true
}
