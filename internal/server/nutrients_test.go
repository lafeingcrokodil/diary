package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/nutrient"
	"github.com/lafeingcrokodil/diary/internal/nutrient/nutrienttest"
	"github.com/lafeingcrokodil/diary/internal/server"
)

const protein = `{"name":"Protein","unit":"g","kcalPerGram":4,"archived":false}`

type nutrientJSON struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Unit        string  `json:"unit"`
	KcalPerGram float64 `json:"kcalPerGram"`
	Archived    bool    `json:"archived"`
}

// newNutrientsAPI returns an api whose nutrients are kept in store.
func newNutrientsAPI(store nutrient.Store) api {
	return newAPI(server.Stores{Nutrients: store})
}

// createProtein adds protein as one of alice's nutrients and returns it.
func (a api) createProtein(t *testing.T) nutrientJSON {
	t.Helper()

	rec := a.do(t, http.MethodPost, "/api/nutrients", alice, protein)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	return decodeNutrient(t, rec)
}

// archive archives one of alice's nutrients.
func (a api) archive(t *testing.T, item nutrientJSON) {
	t.Helper()

	body := `{"name":"` + item.Name + `","unit":"` + item.Unit + `","archived":true}`
	rec := a.do(t, http.MethodPut, "/api/nutrients/"+item.ID, alice, body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

func decodeNutrient(t *testing.T, rec *httptest.ResponseRecorder) nutrientJSON {
	t.Helper()

	var result nutrientJSON
	decode(t, rec, &result)

	return result
}

func TestNutrients(t *testing.T) {
	t.Parallel()

	t.Run("create, get and list", func(t *testing.T) {
		t.Parallel()

		nutrients := newNutrientsAPI(nutrienttest.NewMemoryStore())

		rec := nutrients.do(t, http.MethodPost, "/api/nutrients", alice, protein)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

		created := decodeNutrient(t, rec)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "/api/nutrients/"+created.ID, rec.Header().Get("Location"))
		assert.JSONEq(t, protein, withoutID(t, rec.Body.String()))

		rec = nutrients.do(t, http.MethodGet, "/api/nutrients/"+created.ID, alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, created, decodeNutrient(t, rec))

		rec = nutrients.do(t, http.MethodGet, "/api/nutrients", alice, "")
		require.Equal(t, http.StatusOK, rec.Code)

		var list []nutrientJSON
		decode(t, rec, &list)
		assert.Equal(t, []nutrientJSON{created}, list)
	})

	t.Run("list without nutrients", func(t *testing.T) {
		t.Parallel()

		rec := newNutrientsAPI(nutrienttest.NewMemoryStore()).do(t, http.MethodGet, "/api/nutrients", alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, "[]", rec.Body.String())
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		nutrients := newNutrientsAPI(nutrienttest.NewMemoryStore())
		created := nutrients.createProtein(t)

		// The ID in the body is ignored in favour of the one in the path.
		changed := `{"id":"other","name":"Salt","unit":"mg"}`

		rec := nutrients.do(t, http.MethodPut, "/api/nutrients/"+created.ID, alice, changed)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		updated := decodeNutrient(t, rec)
		assert.Equal(t, nutrientJSON{ID: created.ID, Name: "Salt", Unit: "mg"}, updated)

		rec = nutrients.do(t, http.MethodGet, "/api/nutrients/"+created.ID, alice, "")
		assert.Equal(t, updated, decodeNutrient(t, rec))
	})

	t.Run("archived nutrients are listed and can be unarchived", func(t *testing.T) {
		t.Parallel()

		nutrients := newNutrientsAPI(nutrienttest.NewMemoryStore())
		created := nutrients.createProtein(t)
		nutrients.archive(t, created)

		rec := nutrients.do(t, http.MethodGet, "/api/nutrients", alice, "")

		var list []nutrientJSON
		decode(t, rec, &list)
		require.Len(t, list, 1)
		assert.True(t, list[0].Archived)

		rec = nutrients.do(t, http.MethodPut, "/api/nutrients/"+created.ID, alice, protein)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.False(t, decodeNutrient(t, rec).Archived)
	})

	t.Run("delete an archived nutrient", func(t *testing.T) {
		t.Parallel()

		nutrients := newNutrientsAPI(nutrienttest.NewMemoryStore())
		created := nutrients.createProtein(t)
		nutrients.archive(t, created)

		rec := nutrients.do(t, http.MethodDelete, "/api/nutrients/"+created.ID, alice, "")
		assert.Equal(t, http.StatusNoContent, rec.Code)

		rec = nutrients.do(t, http.MethodGet, "/api/nutrients/"+created.ID, alice, "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("nutrients that aren't archived can't be deleted", func(t *testing.T) {
		t.Parallel()

		nutrients := newNutrientsAPI(nutrienttest.NewMemoryStore())
		created := nutrients.createProtein(t)

		rec := nutrients.do(t, http.MethodDelete, "/api/nutrients/"+created.ID, alice, "")
		assert.Equal(t, http.StatusConflict, rec.Code)

		rec = nutrients.do(t, http.MethodGet, "/api/nutrients/"+created.ID, alice, "")
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing nutrient", func(t *testing.T) {
		t.Parallel()

		nutrients := newNutrientsAPI(nutrienttest.NewMemoryStore())

		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			rec := nutrients.do(t, method, "/api/nutrients/missing", alice, protein)
			assert.Equal(t, http.StatusNotFound, rec.Code, method)
		}
	})

	t.Run("users can't reach each other's nutrients", func(t *testing.T) {
		t.Parallel()

		nutrients := newNutrientsAPI(nutrienttest.NewMemoryStore())
		alicesNutrient := nutrients.createProtein(t)
		nutrients.archive(t, alicesNutrient)
		path := "/api/nutrients/" + alicesNutrient.ID

		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			rec := nutrients.do(t, method, path, bob, protein)
			assert.Equal(t, http.StatusNotFound, rec.Code, method)
		}

		rec := nutrients.do(t, http.MethodGet, "/api/nutrients", bob, "")
		assert.JSONEq(t, "[]", rec.Body.String())

		rec = nutrients.do(t, http.MethodGet, path, alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, decodeNutrient(t, rec).Archived)
	})

	t.Run("not signed in", func(t *testing.T) {
		t.Parallel()

		rec := newNutrientsAPI(nutrienttest.NewMemoryStore()).do(t, http.MethodGet, "/api/nutrients", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestNutrientsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{"not JSON", "protein"},
		{"wrong type", `{"name":"Protein","unit":"g","kcalPerGram":"four"}`},
		{"unknown field", `{"name":"Protein","unit":"g","colour":"brown"}`},
		{"invalid nutrient", `{"name":"Protein","unit":"oz"}`},
		{"too large", `{"name":"` + strings.Repeat("a", 1<<20) + `","unit":"g"}`},
	}

	nutrients := newNutrientsAPI(nutrienttest.NewMemoryStore())
	created := nutrients.createProtein(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rec := nutrients.do(t, http.MethodPost, "/api/nutrients", alice, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			rec = nutrients.do(t, http.MethodPut, "/api/nutrients/"+created.ID, alice, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

// failingNutrientStore is a nutrient.Store whose every method fails.
type failingNutrientStore struct {
	nutrient.Store
}

func (failingNutrientStore) Create(context.Context, string, nutrient.Nutrient) (nutrient.Nutrient, error) {
	return nutrient.Nutrient{}, errStore
}

func (failingNutrientStore) Get(context.Context, string, string) (nutrient.Nutrient, error) {
	return nutrient.Nutrient{}, errStore
}

func (failingNutrientStore) List(context.Context, string) ([]nutrient.Nutrient, error) {
	return nil, errStore
}

func (failingNutrientStore) Update(context.Context, string, nutrient.Nutrient) (nutrient.Nutrient, error) {
	return nutrient.Nutrient{}, errStore
}

// failingDeleteStore is a nutrient store whose Delete fails, so that the
// nutrient can be found and checked first.
type failingDeleteStore struct {
	*nutrienttest.MemoryStore
}

func (failingDeleteStore) Delete(context.Context, string, string) error {
	return errStore
}

func TestNutrientsStoreError(t *testing.T) {
	t.Parallel()

	nutrients := newNutrientsAPI(failingNutrientStore{})

	requests := []struct{ method, path, body string }{
		{http.MethodGet, "/api/nutrients", ""},
		{http.MethodPost, "/api/nutrients", protein},
		{http.MethodGet, "/api/nutrients/some-id", ""},
		{http.MethodPut, "/api/nutrients/some-id", protein},
		{http.MethodDelete, "/api/nutrients/some-id", ""},
	}

	for _, request := range requests {
		rec := nutrients.do(t, request.method, request.path, alice, request.body)
		assert.Equal(t, http.StatusInternalServerError, rec.Code, request.method+" "+request.path)
		assert.NotContains(t, rec.Body.String(), errStore.Error())
	}

	failingDelete := newNutrientsAPI(failingDeleteStore{nutrienttest.NewMemoryStore()})
	created := failingDelete.createProtein(t)
	failingDelete.archive(t, created)

	rec := failingDelete.do(t, http.MethodDelete, "/api/nutrients/"+created.ID, alice, "")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
