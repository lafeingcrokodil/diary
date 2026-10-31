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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)

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
	require.NoError(t, err)

	joseSigner, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: s.key}, nil)
	require.NoError(t, err)

	signed, err := joseSigner.Sign(payloadJSON)
	require.NoError(t, err)

	token, err := signed.CompactSerialize()
	require.NoError(t, err)

	return "Bearer " + token
}

func (s signer) verifier() *oidc.IDTokenVerifier {
	keys := &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&s.key.PublicKey}}

	return oidc.NewVerifier(issuer, keys, &oidc.Config{ClientID: projectID})
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	trusted, untrusted := newSigner(t), newSigner(t)
	expired := time.Now().Add(-time.Minute).Unix()
	// The claims of a user that the middleware should let through, apart from
	// the issuer, audience and expiry claims that every token gets.
	allowedUser := map[string]any{"email": allowed, "email_verified": true}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
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

	handler := auth.Middleware(trusted.verifier(), []string{allowed})(
		http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			email, ok := auth.Email(req.Context())
			assert.True(t, ok, "the request's context has no email")
			assert.NotEmpty(t, email)

			resp.WriteHeader(http.StatusOK)
		}),
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/me", nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Code)

			if rec.Code == http.StatusUnauthorized {
				assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
			}
		})
	}
}
