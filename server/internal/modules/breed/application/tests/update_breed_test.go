package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	breedapplication "server/internal/modules/breed/application"
	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

func TestUpdateBreedServiceExecute(t *testing.T) {
	breedID := uuid.New()
	updatedBy := uuid.New()
	now := time.Date(2026, time.March, 4, 5, 6, 7, 0, time.UTC)

	t.Run("updates name and active", func(t *testing.T) {
		breed := testBreed(breedID)
		name := "  New Name  "
		active := false
		repository := &fakeBreedRepository{getBreed: breed}
		service := breedapplication.NewUpdateBreedService(repository, func() time.Time { return now })

		updated, err := service.Execute(context.Background(), breedID, breedapplication.UpdateBreedCommand{
			Name: &name, Active: &active, UpdatedBy: updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated != breed || breed.Name != "New Name" || breed.Active || breed.UpdatedBy != updatedBy || !breed.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected breed: %#v", breed)
		}
		if repository.updated != breed {
			t.Fatalf("repository did not receive the updated breed")
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		breed := testBreed(breedID)
		active := false
		repository := &fakeBreedRepository{getBreed: breed}
		service := breedapplication.NewUpdateBreedService(repository, time.Now)

		_, err := service.Execute(context.Background(), breedID, breedapplication.UpdateBreedCommand{Active: &active, UpdatedBy: updatedBy})

		if err != nil || breed.Active || breed.Name != "Old Name" {
			t.Fatalf("unexpected result: err=%v breed=%#v", err, breed)
		}
	})

	t.Run("propagates dependency and validation errors", func(t *testing.T) {
		getErr := errors.New("get failed")
		repository := &fakeBreedRepository{getErr: getErr}
		service := breedapplication.NewUpdateBreedService(repository, time.Now)
		_, err := service.Execute(context.Background(), breedID, breedapplication.UpdateBreedCommand{UpdatedBy: updatedBy})
		if !errors.Is(err, getErr) {
			t.Fatalf("get error = %v", err)
		}

		repository = &fakeBreedRepository{getBreed: testBreed(breedID)}
		service = breedapplication.NewUpdateBreedService(repository, time.Now)
		_, err = service.Execute(context.Background(), breedID, breedapplication.UpdateBreedCommand{UpdatedBy: updatedBy})
		if !errors.Is(err, breeddomain.ErrInvalidUpdate) || repository.updated != nil {
			t.Fatalf("empty command error = %v", err)
		}

		name := "New Name"
		repository = &fakeBreedRepository{getBreed: testBreed(breedID), updateErr: ports.ErrBreedNameAlreadyUsed}
		service = breedapplication.NewUpdateBreedService(repository, time.Now)
		_, err = service.Execute(context.Background(), breedID, breedapplication.UpdateBreedCommand{Name: &name, UpdatedBy: updatedBy})
		if !errors.Is(err, ports.ErrBreedNameAlreadyUsed) {
			t.Fatalf("repository error = %v", err)
		}
	})
}
