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
	"encoding/json"
	"path/filepath"
	"runtime"

	"github.com/go-logr/logr"
	"github.com/iter8-tools/etc3/api/v2alpha2"
)

// ContextKey variables can be used to retrieve values from context object
type ContextKey string

// Iter8LogPriority variables and constants can be used for setting Iter8Log priority value
type Iter8LogPriority uint8

// Iter8LogSource variables and constants can be used for setting Iter8Log source value
type Iter8LogSource string

const (
	// LoggerKey is the key used to extract logger from context
	LoggerKey ContextKey = "logger"

	// OriginalStatusKey is the key used to extract the original status from the context
	OriginalStatusKey ContextKey = "originalStatus"

	// Iter8LogPriorityHigh is the high priority value for Iter8Log
	Iter8LogPriorityHigh Iter8LogPriority = 1

	// Iter8LogPriorityMedium is the medium priority value for Iter8Log
	Iter8LogPriorityMedium Iter8LogPriority = 2

	// Iter8LogPriorityLow is the low priority value for Iter8Log
	Iter8LogPriorityLow Iter8LogPriority = 3

	// Iter8LogSourceTR is the task runner source
	Iter8LogSourceTR Iter8LogSource = "task-runner"
)

type Iter8Log struct {
	IsIter8Log          bool             `json:"isIter8Log" yaml:"isIter8Log"`
	ExperimentName      string           `json:"experimentName" yaml:"experimentName"`
	ExperimentNamespace string           `json:"experimentNamespace" yaml:"experimentNamespace"`
	Source              Iter8LogSource   `json:"source" yaml:"source"`
	Priority            Iter8LogPriority `json:"priority" yaml:"priority"`
	Message             string           `json:"message" yaml:"message"`
	// Precedence = 0 ... for start action.
	// Precedence = number of completed loops + 1 ... for loop action, and finish action.
	// Above definition of precedence will evolve as controller and analytics Iter8logs are implemented.
	// Precedence is not intended to be seen/used by the end-user. It is one a field used for ensuring Iter8logs are output in the chronological order.
	Precedence int `json:"precedence" yaml:"precedence"`
}

// JSON returns the JSON string representation of the Iter8log
func (il *Iter8Log) JSON() string {
	bytes, _ := json.Marshal(il)
	return string(bytes)
}

// Logger gets the logger from the context.
func Logger(ctx context.Context) logr.Logger {
	return ctx.Value(LoggerKey).(logr.Logger)
}

// OriginalStatus gets the status from the context
func OriginalStatus(ctx context.Context) *v2alpha2.ExperimentStatus {
	return ctx.Value(OriginalStatusKey).(*v2alpha2.ExperimentStatus)
}

// CompletePath is a helper function for converting file paths, specified relative to the caller of this function, into absolute ones.
// CompletePath is useful in tests and enables deriving the absolute path of experiment YAML files.
func CompletePath(prefix string, suffix string) string {
	_, testFilename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(testFilename), prefix, suffix)
}
