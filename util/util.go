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
package util

import (
	"context"

	"github.com/go-logr/logr"
)

type loggerKeyType string

const (
	// LoggerKey is the key used to extract logger from context
	LoggerKey = loggerKeyType("logger")

	// NamespaceKey is a key used to store/retrieve a namespace from the context
	NamespaceKey = "namespace"
)

// Logger gets the logger from the context.
func Logger(ctx context.Context) logr.Logger {
	return ctx.Value(LoggerKey).(logr.Logger)
}

// func Namespace(ctx context.Context) string {
// 	return string(ctx.Value(NamespaceKey))
// }
