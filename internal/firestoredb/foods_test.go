package firestoredb_test

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/firestoredb"
	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/food/foodtest"
)

// emulatorTimeout is how long the tests wait for the Firestore emulator to
// respond before failing.
const emulatorTimeout = 5 * time.Second

// newClient returns a client for the Firestore emulator, in a project of its
// own, so that every client starts with an empty database. It skips the test if
// FIRESTORE_EMULATOR_HOST isn't set, and fails it if the emulator doesn't
// respond (the client would otherwise keep retrying until the test times out).
func newClient(t *testing.T) *firestore.Client {
	t.Helper()

	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST isn't set, so there's no Firestore emulator to test against")
	}

	requireEmulator(t, host)

	// The emulator accepts any project ID.
	client, err := firestore.NewClient(t.Context(), "test-"+strings.ToLower(rand.Text()))
	require.NoError(t, err)

	t.Cleanup(func() { _ = client.Close() })

	return client
}

func TestFoodStore(t *testing.T) {
	t.Parallel()

	foodtest.CheckStore(t, func() food.Store { return firestoredb.NewFoodStore(newClient(t)) })
}

// requireEmulator fails the test unless the Firestore emulator at host responds
// within a few seconds.
func requireEmulator(t *testing.T, host string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), emulatorTimeout)
	defer cancel()

	// The host comes from the environment of whoever runs the tests, so
	// requesting it isn't a risk.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+host, nil) //nolint:gosec // See above.
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // See above.
	require.NoError(t, err, "the Firestore emulator at %s isn't responding: is it running?", host)

	_ = resp.Body.Close()
}
