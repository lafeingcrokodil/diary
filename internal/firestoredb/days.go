package firestoredb

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lafeingcrokodil/diary/internal/diary"
)

// daysCollection holds every user's days, with each day's owner and date in
// fields.
const daysCollection = "days"

// dayDoc is how a day is stored. Days that nothing is logged for aren't
// stored.
type dayDoc struct {
	Owner        string     `firestore:"owner"`
	Date         string     `firestore:"date"`
	WeightKg     *float64   `firestore:"weightKg"`
	KcalExpended *float64   `firestore:"kcalExpended"`
	Entries      []entryDoc `firestore:"entries"`
}

func (d dayDoc) owner() string { return d.Owner }

type entryDoc struct {
	Time     string  `firestore:"time"`
	FoodID   string  `firestore:"foodId"`
	Serving  string  `firestore:"serving"`
	RecipeID string  `firestore:"recipeId"`
	Amount   float64 `firestore:"amount"`
}

func toDayDoc(owner string, day diary.Day) dayDoc {
	sorted := diary.SortedEntries(day.Entries)

	// Never nil, so that a day without entries is stored as an empty array.
	entries := make([]entryDoc, 0, len(sorted))
	for _, entry := range sorted {
		entries = append(entries, entryDoc(entry))
	}

	return dayDoc{
		Owner:        owner,
		Date:         day.Date,
		WeightKg:     day.WeightKg,
		KcalExpended: day.KcalExpended,
		Entries:      entries,
	}
}

func fromDayDoc(doc dayDoc) diary.Day {
	// Nil if there are no entries, as diary.SortedEntries returns them.
	var entries []diary.Entry
	if len(doc.Entries) > 0 {
		entries = make([]diary.Entry, 0, len(doc.Entries))
		for _, entry := range doc.Entries {
			entries = append(entries, diary.Entry(entry))
		}
	}

	return diary.Day{
		Owner:        doc.Owner,
		Date:         doc.Date,
		WeightKg:     doc.WeightKg,
		KcalExpended: doc.KcalExpended,
		Entries:      entries,
	}
}

// errInvalidOwner is returned for an owner ID that can't be part of a
// document ID.
var errInvalidOwner = errors.New("invalid owner")

// DayStore is a diary.Store that keeps days in Firestore.
type DayStore struct {
	days *firestore.CollectionRef
}

// NewDayStore returns a DayStore that uses client.
func NewDayStore(client *firestore.Client) *DayStore {
	return &DayStore{days: client.Collection(daysCollection)}
}

var _ diary.Store = (*DayStore)(nil)

// Get implements diary.Store.
func (s *DayStore) Get(ctx context.Context, owner, date string) (diary.Day, error) {
	ref, err := s.dayRef(owner, date)
	if err != nil {
		return diary.Day{}, err
	}

	snapshot, err := ref.Get(ctx)
	if status.Code(err) == codes.NotFound {
		return diary.Day{Owner: owner, Date: date}, nil
	}

	var doc dayDoc
	// The ID includes the owner, so the document can only be someone else's
	// if it was written some other way.
	if err := readOwned(owner, snapshot, err, errNotOwned, &doc); err != nil {
		return diary.Day{}, fmt.Errorf("getting day: %w", err)
	}

	return fromDayDoc(doc), nil
}

// errNotOwned is returned when a day's document belongs to someone other than
// the owner in its ID.
var errNotOwned = errors.New("day belongs to someone else")

// Put implements diary.Store.
func (s *DayStore) Put(ctx context.Context, owner string, day diary.Day) (diary.Day, error) {
	ref, err := s.dayRef(owner, day.Date)
	if err != nil {
		return diary.Day{}, err
	}

	doc := toDayDoc(owner, day)

	if day.IsEmpty() {
		// Deleting a document that doesn't exist succeeds.
		if _, err := ref.Delete(ctx); err != nil {
			return diary.Day{}, fmt.Errorf("deleting day: %w", err)
		}
	} else if _, err := ref.Set(ctx, doc); err != nil {
		return diary.Day{}, fmt.Errorf("storing day: %w", err)
	}

	return fromDayDoc(doc), nil
}

// dayRef returns the document for owner's day with the given date. Its ID is
// made of both, so that each user has at most one document per day and it
// can be found without a query. Dates have a fixed length, so no two owners
// and dates give the same ID.
func (s *DayStore) dayRef(owner, date string) (*firestore.DocumentRef, error) {
	if err := diary.ValidateDate(date); err != nil {
		return nil, fmt.Errorf("finding day: %w", err)
	}

	ref := docRef(s.days, owner+"_"+date)
	if owner == "" || ref == nil {
		return nil, fmt.Errorf("%w: %q", errInvalidOwner, owner)
	}

	return ref, nil
}
