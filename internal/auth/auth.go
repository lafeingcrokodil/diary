// Package auth checks that API requests come from users who are allowed to use
// the app, based on the ID tokens that Firebase Authentication issues.
package auth

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	firebaseIssuerPrefix = "https://securetoken.google.com/"
	// Google publishes the public keys that Firebase signs ID tokens with here.
	firebaseKeysURL = "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"
)

// IDTokenVerifier checks an ID token's signature, issuer, audience and expiry.
// *oidc.IDTokenVerifier implements it.
type IDTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

// NewFirebaseVerifier returns a verifier for the ID tokens that Firebase
// Authentication issues for the given project. The context is used for
// fetching the signing keys, which happens on demand, so it should outlive
// individual requests.
func NewFirebaseVerifier(ctx context.Context, projectID string) *oidc.IDTokenVerifier {
	keys := oidc.NewRemoteKeySet(ctx, firebaseKeysURL)

	return oidc.NewVerifier(firebaseIssuerPrefix+projectID, keys, &oidc.Config{ClientID: projectID})
}

// claims are the parts of an ID token's payload that the middleware checks.
type claims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"` //nolint:tagliatelle // The token format sets the claim's name.
}

type emailKey struct{}

// WithEmail returns a copy of ctx that carries the email address of the user
// making the request, which Email returns.
func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailKey{}, email)
}

// Email returns the email address of the user whose request the middleware let
// through.
func Email(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailKey{}).(string)

	return email, ok
}

// Middleware only passes on requests that carry a valid ID token (as
// "Authorization: Bearer <token>") of a user whose verified email address is
// one of allowedEmails. Other requests get a 401 if the token is missing or
// invalid, or a 403 if the user isn't allowed.
func Middleware(verifier IDTokenVerifier, allowedEmails []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			rawToken, ok := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
			if !ok || rawToken == "" {
				unauthorized(resp)
				return
			}

			token, err := verifier.Verify(req.Context(), rawToken)
			if err != nil {
				// Most of these are routine (e.g. an expired token), but logging them
				// also surfaces problems such as failing to fetch the signing keys.
				slog.WarnContext(req.Context(), "rejecting ID token", "err", err)
				unauthorized(resp)
				return
			}

			var tokenClaims claims

			err = token.Claims(&tokenClaims)
			if err != nil {
				slog.WarnContext(req.Context(), "parsing ID token claims", "err", err)
				unauthorized(resp)
				return
			}

			if !tokenClaims.EmailVerified || !isAllowed(tokenClaims.Email, allowedEmails) {
				slog.WarnContext(req.Context(), "rejecting user",
					"email", tokenClaims.Email, "emailVerified", tokenClaims.EmailVerified)
				http.Error(resp, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(resp, req.WithContext(WithEmail(req.Context(), tokenClaims.Email)))
		})
	}
}

func unauthorized(resp http.ResponseWriter) {
	resp.Header().Set("WWW-Authenticate", "Bearer")
	http.Error(resp, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

// isAllowed reports whether email is one of allowedEmails, ignoring case, since
// email providers treat addresses case-insensitively in practice.
func isAllowed(email string, allowedEmails []string) bool {
	return email != "" && slices.ContainsFunc(allowedEmails, func(allowed string) bool {
		return strings.EqualFold(email, allowed)
	})
}
