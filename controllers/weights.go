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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yalp/jsonpath"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

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

// derived from example at https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
func getObj(ctx context.Context, cfg *rest.Config, objRef *corev1.ObjectReference) (*unstructured.Unstructured, error) {
	dr, err := getDynamicResourceInterface(cfg, objRef)
	if err != nil {
		return nil, err
	}

	fmt.Printf("calling Get(%s)\n\n", objRef.Name)
	// result, err := dyn.Resource(res).Namespace(namespace).Get(ctx, objRef.Name, metav1.GetOptions{})
	result, err := dr.Get(ctx, objRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	fmt.Printf("result: %+v\n\n", result)

	// Patch
	// data, err := json.Marshal([]patchStringValue{{
	// 	Op:    "replace",
	// 	Path:  "/spec/type",
	// 	Value: "gauge",
	// }})
	// if err != nil {
	// 	return result, err
	// }
	// fmt.Printf("data = %s\n\n", data)
	// fmt.Printf("name = %s\n\n", objRef.Name)

	// x, err := dr.Patch(ctx, objRef.Name, types.JSONPatchType, data, metav1.PatchOptions{FieldManager: "etc3"})
	// if err != nil {
	// 	return result, err
	// }
	// fmt.Printf("x = %v\n\n", x)

	// // 7. Create or Update the object with SSA
	// //     types.ApplyPatchType indicates SSA.
	// //     FieldManager specifies the field owner ID.
	// _, err = dr.Patch(ctx, objRef.Name, types.ApplyPatchType, data, metav1.PatchOptions{
	// 	FieldManager: "sample-controller",
	// })

	// pretty print the result
	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", "    ")
	// enc.Encode(result)

	return result, nil
}

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func patchObjectDynamic(ctx context.Context, cfg *rest.Config, objRef *corev1.ObjectReference) (*unstructured.Unstructured, error) {
	// dr, err := getDynamicResourceInterface(cfg, objRef)
	// if err != nil {
	// 	return nil, err
	// }

	// data, err := json.Marshal([]patchStringValue{{
	// 	Op:    "replace",
	// 	Path:  "/spec/type",
	// 	Value: "gauge",
	// }})

	// result, err := dr.Patch(ctx, objRef.Name, types.JSONPatchType, data, metav1.PatchOptions{FieldManager: "etc3"})
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Printf("patch result = %v\n\n", result)

	// return result, nil
	return nil, nil
}

func getObjectDynamic(config *rest.Config, objRef *corev1.ObjectReference) (*unstructured.Unstructured, error) {
	obj, err := getObj(context.Background(), config, objRef)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("obj: %+v\n\n", obj)
	return obj, err
}

func getValueDynamic(config *rest.Config, objRef *corev1.ObjectReference) interface{} {
	obj, _ := getObjectDynamic(config, objRef)
	if obj == nil {
		return nil
	}
	resultJson, err := obj.MarshalJSON()
	if err != nil {
		return nil
	}
	fmt.Printf("resultJson = %s\n\n", resultJson)

	resultObj := make(map[string]interface{})
	err = json.Unmarshal(resultJson, &resultObj)
	if err != nil {
		return nil
	}
	// fmt.Printf("resultObj = %v\n\n", resultObj)

	path := fmt.Sprintf("$.%s", objRef.FieldPath)
	fmt.Printf("path = %s\n\n", path)
	value, err := jsonpath.Read(resultObj, path)
	if err != nil {
		return nil
	}
	return value
}
