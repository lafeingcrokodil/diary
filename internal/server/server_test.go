package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lafeingcrokodil/diary/internal/server"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantBody   string
	}{
		{name: "GET succeeds", method: http.MethodGet, wantStatus: http.StatusOK, wantBody: "ok\n"},
		{name: "POST is rejected", method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed, wantBody: ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), testCase.method, "/healthz", nil)
			rec := httptest.NewRecorder()

			server.New().ServeHTTP(rec, req)

			if rec.Code != testCase.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, testCase.wantStatus)
			}

			if testCase.wantBody != "" && rec.Body.String() != testCase.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), testCase.wantBody)
			}
		})
	}
}
