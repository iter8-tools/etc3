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

// config.go - methods to support iter8 install time configuration options

package configuration

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Iter8Config describes structure of configuration file
type Iter8Config struct {
	ExperimentTypes []ExperimentType `yaml:"experimentTypes"`
	Analytics       `json:"analytics" yaml:"analytics"`
	Metrics         `json:"metrics" yaml:"metrics"`
	Namespace       string `envconfig:"MY_POD_NAMESPACE"`
}

// ExperimentType is list of handlers for each supported experiment type
type ExperimentType struct {
	Name     string `yaml:"name"`
	Handlers `yaml:"handlers"`
}

// Handlers is list of default handlers
type Handlers struct {
	Start    string `yaml:"start"`
	Rollback string `yaml:"rollback"`
	Finish   string `yaml:"finish"`
	Failure  string `yaml:"failure"`
}

// Analytics captures details of analytics endpoint(s)
type Analytics struct {
	Endpoint string `yaml:"endpoint" envconfig:"ITER8_ANALYTICS_ENDPOINT"`
}

// Metrics identifies the metric that should be used to count requests
type Metrics struct {
	RequestCount string `yaml:"requestCount" envconfig:"REQUEST_COUNT"`
}

// ReadConfig reads the configuration from a combination of files and the environment
func ReadConfig(cfg *Iter8Config) error {
	file, err := os.Open("default.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(cfg); err != nil {
		return err
	}

	if err = envconfig.Process("", cfg); err != nil {
		return err
	}
	return err
}
