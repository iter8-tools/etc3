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

// client.go - command to invoke call from controller to analystics service

package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	v2beta1 "github.com/iter8-tools/etc3/api/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HTTPTransport ..
type HTTPTransport interface {
	Post(url, contentType string, body []byte) ([]byte, int, error)
}

// Invoke sends payload to endpoint and gets response back
func Invoke(log logr.Logger, endpoint string, payload interface{}, transport HTTPTransport) (*v2beta1.Analysis, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	body, statuscode, _ := transport.Post(endpoint, "application/json", data)

	var prettyBody bytes.Buffer
	json.Indent(&prettyBody, body, "", "  ")
	log.Info("post response", "URL", endpoint)
	log.Info(prettyBody.String())

	if statuscode >= 400 {
		return nil, fmt.Errorf("%v", string(body))
	}

	var response v2beta1.Analysis
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// TODO: fill in provenance, timestamp
	response.AggregatedMetrics.Provenance = endpoint
	response.VersionAssessments.Provenance = endpoint
	response.Weights.Provenance = endpoint
	now := metav1.Now()
	response.AggregatedMetrics.Timestamp = now
	response.VersionAssessments.Timestamp = now
	response.Weights.Timestamp = now

	return &response, nil
}
