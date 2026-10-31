package server_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/lafeingcrokodil/diary/internal/auth"
	"github.com/lafeingcrokodil/diary/internal/server"
)

const (
	indexHTML = "<h1>Diary</h1>"
	// signedIn is the Authorization header that fakeAuth lets through.
	signedIn = "Bearer valid"
	email    = "me@example.com"
)

type testCase struct {
	name       string
	method     string
	path       string
	authHeader string
	wantStatus int
	wantBody   string
}

func TestHealth(t *testing.T) {
	t.Parallel()

	runTests(t, []testCase{
		{name: "GET succeeds", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK, wantBody: "ok\n"},
		{name: "POST is rejected", method: http.MethodPost, path: "/health", wantStatus: http.StatusMethodNotAllowed},
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

func TestMe(t *testing.T) {
	t.Parallel()

	runTests(t, []testCase{
		{
			name: "signed in", path: "/api/me", authHeader: signedIn,
			wantStatus: http.StatusOK, wantBody: `{"email":"me@example.com"}` + "\n",
		},
		{name: "not signed in", path: "/api/me", wantStatus: http.StatusUnauthorized},
		{
			name: "POST is rejected", method: http.MethodPost, path: "/api/me", authHeader: signedIn,
			wantStatus: http.StatusMethodNotAllowed,
		},
	})
}

func TestUnknownAPIRoute(t *testing.T) {
	t.Parallel()

	runTests(t, []testCase{
		{name: "GET", path: "/api/unknown", authHeader: signedIn, wantStatus: http.StatusNotFound},
		{name: "POST", method: http.MethodPost, path: "/api/unknown", authHeader: signedIn, wantStatus: http.StatusNotFound},
		{name: "not signed in", path: "/api/unknown", wantStatus: http.StatusUnauthorized},
	})
}

// fakeAuth stands in for auth.Middleware: it lets through requests with the
// signedIn header, as the user with the address email.
func fakeAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != signedIn {
			http.Error(resp, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(resp, req.WithContext(auth.WithEmail(req.Context(), email)))
	})
}

// runTests sends each test case's request to a server with a small front end,
// and checks the response. An empty method means GET, an empty authHeader
// means that the request isn't signed in, and an empty wantBody means that the
// body isn't checked.
func runTests(t *testing.T, tests []testCase) {
	t.Helper()

	handler := server.New(fstest.MapFS{
		"index.html":     {Data: []byte(indexHTML)},
		"_app/app.js":    {Data: []byte("console.log('hi');")},
		"_app/immutable": {Mode: fs.ModeDir},
	}, fakeAuth)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), test.method, test.path, nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}

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
