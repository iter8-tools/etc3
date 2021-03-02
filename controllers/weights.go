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
	"math"
	"strings"

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

func shouldRedistribute(instance *v2alpha1.Experiment) bool {
	experimentType := instance.Spec.Strategy.TestingPattern
	if experimentType == v2alpha1.TestingPatternConformance {
		return false
	}
	algorithm := instance.Spec.GetDeploymentPattern()
	if algorithm == v2alpha1.DeploymentPatternFixedSplit {
		return false
	}
	return true
}

func redistributeWeight(ctx context.Context, instance *v2alpha1.Experiment, restCfg *rest.Config) error {
	log := util.Logger(ctx)
	log.Info("redistributeWeight called")
	defer log.Info("redistributeWeight ended")

	if !shouldRedistribute(instance) {
		log.Info("No weight redistribution", "strategy", instance.Spec.Strategy.TestingPattern, "algorithm", instance.Spec.GetDeploymentPattern())
		return nil
	}

	// Get spec.versionInfo; it should be present by now
	if versionInfo := instance.Spec.VersionInfo; versionInfo == nil {
		return errors.New("Cannot redistribute weight; no version information present")
	}

	// For each version, get the patch to apply
	// Add to a map of Object --> []patchIntValue
	// Map keys are the kubernetes objects to be modified; values are a list of patches to apply
	patches := map[corev1.ObjectReference][]patchIntValue{}
	if err := addPatch(ctx, instance, instance.Spec.VersionInfo.Baseline, &patches); err != nil {
		return err
	}
	for _, version := range instance.Spec.VersionInfo.Candidates {
		if err := addPatch(ctx, instance, version, &patches); err != nil {
			return err
		}
	}

	// go through map and apply the list of patches to the objects
	for obj, p := range patches {
		_, err := patchWeight(ctx, &obj, p, restCfg)
		log.Info("redistributeWeight", "err", err)
		if err != nil {
			log.Error(err, "Unable to patch", "object", obj, "patch", p)
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

func addPatch(ctx context.Context, instance *v2alpha1.Experiment, version v2alpha1.VersionDetail, patcheMap *map[corev1.ObjectReference][]patchIntValue) error {
	log := util.Logger(ctx)
	//log.Info("addPatch called", "weight recommendations", instance.Status.Analysis.Weights)
	defer log.Info("addPatch completed")

	// verify that there is a weightObjRef; there might not be -- only n-1 versions MUST have one
	if version.WeightObjRef == nil {
		log.Info("Unable to update weight; no weightObjectReference", "version", version)
		return nil
	}
	// verify that the field path is present; again, it might not be -- only n-1 MUST be
	if version.WeightObjRef.FieldPath == "" {
		log.Info("Unable to update weight; no field specified", "version", version)
		return nil
	}

	// get the latest recommended weight from the analytics service (cached in Status)
	var weight *int32
	if instance.Status.Analysis != nil {
		weight = getWeightRecommendation(version.Name, instance.Status.Analysis.Weights.Data)
	}
	if weight == nil {
		log.Info("Unable to find weight recommendation.", "version", version)
		// fatal error; expected a weight recommendation for all versions
		return errors.New("No weight recommendation provided")
	}

	if *weight == *getCurrentWeight(version.Name, instance.Status.CurrentWeightDistribution) {
		log.Info("No change in weight distribution", "version", version.Name)
		return nil
	}

	// create patch
	patch := patchIntValue{
		Op:    "add",
		Path:  version.WeightObjRef.FieldPath,
		Value: *weight,
	}

	// add patch to patchMap
	key := getKey(*version.WeightObjRef)
	if patchList, ok := (*patcheMap)[key]; !ok {
		(*patcheMap)[key] = []patchIntValue{patch}
	} else {
		(*patcheMap)[key] = append(patchList, patch)
	}

	return nil
}

// key is just the obj without the FieldPath
func getKey(obj corev1.ObjectReference) corev1.ObjectReference {
	return corev1.ObjectReference{
		APIVersion: obj.APIVersion,
		Kind:       obj.Kind,
		Namespace:  obj.Namespace,
		Name:       obj.Name,
	}
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

func getCurrentWeight(version string, weights []v2alpha1.WeightData) *int32 {
	zero := int32(0)
	for _, weight := range weights {
		if weight.Name == version {
			return &weight.Value
		}
	}
	return &zero
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

func patchWeight(ctx context.Context, objRef *corev1.ObjectReference, patches []patchIntValue, restCfg *rest.Config) (*unstructured.Unstructured, error) {
	log := util.Logger(ctx)
	log.Info("patchWeight called")
	defer log.Info("patchWeight ended")

	data, err := json.Marshal(patches)
	if err != nil {
		log.Error(err, "Unable to create JSON patch command")
		return nil, err
	}
	log.Info("patchWeight", "marshalled patch", string(data))

	dr, err := getDynamicResourceInterface(restCfg, objRef)
	if err != nil {
		log.Error(err, "Unable to get dynamic resource interface")
		return nil, err
	}

	return dr.Patch(ctx, objRef.Name, types.JSONPatchType, data, metav1.PatchOptions{})
}

func observeWeight(ctx context.Context, objRef *corev1.ObjectReference, restCfg *rest.Config) (*int32, error) {
	log := util.Logger(ctx)
	log.Info("observeWeight called", "objRef", objRef)
	defer log.Info("observeWeight ended")

	dr, err := getDynamicResourceInterface(restCfg, objRef)
	if err != nil {
		log.Error(err, "Unable to get dynamic resource interface")
		return nil, err
	}

	// read object from cluster using unstructured client
	obj, err := dr.Get(ctx, objRef.Name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Unable to read object in cluster", "name", objRef.Name)
		return nil, err
	}
	log.Info("observeWeight", "referenced object", obj)

	// convert unstructured object to JSON object
	resultJSON, err := obj.MarshalJSON()
	if err != nil {
		log.Error(err, "Unable to convert resource to JSON object")
		return nil, err
	}
	log.Info("observeWeight", "as JSON", resultJSON)

	// convert JSON object to Go map
	resultObj := make(map[string]interface{})
	err = json.Unmarshal(resultJSON, &resultObj)
	if err != nil {
		log.Error(err, "Unable to parse JSON object")
		return nil, err
	}
	log.Info("observeWeight", "Go object", resultObj)

	if len(objRef.FieldPath) == 0 {
		log.Error(err, "Unable to read zero length field", "objRef", objRef, "obj", obj)
		return nil, errors.New("No fieldpath specified in referencing object")
	}

	path := objRef.FieldPath
	if objRef.FieldPath[0] == '/' {
		path = path[1:]

	}

	fieldPath := strings.Split(path, "/")
	log.Info("observeWeight", "path", objRef.FieldPath, "path", path)
	fValue, found, err := unstructured.NestedFloat64(resultObj, fieldPath...)
	if !found {
		log.Error(err, "Unable to read field; not found", "objRef", objRef, "obj", obj)
		return nil, errors.New("No such field")
	}
	if err != nil {
		log.Error(err, "Unable to read field; unexpected type", "objRef", objRef, "obj", obj)
		return nil, err
	}

	int32Value := int32(math.Round(fValue))

	return &int32Value, nil
}

func updateObservedWeights(ctx context.Context, instance *v2alpha1.Experiment, restCfg *rest.Config) {
	log := util.Logger(ctx)
	log.Info("updateObservedWeights called")
	defer log.Info("updateObservedWeights  ended")

	// cannot proceed if no version info
	if instance.Spec.VersionInfo == nil {
		return
	}

	observedWeights := make([]v2alpha1.WeightData, 0)
	missing := []string{}
	total := int32(0)

	// baseline
	b := instance.Spec.VersionInfo.Baseline
	if b.WeightObjRef != nil {
		w, err := observeWeight(ctx, b.WeightObjRef, restCfg)
		// if an error occurs, we ignore it (was logged in observeWeight())
		// it just means that no weight was observed for this version
		if err != nil {
			observedWeights = append(observedWeights, v2alpha1.WeightData{Name: b.Name, Value: *w})
			total += *w
		} else if missing == nil {
			missing = append(missing, b.Name)
		}
	}

	// candidates
	for _, c := range instance.Spec.VersionInfo.Candidates {
		if c.WeightObjRef != nil {
			w, err := observeWeight(ctx, c.WeightObjRef, restCfg)
			// if an error occurs, we ignore it (was logged in observeWeight())
			// it just means that no weight was observed for this version
			if err != nil {
				observedWeights = append(observedWeights, v2alpha1.WeightData{Name: c.Name, Value: *w})
				total += *w
			} else if missing == nil {
				missing = append(missing, c.Name)
			}
		}
	}

	// if there was one missing we can compute it; otherwise we'll leave gaps in the observed weights
	if len(missing) == 1 {
		log.Info("Computing weight", "missing", missing[0])
		w := int32(100) - total
		observedWeights = append(observedWeights, v2alpha1.WeightData{Name: missing[0], Value: w})
	} else if len(missing) > 1 {
		log.Info("Multiple weights could not be read from cluster", "missing", missing)
	}

	// assign list of observed weights
	instance.Status.CurrentWeightDistribution = observedWeights
}
