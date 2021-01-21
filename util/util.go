// Package util provides logging and testing related utility functions.
package util

import (
	"context"
	"path/filepath"
	"runtime"

	"github.com/go-logr/logr"
)

type loggerKeyType string

const (
	// LoggerKey is the key used to extract logger from context
	LoggerKey = loggerKeyType("logger")

	// NamespaceKey is a key used to store/retrieve a namespace from the context
	NamespaceKey = "namespace"
)

// util.go - utility methods; currently supporting storing and retreiving values from context.Context

// Logger gets the logger from the context.
func Logger(ctx context.Context) logr.Logger {
	return ctx.Value(LoggerKey).(logr.Logger)
}

// CompletePath is a helper function for converting file paths, specified relative to the caller of this function, into absolute ones.
// CompletePath is useful in tests and enables deriving the absolute path of experiment YAML files.
func CompletePath(prefix string, suffix string) string {
	_, testFilename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(testFilename), prefix, suffix)
}

// func Namespace(ctx context.Context) string {
// 	return string(ctx.Value(NamespaceKey))
// }
