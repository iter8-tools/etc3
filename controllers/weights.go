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

// weights.go - implements interactions with dynamic client to read and update weights of objects in spec.versionInfo

package controllers

import (
	"context"
	"encoding/json"

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

// weights.go - logic to redistribute weights in domain objects using dynamic client
// derived from example at https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go

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

	dr, err := getDynamicResourceInterface(r.Config, objRef)
	if err != nil {
		log.Error(err, "Unable to get dynamic resource interface")
		return nil, err
	}

	return dr.Patch(ctx, objRef.Name, types.JSONPatchType, data, metav1.PatchOptions{FieldManager: "etc3"})
}
