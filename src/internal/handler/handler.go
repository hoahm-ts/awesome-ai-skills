// Package handler contains HTTP route handlers.
//
// Handlers are thin delivery-layer components: they parse and validate input,
// delegate to a domain service via an interface, and write a standard response
// using the response package. Business logic must not live here.
package handler

import (
	"net/http"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/response"
)

// Ping handles GET /ping. It returns a simple liveness response confirming
// the API server is reachable.
func Ping(w http.ResponseWriter, _ *http.Request) {
	response.OKMsg(w, struct{}{}, "pong")
}
