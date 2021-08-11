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

// experiment_types.go - go model for experiment CRD

package v2beta1

// NamedValue name/value to be used in constructing a REST query to backend metrics server
type NamedValue struct {
	// Name of parameter
	Name string `json:"name" yaml:"name"`

	// Value of parameter
	Value string `json:"value" yaml:"value"`
}
