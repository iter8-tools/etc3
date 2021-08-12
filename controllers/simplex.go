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

// handlers.go implements code to start jobs

package controllers

import (
	"context"
	"io/ioutil"
	"path"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// SimplexYaml is the name of the YAML file that describes the GVKs to be watched
	SimplexYaml = "simplex.yaml"
	// SimplexAnnotationPrefix is the prefix used by all simplex annotation keys
	SimplexAnnotationPrefix = "iter8.tools/simplex."
)

// SimplexGVKs describes the object kinds to be watched by simplex controllers
type SimplexGVKs struct {
	GVKs []schema.GroupVersionKind `json:"gvks,omitempty" yaml:"gvks,omitempty"`
}

// readSimplexGVKs reads a simplex file spec to a SimplexGVKs struct
func readSimplexGVKs(simplexFile string, simplex *SimplexGVKs) error {
	yamlFile, err := ioutil.ReadFile(simplexFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, simplex); err == nil {
		return err
	}

	return nil
}

// Simplex holds contextual info used by a Simplex controller during reconcile
type Simplex struct {
	GVK           schema.GroupVersionKind
	Log           logr.Logger
	EventRecorder record.EventRecorder
}

// GetSimplexes returns a slice of Simplexes
func GetSimplexes(cfg *Iter8Config) ([]Simplex, error) {
	log := GetLogger()
	log.Info("calling GetSimplexes")
	defer log.Info("completing GetSimplexes")

	if cfg == nil || cfg.SimplexDir == "" {
		log.Info("no simplex configuration found")
		return nil, nil
	}

	var sr SimplexGVKs

	sy := path.Join(cfg.SimplexDir, SimplexYaml)
	err := readSimplexGVKs(sy, &sr)
	if err != nil {
		return nil, err
	}

	log.Info("object kinds", sr.GVKs)

	simplexes := []Simplex{}
	for _, gvk := range sr.GVKs {
		simplexes = append(simplexes, Simplex{
			GVK: gvk,
		})
	}
	return simplexes, nil
}

// Reconcile attempts to reconcile an object
// Implementation based on https://firehydrant.io/blog/dynamic-kubernetes-informers/
func (s *Simplex) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := s.Log.WithValues("simplex object", req.NamespacedName).WithValues("kind", s.GVK.Kind)

	log.Info("simplex reconcile called")
	defer log.Info("simplex reconcile completed")

	// get the object
	// if no such object, proceed to helm uninstall below

	// if there is an object
	// a := object.GetAnnotations()
	// var sa map[string]string = getSimplexAnnotations(a)
	// if validSimplexAnnotations(sa) {
	// // valid simplex.values.candId annotation
	// // valid simplex.values.target annotation
	// // helm upgrade --install
	// // // release name is hashed from simplex.values.target
	// // // experiment name suffix is hashed from simplex.values.candId and simplex.values.target
	// // // release namespace equals simplex resource namespace
	// }
	// if invalid or error or no such object {
	// // helm uninstall release --ignore errors
	// }
	// create iter8ctl assertable logs every step of the way
	// create events using event recorder
	return ctrl.Result{}, nil
}

//getSimplexAnnotations returns the subset of annotations relevant to simplex
func getSimplexAnnotations(a map[string]string) map[string]string {
	m := map[string]string{}
	for k, v := range a {
		if strings.HasPrefix(k, SimplexAnnotationPrefix) {
			m[k] = v
		}
	}
	return m
}

// SetupWithManager is the method called when setting up the simplex reconciler with the controller manager.
func (s *Simplex) SetupWithManager(mgr ctrl.Manager) error {

	predicateFuncs := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			// return true if the object has simplex annotation
			a := e.Object.GetAnnotations()
			var sa map[string]string = getSimplexAnnotations(a)
			return len(sa) > 0
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			a := e.ObjectOld.GetAnnotations()
			var sa map[string]string = getSimplexAnnotations(a)
			b := e.ObjectNew.GetAnnotations()
			var sb map[string]string = getSimplexAnnotations(b)

			// nothing to do, if no change in simplex annotations
			return !reflect.DeepEqual(sa, sb)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
	}

	u := unstructured.Unstructured{}
	u.SetGroupVersionKind(s.GVK)
	return ctrl.NewControllerManagedBy(mgr).
		Watches(&source.Kind{Type: &batchv1.Job{}},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicateFuncs)).
		Complete(s)
}
