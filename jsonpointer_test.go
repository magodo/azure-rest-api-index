package main

import (
	"testing"

	"github.com/go-openapi/jsonpointer"
	"github.com/stretchr/testify/assert"
)

func TestJSONPointerOffset(t *testing.T) {
	cases := []struct {
		name     string
		ptr      string
		input    string
		offset   int64
		hasError bool
	}{
		{
			name:   "object key",
			ptr:    "/foo/bar",
			input:  `{"foo": {"bar": 21}}`,
			offset: 14,
		},
		{
			name:   "array index",
			ptr:    "/0/1",
			input:  `[[1,2], [3,4]]`,
			offset: 3,
		},
		{
			name:   "mix array index and object key",
			ptr:    "/0/1/foo/0",
			input:  `[[1, {"foo": ["a", "b"]}], [3, 4]]`,
			offset: 14,
		},
		{
			name:     "nonexist object key",
			ptr:      "/foo/baz",
			input:    `{"foo": {"bar": 21}}`,
			hasError: true,
		},
		{
			name:     "nonexist array index",
			ptr:      "/0/2",
			input:    `[[1,2], [3,4]]`,
			hasError: true,
		},
		{
			name:   "encoded reference",
			ptr:    "/paths/~1p~1{}/get",
			input:  `{"paths": {"foo": {"bar": 123, "baz": {}}, "/p/{}": {"get": {}}}}`,
			offset: 58,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ptr, err := jsonpointer.New(tt.ptr)
			assert.NoError(t, err)
			offset, err := JSONPointerOffset(ptr, tt.input)
			if tt.hasError {
				assert.Error(t, err)
				return
			}
			t.Log(offset, err)
			assert.NoError(t, err)
			assert.Equal(t, tt.offset, offset)
		})
	}
}
