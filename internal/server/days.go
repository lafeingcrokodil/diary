package server

import (
	"net/http"

	"github.com/lafeingcrokodil/diary/internal/diary"
)

// dayJSON is how the API represents a day. The owner is left out: it's always
// the signed-in user. Values that aren't logged are null.
type dayJSON struct {
	Date         string   `json:"date"`
	WeightKg     *float64 `json:"weightKg"`
	KcalExpended *float64 `json:"kcalExpended"`
}

func toDayJSON(day diary.Day) dayJSON {
	return dayJSON{
		Date:         day.Date,
		WeightKg:     day.WeightKg,
		KcalExpended: day.KcalExpended,
	}
}

func fromDayJSON(day dayJSON) diary.Day {
	return diary.Day{
		Date:         day.Date,
		WeightKg:     day.WeightKg,
		KcalExpended: day.KcalExpended,
	}
}

// handleDays adds the routes for the signed-in user's days to mux. Every date
// has a day, which is empty if nothing is logged for it, so days are only
// read and replaced, never created or deleted.
func handleDays(mux *http.ServeMux, days diary.Store) {
	handlers := dayHandlers{days: days}
	mux.HandleFunc("GET /api/days/{date}", handlers.get)
	mux.HandleFunc("PUT /api/days/{date}", handlers.put)
}

type dayHandlers struct {
	days diary.Store
}

func (h dayHandlers) get(resp http.ResponseWriter, req *http.Request) {
	date, ok := readDate(resp, req)
	if !ok {
		return
	}

	got, err := h.days.Get(req.Context(), owner(req), date)
	if err != nil {
		storeError(resp, req, err)
		return
	}

	writeJSON(resp, req, http.StatusOK, toDayJSON(got))
}

func (h dayHandlers) put(resp http.ResponseWriter, req *http.Request) {
	date, ok := readDate(resp, req)
	if !ok {
		return
	}

	var input dayJSON
	if !decodeBody(resp, req, &input, "day") {
		return
	}

	// The path says which day to replace, whatever the body says.
	input.Date = date

	parsed := fromDayJSON(input)
	if err := parsed.Validate(); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	stored, err := h.days.Put(req.Context(), owner(req), parsed)
	if err != nil {
		storeError(resp, req, err)
		return
	}

	writeJSON(resp, req, http.StatusOK, toDayJSON(stored))
}

// readDate returns the date in a request's path. If it isn't a valid date, it
// responds with a 400 and returns false.
func readDate(resp http.ResponseWriter, req *http.Request) (string, bool) {
	date := req.PathValue("date")
	if err := diary.ValidateDate(date); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return "", false
	}

	return date, true
}
