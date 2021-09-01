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

func TestNoNamespaceInterpolation(t *testing.T) {
	os.Setenv("ITER8_NAMESPACE", "namespace")
	os.Setenv("ITER8_ANALYTICS_ENDPOINT", "endpoint")
	os.Setenv("ITER8_TASKRUNNER_IMAGE", "taskrunner")

	cfg := Iter8Config{}
	err := ReadConfig(&cfg)
	if err != nil {
		t.Error("Unable to read configuration")
	}

	// verify values
	if cfg.Namespace != "namespace" {
		t.Errorf("cfg.Namespace incorrect. Expected: %s, got: %s", "namespace", cfg.Namespace)
	}
	if cfg.TaskRunnerImage != "taskrunner" {
		t.Errorf("cfg.TaskRunner incorrect. Expected: %s, got: %s", "taskrunner", cfg.TaskRunnerImage)
	}
	if cfg.AnalyticsEndpoint != "endpoint" {
		t.Errorf("cfg.Endpoint incorrect. Expected: %s, got: %s", "endpoint", cfg.AnalyticsEndpoint)
	}
}

func TestNamespaceInterpolation(t *testing.T) {
	os.Setenv("ITER8_NAMESPACE", "namespace")
	os.Setenv("ITER8_ANALYTICS_ENDPOINT", "ITER8_NAMESPACE/endpoint")
	os.Setenv("ITER8_TASKRUNNER_IMAGE", "taskrunner")

	cfg := Iter8Config{}
	err := ReadConfig(&cfg)
	if err != nil {
		t.Error("Unable to read configuration")
	}

	// verify values
	if cfg.Namespace != "namespace" {
		t.Errorf("cfg.Namespace incorrect. Expected: %s, got: %s", "namespace", cfg.Namespace)
	}
	if cfg.TaskRunnerImage != "taskrunner" {
		t.Errorf("cfg.TaskRunner incorrect. Expected: %s, got: %s", "taskrunner", cfg.TaskRunnerImage)
	}
	if cfg.AnalyticsEndpoint != "namespace/endpoint" {
		t.Errorf("cfg.Analytics.Endpoint incorrect. Expected: %s, got: %s", "namespace/endpoint", cfg.AnalyticsEndpoint)
	}
}

func TestIter8Config(t *testing.T) {
	config := NewIter8Config().
		WithEndpoint("endpoint").
		WithNamespace("namespace").
		WithTaskRunnerImage("tRunner").
		Build()
	assert.Equal(t, "endpoint", config.AnalyticsEndpoint)
	assert.Equal(t, "namespace", config.Namespace)
	assert.Equal(t, "tRunner", config.TaskRunnerImage)
}
