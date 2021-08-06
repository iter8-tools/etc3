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

	"github.com/ghodss/yaml"
)

const (
	// SimplexYaml is the name of the yaml file that describes the resources to be watched
	SimplexYaml = "simplex.yaml"
)

/*
An example SimpleX spec might look as follows.
// namespaces: # optional; if a list is not specified, all namespaces will be watched
// namespace support is todo
resources: # list of resources to watch; required;
- deployments.v1.apps
- configmaps
- mutatingwebhookconfigurations.v1beta1.admissionregistration.k8s.io
*/

type Simplex struct {
	Resources []string `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// readSimplexSpec reads a simplex file spec to a Simplex object
func readSimplexSpec(simplexFile string, simplex *Simplex) error {
	yamlFile, err := ioutil.ReadFile(simplexFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, simplex); err == nil {
		return err
	}

	return nil
}

// LaunchSimplex lauches all the Simplex watches
func LaunchSimplex(ctx context.Context, cfg *Iter8Config) error {
	log := GetLogger()
	log.Info("LaunchSimplex called")
	defer log.Info("LaunchSimplex completed")

	if cfg == nil || cfg.SimplexDir == "" {
		log.Info("simplex is not configured")
		return nil
	}

	var simplex Simplex

	sy := path.Join(cfg.SimplexDir, SimplexYaml)
	err := readSimplexSpec(sy, &simplex)
	if err != nil {
		return err
	}

	return nil
}
