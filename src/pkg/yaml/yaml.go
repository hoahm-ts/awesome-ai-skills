// Package yaml provides generic utilities for parsing YAML files.
package yaml

import (
	"bytes"
	"fmt"
	"os"

	goyaml "gopkg.in/yaml.v3"
)

// ParseFile reads the YAML file at path and unmarshals its contents into a new value of type T.
// It returns a pointer to the populated value, or an error if the file cannot be read or parsed.
// Unknown fields are rejected so that typos in config keys (e.g. read_tiemout_seconds) fail fast
// instead of silently applying unintended defaults.
func ParseFile[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read yaml file %q: %w", path, err)
	}
	var v T
	dec := goyaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("parse yaml file %q: %w", path, err)
	}
	return &v, nil
}
