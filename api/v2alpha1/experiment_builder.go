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

package v2alpha1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
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

// WithTarget sets spec.Target
func (b *ExperimentBuilder) WithTarget(target string) *ExperimentBuilder {
	b.Spec.Target = target

	return b
}

// WithStrategy ..
func (b *ExperimentBuilder) WithStrategy(strategy StrategyType) *ExperimentBuilder {
	b.Spec.Strategy.Type = strategy

	return b
}

// WithDuration ..
func (b *ExperimentBuilder) WithDuration(interval int32, maxIterations int32) *ExperimentBuilder {

	if b.Spec.Duration == nil {
		b.Spec.Duration = &Duration{}
	}

	b.Spec.Duration.IntervalSeconds = &interval
	b.Spec.Duration.MaxIterations = &maxIterations

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

// WithHandlers ..
func (b *ExperimentBuilder) WithHandlers(handlers map[string]string) *ExperimentBuilder {

	if b.Spec.Strategy.Handlers == nil {
		b.Spec.Strategy.Handlers = &Handlers{}
	}

	for key, value := range handlers {
		hdlr := value
		switch strings.ToLower(key) {
		case "start":
			b.Spec.Strategy.Handlers.Start = &hdlr
		case "finish":
			b.Spec.Strategy.Handlers.Finish = &hdlr
		case "failure":
			b.Spec.Strategy.Handlers.Failure = &hdlr
		case "rollback":
			b.Spec.Strategy.Handlers.Rollback = &hdlr
		default:
		}
	}

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

	for _, w := range b.Status.CurrentWeightDistribution {
		if w.Name == name {
			w.Value = weight
			return b
		}
	}
	b.Status.CurrentWeightDistribution = append(
		b.Status.CurrentWeightDistribution,
		WeightData{Name: name, Value: weight},
	)
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
