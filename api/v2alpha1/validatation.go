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

// validation.go - methods to validate an experiment resource

package v2alpha1

// // IsValidSpec determines if the spec is consistent and valid for an experiment
// // If not, an error is returned
// func (s *ExperimentSpec) IsValidSpec() bool, string {

// 	// There should always be a baseline; we do not need to check this because the the
// 	// versionInfo cannot be created without it (ie, an error would already have been registered)
// 	// The number of candidates depends on the experiment type
// 	// This can only be checked after the start handler has been successfully called
// 	// (the start handler populates spec.versionInfo)

// 	// if start handler completed:
// 	numberCandidates := s.GetNumberOfCandidates()
// 	switch s.Strategy.Type {
// 	case Performance:
// 		if numberCandidates > 1 {
// 			return false, "Performance tests should not define candidates"
// 		}
// 	case Canary:
// 		if numberCandidates != 1 {
// 			return false, "Canary tests require exactly 1 candidate"
// 		}
// 	case AB:
// 		if numberCandidates != 1 {
// 			return false, "A/B tests require exactly 1 candidate"
// 		}
// 	case BlueGreen:
// 		if numberCandidates != 1 {
// 			return false, "BlueGreen tests require exactly 1 candidate"
// 		}
// 	case ABN:
// 		if numberCandidates < 2 {
// 			return false, "A/B/n tests require at least 2 candidate versions"
// 		}
// 	}

// 	//

// 	return true, ""
// }
