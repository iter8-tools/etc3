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

// metrics.go - methods to read metrics specified by criteria into experiment spec

package controllers

import (
	"context"
	"strings"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// ReadMetric reads a metric from the cluster using the name as the key
// If the name is of the form "namespace/name", look in namespace for name.
// Otherwise look for name. If not found, look in util.Iter8InstallNamespace() for name.
// If not found return NotFound error
func (r *ExperimentReconciler) ReadMetric(ctx context.Context, instance *v2alpha2.Experiment, name string, metricMap map[string]*v2alpha2.Metric) bool {
	key := name
	// default namespace to use is the experiment namespace
	namespace := instance.GetObjectMeta().GetNamespace()

	// If the metric name includes a "/" then use the prefix as the namespace
	splt := strings.Split(name, "/")
	if len(splt) == 2 {
		namespace = splt[0]
		name = splt[1]
	}

	// if we've already read the metric, then we don't need to proceed; just return true
	if _, ok := metricMap[key]; ok {
		return true
	}

	metric := &v2alpha2.Metric{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, metric)
	if err != nil {
		// could not read metric; record the problem and indicate that the read did not succeed
		if errors.IsNotFound(err) {
			r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonMetricUnavailable, "Unable to find metric %s", name)
		} else {
			r.recordExperimentFailed(ctx, instance, v2alpha2.ReasonMetricsUnreadable, "Unable to load metric %s", name)
		}
		return false // not ok
	}

	// add to the map
	metricMap[key] = metric

	// check if this metric references another metric. If so, read it too
	if metric.Spec.SampleSize != nil {
		referencedNamespace := metric.GetObjectMeta().GetNamespace()
		if metric.Spec.SampleSize.Namespace != nil {
			referencedNamespace = *metric.Spec.SampleSize.Namespace
		}
		return r.ReadMetric(ctx, instance, referencedNamespace+"/"+metric.Spec.SampleSize.Name, metricMap)
	}

	// must be ok
	return true
}

// ReadMetrics reads needed metrics from cluster and caches them in the experiment
// result is false if an error occurred reading metrics
func (r *ExperimentReconciler) ReadMetrics(ctx context.Context, instance *v2alpha2.Experiment) bool {
	log := util.Logger(ctx)
	log.Info("ReadMetrics called")
	defer log.Info("ReadMetrics completed")

	criteria := instance.Spec.Criteria
	if len(instance.Spec.Metrics) > 0 || criteria == nil {
		return true
	}

	metricsCache := make(map[string]*v2alpha2.Metric)

	// name of request counter
	requestCount := instance.Spec.GetRequestCount(r.Iter8Config)
	if ok := r.ReadMetric(ctx, instance, *requestCount, metricsCache); !ok {
		return ok
	}

	// name of reward, if any
	reward := instance.Spec.GetReward()
	if reward != nil {
		if ok := r.ReadMetric(ctx, instance, reward.Metric, metricsCache); !ok {
			return ok
		}
	}

	// indicators
	for _, indicator := range criteria.Indicators {
		if metricsCache[indicator] == nil {
			if ok := r.ReadMetric(ctx, instance, indicator, metricsCache); !ok {
				return ok
			}
		}
	}

	for _, objective := range criteria.Objectives {
		if metricsCache[objective.Metric] == nil {
			if ok := r.ReadMetric(ctx, instance, objective.Metric, metricsCache); !ok {
				return ok
			}
		}
	}

	// found all metrics; copy into instance.Spec
	for name, obj := range metricsCache {
		instance.Spec.Metrics = append(instance.Spec.Metrics,
			v2alpha2.MetricInfo{Name: name, MetricObj: *obj})
	}
	return true
}
