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
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadSimplexGVKs(t *testing.T) {
	simplexDir := CompletePath("../test/data", "")
	simplexFile := path.Join(simplexDir, SimplexYaml)
	var sgvks SimplexGVKs
	err := readSimplexGVKs(simplexFile, &sgvks)

	assert.NoError(t, err)
	assert.Len(t, sgvks.GVKs, 2)
	assert.Equal(t, "apps", sgvks.GVKs[0].Group)
	assert.Equal(t, "v1", sgvks.GVKs[0].Version)
	assert.Equal(t, "deployments", sgvks.GVKs[0].Kind)
	assert.Equal(t, "", sgvks.GVKs[1].Group)
	assert.Equal(t, "v1", sgvks.GVKs[1].Version)
	assert.Equal(t, "secrets", sgvks.GVKs[1].Kind)
}

func TestGetSimplexes(t *testing.T) {
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

	s, err := GetSimplexes(&cfg)
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Len(t, s, 2)
	assert.Equal(t, "apps", s[0].GVK.Group)
	assert.Equal(t, "v1", s[0].GVK.Version)
	assert.Equal(t, "deployments", s[0].GVK.Kind)
	assert.Equal(t, "", s[1].GVK.Group)
	assert.Equal(t, "v1", s[1].GVK.Version)
	assert.Equal(t, "secrets", s[1].GVK.Kind)
}
