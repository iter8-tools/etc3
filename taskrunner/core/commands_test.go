package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetActionSpec(t *testing.T) {
	var nilExp *Experiment = nil
	_, err := nilExp.GetActionSpec("stay-calm")
	assert.Error(t, err)

	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment8.yaml")).Build()
	assert.NoError(t, err)
	_, err = exp.GetActionSpec("stay-calm")
	assert.Error(t, err)

	a, err := exp.GetActionSpec("start")
	assert.NoError(t, err)
	assert.NotEmpty(t, a)

	exp, err = (&Builder{}).FromFile(CompletePath("../", "testdata/experiment4.yaml")).Build()
	assert.NoError(t, err)
	_, err = exp.GetActionSpec("start")
	assert.Error(t, err)
}
