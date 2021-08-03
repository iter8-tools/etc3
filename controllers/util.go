// Package util provides logging and testing related utility functions.
package controllers

// util.go - utility methods
//    including methods to retrieve values from context.Context

import (
	"context"
	"path/filepath"
	"runtime"

	"github.com/go-logr/logr"
	"github.com/iter8-tools/etc3/api/v2alpha2"
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
func OriginalStatus(ctx context.Context) *v2alpha2.ExperimentStatus {
	return ctx.Value(OriginalStatusKey).(*v2alpha2.ExperimentStatus)
}

// CompletePath is a helper function for converting file paths, specified relative to the caller of this function, into absolute ones.
// CompletePath is useful in tests and enables deriving the absolute path of experiment YAML files.
func CompletePath(prefix string, suffix string) string {
	_, testFilename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(testFilename), prefix, suffix)
}
