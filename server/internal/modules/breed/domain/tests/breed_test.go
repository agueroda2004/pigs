package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
)

func TestNewBreed(t *testing.T) {
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	breedID := uuid.New()
	createdBy := uuid.New()

	t.Run("creates a valid active breed", func(t *testing.T) {
		breed, err := breeddomain.NewBreed(breedID, "  Duroc  ", createdBy, now)

		if err != nil {
			t.Fatalf("NewBreed() error = %v", err)
		}
		if breed.ID != breedID || breed.Name != "Duroc" || !breed.Active {
			t.Fatalf("unexpected breed: %#v", breed)
		}
		if breed.CreatedBy != createdBy || breed.UpdatedBy != createdBy {
			t.Fatalf("unexpected audit fields: %#v", breed)
		}
		if !breed.CreatedAt.Equal(now) || !breed.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", breed)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		_, err := breeddomain.NewBreed(uuid.Nil, "Duroc", createdBy, now)
		if !errors.Is(err, breeddomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid name", func(t *testing.T) {
		_, err := breeddomain.NewBreed(breedID, "  ", createdBy, now)
		if !errors.Is(err, breeddomain.ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}

		_, err = breeddomain.NewBreed(breedID, strings.Repeat("a", 101), createdBy, now)
		if !errors.Is(err, breeddomain.ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}
	})

	t.Run("accepts name at exactly max length", func(t *testing.T) {
		name := strings.Repeat("a", 100)
		breed, err := breeddomain.NewBreed(breedID, name, createdBy, now)
		if err != nil || breed.Name != name {
			t.Fatalf("unexpected result: err=%v breed=%#v", err, breed)
		}
	})

	t.Run("rejects nil created by", func(t *testing.T) {
		_, err := breeddomain.NewBreed(breedID, "Duroc", uuid.Nil, now)
		if !errors.Is(err, breeddomain.ErrInvalidCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidCreatedBy", err)
		}
	})
}

func TestBreedUpdate(t *testing.T) {
	breedID := uuid.New()
	updatedBy := uuid.New()
	now := time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)

	t.Run("updates name and active", func(t *testing.T) {
		breed := &breeddomain.Breed{ID: breedID, Name: "Old", Active: true}
		name := "  New Name  "
		active := false

		err := breed.Update(&name, &active, updatedBy, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if breed.Name != "New Name" || breed.Active || breed.UpdatedBy != updatedBy || !breed.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected breed: %#v", breed)
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		breed := &breeddomain.Breed{ID: breedID, Name: "Old", Active: true}
		name := "Updated"

		err := breed.Update(&name, nil, updatedBy, now)

		if err != nil || breed.Name != name || !breed.Active {
			t.Fatalf("unexpected result: err=%v breed=%#v", err, breed)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		breed := &breeddomain.Breed{ID: breedID, Name: "Old", Active: true}

		err := breed.Update(nil, nil, updatedBy, now)

		if !errors.Is(err, breeddomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects nil updated by", func(t *testing.T) {
		breed := &breeddomain.Breed{ID: breedID, Name: "Old", Active: true}
		name := "New Name"

		err := breed.Update(&name, nil, uuid.Nil, now)

		if !errors.Is(err, breeddomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects invalid name without mutating the breed", func(t *testing.T) {
		breed := &breeddomain.Breed{ID: breedID, Name: "Old", Active: true}
		name := "  "
		active := false

		err := breed.Update(&name, &active, updatedBy, now)

		if !errors.Is(err, breeddomain.ErrInvalidName) || breed.Name != "Old" || !breed.Active {
			t.Fatalf("unexpected result: err=%v breed=%#v", err, breed)
		}
	})

	t.Run("rejects nil receiver", func(t *testing.T) {
		var breed *breeddomain.Breed
		name := "New Name"

		err := breed.Update(&name, nil, updatedBy, now)

		if !errors.Is(err, breeddomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}
