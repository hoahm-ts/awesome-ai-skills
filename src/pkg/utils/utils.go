// Package utils provides small general-purpose helper functions with no business logic.
// Do not add domain behaviour here.
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID returns a cryptographically random hex-encoded identifier of the given byte length.
// For example, GenerateID(16) returns a 32-character hex string.
func GenerateID(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Ptr returns a pointer to the value v.
// Useful when you need a pointer to a literal or computed value.
func Ptr[T any](v T) *T {
	return &v
}
