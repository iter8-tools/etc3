package controllers

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/iter8-tools/etc3/api/v2beta1"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestCompletePath(t *testing.T) {
	p1 := CompletePath("", "a")
	p2 := CompletePath("../", "controllers/a")
	p3 := CompletePath("", "b")
	assert.Equal(t, p1, p2)
	assert.NotEqual(t, p2, p3)
}

func ExampleCompletePath() {
	filePath := CompletePath("../test/data", "expwithextrafields.yaml")
	_, _ = ioutil.ReadFile(filePath)
}

func TestContext(t *testing.T) {
	ctx := context.Background()

	lg := ctrl.Log.WithName("etc3").WithName("util").WithName("test")
	ctx = context.WithValue(ctx, LoggerKey, lg)

	iterations := int32(5)
	loops := int32(2)
	message := "message"
	status := v2beta1.ExperimentStatus{
		CompletedIterations: &iterations,
		CompletedLoops:      &loops,
		Message:             &message,
	}
	ctx = context.WithValue(ctx, OriginalStatusKey, &status)

	assert.Equal(t, lg, Logger(ctx))
	assert.True(t, reflect.DeepEqual(OriginalStatus(ctx), &status))
}

func TestVersionIndex(t *testing.T) {
	e := v2beta1.NewExperiment("test", "namespace").WithVersion("v1").WithVersion("v2").Build()
	i, ok := versionIndex("v1", e)
	assert.True(t, ok)
	assert.Equal(t, i, 0)

	i, ok = versionIndex("v2", e)
	assert.True(t, ok)
	assert.Equal(t, i, 1)

	i, ok = versionIndex("v3", e)
	assert.False(t, ok)
	assert.Equal(t, i, -1)

}
