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

package v2beta1

import (
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExperimentBuilder ..
type ExperimentBuilder Experiment

// NewExperiment returns an iter8 experiment
func NewExperiment(name, namespace string) *ExperimentBuilder {
	e := &Experiment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Experiment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return (*ExperimentBuilder)(e)
}

// Build the experiment object
func (b *ExperimentBuilder) Build() *Experiment {
	return (*Experiment)(b)
}

// With Version ..
func (b *ExperimentBuilder) WithVersion(version string) *ExperimentBuilder {
	b.Spec.Versions = append(b.Spec.Versions, version)

	return b
}

// WithDeploymentPattern ..
func (b *ExperimentBuilder) WithDeploymentPattern(deploymentPattern DeploymentPatternType) *ExperimentBuilder {
	b.Spec.Strategy.DeploymentPattern = &deploymentPattern

	return b
}

// WithDuration ..
func (b *ExperimentBuilder) WithDuration(interval int32, iterationsPerLoop int32, maxLoops int32) *ExperimentBuilder {

	if b.Spec.Duration == nil {
		b.Spec.Duration = &Duration{}
	}

	b.Spec.Duration.IntervalSeconds = &interval
	b.Spec.Duration.IterationsPerLoop = &iterationsPerLoop
	b.Spec.Duration.MaxLoops = &maxLoops

	return b
}

// WithRequestCount ..
func (b *ExperimentBuilder) WithRequestCount(requestCount string) *ExperimentBuilder {

	if b.Spec.Criteria == nil {
		b.Spec.Criteria = &Criteria{}
	}

	b.Spec.Criteria.RequestCount = &requestCount

	return b
}

// WithBaselineVersion ..
func (b *ExperimentBuilder) WithBaselineVersion(name string, objRef *corev1.ObjectReference) *ExperimentBuilder {

	if b.Spec.VersionInfo == nil {
		b.Spec.VersionInfo = &VersionInfo{
			Baseline: VersionDetail{
				Name: name,
			},
			Candidates: []VersionDetail{},
		}
	} else {
		// override whatever candidate is there
		b.Spec.VersionInfo.Baseline = VersionDetail{
			Name: name,
		}
	}

	if objRef != nil {
		b.Spec.VersionInfo.Baseline.WeightObjRef = objRef
	}

	return b
}

// WithCandidateVersion ..
// Expects VersionInfo to be defined already via WithBaselineVersion()
func (b *ExperimentBuilder) WithCandidateVersion(name string, objRef *corev1.ObjectReference) *ExperimentBuilder {

	candidate := VersionDetail{
		Name: name,
	}

	if objRef != nil {
		candidate.WeightObjRef = objRef
	}

	if b.Spec.VersionInfo == nil {
		b.Spec.VersionInfo = &VersionInfo{}
	}

	for _, c := range b.Spec.VersionInfo.Candidates {
		if c.Name == name {
			// overwrite
			c.WeightObjRef = candidate.WeightObjRef
			return b
		}
	}

	// if not present, append candidate
	b.Spec.VersionInfo.Candidates = append(b.Spec.VersionInfo.Candidates, candidate)
	return b
}

// WithCurrentWeight ..
func (b *ExperimentBuilder) WithCurrentWeight(name string, weight int32) *ExperimentBuilder {

	if len(b.Status.CurrentWeightDistribution) == 0 {
		b.Status.CurrentWeightDistribution = make([]int32, len(b.Spec.Versions))
	}

	for i, w := range b.Spec.Versions {
		if w == name {
			b.Status.CurrentWeightDistribution[i] = weight
			return b
		}
	}

	return b
}

// WithRecommendedWeight ..
func (b *ExperimentBuilder) WithRecommendedWeight(name string, weight int32) *ExperimentBuilder {

	if b.Status.Analysis == nil {
		b.Status.Analysis = &Analysis{}
	}
	if b.Status.Analysis.Weights == nil {
		now := metav1.Now()
		b.Status.Analysis.Weights = &WeightsAnalysis{
			AnalysisMetaData: AnalysisMetaData{
				Provenance: "provenance",
				Timestamp:  now,
			},
			Data: []WeightData{},
		}
	}

	for _, w := range b.Status.Analysis.Weights.Data {
		if w.Name == name {
			w.Value = weight
			return b
		}
	}
	b.Status.Analysis.Weights.Data = append(
		b.Status.Analysis.Weights.Data,
		WeightData{Name: name, Value: weight},
	)

	return b
}

// WithCondition ..
func (b *ExperimentBuilder) WithCondition(condition ExperimentConditionType, status corev1.ConditionStatus, reason string, messageFormat string, messageA ...interface{}) *ExperimentBuilder {
	b.Status.MarkCondition(condition, status, reason, messageFormat, messageA...)
	return b
}

// WithAction ..
func (b *ExperimentBuilder) WithAction(key string, tasks []TaskSpec) *ExperimentBuilder {
	if b.Spec.Strategy.Actions == nil {
		b.Spec.Strategy.Actions = make(ActionMap)
	}
	b.Spec.Strategy.Actions[key] = tasks
	return b
}

// WithReward ..
func (b *ExperimentBuilder) WithReward(metric Metric, preferredDirection PreferredDirectionType) *ExperimentBuilder {
	if b.Spec.Criteria == nil {
		b.Spec.Criteria = &Criteria{}
	}
	name := metric.Namespace + "/" + metric.Name
	b.Spec.Criteria.Rewards = append(b.Spec.Criteria.Rewards, Reward{
		Metric:             name,
		PreferredDirection: preferredDirection,
	})
	return b
}

// WithIndicator ..
func (b *ExperimentBuilder) WithIndicator(metric Metric) *ExperimentBuilder {
	if b.Spec.Criteria == nil {
		b.Spec.Criteria = &Criteria{}
	}
	name := metric.Namespace + "/" + metric.Name
	b.Spec.Criteria.Indicators = append(b.Spec.Criteria.Indicators, name)
	return b
}

// WithObjective ..
func (b *ExperimentBuilder) WithObjective(metric Metric, upper *resource.Quantity, lower *resource.Quantity, rollback bool) *ExperimentBuilder {
	if b.Spec.Criteria == nil {
		b.Spec.Criteria = &Criteria{}
	}
	name := metric.Namespace + "/" + metric.Name
	b.Spec.Criteria.Objectives = append(b.Spec.Criteria.Objectives, Objective{
		Metric:            name,
		UpperLimit:        upper,
		LowerLimit:        lower,
		RollbackOnFailure: &rollback,
	})
	return b
}
