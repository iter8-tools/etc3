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
	"context"
	"strings"

	"github.com/iter8-tools/etc3/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetRequiredMetrics returns a list of the metrics required by the criteria
func (s *ExperimentSpec) GetRequiredMetrics() []string {
	if s.Criteria == nil {
		return make([]string, 0)
	}

	metricsmap := make(map[string]struct{})
	metricsmap[*s.GetRequestCount()] = struct{}{}
	for _, i := range s.Criteria.Indicators {
		metricsmap[i] = struct{}{}
	}
	for _, o := range s.Criteria.Objectives {
		metricsmap[o.Metric] = struct{}{}
	}

	metrics := make([]string, 0, len(metricsmap))
	for k := range metricsmap {
		metrics = append(metrics, k)
	}

	return metrics
}

// ReadMetric reads a metric from the cluster using the name as the
// If the name is of the form "namespace/name", look in namespace for name.
// Otherwise look for name. If not found, look in util.Iter8InstallNamespace() for name.
// If not found returnd NotFound error
func ReadMetric(ctx context.Context, c client.Client, namespace string, name string) (*Metric, error) {
	// if name has a "/" then split into namespace / name
	explicitNamespaceProvided := false
	splt := strings.Split(name, "/")
	if len(splt) == 2 {
		explicitNamespaceProvided = true
		namespace = splt[0]
		name = splt[1]
	}

	metric := &Metric{}
	err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, metric)
	if err != nil {
		if errors.IsNotFound(err) && !explicitNamespaceProvided {
			// can we try a different namespace?
			namespace = util.GetIter8InstallNamespace()
			err = c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, metric)
		}
	}
	if err != nil {
		// could not read metric; skip it
		return nil, err
	}

	return metric, nil
}

// func (s *ExperimentSpec) ReadMetrics(ctx context.Context, c client.Client, namespace string) {

// 	success := true

// 	for _, m := range metrics {
// 		m.Name, err := ReadMetric (ctx, c, namespace, name)
// 		// do we already have this one?
// 		if name
// 		if err != nil {
// 			success = false
// 			message = err.Error()
// 		} else {
// 			//
// 		}
// 	}
// 	if success {
// 		// set Condition True
// 		// set Reason AllMetricsRead
// 		// set Message all metics found
// 	}
// 	// set Condition False
// 	// set Reason MetricNotFound
// 	// set Message message

// }
