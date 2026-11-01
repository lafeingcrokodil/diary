package server

import (
	"errors"
	"net/http"

	"github.com/lafeingcrokodil/diary/internal/nutrient"
)

// nutrientJSON is how the API represents a nutrient. The owner is left out:
// it's always the signed-in user.
type nutrientJSON struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Unit        nutrient.Unit `json:"unit"`
	KcalPerGram float64       `json:"kcalPerGram"`
	Archived    bool          `json:"archived"`
}

func toNutrientJSON(item nutrient.Nutrient) nutrientJSON {
	return nutrientJSON{
		ID:          item.ID,
		Name:        item.Name,
		Unit:        item.Unit,
		KcalPerGram: item.KcalPerGram,
		Archived:    item.Archived,
	}
}

func fromNutrientJSON(item nutrientJSON) nutrient.Nutrient {
	return nutrient.Nutrient{
		ID:          item.ID,
		Name:        item.Name,
		Unit:        item.Unit,
		KcalPerGram: item.KcalPerGram,
		Archived:    item.Archived,
	}
}

// handleNutrients adds the routes for the signed-in user's nutrients to mux.
// Nutrients are archived and unarchived by updating them, and only archived
// nutrients can be deleted (see nutrient.Store.Delete).
func handleNutrients(mux *http.ServeMux, nutrients nutrient.Store) {
	handlers := nutrientHandlers{nutrients: nutrients}
	mux.HandleFunc("GET /api/nutrients", handlers.list)
	mux.HandleFunc("POST /api/nutrients", handlers.create)
	mux.HandleFunc("GET /api/nutrients/{id}", handlers.get)
	mux.HandleFunc("PUT /api/nutrients/{id}", handlers.update)
	mux.HandleFunc("DELETE /api/nutrients/{id}", handlers.delete)
}

type nutrientHandlers struct {
	nutrients nutrient.Store
}

func (h nutrientHandlers) list(resp http.ResponseWriter, req *http.Request) {
	list, err := h.nutrients.List(req.Context(), owner(req))
	if err != nil {
		storeError(resp, req, err)
		return
	}

	result := make([]nutrientJSON, 0, len(list))
	for _, item := range list {
		result = append(result, toNutrientJSON(item))
	}

	writeJSON(resp, req, http.StatusOK, result)
}

func (h nutrientHandlers) create(resp http.ResponseWriter, req *http.Request) {
	input, ok := readNutrient(resp, req)
	if !ok {
		return
	}

	created, err := h.nutrients.Create(req.Context(), owner(req), input)
	if err != nil {
		storeError(resp, req, err)
		return
	}

	resp.Header().Set("Location", "/api/nutrients/"+created.ID)
	writeJSON(resp, req, http.StatusCreated, toNutrientJSON(created))
}

func (h nutrientHandlers) get(resp http.ResponseWriter, req *http.Request) {
	got, err := h.nutrients.Get(req.Context(), owner(req), req.PathValue("id"))
	if err != nil {
		storeError(resp, req, err)
		return
	}

	writeJSON(resp, req, http.StatusOK, toNutrientJSON(got))
}

func (h nutrientHandlers) update(resp http.ResponseWriter, req *http.Request) {
	input, ok := readNutrient(resp, req)
	if !ok {
		return
	}

	// The path says which nutrient to update, whatever the body says.
	input.ID = req.PathValue("id")

	updated, err := h.nutrients.Update(req.Context(), owner(req), input)
	if err != nil {
		storeError(resp, req, err)
		return
	}

	writeJSON(resp, req, http.StatusOK, toNutrientJSON(updated))
}

func (h nutrientHandlers) delete(resp http.ResponseWriter, req *http.Request) {
	err := h.nutrients.Delete(req.Context(), owner(req), req.PathValue("id"))
	if errors.Is(err, nutrient.ErrNotArchived) {
		http.Error(resp, "Only archived nutrients can be deleted.", http.StatusConflict)
		return
	}

	if err != nil {
		storeError(resp, req, err)
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

// readNutrient decodes and validates the nutrient in a request's body. If that
// fails, it responds with a 400 and returns false.
func readNutrient(resp http.ResponseWriter, req *http.Request) (nutrient.Nutrient, bool) {
	var input nutrientJSON
	if !decodeBody(resp, req, &input, "nutrient") {
		return nutrient.Nutrient{}, false
	}

	parsed := fromNutrientJSON(input)
	if err := parsed.Validate(); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return nutrient.Nutrient{}, false
	}

	return parsed, true
}
