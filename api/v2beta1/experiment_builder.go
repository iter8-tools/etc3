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
	b.Spec.VersionInfo = append(b.Spec.VersionInfo, version)

	return b
}

// WithDuration ..
func (b *ExperimentBuilder) WithDuration(interval int32, maxLoops int32) *ExperimentBuilder {

	if b.Spec.Duration == nil {
		b.Spec.Duration = &Duration{}
	}

	b.Spec.Duration.MinIntervalBetweenLoops = &interval
	b.Spec.Duration.MaxLoops = &maxLoops

	return b
}

// WithCurrentWeight ..
func (b *ExperimentBuilder) WithCurrentWeight(name string, weight int32) *ExperimentBuilder {

	if len(b.Status.CurrentWeightDistribution) == 0 {
		b.Status.CurrentWeightDistribution = make([]int32, len(b.Spec.VersionInfo))
	}

	for i, w := range b.Spec.VersionInfo {
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

	if len(b.Status.Analysis.Weights) == 0 {
		b.Status.Analysis.Weights = make([]int32, len(b.Spec.VersionInfo))
	}

	for i, w := range b.Spec.VersionInfo {
		if w == name {
			b.Status.Analysis.Weights[i] = weight
			return b
		}
	}

	return b
}

// WithCondition ..
func (b *ExperimentBuilder) WithCondition(condition ExperimentConditionType, status corev1.ConditionStatus, reason string, messageFormat string, messageA ...interface{}) *ExperimentBuilder {
	b.Status.MarkCondition(condition, status, reason, messageFormat, messageA...)
	return b
}

// WithAction ..
func (b *ExperimentBuilder) WithAction(key string, tasks []TaskSpec) *ExperimentBuilder {
	if b.Spec.Actions == nil {
		b.Spec.Actions = make(ActionMap)
	}
	b.Spec.Actions[key] = tasks
	return b
}

// WithReward ..
func (b *ExperimentBuilder) WithReward(metric string, preferredDirection PreferredDirectionType) *ExperimentBuilder {
	if b.Spec.Criteria == nil {
		b.Spec.Criteria = &Criteria{}
	}
	b.Spec.Criteria.Rewards = append(b.Spec.Criteria.Rewards, Reward{
		Metric:             metric,
		PreferredDirection: preferredDirection,
	})
	return b
}

// WithObjective ..
func (b *ExperimentBuilder) WithObjective(metric string, upper *resource.Quantity, lower *resource.Quantity) *ExperimentBuilder {
	if b.Spec.Criteria == nil {
		b.Spec.Criteria = &Criteria{}
	}
	b.Spec.Criteria.Objectives = append(b.Spec.Criteria.Objectives, Objective{
		Metric:     metric,
		UpperLimit: upper,
		LowerLimit: lower,
	})
	return b
}
