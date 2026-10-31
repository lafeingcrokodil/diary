// Package server wires up the HTTP routes for the diary app.
package server

import (
	"net/http"
)

// New returns an http.Handler that serves all of the app's routes.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	return mux
}

// handleHealthz reports that the server is up and able to handle requests.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}
