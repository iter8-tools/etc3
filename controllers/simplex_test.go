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
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLaunchSimplex(t *testing.T) {
	os.Setenv("ITER8_NAMESPACE", "namespace")
	os.Setenv("ITER8_ANALYTICS_ENDPOINT", "endpoint")
	os.Setenv("HANDLERS_DIR", "dir")
	simplexDir := CompletePath("../test/data", "")
	os.Setenv("SIMPLEX_DIR", simplexDir)

	cfg := Iter8Config{}
	err := ReadConfig(&cfg)
	if err != nil {
		t.Error("Unable to read configuration")
	}

	err = LaunchSimplex(context.Background(), &cfg)
	assert.NoError(t, err)
}
