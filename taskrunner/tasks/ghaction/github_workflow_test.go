package ghaction

import (
	"encoding/json"
	"testing"

	iter8 "github.com/iter8-tools/etc3/api/v2beta1"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeFakeGHWorkflowTask(t *testing.T) {
	_, err := Make(&iter8.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeGHWorkflowTask1(t *testing.T) {
	repository, _ := json.Marshal("iter8-tools/handler")
	workflow, _ := json.Marshal("workflow.yaml")
	secret, _ := json.Marshal("mysecret")
	task, err := Make(&iter8.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"repository": {Raw: repository},
			"workflow":   {Raw: workflow},
			"secret":     {Raw: secret},
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, task)

	httpTask := task.(*Task).ToHTTPTask()
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	assert.Equal(t, "https://api.github.com/repos/iter8-tools/handler/actions/workflows/workflow.yaml/dispatches", httpTask.With.URL)
	assert.Equal(t, "mysecret", *httpTask.With.Secret)
	assert.Equal(t, iter8.BearerAuthType, *httpTask.With.AuthType)
	assert.Equal(t, "{\"ref\": \"master\",\"inputs\": {}}", *httpTask.With.Body)
}

func TestMakeGHWorkflowTask2(t *testing.T) {
	repository, _ := json.Marshal("iter8-tools/handler")
	workflow, _ := json.Marshal("workflow.yaml")
	secret, _ := json.Marshal("mysecret")
	inputs, _ := json.Marshal([]core.NamedValue{{
		Name:  "arg1",
		Value: "value1",
	}, {
		Name:  "arg2",
		Value: "value2",
	}})
	task, err := Make(&iter8.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"repository": {Raw: repository},
			"workflow":   {Raw: workflow},
			"secret":     {Raw: secret},
			"inputs":     {Raw: inputs},
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, task)

	httpTask := task.(*Task).ToHTTPTask()
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	assert.Equal(t, "https://api.github.com/repos/iter8-tools/handler/actions/workflows/workflow.yaml/dispatches", httpTask.With.URL)
	assert.Equal(t, "mysecret", *httpTask.With.Secret)
	assert.Equal(t, iter8.BearerAuthType, *httpTask.With.AuthType)
	assert.Equal(t, "{\"ref\": \"master\",\"inputs\": {\"arg1\": \"value1\",\"arg2\": \"value2\"}}", *httpTask.With.Body)
}
