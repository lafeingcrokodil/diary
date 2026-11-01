package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lafeingcrokodil/diary/internal/auth"
	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/nutrient"
)

// maxBodyBytes limits the size of request bodies, which are small JSON
// documents.
const maxBodyBytes = 1 << 20

// owner returns the ID of the signed-in user, whose data a request is about.
// The auth middleware guarantees that there's a user.
func owner(req *http.Request) string {
	user, _ := auth.UserFrom(req.Context())

	return user.ID
}

// decodeBody decodes the JSON in a request's body into value, which must be a
// pointer. If that fails, it responds with a 400 that names what the body
// should have been, and returns false.
func decodeBody(resp http.ResponseWriter, req *http.Request, value any, what string) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(resp, req.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(value); err != nil {
		http.Error(resp, "The request body isn't a valid "+what+": "+err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

// storeError responds to an error from a store: a 404 if the item doesn't
// exist (or belongs to someone else), and a 500 otherwise.
func storeError(resp http.ResponseWriter, req *http.Request, err error) {
	if errors.Is(err, food.ErrNotFound) || errors.Is(err, nutrient.ErrNotFound) {
		http.NotFound(resp, req)
		return
	}

	slog.ErrorContext(req.Context(), "accessing a store", "err", err)
	http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// writeJSON responds with value encoded as JSON.
func writeJSON(resp http.ResponseWriter, req *http.Request, status int, value any) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(status)

	if err := json.NewEncoder(resp).Encode(value); err != nil {
		slog.ErrorContext(req.Context(), "writing response", "err", err)
	}
}
