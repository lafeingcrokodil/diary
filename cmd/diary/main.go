// Command diary runs the diary web server.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lafeingcrokodil/diary/internal/server"
)

// readHeaderTimeout limits how long a client may take to send request headers,
// which protects against Slowloris-style attacks.
const readHeaderTimeout = 10 * time.Second

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           server.New(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	slog.Info("starting server", "addr", srv.Addr)

	// ListenAndServe only ever returns a non-nil error.
	err := srv.ListenAndServe()
	slog.Error("server stopped", "err", err)
	os.Exit(1)
}
