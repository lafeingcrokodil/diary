// Package server wires up the HTTP routes for the diary app.
package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// New returns an http.Handler that serves all of the app's routes, plus the
// front end from the given files, which must include an index.html.
func New(frontend fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	// Deprecated: the old path of the health check, to be removed.
	mux.HandleFunc("GET /healthz", handleHealth)
	// Unknown API routes get a 404 rather than the front end.
	mux.Handle("/api/", http.NotFoundHandler())
	// This can't be "GET /", because ServeMux would reject it as conflicting
	// with "/api/", so handleFrontend checks the method itself.
	mux.Handle("/", handleFrontend(frontend))
	return mux
}

// handleHealth reports that the server is up and able to handle requests.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

// handleFrontend serves a single-page app: requests for existing files get
// those files, while any other path without a file extension gets index.html,
// so that the app can handle the route in the browser. Requests for missing
// files with an extension (e.g. an outdated script) get a 404 instead of HTML.
func handleFrontend(files fs.FS) http.Handler {
	fileServer := http.FileServerFS(files)

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			resp.Header().Set("Allow", "GET, HEAD")
			http.Error(resp, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(path.Clean(req.URL.Path), "/")

		info, err := fs.Stat(files, name)
		if err == nil && !info.IsDir() {
			fileServer.ServeHTTP(resp, req)
			return
		}

		if path.Ext(name) != "" {
			http.NotFound(resp, req)
			return
		}

		http.ServeFileFS(resp, req, files, "index.html")
	})
}
