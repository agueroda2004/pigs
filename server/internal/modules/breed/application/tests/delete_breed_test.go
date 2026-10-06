package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	breedapplication "server/internal/modules/breed/application"
	"server/internal/modules/breed/ports"
)

func TestDeleteBreedServiceExecute(t *testing.T) {
	breedID := uuid.New()

	t.Run("deletes the breed by its identifier", func(t *testing.T) {
		repository := &fakeBreedRepository{}
		service := breedapplication.NewDeleteBreedService(repository)

		if err := service.Execute(context.Background(), breedID); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deletedID != breedID {
			t.Fatalf("deleted id = %v, want %v", repository.deletedID, breedID)
		}
	})

	t.Run("propagates a missing breed", func(t *testing.T) {
		repository := &fakeBreedRepository{deleteErr: ports.ErrBreedNotFound}
		service := breedapplication.NewDeleteBreedService(repository)

		if err := service.Execute(context.Background(), breedID); !errors.Is(err, ports.ErrBreedNotFound) {
			t.Fatalf("error = %v, want ErrBreedNotFound", err)
		}
	})

	t.Run("propagates a breed in use", func(t *testing.T) {
		repository := &fakeBreedRepository{deleteErr: ports.ErrBreedInUse}
		service := breedapplication.NewDeleteBreedService(repository)

		if err := service.Execute(context.Background(), breedID); !errors.Is(err, ports.ErrBreedInUse) {
			t.Fatalf("error = %v, want ErrBreedInUse", err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeBreedRepository{deleteErr: expected}
		service := breedapplication.NewDeleteBreedService(repository)

		if err := service.Execute(context.Background(), breedID); !errors.Is(err, expected) {
			t.Fatalf("error = %v, want %v", err, expected)
		}
	})
}
