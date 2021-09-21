package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	iter8 "github.com/iter8-tools/etc3/api/v2beta1"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeFakeNotificationTask(t *testing.T) {
	_, err := Make(&iter8.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeFakeHTTPTask(t *testing.T) {
	_, err := Make(&iter8.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeHttpTask(t *testing.T) {
	url, _ := json.Marshal("http://postman-echo.com/post")
	body, _ := json.Marshal("{\"hello\":\"world\"}")
	headers, _ := json.Marshal([]core.NamedValue{{
		Name:  "x-foo",
		Value: "bar",
	}, {
		Name:  "Authentication",
		Value: "Basic: dXNlcm5hbWU6cGFzc3dvcmQK",
	}})
	task, err := Make(&iter8.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"URL":     {Raw: url},
			"body":    {Raw: body},
			"headers": {Raw: headers},
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, task)
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	req, err := task.(*Task).prepareRequest(ctx)
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	assert.Equal(t, "http://postman-echo.com/post", req.URL.String())
	assert.Equal(t, "bar", req.Header.Get("x-foo"))

	err = task.Run(ctx)
	assert.NoError(t, err)
}

func TestMakeHttpTaskDefaults(t *testing.T) {
	url, _ := json.Marshal("http://target")
	task, err := Make(&iter8.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"URL": {Raw: url},
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, task)

	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	req, err := task.(*Task).prepareRequest(ctx)
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, 1, len(req.Header))
	assert.Equal(t, "application/json", req.Header.Get("Content-type"))

	data, err := ioutil.ReadAll(req.Body)
	assert.NoError(t, err)

	expectedBody := `{"summary":{"winnerFound":false},"experiment":{"kind":"Experiment","apiVersion":"iter8.tools/v2beta1","metadata":{"name":"test-experiment-1","namespace":"default","creationTimestamp":null},"spec":{"versionInfo":["default","canary"],"criteria":{"objectives":[{"metric":"mean-latency","upperLimit":"1k"},{"metric":"error-rate","upperLimit":"10m"}]},"duration":{"minIntervalBetweenLoops":15,"maxLoops":10},"backends":[{"name":"backend","description":"backend description","method":"POST","provider":"provider","jqExpression":"jqExpression","headers":{"header":"{{.variable-1}}::{{.variable-2}}"},"url":"https://provider.url","versionInfo":[{"interval":"interval-v1","name":"name-v1"},{"interval":"interval-v2","name":"name-v2"}],"metrics":[{"name":"mean-latency","description":"Mean latency","params":{"query":"(sum(increase(revision_app_request_latencies_sum{service_name=~'.*$name'}[$interval]))or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{service_name=~'.*$name'}[$interval])) or on() vector(0))"},"units":"milliseconds","type":"Gauge"},{"name":"error-rate","description":"Fraction of requests with error responses","params":{"query":"(sum(increase(revision_app_request_latencies_count{response_code_class!='2xx',service_name=~'.*$name'}[$interval])) or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{service_name=~'.*$name'}[$interval])) or on() vector(0))"},"type":"Gauge"},{"name":"request-count","description":"Number of requests","params":{"query":"sum(increase(revision_app_request_latencies_count{service_name=~'.*$name'}[$interval])) or on() vector(0)"},"type":"Counter"},{"name":"95th-percentile-tail-latency","description":"95th percentile tail latency","params":{"query":"histogram_quantile(0.95, sum(rate(revision_app_request_latencies_bucket{service_name=~'.*$name'}[$interval])) by (le))"},"units":"milliseconds","type":"Gauge"}]}]},"status":{"conditions":[{"type":"Completed","status":"False","lastTransitionTime":"2020-12-27T21:55:49Z","reason":"StartHandlerLaunched","message":"Start handler 'start' launched"},{"type":"Failed","status":"False","lastTransitionTime":"2020-12-27T21:55:48Z"}],"startTime":"2020-12-27T21:55:48Z","lastUpdateTime":"2020-12-27T21:55:48Z","stage":"Initializing","testingPattern":"SLOValidation","completedLoops":0,"message":"StartHandlerLaunched: Start handler 'start' launched"}}}`
	assert.Equal(t, expectedBody, string(data))
}
