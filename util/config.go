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

package util

const (
	// DefaultRequestCounter is the default request counter metric
	DefaultRequestCounter = "request-count"

	// DefaultAnalyticsServiceEndpoint is the default analytics service endpoint
	DefaultAnalyticsServiceEndpoint = "http://127.0.0.1:8080/v2/analytics_results"

	// DefaultIter8Namespace is the default namespace where iter8 is installed
	DefaultIter8Namespace = "default"
)

// GetRequestCount returns the default request count metric identified by the install config
func GetRequestCount() string {
	return DefaultRequestCounter
}

// GetAnalyticsService returns the default analytics service defined by the install config
func GetAnalyticsService() string {
	return DefaultAnalyticsServiceEndpoint
}

// GetIter8InstallNamespace returns the namespace where iter8 is deployed
func GetIter8InstallNamespace() string {
	return DefaultIter8Namespace
}
