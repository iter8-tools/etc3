package core

import (
	"context"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildErrorGarbageYAML(t *testing.T) {
	_, err := (&Builder{}).FromFile(CompletePath("../", "testdata/garbage.yaml")).Build()
	assert.Error(t, err)
}

func TestInvalidAction(t *testing.T) {
	_, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment3.yaml")).Build()
	assert.Error(t, err)
}

func TestInvalidActions(t *testing.T) {
	_, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment5.yaml")).Build()
	assert.Error(t, err)
}

func TestStringAction(t *testing.T) {
	_, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment9.yaml")).Build()
	assert.Error(t, err)
}

func TestGetRecommendedBaseline(t *testing.T) {
	var err error
	var exp *Experiment
	var b string
	exp, err = (&Builder{}).FromFile(CompletePath("../", "testdata/experiment6.yaml")).Build()
	assert.NoError(t, err)
	b, err = exp.GetVersionRecommendedForPromotion()
	assert.NoError(t, err)
	assert.Equal(t, "default", b)

	exp, err = (&Builder{}).FromFile(CompletePath("../", "testdata/experiment2.yaml")).Build()
	assert.NoError(t, err)
	_, err = exp.GetVersionRecommendedForPromotion()
	assert.Error(t, err)

	exp = nil
	_, err = exp.GetVersionRecommendedForPromotion()
	assert.Error(t, err)
}

func TestGetExperimentFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKey("experiment"), "hello world")
	_, err := GetExperimentFromContext(ctx)
	assert.Error(t, err)

	_, err = GetExperimentFromContext(context.Background())
	assert.Error(t, err)

	ctx = context.WithValue(context.Background(), ContextKey("experiment"), &Experiment{
		Experiment: v2alpha2.Experiment{
			TypeMeta:   v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{},
			Spec:       v2alpha2.ExperimentSpec{},
			Status:     v2alpha2.ExperimentStatus{},
		},
	})

	exp, err := GetExperimentFromContext(ctx)
	assert.NotNil(t, exp)
	assert.NoError(t, err)
}

func TestInterpolateWithExperiment(t *testing.T) {
	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment6.yaml")).Build()
	assert.NoError(t, err)
	e, err := exp.ToMap()
	assert.NoError(t, err)
	tags := NewTags().With("this", e)
	str := "{{.this.metadata.namespace}}"
	v, err := tags.Interpolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, "default", v)
}
