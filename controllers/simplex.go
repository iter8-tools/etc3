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

// handlers.go implements code to start jobs

package controllers

const (
	// SimplexYaml is the name of the yaml file that describes the resources to be watched
	SimplexYaml = "simplex.yaml"
)

// // LaunchSimplex lauches all the Simplex resource watches
// func LaunchSimplex(ctx context.Context, simplexResources *SimplexResources) error {
// 	log := Logger(ctx)
// 	log.Info("LaunchSimplex called")
// 	defer log.Info("LaunchSimplex completed")

// 	return nil
// }
