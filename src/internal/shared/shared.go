// Package shared is the shared kernel: common types, sentinel errors, and port interfaces
// that are used by multiple domain modules.
//
// Rules:
//   - May contain value objects, IDs, error sentinels, and port/interface definitions.
//   - Must NOT contain business logic or domain rules.
//   - Must NOT import from any internal domain package.
package shared

import "errors"

// Sentinel errors returned by repository implementations and mapped to HTTP status codes
// at the handler boundary.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidInput  = errors.New("invalid input")
)

// Page holds cursor-based pagination parameters.
// Do not use offset-based pagination — see AGENTS.md for rationale.
type Page struct {
	Limit  int
	Cursor string
}

// PageResult wraps a slice of results with pagination metadata.
type PageResult[T any] struct {
	Items      []T
	NextCursor string
	HasMore    bool
}
