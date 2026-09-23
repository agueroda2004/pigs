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

func TestCreateBreedServiceExecute(t *testing.T) {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	createdBy := uuid.New()

	t.Run("creates a breed with a normalized name", func(t *testing.T) {
		repository := &fakeBreedRepository{}
		service := breedapplication.NewCreateBreedService(repository, func() time.Time { return createdAt })

		breed, err := service.Execute(context.Background(), breedapplication.CreateBreedCommand{
			Name:      "  Duroc  ",
			CreatedBy: createdBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if breed.ID == uuid.Nil || breed.Name != "Duroc" || !breed.Active {
			t.Fatalf("unexpected breed: %#v", breed)
		}
		if breed.CreatedBy != createdBy || breed.UpdatedBy != createdBy || !breed.CreatedAt.Equal(createdAt) {
			t.Fatalf("unexpected audit fields: %#v", breed)
		}
		if repository.created != breed {
			t.Fatalf("repository did not receive the created breed")
		}
	})

	t.Run("returns duplicate name error", func(t *testing.T) {
		repository := &fakeBreedRepository{exists: true}
		service := breedapplication.NewCreateBreedService(repository, time.Now)

		_, err := service.Execute(context.Background(), breedapplication.CreateBreedCommand{Name: "Duroc", CreatedBy: createdBy})

		if !errors.Is(err, ports.ErrBreedNameAlreadyUsed) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%v", err, repository.created)
		}
	})

	t.Run("propagates repository existence error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeBreedRepository{existsErr: expected}
		service := breedapplication.NewCreateBreedService(repository, time.Now)

		_, err := service.Execute(context.Background(), breedapplication.CreateBreedCommand{Name: "Duroc", CreatedBy: createdBy})

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})

	t.Run("propagates validation and create errors", func(t *testing.T) {
		repository := &fakeBreedRepository{createErr: errors.New("create failed")}
		service := breedapplication.NewCreateBreedService(repository, time.Now)

		_, err := service.Execute(context.Background(), breedapplication.CreateBreedCommand{Name: "  ", CreatedBy: createdBy})
		if !errors.Is(err, breeddomain.ErrInvalidName) || repository.created != nil {
			t.Fatalf("unexpected validation result: %v", err)
		}

		_, err = service.Execute(context.Background(), breedapplication.CreateBreedCommand{Name: "Duroc", CreatedBy: createdBy})
		if !errors.Is(err, repository.createErr) {
			t.Fatalf("unexpected create result: %v", err)
		}
	})
}
