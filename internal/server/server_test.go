package server_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/lafeingcrokodil/diary/internal/server"
)

const indexHTML = "<h1>Diary</h1>"

type testCase struct {
	name       string
	method     string
	path       string
	wantStatus int
	wantBody   string
}

func TestHealth(t *testing.T) {
	t.Parallel()

	runTests(t, []testCase{
		{name: "GET succeeds", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK, wantBody: "ok\n"},
		{name: "POST is rejected", method: http.MethodPost, path: "/health", wantStatus: http.StatusMethodNotAllowed},
		{name: "old path still works", method: http.MethodGet, path: "/healthz", wantStatus: http.StatusOK, wantBody: "ok\n"},
	})
}

func TestFrontend(t *testing.T) {
	t.Parallel()

	runTests(t, []testCase{
		{name: "root serves index.html", path: "/", wantStatus: http.StatusOK, wantBody: indexHTML},
		{name: "app route serves index.html", path: "/recipes/123", wantStatus: http.StatusOK, wantBody: indexHTML},
		{name: "directory serves index.html", path: "/_app/immutable", wantStatus: http.StatusOK, wantBody: indexHTML},
		{name: "existing file", path: "/_app/app.js", wantStatus: http.StatusOK, wantBody: "console.log('hi');"},
		{name: "missing file", path: "/_app/old.js", wantStatus: http.StatusNotFound},
		{name: "POST is rejected", method: http.MethodPost, path: "/recipes/123", wantStatus: http.StatusMethodNotAllowed},
	})
}

func TestUnknownAPIRoute(t *testing.T) {
	t.Parallel()

	runTests(t, []testCase{
		{name: "GET", path: "/api/unknown", wantStatus: http.StatusNotFound},
		{name: "POST", method: http.MethodPost, path: "/api/unknown", wantStatus: http.StatusNotFound},
	})
}

// runTests sends each test case's request to a server with a small front end,
// and checks the response. An empty method means GET, and an empty wantBody
// means that the body isn't checked.
func runTests(t *testing.T, tests []testCase) {
	t.Helper()

	handler := server.New(fstest.MapFS{
		"index.html":     {Data: []byte(indexHTML)},
		"_app/app.js":    {Data: []byte("console.log('hi');")},
		"_app/immutable": {Mode: fs.ModeDir},
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), test.method, test.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, test.wantStatus)
			}

			if test.wantBody != "" && rec.Body.String() != test.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), test.wantBody)
			}
		})
	}
}
