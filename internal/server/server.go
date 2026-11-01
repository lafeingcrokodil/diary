// Package server wires up the HTTP routes for the diary app.
package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/lafeingcrokodil/diary/internal/auth"
	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/nutrient"
	"github.com/lafeingcrokodil/diary/internal/recipe"
)

// Stores holds the stores that keep the app's data.
type Stores struct {
	Foods     food.Store
	Nutrients nutrient.Store
	Recipes   recipe.Store
}

// New returns an http.Handler that serves all of the app's routes, plus the
// front end from the given files, which must include an index.html.
//
// Every request under /api/ goes through authenticate, which must only pass on
// requests from allowed users, with the user in the context (see
// auth.WithUser). The app's data is kept in stores.
func New(frontend fs.FS, authenticate func(http.Handler) http.Handler, stores Stores) http.Handler {
	// The API's own mux responds with a 404 to unknown routes, and a 405 to
	// known routes with other methods.
	api := http.NewServeMux()
	api.HandleFunc("GET /api/me", handleMe)
	handleFoods(api, stores.Foods, stores.Nutrients, stores.Recipes)
	handleNutrients(api, stores.Nutrients)
	handleRecipes(api, stores.Recipes, stores.Foods)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	// Unknown API routes get a 404 rather than the front end.
	mux.Handle("/api/", authenticate(api))
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

// me describes the signed-in user.
type me struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// handleMe responds with the signed-in user.
func handleMe(resp http.ResponseWriter, req *http.Request) {
	user, ok := auth.UserFrom(req.Context())
	if !ok {
		slog.ErrorContext(req.Context(), "no user in the context of an authenticated request")
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	writeJSON(resp, req, http.StatusOK, me{ID: user.ID, Email: user.Email})
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
