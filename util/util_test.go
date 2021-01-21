package util

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletePath(t *testing.T) {
	p1 := CompletePath("", "a")
	p2 := CompletePath("../", "util/a")
	p3 := CompletePath("", "b")
	assert.Equal(t, p1, p2)
	assert.NotEqual(t, p2, p3)
}

func ExampleCompletePath() {
	filePath := CompletePath("../test/data", "expwithextrafields.yaml")
	_, _ = ioutil.ReadFile(filePath)
}
