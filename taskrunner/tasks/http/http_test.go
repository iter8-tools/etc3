package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeFakeNotificationTask(t *testing.T) {
	_, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeFakeHTTPTask(t *testing.T) {
	_, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeHttpTask(t *testing.T) {
	url, _ := json.Marshal("http://postman-echo.com/post")
	body, _ := json.Marshal("{\"hello\":\"world\"}")
	headers, _ := json.Marshal([]v2alpha2.NamedValue{{
		Name:  "x-foo",
		Value: "bar",
	}, {
		Name:  "Authentication",
		Value: "Basic: dXNlcm5hbWU6cGFzc3dvcmQK",
	}})
	task, err := Make(&v2alpha2.TaskSpec{
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
	task, err := Make(&v2alpha2.TaskSpec{
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

	_, err = ioutil.ReadAll(req.Body)
	assert.NoError(t, err)

}
