// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v2alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AggregatedMetricsAnalysis) DeepCopyInto(out *AggregatedMetricsAnalysis) {
	*out = *in
	in.AnalysisMetaData.DeepCopyInto(&out.AnalysisMetaData)
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make(map[string]AggregatedMetricsData, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AggregatedMetricsAnalysis.
func (in *AggregatedMetricsAnalysis) DeepCopy() *AggregatedMetricsAnalysis {
	if in == nil {
		return nil
	}
	out := new(AggregatedMetricsAnalysis)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AggregatedMetricsData) DeepCopyInto(out *AggregatedMetricsData) {
	*out = *in
	if in.Max != nil {
		in, out := &in.Max, &out.Max
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.Min != nil {
		in, out := &in.Min, &out.Min
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make(map[string]AggregatedMetricsVersionData, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AggregatedMetricsData.
func (in *AggregatedMetricsData) DeepCopy() *AggregatedMetricsData {
	if in == nil {
		return nil
	}
	out := new(AggregatedMetricsData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AggregatedMetricsVersionData) DeepCopyInto(out *AggregatedMetricsVersionData) {
	*out = *in
	if in.Max != nil {
		in, out := &in.Max, &out.Max
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.Min != nil {
		in, out := &in.Min, &out.Min
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.SampleSize != nil {
		in, out := &in.SampleSize, &out.SampleSize
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AggregatedMetricsVersionData.
func (in *AggregatedMetricsVersionData) DeepCopy() *AggregatedMetricsVersionData {
	if in == nil {
		return nil
	}
	out := new(AggregatedMetricsVersionData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Analysis) DeepCopyInto(out *Analysis) {
	*out = *in
	if in.AggregatedMetrics != nil {
		in, out := &in.AggregatedMetrics, &out.AggregatedMetrics
		*out = new(AggregatedMetricsAnalysis)
		(*in).DeepCopyInto(*out)
	}
	if in.WinnerAssessment != nil {
		in, out := &in.WinnerAssessment, &out.WinnerAssessment
		*out = new(WinnerAssessmentAnalysis)
		(*in).DeepCopyInto(*out)
	}
	if in.VersionAssessments != nil {
		in, out := &in.VersionAssessments, &out.VersionAssessments
		*out = new(VersionAssessmentAnalysis)
		(*in).DeepCopyInto(*out)
	}
	if in.Weights != nil {
		in, out := &in.Weights, &out.Weights
		*out = new(WeightsAnalysis)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Analysis.
func (in *Analysis) DeepCopy() *Analysis {
	if in == nil {
		return nil
	}
	out := new(Analysis)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AnalysisMetaData) DeepCopyInto(out *AnalysisMetaData) {
	*out = *in
	in.Timestamp.DeepCopyInto(&out.Timestamp)
	if in.Message != nil {
		in, out := &in.Message, &out.Message
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AnalysisMetaData.
func (in *AnalysisMetaData) DeepCopy() *AnalysisMetaData {
	if in == nil {
		return nil
	}
	out := new(AnalysisMetaData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in BooleanList) DeepCopyInto(out *BooleanList) {
	{
		in := &in
		*out = make(BooleanList, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BooleanList.
func (in BooleanList) DeepCopy() BooleanList {
	if in == nil {
		return nil
	}
	out := new(BooleanList)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Criteria) DeepCopyInto(out *Criteria) {
	*out = *in
	if in.RequestCount != nil {
		in, out := &in.RequestCount, &out.RequestCount
		*out = new(string)
		**out = **in
	}
	if in.Reward != nil {
		in, out := &in.Reward, &out.Reward
		*out = new(Reward)
		**out = **in
	}
	if in.Indicators != nil {
		in, out := &in.Indicators, &out.Indicators
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Objectives != nil {
		in, out := &in.Objectives, &out.Objectives
		*out = make([]Objective, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Criteria.
func (in *Criteria) DeepCopy() *Criteria {
	if in == nil {
		return nil
	}
	out := new(Criteria)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Duration) DeepCopyInto(out *Duration) {
	*out = *in
	if in.Interval != nil {
		in, out := &in.Interval, &out.Interval
		*out = new(string)
		**out = **in
	}
	if in.MaxIterations != nil {
		in, out := &in.MaxIterations, &out.MaxIterations
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Duration.
func (in *Duration) DeepCopy() *Duration {
	if in == nil {
		return nil
	}
	out := new(Duration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Experiment) DeepCopyInto(out *Experiment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Experiment.
func (in *Experiment) DeepCopy() *Experiment {
	if in == nil {
		return nil
	}
	out := new(Experiment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Experiment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExperimentCondition) DeepCopyInto(out *ExperimentCondition) {
	*out = *in
	if in.LastTransitionTime != nil {
		in, out := &in.LastTransitionTime, &out.LastTransitionTime
		*out = (*in).DeepCopy()
	}
	if in.Reason != nil {
		in, out := &in.Reason, &out.Reason
		*out = new(string)
		**out = **in
	}
	if in.Message != nil {
		in, out := &in.Message, &out.Message
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExperimentCondition.
func (in *ExperimentCondition) DeepCopy() *ExperimentCondition {
	if in == nil {
		return nil
	}
	out := new(ExperimentCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExperimentList) DeepCopyInto(out *ExperimentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Experiment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExperimentList.
func (in *ExperimentList) DeepCopy() *ExperimentList {
	if in == nil {
		return nil
	}
	out := new(ExperimentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExperimentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExperimentSpec) DeepCopyInto(out *ExperimentSpec) {
	*out = *in
	if in.Target != nil {
		in, out := &in.Target, &out.Target
		*out = new(string)
		**out = **in
	}
	if in.VersionInfo != nil {
		in, out := &in.VersionInfo, &out.VersionInfo
		*out = new(VersionInfo)
		(*in).DeepCopyInto(*out)
	}
	in.Strategy.DeepCopyInto(&out.Strategy)
	if in.Criteria != nil {
		in, out := &in.Criteria, &out.Criteria
		*out = new(Criteria)
		(*in).DeepCopyInto(*out)
	}
	if in.Duration != nil {
		in, out := &in.Duration, &out.Duration
		*out = new(Duration)
		(*in).DeepCopyInto(*out)
	}
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = new(map[string]Metric)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]Metric, len(*in))
			for key, val := range *in {
				(*out)[key] = *val.DeepCopy()
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExperimentSpec.
func (in *ExperimentSpec) DeepCopy() *ExperimentSpec {
	if in == nil {
		return nil
	}
	out := new(ExperimentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExperimentStatus) DeepCopyInto(out *ExperimentStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]*ExperimentCondition, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ExperimentCondition)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.InitTime != nil {
		in, out := &in.InitTime, &out.InitTime
		*out = (*in).DeepCopy()
	}
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.EndTime != nil {
		in, out := &in.EndTime, &out.EndTime
		*out = (*in).DeepCopy()
	}
	if in.LastUpdateTime != nil {
		in, out := &in.LastUpdateTime, &out.LastUpdateTime
		*out = (*in).DeepCopy()
	}
	if in.CurrentIteration != nil {
		in, out := &in.CurrentIteration, &out.CurrentIteration
		*out = new(int32)
		**out = **in
	}
	if in.CurrentWeightDistribution != nil {
		in, out := &in.CurrentWeightDistribution, &out.CurrentWeightDistribution
		*out = make([]WeightData, len(*in))
		copy(*out, *in)
	}
	if in.Analysis != nil {
		in, out := &in.Analysis, &out.Analysis
		*out = new(Analysis)
		(*in).DeepCopyInto(*out)
	}
	if in.RecommendedBaseline != nil {
		in, out := &in.RecommendedBaseline, &out.RecommendedBaseline
		*out = new(string)
		**out = **in
	}
	if in.Message != nil {
		in, out := &in.Message, &out.Message
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExperimentStatus.
func (in *ExperimentStatus) DeepCopy() *ExperimentStatus {
	if in == nil {
		return nil
	}
	out := new(ExperimentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Handlers) DeepCopyInto(out *Handlers) {
	*out = *in
	if in.Start != nil {
		in, out := &in.Start, &out.Start
		*out = new(string)
		**out = **in
	}
	if in.Finish != nil {
		in, out := &in.Finish, &out.Finish
		*out = new(string)
		**out = **in
	}
	if in.Rollback != nil {
		in, out := &in.Rollback, &out.Rollback
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Handlers.
func (in *Handlers) DeepCopy() *Handlers {
	if in == nil {
		return nil
	}
	out := new(Handlers)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Metric) DeepCopyInto(out *Metric) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Metric.
func (in *Metric) DeepCopy() *Metric {
	if in == nil {
		return nil
	}
	out := new(Metric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Metric) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricList) DeepCopyInto(out *MetricList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Metric, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricList.
func (in *MetricList) DeepCopy() *MetricList {
	if in == nil {
		return nil
	}
	out := new(MetricList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MetricList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricReference) DeepCopyInto(out *MetricReference) {
	*out = *in
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricReference.
func (in *MetricReference) DeepCopy() *MetricReference {
	if in == nil {
		return nil
	}
	out := new(MetricReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricSpec) DeepCopyInto(out *MetricSpec) {
	*out = *in
	if in.Params != nil {
		in, out := &in.Params, &out.Params
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.Description != nil {
		in, out := &in.Description, &out.Description
		*out = new(string)
		**out = **in
	}
	if in.Units != nil {
		in, out := &in.Units, &out.Units
		*out = new(string)
		**out = **in
	}
	if in.SampleSize != nil {
		in, out := &in.SampleSize, &out.SampleSize
		*out = new(MetricReference)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricSpec.
func (in *MetricSpec) DeepCopy() *MetricSpec {
	if in == nil {
		return nil
	}
	out := new(MetricSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricStatus) DeepCopyInto(out *MetricStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricStatus.
func (in *MetricStatus) DeepCopy() *MetricStatus {
	if in == nil {
		return nil
	}
	out := new(MetricStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Objective) DeepCopyInto(out *Objective) {
	*out = *in
	if in.UpperLimit != nil {
		in, out := &in.UpperLimit, &out.UpperLimit
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.LowerLimit != nil {
		in, out := &in.LowerLimit, &out.LowerLimit
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.RollbackOnFailure != nil {
		in, out := &in.RollbackOnFailure, &out.RollbackOnFailure
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Objective.
func (in *Objective) DeepCopy() *Objective {
	if in == nil {
		return nil
	}
	out := new(Objective)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Reward) DeepCopyInto(out *Reward) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Reward.
func (in *Reward) DeepCopy() *Reward {
	if in == nil {
		return nil
	}
	out := new(Reward)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Strategy) DeepCopyInto(out *Strategy) {
	*out = *in
	if in.Handlers != nil {
		in, out := &in.Handlers, &out.Handlers
		*out = new(Handlers)
		(*in).DeepCopyInto(*out)
	}
	if in.Weights != nil {
		in, out := &in.Weights, &out.Weights
		*out = new(Weights)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Strategy.
func (in *Strategy) DeepCopy() *Strategy {
	if in == nil {
		return nil
	}
	out := new(Strategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionAssessmentAnalysis) DeepCopyInto(out *VersionAssessmentAnalysis) {
	*out = *in
	in.AnalysisMetaData.DeepCopyInto(&out.AnalysisMetaData)
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make(map[string]BooleanList, len(*in))
		for key, val := range *in {
			var outVal []bool
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make(BooleanList, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionAssessmentAnalysis.
func (in *VersionAssessmentAnalysis) DeepCopy() *VersionAssessmentAnalysis {
	if in == nil {
		return nil
	}
	out := new(VersionAssessmentAnalysis)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionDetail) DeepCopyInto(out *VersionDetail) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.WeightObjRef != nil {
		in, out := &in.WeightObjRef, &out.WeightObjRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionDetail.
func (in *VersionDetail) DeepCopy() *VersionDetail {
	if in == nil {
		return nil
	}
	out := new(VersionDetail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionInfo) DeepCopyInto(out *VersionInfo) {
	*out = *in
	in.Baseline.DeepCopyInto(&out.Baseline)
	if in.Candidates != nil {
		in, out := &in.Candidates, &out.Candidates
		*out = make([]VersionDetail, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionInfo.
func (in *VersionInfo) DeepCopy() *VersionInfo {
	if in == nil {
		return nil
	}
	out := new(VersionInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WeightData) DeepCopyInto(out *WeightData) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WeightData.
func (in *WeightData) DeepCopy() *WeightData {
	if in == nil {
		return nil
	}
	out := new(WeightData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Weights) DeepCopyInto(out *Weights) {
	*out = *in
	if in.MaxCandidateWeight != nil {
		in, out := &in.MaxCandidateWeight, &out.MaxCandidateWeight
		*out = new(int32)
		**out = **in
	}
	if in.MaxCandidateWeightIncrement != nil {
		in, out := &in.MaxCandidateWeightIncrement, &out.MaxCandidateWeightIncrement
		*out = new(int32)
		**out = **in
	}
	if in.Algorithm != nil {
		in, out := &in.Algorithm, &out.Algorithm
		*out = new(AlgorithmType)
		**out = **in
	}
	if in.WeightDistribution != nil {
		in, out := &in.WeightDistribution, &out.WeightDistribution
		*out = make([]int32, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Weights.
func (in *Weights) DeepCopy() *Weights {
	if in == nil {
		return nil
	}
	out := new(Weights)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WeightsAnalysis) DeepCopyInto(out *WeightsAnalysis) {
	*out = *in
	in.AnalysisMetaData.DeepCopyInto(&out.AnalysisMetaData)
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make([]WeightData, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WeightsAnalysis.
func (in *WeightsAnalysis) DeepCopy() *WeightsAnalysis {
	if in == nil {
		return nil
	}
	out := new(WeightsAnalysis)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WinnerAssessmentAnalysis) DeepCopyInto(out *WinnerAssessmentAnalysis) {
	*out = *in
	in.AnalysisMetaData.DeepCopyInto(&out.AnalysisMetaData)
	in.Data.DeepCopyInto(&out.Data)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WinnerAssessmentAnalysis.
func (in *WinnerAssessmentAnalysis) DeepCopy() *WinnerAssessmentAnalysis {
	if in == nil {
		return nil
	}
	out := new(WinnerAssessmentAnalysis)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WinnerAssessmentData) DeepCopyInto(out *WinnerAssessmentData) {
	*out = *in
	if in.Winner != nil {
		in, out := &in.Winner, &out.Winner
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WinnerAssessmentData.
func (in *WinnerAssessmentData) DeepCopy() *WinnerAssessmentData {
	if in == nil {
		return nil
	}
	out := new(WinnerAssessmentData)
	in.DeepCopyInto(out)
	return out
}
