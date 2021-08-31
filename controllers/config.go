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

package controllers

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Iter8Config describes structure of configuration file
type Iter8Config struct {
	AnalyticsEndpoint string `json:"analyticsEndpoint" yaml:"analyticsEndpoint" envconfig:"ITER8_ANALYTICS_ENDPOINT"`
	Namespace         string `json:"namespace" yaml:"namespace" envconfig:"ITER8_NAMESPACE"`
	TaskRunnerImage   string `json:"taskRunnerImage" yaml:"taskRunnerImage" envconfig:"ITER8_TASKRUNNER_IMAGE"`
}

// ReadConfig reads the configuration from a combination of files and the environment
// In our case, no config file is provided; so read from environment.
func ReadConfig(cfg *Iter8Config) error {
	if err := envconfig.Process("", cfg); err != nil {
		return err
	}

	// overwrite AnalyticsEndpoint if it has the string "ITER8_NAMESPACE" in the value
	cfg.AnalyticsEndpoint = strings.Replace(cfg.AnalyticsEndpoint, "ITER8_NAMESPACE", cfg.Namespace, 1)

	return nil
}

// Iter8ConfigBuilder type for building new Iter8Config by hand. Used for testing.
type Iter8ConfigBuilder Iter8Config

// NewIter8Config returns a new config builder
func NewIter8Config() Iter8ConfigBuilder {
	cfg := Iter8Config{}
	return (Iter8ConfigBuilder)(cfg)
}

// WithEndpoint adds an endpoint to an Iter8Config. Used for testing.
func (b Iter8ConfigBuilder) WithEndpoint(endpoint string) Iter8ConfigBuilder {
	b.AnalyticsEndpoint = endpoint
	return b
}

// WithNamespace adds a namespace to an Iter8Config. Used for testing.
func (b Iter8ConfigBuilder) WithNamespace(namespace string) Iter8ConfigBuilder {
	b.Namespace = namespace
	return b
}

// WithTaskRunnerImage adds a task runner image to an Iter8Config. Used for testing.
func (b Iter8ConfigBuilder) WithTaskRunnerImage(taskRunnerImage string) Iter8ConfigBuilder {
	b.TaskRunnerImage = taskRunnerImage
	return b
}

// Build creates an Iter8Config from using builder pattern. Used for testing.
func (b Iter8ConfigBuilder) Build() Iter8Config {
	return (Iter8Config)(b)
}
