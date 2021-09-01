package core

import (
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetJsonBytes(t *testing.T) {
	// valid
	_, err := GetJSONBytes("https://httpbin.org/stream/1")
	assert.NoError(t, err)

	// invalid
	_, err = GetJSONBytes("https://httpbin.org/undef")
	assert.Error(t, err)
}

func TestPointers(t *testing.T) {
	assert.Equal(t, int32(1), *Int32Pointer(1))
	assert.Equal(t, float32(0.1), *Float32Pointer(0.1))
	assert.Equal(t, float64(0.1), *Float64Pointer(0.1))
	assert.Equal(t, "hello", *StringPointer("hello"))
	assert.Equal(t, false, *BoolPointer(false))
	assert.Equal(t, GET, *HTTPMethodPointer(GET))
}

func TestWait(t *testing.T) {
	errCh := make(chan error)
	defer close(errCh)

	var wg sync.WaitGroup
	for j := range []int{0, 1, 2, 3} {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			time.Sleep(10 * time.Second)
		}(j)
	}

	err := WaitTimeoutOrError(&wg, 30*time.Second, errCh)
	assert.NoError(t, err)
}

func TestWaitTimeout(t *testing.T) {
	errCh := make(chan error)
	defer close(errCh)

	var wg sync.WaitGroup
	for j := range []int{0, 1, 2, 3} {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			time.Sleep(10 * time.Second)
		}(j)
	}

	err := WaitTimeoutOrError(&wg, 5*time.Second, errCh)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for go routines to complete")
}

func TestSetLogLevel(t *testing.T) {
	SetLogLevel(logrus.InfoLevel)
	assert.Equal(t, logrus.InfoLevel, log.GetLevel())
}
