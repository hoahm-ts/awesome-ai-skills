// Package handler contains HTTP route handlers.
//
// Handlers are thin delivery-layer components: they parse and validate input,
// delegate to a domain service via an interface, and write a standard response
// using the response package. Business logic must not live here.
package handler
