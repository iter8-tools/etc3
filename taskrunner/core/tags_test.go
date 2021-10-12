package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInterpolate(t *testing.T) {
	tags := NewTags().
		With("name", "tester").
		With("revision", "revision1").
		With("container", "super-container")

	// success cases
	inputs := []string{
		// `hello {{index . "name"}}`,
		// "hello {{index .name}}",
		"hello @<.name>@",
		"hello @<.name>@@<.other>@",
	}
	for _, str := range inputs {
		interpolated, err := tags.Interpolate(&str)
		assert.NoError(t, err)
		assert.Equal(t, "hello tester", interpolated)
	}

	// failure cases
	inputs = []string{
		// bad delimiters,
		"hello @{index .name>@",
		// missing '.'
		"hello @<name>@",
	}
	for _, str := range inputs {
		_, err := tags.Interpolate(&str)
		assert.Error(t, err)
	}

	// empty tags (success cases)
	str := "hello @<.name>@"
	tags = NewTags()
	interpolated, err := tags.Interpolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, "hello ", interpolated)

	// secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"secretName": []byte("tester"),
		},
	}

	str = "hello @<.secret.secretName>@"
	tags = NewTags().WithSecret("secret", &secret)
	assert.Contains(t, tags.M, "secret")
	interpolated, err = tags.Interpolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, "hello tester", interpolated)
}
