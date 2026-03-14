// Package security provides helpers for password hashing and verification.
package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const _defaultCost = bcrypt.DefaultCost

// HashPassword hashes the plain-text password using bcrypt with the default cost.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), _defaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword reports whether plain matches the stored bcrypt hash.
// Returns nil on success, bcrypt.ErrMismatchedHashAndPassword on mismatch.
func CheckPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
