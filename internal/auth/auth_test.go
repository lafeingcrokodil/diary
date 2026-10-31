package auth_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"

	"github.com/lafeingcrokodil/diary/internal/auth"
)

const (
	projectID = "test-project"
	issuer    = "https://securetoken.google.com/" + projectID
	allowed   = "allowed@example.com"
)

// with returns a copy of claims with one claim set to value.
func with(claims map[string]any, name string, value any) map[string]any {
	result := maps.Clone(claims)
	result[name] = value

	return result
}

// signer creates ID tokens signed with a key that's generated for the test.
type signer struct {
	key *rsa.PrivateKey
}

func newSigner(t *testing.T) signer {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}

	return signer{key: key}
}

// bearer returns an Authorization header with a signed ID token that has the
// given claims, on top of valid issuer, audience and expiry claims.
func (s signer) bearer(t *testing.T, claims map[string]any) string {
	t.Helper()

	payload := map[string]any{
		"iss": issuer,
		"aud": projectID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	maps.Copy(payload, claims)

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encoding claims: %v", err)
	}

	joseSigner, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: s.key}, nil)
	if err != nil {
		t.Fatalf("creating signer: %v", err)
	}

	signed, err := joseSigner.Sign(payloadJSON)
	if err != nil {
		t.Fatalf("signing token: %v", err)
	}

	token, err := signed.CompactSerialize()
	if err != nil {
		t.Fatalf("serializing token: %v", err)
	}

	return "Bearer " + token
}

func (s signer) verifier() *oidc.IDTokenVerifier {
	keys := &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&s.key.PublicKey}}

	return oidc.NewVerifier(issuer, keys, &oidc.Config{ClientID: projectID})
}

type testCase struct {
	name       string
	authHeader string
	wantStatus int
}

func testCases(t *testing.T, trusted, untrusted signer) []testCase {
	t.Helper()

	expired := time.Now().Add(-time.Minute).Unix()
	// The claims of a user that the middleware should let through, apart from
	// the issuer, audience and expiry claims that every token gets.
	allowedUser := map[string]any{"email": allowed, "email_verified": true}

	return []testCase{
		{"allowed user", trusted.bearer(t, allowedUser), http.StatusOK},
		{"email in different case", trusted.bearer(t, with(allowedUser, "email", "Allowed@Example.com")), http.StatusOK},
		{"no header", "", http.StatusUnauthorized},
		{"not a bearer token", "Basic dXNlcjpwYXNz", http.StatusUnauthorized},
		{"malformed token", "Bearer not-a-token", http.StatusUnauthorized},
		{"signed by another key", untrusted.bearer(t, allowedUser), http.StatusUnauthorized},
		{"expired", trusted.bearer(t, with(allowedUser, "exp", expired)), http.StatusUnauthorized},
		{"other project", trusted.bearer(t, with(allowedUser, "aud", "other-project")), http.StatusUnauthorized},
		{"unverified email", trusted.bearer(t, with(allowedUser, "email_verified", false)), http.StatusForbidden},
		{"other user", trusted.bearer(t, with(allowedUser, "email", "other@example.com")), http.StatusForbidden},
		{"no email", trusted.bearer(t, map[string]any{}), http.StatusForbidden},
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	trusted := newSigner(t)

	handler := auth.Middleware(trusted.verifier(), []string{allowed})(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			email, ok := auth.Email(req.Context())
			if !ok || email == "" {
				t.Error("the request's context has no email")
			}

			resp.WriteHeader(http.StatusOK)
		}),
	)

	for _, test := range testCases(t, trusted, newSigner(t)) {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/me", nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, test.wantStatus)
			}

			if rec.Code == http.StatusUnauthorized && rec.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Errorf("WWW-Authenticate = %q, want %q", rec.Header().Get("WWW-Authenticate"), "Bearer")
			}
		})
	}
}
