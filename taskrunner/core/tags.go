package core

import (
	"bytes"
	"errors"
	"html/template"

	corev1 "k8s.io/api/core/v1"
)

// Tags supports string extrapolation using tags.
type Tags struct {
	M map[string]interface{}
}

// NewTags creates an empty instance of Tags
func NewTags() Tags {
	return Tags{M: make(map[string]interface{})}
}

// WithSecret adds the fields in secret to tags
func (tags Tags) WithSecret(label string, secret *corev1.Secret) Tags {
	if secret != nil {
		obj := make(map[string]interface{})
		for n, v := range secret.Data {
			obj[n] = string(v)
		}
		tags = tags.With(label, obj)
	}
	return tags
}

// With adds obj to tags
func (tags Tags) With(label string, obj interface{}) Tags {
	if obj != nil {
		tags.M[label] = obj
	}
	return tags
}

// Interpolate str using tags.
func (tags *Tags) Interpolate(str *string) (string, error) {
	if tags == nil || tags.M == nil { // return a copy of the string
		return *str, nil
	}
	var err error
	var templ *template.Template
	if templ, err = template.New("").Delims(LeftDelim, RightDelim).Parse(*str); err == nil {
		buf := bytes.Buffer{}
		if err = templ.Execute(&buf, tags.M); err == nil {
			return buf.String(), nil
		}
		log.Error("template execution error: ", err)
		return "", errors.New("cannot interpolate string")
	}
	log.Error("template creation error: ", err)
	return "", errors.New("cannot interpolate string")
}
