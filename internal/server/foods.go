package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lafeingcrokodil/diary/internal/auth"
	"github.com/lafeingcrokodil/diary/internal/food"
)

// maxBodyBytes limits the size of request bodies, which are small JSON
// documents.
const maxBodyBytes = 1 << 20

// foodJSON is how the API represents a food. The owner is left out: it's
// always the signed-in user.
type foodJSON struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Brand      string        `json:"brand"`
	Unit       food.Unit     `json:"unit"`
	KcalPer100 float64       `json:"kcalPer100"`
	Servings   []servingJSON `json:"servings"`
}

type servingJSON struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

func toJSON(item food.Food) foodJSON {
	servings := make([]servingJSON, 0, len(item.Servings))
	for _, serving := range item.Servings {
		servings = append(servings, servingJSON(serving))
	}

	return foodJSON{
		ID:         item.ID,
		Name:       item.Name,
		Brand:      item.Brand,
		Unit:       item.Unit,
		KcalPer100: item.KcalPer100,
		Servings:   servings,
	}
}

func fromJSON(item foodJSON) food.Food {
	servings := make([]food.Serving, 0, len(item.Servings))
	for _, serving := range item.Servings {
		servings = append(servings, food.Serving(serving))
	}

	return food.Food{
		ID:         item.ID,
		Name:       item.Name,
		Brand:      item.Brand,
		Unit:       item.Unit,
		KcalPer100: item.KcalPer100,
		Servings:   servings,
	}
}

// handleFoods adds the routes for the signed-in user's foods to mux.
func handleFoods(mux *http.ServeMux, foods food.Store) {
	handlers := foodHandlers{foods: foods}
	mux.HandleFunc("GET /api/foods", handlers.list)
	mux.HandleFunc("POST /api/foods", handlers.create)
	mux.HandleFunc("GET /api/foods/{id}", handlers.get)
	mux.HandleFunc("PUT /api/foods/{id}", handlers.update)
	mux.HandleFunc("DELETE /api/foods/{id}", handlers.delete)
}

type foodHandlers struct {
	foods food.Store
}

func (h foodHandlers) list(resp http.ResponseWriter, req *http.Request) {
	list, err := h.foods.List(req.Context(), owner(req))
	if err != nil {
		storeError(resp, req, err)
		return
	}

	result := make([]foodJSON, 0, len(list))
	for _, item := range list {
		result = append(result, toJSON(item))
	}

	writeJSON(resp, req, http.StatusOK, result)
}

func (h foodHandlers) create(resp http.ResponseWriter, req *http.Request) {
	input, ok := readFood(resp, req)
	if !ok {
		return
	}

	created, err := h.foods.Create(req.Context(), owner(req), input)
	if err != nil {
		storeError(resp, req, err)
		return
	}

	resp.Header().Set("Location", "/api/foods/"+created.ID)
	writeJSON(resp, req, http.StatusCreated, toJSON(created))
}

func (h foodHandlers) get(resp http.ResponseWriter, req *http.Request) {
	got, err := h.foods.Get(req.Context(), owner(req), req.PathValue("id"))
	if err != nil {
		storeError(resp, req, err)
		return
	}

	writeJSON(resp, req, http.StatusOK, toJSON(got))
}

func (h foodHandlers) update(resp http.ResponseWriter, req *http.Request) {
	input, ok := readFood(resp, req)
	if !ok {
		return
	}

	// The path says which food to update, whatever the body says.
	input.ID = req.PathValue("id")

	updated, err := h.foods.Update(req.Context(), owner(req), input)
	if err != nil {
		storeError(resp, req, err)
		return
	}

	writeJSON(resp, req, http.StatusOK, toJSON(updated))
}

func (h foodHandlers) delete(resp http.ResponseWriter, req *http.Request) {
	if err := h.foods.Delete(req.Context(), owner(req), req.PathValue("id")); err != nil {
		storeError(resp, req, err)
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

// owner returns the ID of the signed-in user, whose foods a request is about.
// The auth middleware guarantees that there's a user.
func owner(req *http.Request) string {
	user, _ := auth.UserFrom(req.Context())

	return user.ID
}

// readFood decodes and validates the food in a request's body. If that fails,
// it responds with a 400 and returns false.
func readFood(resp http.ResponseWriter, req *http.Request) (food.Food, bool) {
	decoder := json.NewDecoder(http.MaxBytesReader(resp, req.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()

	var input foodJSON
	if err := decoder.Decode(&input); err != nil {
		http.Error(resp, "The request body isn't a valid food: "+err.Error(), http.StatusBadRequest)
		return food.Food{}, false
	}

	parsed := fromJSON(input)
	if err := parsed.Validate(); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return food.Food{}, false
	}

	return parsed, true
}

// storeError responds to an error from the food store: a 404 if the food
// doesn't exist (or belongs to someone else), and a 500 otherwise.
func storeError(resp http.ResponseWriter, req *http.Request, err error) {
	if errors.Is(err, food.ErrNotFound) {
		http.NotFound(resp, req)
		return
	}

	slog.ErrorContext(req.Context(), "accessing foods", "err", err)
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
