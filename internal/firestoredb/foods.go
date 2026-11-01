package firestoredb

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"

	"github.com/lafeingcrokodil/diary/internal/food"
)

// foodsCollection holds every user's foods, with each food's owner in a field.
const foodsCollection = "foods"

// foodDoc is how a food is stored. The food's ID is the document's ID.
type foodDoc struct {
	Owner      string       `firestore:"owner"`
	Name       string       `firestore:"name"`
	Brand      string       `firestore:"brand"`
	Unit       string       `firestore:"unit"`
	KcalPer100 float64      `firestore:"kcalPer100"`
	Servings   []servingDoc `firestore:"servings"`
	// Nutrients is keyed by nutrient ID.
	Nutrients map[string]float64 `firestore:"nutrients"`
}

func (d foodDoc) owner() string { return d.Owner }

type servingDoc struct {
	Name   string  `firestore:"name"`
	Amount float64 `firestore:"amount"`
}

func toDoc(owner string, item food.Food) foodDoc {
	servings := make([]servingDoc, 0, len(item.Servings))
	for _, serving := range item.Servings {
		servings = append(servings, servingDoc(serving))
	}

	return foodDoc{
		Owner:      owner,
		Name:       item.Name,
		Brand:      item.Brand,
		Unit:       string(item.Unit),
		KcalPer100: item.KcalPer100,
		Servings:   servings,
		Nutrients:  item.Nutrients,
	}
}

func fromDoc(id string, doc foodDoc) food.Food {
	servings := make([]food.Serving, 0, len(doc.Servings))
	for _, serving := range doc.Servings {
		servings = append(servings, food.Serving(serving))
	}

	return food.Food{
		ID:         id,
		Owner:      doc.Owner,
		Name:       doc.Name,
		Brand:      doc.Brand,
		Unit:       food.Unit(doc.Unit),
		KcalPer100: doc.KcalPer100,
		Servings:   servings,
		Nutrients:  doc.Nutrients,
	}
}

// FoodStore is a food.Store that keeps foods in Firestore.
type FoodStore struct {
	client *firestore.Client
	foods  *firestore.CollectionRef
}

// NewFoodStore returns a FoodStore that uses client.
func NewFoodStore(client *firestore.Client) *FoodStore {
	return &FoodStore{client: client, foods: client.Collection(foodsCollection)}
}

var _ food.Store = (*FoodStore)(nil)

// Create implements food.Store.
func (s *FoodStore) Create(ctx context.Context, owner string, item food.Food) (food.Food, error) {
	ref := s.foods.NewDoc()

	doc := toDoc(owner, item)
	if _, err := ref.Create(ctx, doc); err != nil {
		return food.Food{}, fmt.Errorf("creating food: %w", err)
	}

	return fromDoc(ref.ID, doc), nil
}

// Get implements food.Store.
func (s *FoodStore) Get(ctx context.Context, owner, id string) (food.Food, error) {
	ref := docRef(s.foods, id)
	if ref == nil {
		return food.Food{}, food.ErrNotFound
	}

	snapshot, err := ref.Get(ctx)

	var doc foodDoc
	if err := readOwned(owner, snapshot, err, food.ErrNotFound, &doc); err != nil {
		return food.Food{}, err
	}

	return fromDoc(id, doc), nil
}

// List implements food.Store.
func (s *FoodStore) List(ctx context.Context, owner string) ([]food.Food, error) {
	snapshots, err := s.foods.Where("owner", "==", owner).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("listing foods: %w", err)
	}

	foods := make([]food.Food, 0, len(snapshots))

	for _, snapshot := range snapshots {
		var doc foodDoc
		if err := snapshot.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("reading food %s: %w", snapshot.Ref.ID, err)
		}

		foods = append(foods, fromDoc(snapshot.Ref.ID, doc))
	}

	// Firestore sorts strings case-sensitively, so the store sorts them itself.
	food.SortByName(foods)

	return foods, nil
}

// Update implements food.Store.
func (s *FoodStore) Update(ctx context.Context, owner string, item food.Food) (food.Food, error) {
	ref := docRef(s.foods, item.ID)
	if ref == nil {
		return food.Food{}, food.ErrNotFound
	}

	doc := toDoc(owner, item)

	// Checking the owner and writing happen in one transaction, so that the
	// food can't change hands in between.
	err := s.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		snapshot, err := tx.Get(ref)
		if err := readOwned(owner, snapshot, err, food.ErrNotFound, &foodDoc{}); err != nil {
			return err
		}

		return tx.Set(ref, doc)
	})
	if err != nil {
		return food.Food{}, wrap("updating food", err, food.ErrNotFound)
	}

	return fromDoc(item.ID, doc), nil
}

// Delete implements food.Store.
func (s *FoodStore) Delete(ctx context.Context, owner, id string) error {
	ref := docRef(s.foods, id)
	if ref == nil {
		return food.ErrNotFound
	}

	err := s.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		snapshot, err := tx.Get(ref)
		if err := readOwned(owner, snapshot, err, food.ErrNotFound, &foodDoc{}); err != nil {
			return err
		}

		return tx.Delete(ref)
	})

	return wrap("deleting food", err, food.ErrNotFound)
}
