// Package integration contains external service clients and adapters.
//
// Each adapter implements a port interface defined in the relevant domain package.
// External service SDKs (HTTP clients, gRPC stubs, vendor SDKs) must not
// leak beyond this package boundary.
package integration
