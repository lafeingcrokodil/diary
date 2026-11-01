// Command diary runs the diary web server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/lafeingcrokodil/diary/internal/auth"
	"github.com/lafeingcrokodil/diary/internal/firestoredb"
	"github.com/lafeingcrokodil/diary/internal/logging"
	"github.com/lafeingcrokodil/diary/internal/server"
	"github.com/lafeingcrokodil/diary/web"
)

// readHeaderTimeout limits how long a client may take to send request headers,
// which protects against Slowloris-style attacks.
const readHeaderTimeout = 10 * time.Second

var errMissingEnv = errors.New("required environment variable not set")

func main() {
	slog.SetDefault(slog.New(logging.NewHandler(os.Stdout, slog.LevelInfo)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	frontend, err := web.Build()
	if err != nil {
		slog.Error("loading front end", "err", err)
		os.Exit(1)
	}

	projectID, allowedEmails, err := authConfig()
	if err != nil {
		slog.Error("loading auth configuration", "err", err)
		os.Exit(1)
	}

	// The verifier fetches signing keys as needed for as long as the server runs.
	verifier := auth.NewFirebaseVerifier(context.Background(), projectID)

	// The app's data is in the same project's Firestore database. The client
	// talks to the emulator instead if FIRESTORE_EMULATOR_HOST is set.
	firestoreClient, err := firestore.NewClient(context.Background(), projectID)
	if err != nil {
		slog.Error("connecting to Firestore", "err", err)
		os.Exit(1)
	}

	stores := server.Stores{
		Foods:     firestoredb.NewFoodStore(firestoreClient),
		Nutrients: firestoredb.NewNutrientStore(firestoreClient),
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           server.New(frontend, auth.Middleware(verifier, allowedEmails), stores),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	slog.Info("starting server", "addr", srv.Addr)

	// ListenAndServe only ever returns a non-nil error.
	err = srv.ListenAndServe()
	slog.Error("server stopped", "err", err)
	os.Exit(1)
}

// authConfig reads the ID of the Firebase project whose users can sign in, and
// whose Firestore database the app uses (FIREBASE_PROJECT_ID), and the
// comma-separated email addresses of the users who are allowed to use the app
// (ALLOWED_EMAILS). Both are required.
func authConfig() (string, []string, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		return "", nil, fmt.Errorf("%w: FIREBASE_PROJECT_ID", errMissingEnv)
	}

	var allowedEmails []string

	for email := range strings.SplitSeq(os.Getenv("ALLOWED_EMAILS"), ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			allowedEmails = append(allowedEmails, email)
		}
	}

	if len(allowedEmails) == 0 {
		return "", nil, fmt.Errorf("%w: ALLOWED_EMAILS", errMissingEnv)
	}

	return projectID, allowedEmails, nil
}
