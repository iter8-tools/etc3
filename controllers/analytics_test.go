package controllers

import (
	"io/ioutil"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

type HTTPMock struct{}

func (HTTPMock) Post(url, contentType string, body []byte) ([]byte, int, error) {
	filePath := CompletePath("../test/data", "analyticsresponse.json")
	bytes, err := ioutil.ReadFile(filePath)
	return bytes, 200, err
}

func TestInvoke(t *testing.T) {
	hm := HTTPMock{}
	log := logr.Discard()
	_, err := Invoke(log, "https://iter8.tools", "hello", hm)
	assert.NoError(t, err)
}
