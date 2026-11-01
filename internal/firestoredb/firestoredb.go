// Package firestoredb stores the app's data in Firestore.
package firestoredb

import (
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ownedDoc is a document that belongs to a user.
type ownedDoc interface {
	owner() string
}

// docRef returns the document with the given ID in collection, or nil if no
// document can have that ID: Firestore rejects empty IDs, and a slash would
// make the ID a path to a document somewhere else.
func docRef(collection *firestore.CollectionRef, id string) *firestore.DocumentRef {
	if id == "" || strings.Contains(id, "/") {
		return nil
	}

	return collection.Doc(id)
}

// readOwned decodes a document that was read from Firestore with the given
// error into doc, which must be a pointer, and returns errNotFound if the
// document doesn't exist or belongs to someone other than owner.
func readOwned(owner string, snapshot *firestore.DocumentSnapshot, err, errNotFound error, doc ownedDoc) error {
	if status.Code(err) == codes.NotFound {
		return errNotFound
	}

	if err != nil {
		return fmt.Errorf("reading document: %w", err)
	}

	if err := snapshot.DataTo(doc); err != nil {
		return fmt.Errorf("reading document %s: %w", snapshot.Ref.ID, err)
	}

	if doc.owner() != owner {
		return errNotFound
	}

	return nil
}

// wrap adds context to err, unless it's nil or errNotFound, which callers
// check for as it is.
func wrap(action string, err, errNotFound error) error {
	if err == nil || errors.Is(err, errNotFound) {
		return err
	}

	return fmt.Errorf("%s: %w", action, err)
}
