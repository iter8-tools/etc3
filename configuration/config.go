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
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Iter8Config describes structure of configuration file
type Iter8Config struct {
	Analytics   `json:"analytics" yaml:"analytics"`
	Namespace   string `envconfig:"ITER8_NAMESPACE"`
	HandlersDir string `envconfig:"HANDLERS_DIR"`
}

// Analytics captures details of analytics endpoint(s)
type Analytics struct {
	Endpoint string `yaml:"endpoint" envconfig:"ITER8_ANALYTICS_ENDPOINT"`
}

// ReadConfig reads the configuration from a combination of files and the environment
func ReadConfig(cfg *Iter8Config) error {
	if err := envconfig.Process("", cfg); err != nil {
		return err
	}

	cfg.Analytics.Endpoint = strings.Replace(cfg.Analytics.Endpoint, "ITER8_NAMESPACE", cfg.Namespace, 1)

	return nil
}

// Iter8ConfigBuilder type for building new config by hand
type Iter8ConfigBuilder Iter8Config

// NewIter8Config returns a new config builder
func NewIter8Config() Iter8ConfigBuilder {
	cfg := Iter8Config{}
	return (Iter8ConfigBuilder)(cfg)
}

// WithEndpoint ..
func (b Iter8ConfigBuilder) WithEndpoint(endpoint string) Iter8ConfigBuilder {
	b.Analytics.Endpoint = endpoint
	return b
}

// WithNamespace ..
func (b Iter8ConfigBuilder) WithNamespace(namespace string) Iter8ConfigBuilder {
	b.Namespace = namespace
	return b
}

// WithHandlersDir ..
func (b Iter8ConfigBuilder) WithHandlersDir(handlersDir string) Iter8ConfigBuilder {
	b.HandlersDir = handlersDir
	return b
}

// Build ..
func (b Iter8ConfigBuilder) Build() Iter8Config {
	return (Iter8Config)(b)
}
