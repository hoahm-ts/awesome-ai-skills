// Package lifecycle manages application startup and graceful shutdown sequences.
//
// Components that require an explicit stop signal (background goroutines, connection
// pools, etc.) must expose a Close/Shutdown method and register it here.
package lifecycle
