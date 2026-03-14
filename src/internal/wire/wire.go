// Package wire is the single composition root for the application.
//
// All dependency wiring happens here via Google Wire provider sets.
// Domain packages must not import this package.
//
// To regenerate the wiring: run `wire gen ./internal/wire/...` from the src/ directory.
package wire
