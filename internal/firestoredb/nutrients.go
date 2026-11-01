package firestoredb

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"

	"github.com/lafeingcrokodil/diary/internal/nutrient"
)

// nutrientsCollection holds every user's nutrients, with each nutrient's owner
// in a field.
const nutrientsCollection = "nutrients"

// nutrientDoc is how a nutrient is stored. The nutrient's ID is the document's
// ID.
type nutrientDoc struct {
	Owner       string  `firestore:"owner"`
	Name        string  `firestore:"name"`
	Unit        string  `firestore:"unit"`
	KcalPerGram float64 `firestore:"kcalPerGram"`
	Archived    bool    `firestore:"archived"`
}

func (d nutrientDoc) owner() string { return d.Owner }

func toNutrientDoc(owner string, item nutrient.Nutrient) nutrientDoc {
	return nutrientDoc{
		Owner:       owner,
		Name:        item.Name,
		Unit:        string(item.Unit),
		KcalPerGram: item.KcalPerGram,
		Archived:    item.Archived,
	}
}

func fromNutrientDoc(id string, doc nutrientDoc) nutrient.Nutrient {
	return nutrient.Nutrient{
		ID:          id,
		Owner:       doc.Owner,
		Name:        doc.Name,
		Unit:        nutrient.Unit(doc.Unit),
		KcalPerGram: doc.KcalPerGram,
		Archived:    doc.Archived,
	}
}

// NutrientStore is a nutrient.Store that keeps nutrients in Firestore.
type NutrientStore struct {
	client    *firestore.Client
	nutrients *firestore.CollectionRef
}

// NewNutrientStore returns a NutrientStore that uses client.
func NewNutrientStore(client *firestore.Client) *NutrientStore {
	return &NutrientStore{client: client, nutrients: client.Collection(nutrientsCollection)}
}

var _ nutrient.Store = (*NutrientStore)(nil)

// Create implements nutrient.Store.
func (s *NutrientStore) Create(ctx context.Context, owner string, item nutrient.Nutrient) (nutrient.Nutrient, error) {
	ref := s.nutrients.NewDoc()

	doc := toNutrientDoc(owner, item)
	if _, err := ref.Create(ctx, doc); err != nil {
		return nutrient.Nutrient{}, fmt.Errorf("creating nutrient: %w", err)
	}

	return fromNutrientDoc(ref.ID, doc), nil
}

// Get implements nutrient.Store.
func (s *NutrientStore) Get(ctx context.Context, owner, id string) (nutrient.Nutrient, error) {
	ref := docRef(s.nutrients, id)
	if ref == nil {
		return nutrient.Nutrient{}, nutrient.ErrNotFound
	}

	snapshot, err := ref.Get(ctx)

	var doc nutrientDoc
	if err := readOwned(owner, snapshot, err, nutrient.ErrNotFound, &doc); err != nil {
		return nutrient.Nutrient{}, err
	}

	return fromNutrientDoc(id, doc), nil
}

// List implements nutrient.Store.
func (s *NutrientStore) List(ctx context.Context, owner string) ([]nutrient.Nutrient, error) {
	snapshots, err := s.nutrients.Where("owner", "==", owner).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("listing nutrients: %w", err)
	}

	nutrients := make([]nutrient.Nutrient, 0, len(snapshots))

	for _, snapshot := range snapshots {
		var doc nutrientDoc
		if err := snapshot.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("reading nutrient %s: %w", snapshot.Ref.ID, err)
		}

		nutrients = append(nutrients, fromNutrientDoc(snapshot.Ref.ID, doc))
	}

	// Firestore sorts strings case-sensitively, so the store sorts them itself.
	nutrient.SortByName(nutrients)

	return nutrients, nil
}

// Update implements nutrient.Store.
func (s *NutrientStore) Update(ctx context.Context, owner string, item nutrient.Nutrient) (nutrient.Nutrient, error) {
	ref := docRef(s.nutrients, item.ID)
	if ref == nil {
		return nutrient.Nutrient{}, nutrient.ErrNotFound
	}

	doc := toNutrientDoc(owner, item)

	// Checking the owner and writing happen in one transaction, so that the
	// nutrient can't change hands in between.
	err := s.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		snapshot, err := tx.Get(ref)
		if err := readOwned(owner, snapshot, err, nutrient.ErrNotFound, &nutrientDoc{}); err != nil {
			return err
		}

		return tx.Set(ref, doc)
	})
	if err != nil {
		return nutrient.Nutrient{}, wrap("updating nutrient", err, nutrient.ErrNotFound)
	}

	return fromNutrientDoc(item.ID, doc), nil
}

// Delete implements nutrient.Store.
func (s *NutrientStore) Delete(ctx context.Context, owner, id string) error {
	ref := docRef(s.nutrients, id)
	if ref == nil {
		return nutrient.ErrNotFound
	}

	// Checking that the nutrient is archived and deleting it happen in one
	// transaction, so that it can't be unarchived in between.
	err := s.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		snapshot, err := tx.Get(ref)

		var doc nutrientDoc
		if err := readOwned(owner, snapshot, err, nutrient.ErrNotFound, &doc); err != nil {
			return err
		}

		if !doc.Archived {
			return nutrient.ErrNotArchived
		}

		return tx.Delete(ref)
	})

	return wrap("deleting nutrient", err, nutrient.ErrNotFound)
}
