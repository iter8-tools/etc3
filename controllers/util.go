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

// util.go - utility methods
//    including methods to retrieve values from context.Context

package controllers

import (
	"context"
	"path/filepath"
	"runtime"

	"github.com/go-logr/logr"
	"github.com/iter8-tools/etc3/api/v2beta1"
)

// ContextKey variables can be used to retrieve values from context object
type ContextKey string

const (
	// LoggerKey is the key used to extract logger from context
	LoggerKey ContextKey = "logger"

	// OriginalStatusKey is the key used to extract the original status from the context
	OriginalStatusKey ContextKey = "originalStatus"
)

// Logger gets the logger from the context.
func Logger(ctx context.Context) logr.Logger {
	return ctx.Value(LoggerKey).(logr.Logger)
}

// OriginalStatus gets the status from the context
func OriginalStatus(ctx context.Context) *v2beta1.ExperimentStatus {
	return ctx.Value(OriginalStatusKey).(*v2beta1.ExperimentStatus)
}

// CompletePath is a helper function for converting file paths, specified relative to the caller of this function, into absolute ones.
// CompletePath is useful in tests and enables deriving the absolute path of experiment YAML files.
func CompletePath(prefix string, suffix string) string {
	_, testFilename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(testFilename), prefix, suffix)
}

func versionIndex(version string, instance *v2beta1.Experiment) (index int, ok bool) {
	for i, v := range instance.Spec.VersionInfo {
		if v == version {
			return i, true
		}
	}
	return -1, false
}
