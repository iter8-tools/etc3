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

// weights.go - logic to redistribute weights in domain objects using dynamic client
// derived from example at https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go

package controllers

import (
	"context"
	"encoding/json"
	"errors"

	v2alpha1 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/etc3/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

func (r *ExperimentReconciler) redistributeWeight(ctx context.Context, instance *v2alpha1.Experiment) error {
	log := util.Logger(ctx)
	log.Info("redistributeWeight() called")
	defer log.Info("redistributeWeight() ended")

	if versionInfo := instance.Spec.VersionInfo; versionInfo == nil {
		// should never get here; should have been validated before this
		log.Info("Cannot redistribute weight; no version information present.")
		return nil
	}

	// TODO collect multiple changes for each object
	if err := r.updateWeightForVersion(ctx, instance, instance.Spec.VersionInfo.Baseline); err != nil {
		return err
	}
	for _, version := range instance.Spec.VersionInfo.Candidates {
		if err := r.updateWeightForVersion(ctx, instance, version); err != nil {
			return err
		}
	}

	// set status.currentWeightDistribution to match set weights
	// for now copy from status.analysis.weights
	instance.Status.CurrentWeightDistribution = make([]v2alpha1.WeightData, len(instance.Status.Analysis.Weights.Data))
	for i, w := range instance.Status.Analysis.Weights.Data {
		instance.Status.CurrentWeightDistribution[i] = w
	}
	return nil
}

func (r *ExperimentReconciler) updateWeightForVersion(ctx context.Context, instance *v2alpha1.Experiment, version v2alpha1.VersionDetail) error {
	log := util.Logger(ctx)

	// n-1 versions should have a weightObjRef.fieldPath; should be caught by validation
	if version.WeightObjRef == nil {
		log.Info("Unable to update weight; no weightObjectReference", "version", version)
		return nil
	}
	if version.WeightObjRef.FieldPath == "" {
		log.Info("Unable to update weight; no field specified", "version", version)
		return nil
	}

	weight := getWeightRecommendation(version.Name, instance.Status.Analysis.Weights.Data)
	if weight == nil {
		log.Info("Unable to find weight recommendation.", "version", version, "weights", instance.Status.Analysis.Weights)
		// fatal error; expected a weight recommendation for all versions
		return errors.New("Unable to find weight recommendation")
	}
	_, err := r.patchWeight(ctx, version.WeightObjRef, *weight)
	if err != nil {
		log.Error(err, "Failed to update weight", "version", version)
		return err
	}

	return nil
}

func getWeightRecommendation(version string, weights []v2alpha1.WeightData) *int32 {
	for _, w := range weights {
		if w.Name == version {
			weight := w.Value
			return &weight
		}
	}
	return nil
}

func getDynamicResourceInterface(cfg *rest.Config, objRef *corev1.ObjectReference) (dynamic.ResourceInterface, error) {
	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	gvk := schema.FromAPIVersionAndKind(objRef.APIVersion, objRef.Kind)

	// 3. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	// 4. Obtain REST interface for the GVR
	namespace := objRef.Namespace
	if namespace == "" {
		namespace = "default"
	}
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	return dr, nil
}

type patchIntValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value int32  `json:"value"`
}

func (r *ExperimentReconciler) patchWeight(ctx context.Context, objRef *corev1.ObjectReference, weight int32) (*unstructured.Unstructured, error) {
	log := util.Logger(ctx)
	log.Info("patchWeight() called")
	defer log.Info("patchWeight() ended")

	data, err := json.Marshal([]patchIntValue{{
		Op:    "add",
		Path:  objRef.FieldPath,
		Value: weight,
	}})
	if err != nil {
		log.Error(err, "Unable to create JSON patch command")
		return nil, err
	}
	log.Info("marshalled patch", "patch", string(data))

	dr, err := getDynamicResourceInterface(r.RestConfig, objRef)
	if err != nil {
		log.Error(err, "Unable to get dynamic resource interface")
		return nil, err
	}

	return dr.Patch(ctx, objRef.Name, types.JSONPatchType, data, metav1.PatchOptions{FieldManager: "etc3"})
}
