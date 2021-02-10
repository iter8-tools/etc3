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
	"strings"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Iter8Config describes structure of configuration file
type Iter8Config struct {
	ExperimentTypes []ExperimentType `yaml:"experimentTypes"`
	Analytics       `json:"analytics" yaml:"analytics"`
	Metrics         `json:"metrics" yaml:"metrics"`
	Namespace       string `envconfig:"ITER8_NAMESPACE"`
	HandlersDir     string `envconfig:"HANDLERS_DIR"`
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
func ReadConfig(configFile string, cfg *Iter8Config) error {
	file, err := os.Open(configFile)
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

	return nil
}

// Iter8ConfigBuilder type for building new config by hand
type Iter8ConfigBuilder Iter8Config

// NewIter8Config returns a new config builder
func NewIter8Config() Iter8ConfigBuilder {
	cfg := Iter8Config{}
	return (Iter8ConfigBuilder)(cfg)
}

// WithTestingPattern ..
func (b Iter8ConfigBuilder) WithTestingPattern(testingPattern string, handlers map[string]string) Iter8ConfigBuilder {
	s := ExperimentType{Name: testingPattern}
	for key, value := range handlers {
		hdlr := value
		switch strings.ToLower(key) {
		case "start":
			s.Handlers.Start = hdlr
		case "finish":
			s.Handlers.Finish = hdlr
		case "failure":
			s.Handlers.Failure = hdlr
		case "rollback":
			s.Handlers.Rollback = hdlr
		default:
		}
	}

	for i, exType := range b.ExperimentTypes {
		if exType.Name == s.Name {
			b.ExperimentTypes[i] = s
			return b
		}
	}

	b.ExperimentTypes = append(b.ExperimentTypes, s)

	return b
}

// WithEndpoint ..
func (b Iter8ConfigBuilder) WithEndpoint(endpoint string) Iter8ConfigBuilder {
	b.Analytics.Endpoint = endpoint
	return b
}

// WithRequestCount ..
func (b Iter8ConfigBuilder) WithRequestCount(requestCount string) Iter8ConfigBuilder {
	b.Metrics.RequestCount = requestCount
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
