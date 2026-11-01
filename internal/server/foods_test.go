package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/food/foodtest"
	"github.com/lafeingcrokodil/diary/internal/server"
)

const (
	alice = "alice"
	bob   = "bob"
	bread = `{"name":"Rye bread","brand":"Bakery","unit":"g","kcalPer100":210,` +
		`"servings":[{"name":"slice","amount":40}]}`
)

type foodJSON struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Brand      string  `json:"brand"`
	Unit       string  `json:"unit"`
	KcalPer100 float64 `json:"kcalPer100"`
	Servings   []struct {
		Name   string  `json:"name"`
		Amount float64 `json:"amount"`
	} `json:"servings"`
}

// api sends requests to a server whose data is kept in stores.
type api struct {
	handler http.Handler
}

func newAPI(stores server.Stores) api {
	return api{handler: server.New(fstest.MapFS{"index.html": {Data: []byte(indexHTML)}}, fakeAuth, stores)}
}

// newFoodsAPI returns an api whose foods are kept in store.
func newFoodsAPI(store food.Store) api {
	return newAPI(server.Stores{Foods: store})
}

// do sends a request as user (none if empty) and returns the response.
func (a api) do(t *testing.T, method, path, user, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	if user != "" {
		req.Header.Set("Authorization", "Bearer "+user)
	}

	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)

	return rec
}

// createBread adds bread as one of alice's foods and returns it.
func (a api) createBread(t *testing.T) foodJSON {
	t.Helper()

	rec := a.do(t, http.MethodPost, "/api/foods", alice, bread)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	return decodeFood(t, rec)
}

// decode checks that a response is JSON, and decodes it into value.
func decode(t *testing.T, rec *httptest.ResponseRecorder, value any) {
	t.Helper()

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), value))
}

func decodeFood(t *testing.T, rec *httptest.ResponseRecorder) foodJSON {
	t.Helper()

	var result foodJSON
	decode(t, rec, &result)

	return result
}

func TestFoods(t *testing.T) {
	t.Parallel()

	t.Run("create, get and list", func(t *testing.T) {
		t.Parallel()

		foods := newFoodsAPI(foodtest.NewMemoryStore())

		rec := foods.do(t, http.MethodPost, "/api/foods", alice, bread)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

		created := decodeFood(t, rec)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "/api/foods/"+created.ID, rec.Header().Get("Location"))
		assert.JSONEq(t, bread, withoutID(t, rec.Body.String()))

		rec = foods.do(t, http.MethodGet, "/api/foods/"+created.ID, alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, created, decodeFood(t, rec))

		rec = foods.do(t, http.MethodGet, "/api/foods", alice, "")
		require.Equal(t, http.StatusOK, rec.Code)

		var list []foodJSON
		decode(t, rec, &list)
		assert.Equal(t, []foodJSON{created}, list)
	})

	t.Run("list without foods", func(t *testing.T) {
		t.Parallel()

		rec := newFoodsAPI(foodtest.NewMemoryStore()).do(t, http.MethodGet, "/api/foods", alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, "[]", rec.Body.String())
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		foods := newFoodsAPI(foodtest.NewMemoryStore())
		created := foods.createBread(t)

		// The ID in the body is ignored in favour of the one in the path.
		changed := `{"id":"other","name":"Milk","unit":"ml","kcalPer100":64,"servings":[]}`

		rec := foods.do(t, http.MethodPut, "/api/foods/"+created.ID, alice, changed)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		updated := decodeFood(t, rec)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Milk", updated.Name)
		assert.Empty(t, updated.Brand)

		rec = foods.do(t, http.MethodGet, "/api/foods/"+created.ID, alice, "")
		assert.Equal(t, updated, decodeFood(t, rec))
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		foods := newFoodsAPI(foodtest.NewMemoryStore())
		created := foods.createBread(t)

		rec := foods.do(t, http.MethodDelete, "/api/foods/"+created.ID, alice, "")
		assert.Equal(t, http.StatusNoContent, rec.Code)

		rec = foods.do(t, http.MethodGet, "/api/foods/"+created.ID, alice, "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("missing food", func(t *testing.T) {
		t.Parallel()

		foods := newFoodsAPI(foodtest.NewMemoryStore())

		assert.Equal(t, http.StatusNotFound, foods.do(t, http.MethodGet, "/api/foods/missing", alice, "").Code)
		assert.Equal(t, http.StatusNotFound, foods.do(t, http.MethodPut, "/api/foods/missing", alice, bread).Code)
		assert.Equal(t, http.StatusNotFound, foods.do(t, http.MethodDelete, "/api/foods/missing", alice, "").Code)
	})

	t.Run("users can't reach each other's foods", func(t *testing.T) {
		t.Parallel()

		foods := newFoodsAPI(foodtest.NewMemoryStore())
		alicesFood := foods.createBread(t)
		path := "/api/foods/" + alicesFood.ID

		assert.Equal(t, http.StatusNotFound, foods.do(t, http.MethodGet, path, bob, "").Code)
		assert.Equal(t, http.StatusNotFound, foods.do(t, http.MethodPut, path, bob, bread).Code)
		assert.Equal(t, http.StatusNotFound, foods.do(t, http.MethodDelete, path, bob, "").Code)

		rec := foods.do(t, http.MethodGet, "/api/foods", bob, "")
		assert.JSONEq(t, "[]", rec.Body.String())

		rec = foods.do(t, http.MethodGet, path, alice, "")
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("not signed in", func(t *testing.T) {
		t.Parallel()

		rec := newFoodsAPI(foodtest.NewMemoryStore()).do(t, http.MethodGet, "/api/foods", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestFoodsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{"not JSON", "bread"},
		{"wrong type", `{"name":"Rye bread","unit":"g","kcalPer100":"lots"}`},
		{"unknown field", `{"name":"Rye bread","unit":"g","kcalPer100":210,"colour":"brown"}`},
		{"invalid food", `{"name":"","unit":"g","kcalPer100":210}`},
		{"too large", `{"name":"` + strings.Repeat("a", 1<<20) + `","unit":"g"}`},
	}

	foods := newFoodsAPI(foodtest.NewMemoryStore())
	created := foods.createBread(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rec := foods.do(t, http.MethodPost, "/api/foods", alice, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			rec = foods.do(t, http.MethodPut, "/api/foods/"+created.ID, alice, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

// failingFoodStore is a food.Store whose List fails.
type failingFoodStore struct {
	food.Store
}

var errStore = errors.New("store unavailable")

func (failingFoodStore) List(_ context.Context, _ string) ([]food.Food, error) {
	return nil, errStore
}

func TestFoodsStoreError(t *testing.T) {
	t.Parallel()

	rec := newFoodsAPI(failingFoodStore{}).do(t, http.MethodGet, "/api/foods", alice, "")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), errStore.Error())
}

// withoutID returns body, a JSON object, without its "id" field.
func withoutID(t *testing.T, body string) string {
	t.Helper()

	var fields map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &fields))
	delete(fields, "id")

	result, err := json.Marshal(fields)
	require.NoError(t, err)

	return string(result)
}
