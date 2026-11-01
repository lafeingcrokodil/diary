package firestoredb

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"

	"github.com/lafeingcrokodil/diary/internal/recipe"
)

// recipesCollection holds every user's recipes, with each recipe's owner in a
// field.
const recipesCollection = "recipes"

// recipeDoc is how a recipe is stored. The recipe's ID is the document's ID.
type recipeDoc struct {
	Owner       string          `firestore:"owner"`
	Name        string          `firestore:"name"`
	SourceURL   string          `firestore:"sourceUrl"`
	Ingredients []ingredientDoc `firestore:"ingredients"`
	Steps       []string        `firestore:"steps"`
	Notes       string          `firestore:"notes"`
	Archived    bool            `firestore:"archived"`
}

func (d recipeDoc) owner() string { return d.Owner }

type ingredientDoc struct {
	FoodID  string  `firestore:"foodId"`
	Serving string  `firestore:"serving"`
	Amount  float64 `firestore:"amount"`
	Note    string  `firestore:"note"`
}

func toRecipeDoc(owner string, item recipe.Recipe) recipeDoc {
	ingredients := make([]ingredientDoc, 0, len(item.Ingredients))
	for _, ingredient := range item.Ingredients {
		ingredients = append(ingredients, ingredientDoc(ingredient))
	}

	return recipeDoc{
		Owner:       owner,
		Name:        item.Name,
		SourceURL:   item.SourceURL,
		Ingredients: ingredients,
		// Never nil, so that a recipe without steps is stored as an empty
		// array, like one without ingredients.
		Steps:    append([]string{}, item.Steps...),
		Notes:    item.Notes,
		Archived: item.Archived,
	}
}

func fromRecipeDoc(id string, doc recipeDoc) recipe.Recipe {
	ingredients := make([]recipe.Ingredient, 0, len(doc.Ingredients))
	for _, ingredient := range doc.Ingredients {
		ingredients = append(ingredients, recipe.Ingredient(ingredient))
	}

	return recipe.Recipe{
		ID:          id,
		Owner:       doc.Owner,
		Name:        doc.Name,
		SourceURL:   doc.SourceURL,
		Ingredients: ingredients,
		Steps:       doc.Steps,
		Notes:       doc.Notes,
		Archived:    doc.Archived,
	}
}

// RecipeStore is a recipe.Store that keeps recipes in Firestore.
type RecipeStore struct {
	client  *firestore.Client
	recipes *firestore.CollectionRef
}

// NewRecipeStore returns a RecipeStore that uses client.
func NewRecipeStore(client *firestore.Client) *RecipeStore {
	return &RecipeStore{client: client, recipes: client.Collection(recipesCollection)}
}

var _ recipe.Store = (*RecipeStore)(nil)

// Create implements recipe.Store.
func (s *RecipeStore) Create(ctx context.Context, owner string, item recipe.Recipe) (recipe.Recipe, error) {
	ref := s.recipes.NewDoc()

	doc := toRecipeDoc(owner, item)
	if _, err := ref.Create(ctx, doc); err != nil {
		return recipe.Recipe{}, fmt.Errorf("creating recipe: %w", err)
	}

	return fromRecipeDoc(ref.ID, doc), nil
}

// Get implements recipe.Store.
func (s *RecipeStore) Get(ctx context.Context, owner, id string) (recipe.Recipe, error) {
	ref := docRef(s.recipes, id)
	if ref == nil {
		return recipe.Recipe{}, recipe.ErrNotFound
	}

	snapshot, err := ref.Get(ctx)

	var doc recipeDoc
	if err := readOwned(owner, snapshot, err, recipe.ErrNotFound, &doc); err != nil {
		return recipe.Recipe{}, err
	}

	return fromRecipeDoc(id, doc), nil
}

// List implements recipe.Store.
func (s *RecipeStore) List(ctx context.Context, owner string) ([]recipe.Recipe, error) {
	snapshots, err := s.recipes.Where("owner", "==", owner).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("listing recipes: %w", err)
	}

	recipes := make([]recipe.Recipe, 0, len(snapshots))

	for _, snapshot := range snapshots {
		var doc recipeDoc
		if err := snapshot.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("reading recipe %s: %w", snapshot.Ref.ID, err)
		}

		recipes = append(recipes, fromRecipeDoc(snapshot.Ref.ID, doc))
	}

	// Firestore sorts strings case-sensitively, so the store sorts them itself.
	recipe.SortByName(recipes)

	return recipes, nil
}

// Update implements recipe.Store.
func (s *RecipeStore) Update(ctx context.Context, owner string, item recipe.Recipe) (recipe.Recipe, error) {
	ref := docRef(s.recipes, item.ID)
	if ref == nil {
		return recipe.Recipe{}, recipe.ErrNotFound
	}

	doc := toRecipeDoc(owner, item)

	// Checking the owner and writing happen in one transaction, so that the
	// recipe can't change hands in between.
	err := s.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		snapshot, err := tx.Get(ref)
		if err := readOwned(owner, snapshot, err, recipe.ErrNotFound, &recipeDoc{}); err != nil {
			return err
		}

		return tx.Set(ref, doc)
	})
	if err != nil {
		return recipe.Recipe{}, wrap("updating recipe", err, recipe.ErrNotFound)
	}

	return fromRecipeDoc(item.ID, doc), nil
}

// Delete implements recipe.Store.
func (s *RecipeStore) Delete(ctx context.Context, owner, id string) error {
	ref := docRef(s.recipes, id)
	if ref == nil {
		return recipe.ErrNotFound
	}

	// Checking that the recipe is archived and deleting it happen in one
	// transaction, so that it can't be unarchived in between.
	err := s.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		snapshot, err := tx.Get(ref)

		var doc recipeDoc
		if err := readOwned(owner, snapshot, err, recipe.ErrNotFound, &doc); err != nil {
			return err
		}

		if !doc.Archived {
			return recipe.ErrNotArchived
		}

		return tx.Delete(ref)
	})

	return wrap("deleting recipe", err, recipe.ErrNotFound)
}
