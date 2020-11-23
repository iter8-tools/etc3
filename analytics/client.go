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

package analytics

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/iter8-tools/etc3/api/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mock() (*v2alpha1.Analysis, error) {

	now := metav1.Now()
	return &v2alpha1.Analysis{
		AggregatedMetrics: &v2alpha1.AggregatedMetricsAnalysis{
			AnalysisMetaData: v2alpha1.AnalysisMetaData{
				Provenance: "mock",
				Timestamp:  now,
				Message:    nil,
			},
			Data: make(map[string]v2alpha1.AggregatedMetricsData),
		},
		WinnerAssessment: &v2alpha1.WinnerAssessmentAnalysis{
			AnalysisMetaData: v2alpha1.AnalysisMetaData{
				Provenance: "mock",
				Timestamp:  now,
				Message:    nil,
			},
			Data: v2alpha1.WinnerAssessmentData{
				WinnerFound: false,
			},
		},
		VersionAssessments: &v2alpha1.VersionAssessmentAnalysis{
			AnalysisMetaData: v2alpha1.AnalysisMetaData{
				Provenance: "mock",
				Timestamp:  now,
				Message:    nil,
			},
			Data: make(map[string]v2alpha1.BooleanList),
		},
		Weights: &v2alpha1.WeightsAnalysis{
			AnalysisMetaData: v2alpha1.AnalysisMetaData{
				Provenance: "mock",
				Timestamp:  metav1.Now(),
				Message:    nil,
			},
			Data: []v2alpha1.WeightData{{
				Name:  "baseline",
				Value: 25,
			}, {
				Name:  "canary",
				Value: 75,
			}},
		},
	}, nil
}

// Invoke sends payload to endpoint and gets response back
func Invoke(log logr.Logger, endpoint string, payload interface{}) (*v2alpha1.Analysis, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, data, "", "  ")
	log.Info("post request", "URL", endpoint)
	log.Info(string(prettyJSON.Bytes()))
	raw, err := http.Post(endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return mock()
		// return nil, err
	}

	defer raw.Body.Close()
	body, err := ioutil.ReadAll(raw.Body)

	//var prettyBody bytes.Buffer
	//json.Indent(&prettyBody, body, "", "  ")
	log.Info("post response", "URL", endpoint)
	//log.Info(string(prettyBody.Bytes()))

	if raw.StatusCode >= 400 {
		log.Info("post error response", "response", string(body))
		return mock()
		// return nil, fmt.Errorf("%v", string(body))
	}

	var response v2alpha1.Analysis
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// TODO: fill in provenance, timestamp
	response.AggregatedMetrics.Provenance = endpoint
	response.VersionAssessments.Provenance = endpoint
	response.WinnerAssessment.Provenance = endpoint
	response.Weights.Provenance = endpoint
	now := metav1.Now()
	response.AggregatedMetrics.Timestamp = now
	response.VersionAssessments.Timestamp = now
	response.WinnerAssessment.Timestamp = now
	response.Weights.Timestamp = now

	return &response, nil
}
