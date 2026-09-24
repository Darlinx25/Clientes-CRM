//go:build !web_embed

package web

import "net/http"

// Available reports whether the frontend build is embedded in this binary.
// Without the web_embed tag the frontend is served by nginx (Docker flow).
func Available() bool { return false }

// Handler is only reachable when Available() is true.
func Handler() http.Handler {
	return http.NotFoundHandler()
}