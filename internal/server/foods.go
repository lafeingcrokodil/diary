package server

import (
	"net/http"

	"github.com/lafeingcrokodil/diary/internal/food"
)

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

func toFoodJSON(item food.Food) foodJSON {
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

func fromFoodJSON(item foodJSON) food.Food {
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
		result = append(result, toFoodJSON(item))
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
	writeJSON(resp, req, http.StatusCreated, toFoodJSON(created))
}

func (h foodHandlers) get(resp http.ResponseWriter, req *http.Request) {
	got, err := h.foods.Get(req.Context(), owner(req), req.PathValue("id"))
	if err != nil {
		storeError(resp, req, err)
		return
	}

	writeJSON(resp, req, http.StatusOK, toFoodJSON(got))
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

	writeJSON(resp, req, http.StatusOK, toFoodJSON(updated))
}

func (h foodHandlers) delete(resp http.ResponseWriter, req *http.Request) {
	if err := h.foods.Delete(req.Context(), owner(req), req.PathValue("id")); err != nil {
		storeError(resp, req, err)
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

// readFood decodes and validates the food in a request's body. If that fails,
// it responds with a 400 and returns false.
func readFood(resp http.ResponseWriter, req *http.Request) (food.Food, bool) {
	var input foodJSON
	if !decodeBody(resp, req, &input, "food") {
		return food.Food{}, false
	}

	parsed := fromFoodJSON(input)
	if err := parsed.Validate(); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return food.Food{}, false
	}

	return parsed, true
}
