package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/diary/diarytest"
	"github.com/lafeingcrokodil/diary/internal/server"
)

const monday = `{"date":"2026-11-02","weightKg":70.5,"kcalExpended":2300}`

// newDaysAPI returns an api whose days are kept in memory.
func newDaysAPI() api {
	return newAPI(server.Stores{Days: diarytest.NewMemoryStore()})
}

func TestDays(t *testing.T) {
	t.Parallel()

	t.Run("put and get", func(t *testing.T) {
		t.Parallel()

		days := newDaysAPI()

		rec := days.do(t, http.MethodPut, "/api/days/2026-11-02", alice, monday)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.JSONEq(t, monday, rec.Body.String())

		rec = days.do(t, http.MethodGet, "/api/days/2026-11-02", alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, monday, rec.Body.String())
	})

	t.Run("a day that nothing is logged for", func(t *testing.T) {
		t.Parallel()

		rec := newDaysAPI().do(t, http.MethodGet, "/api/days/2026-11-02", alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"date":"2026-11-02","weightKg":null,"kcalExpended":null}`, rec.Body.String())
	})

	t.Run("values that are left out or null aren't logged", func(t *testing.T) {
		t.Parallel()

		days := newDaysAPI()
		days.putMonday(t)

		rec := days.do(t, http.MethodPut, "/api/days/2026-11-02", alice, `{"weightKg":null}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		want := `{"date":"2026-11-02","weightKg":null,"kcalExpended":null}`
		assert.JSONEq(t, want, rec.Body.String())

		rec = days.do(t, http.MethodGet, "/api/days/2026-11-02", alice, "")
		assert.JSONEq(t, want, rec.Body.String())
	})

	t.Run("the date in the body is ignored in favour of the one in the path", func(t *testing.T) {
		t.Parallel()

		days := newDaysAPI()

		rec := days.do(t, http.MethodPut, "/api/days/2026-11-03", alice, monday)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, "2026-11-03", decodeDay(t, rec).Date)

		rec = days.do(t, http.MethodGet, "/api/days/2026-11-02", alice, "")
		assert.Nil(t, decodeDay(t, rec).WeightKg)
	})

	t.Run("invalid days are rejected", func(t *testing.T) {
		t.Parallel()

		days := newDaysAPI()
		days.putMonday(t)

		for _, body := range []string{
			`{"weightKg":0}`,
			`{"weightKg":-70}`,
			`{"kcalExpended":-1}`,
			`{"weightKg":"70"}`,
			`{"weight":70}`,
			`not JSON`,
		} {
			rec := days.do(t, http.MethodPut, "/api/days/2026-11-02", alice, body)
			assert.Equal(t, http.StatusBadRequest, rec.Code, body)
		}

		rec := days.do(t, http.MethodGet, "/api/days/2026-11-02", alice, "")
		assert.JSONEq(t, monday, rec.Body.String())
	})

	t.Run("invalid dates are rejected", func(t *testing.T) {
		t.Parallel()

		days := newDaysAPI()

		for _, date := range []string{"2026-02-30", "2026-1-5", "today", "2026-11-02T00:00:00Z"} {
			for _, method := range []string{http.MethodGet, http.MethodPut} {
				rec := days.do(t, method, "/api/days/"+date, alice, monday)
				assert.Equal(t, http.StatusBadRequest, rec.Code, "%s %s", method, date)
			}
		}
	})

	t.Run("users can't reach each other's days", func(t *testing.T) {
		t.Parallel()

		days := newDaysAPI()
		days.putMonday(t)

		rec := days.do(t, http.MethodGet, "/api/days/2026-11-02", bob, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Nil(t, decodeDay(t, rec).WeightKg)

		rec = days.do(t, http.MethodPut, "/api/days/2026-11-02", bob, `{"weightKg":90}`)
		require.Equal(t, http.StatusOK, rec.Code)

		rec = days.do(t, http.MethodGet, "/api/days/2026-11-02", alice, "")
		assert.JSONEq(t, monday, rec.Body.String())
	})

	t.Run("not signed in", func(t *testing.T) {
		t.Parallel()

		rec := newDaysAPI().do(t, http.MethodGet, "/api/days/2026-11-02", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

type dayJSON struct {
	Date         string   `json:"date"`
	WeightKg     *float64 `json:"weightKg"`
	KcalExpended *float64 `json:"kcalExpended"`
}

// putMonday logs monday as one of alice's days.
func (a api) putMonday(t *testing.T) {
	t.Helper()

	rec := a.do(t, http.MethodPut, "/api/days/2026-11-02", alice, monday)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

func decodeDay(t *testing.T, rec *httptest.ResponseRecorder) dayJSON {
	t.Helper()

	var result dayJSON
	decode(t, rec, &result)

	return result
}
