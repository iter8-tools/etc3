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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoInterpolation(t *testing.T) {
	os.Setenv("ITER8_NAMESPACE", "namespace")
	os.Setenv("ITER8_ANALYTICS_ENDPOINT", "endpoint")
	os.Setenv("HANDLERS_DIR", "dir")

	cfg := Iter8Config{}
	err := ReadConfig(&cfg)
	if err != nil {
		t.Error("Unable to read configuration")
	}

	// verify values
	if cfg.Namespace != "namespace" {
		t.Errorf("cfg.Namespace incorrect. Expected: %s, got: %s", "namespace", cfg.Namespace)
	}
	if cfg.HandlersDir != "dir" {
		t.Errorf("cfg.HandlersDir incorrect. Expected: %s, got: %s", "dir", cfg.HandlersDir)
	}
	if cfg.Analytics.Endpoint != "endpoint" {
		t.Errorf("cfg.Analytics.Endpoint incorrect. Expected: %s, got: %s", "endpoint", cfg.Analytics.Endpoint)
	}
}

func TestInterpolation(t *testing.T) {
	os.Setenv("ITER8_NAMESPACE", "namespace")
	os.Setenv("ITER8_ANALYTICS_ENDPOINT", "ITER8_NAMESPACE/endpoint")
	os.Setenv("HANDLERS_DIR", "dir")

	cfg := Iter8Config{}
	err := ReadConfig(&cfg)
	if err != nil {
		t.Error("Unable to read configuration")
	}

	// verify values
	if cfg.Namespace != "namespace" {
		t.Errorf("cfg.Namespace incorrect. Expected: %s, got: %s", "namespace", cfg.Namespace)
	}
	if cfg.HandlersDir != "dir" {
		t.Errorf("cfg.HandlersDir incorrect. Expected: %s, got: %s", "dir", cfg.HandlersDir)
	}
	if cfg.Analytics.Endpoint != "namespace/endpoint" {
		t.Errorf("cfg.Analytics.Endpoint incorrect. Expected: %s, got: %s", "namespace/endpoint", cfg.Analytics.Endpoint)
	}
}

func TestIter8Config(t *testing.T) {
	config := NewIter8Config().
		WithEndpoint("endpoint").
		WithNamespace("namespace").
		WithHandlersDir("hDir").
		Build()
	assert.Equal(t, "endpoint", config.Analytics.Endpoint)
	assert.Equal(t, "namespace", config.Namespace)
	assert.Equal(t, "hDir", config.HandlersDir)
}
