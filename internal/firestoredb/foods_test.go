package firestoredb_test

import (
	"crypto/rand"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/firestoredb"
	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/food/foodtest"
)

// newClient returns a client for the Firestore emulator, in a project of its
// own, so that every client starts with an empty database. It skips the test if
// the emulator isn't running.
func newClient(t *testing.T) *firestore.Client {
	t.Helper()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST isn't set, so there's no Firestore emulator to test against")
	}

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
