package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	boarapplication "server/internal/modules/boar/application"
	"server/internal/modules/boar/ports"
)

func TestDeleteBoarServiceExecute(t *testing.T) {
	boarID := uuid.New()

	t.Run("deletes the boar by its identifier", func(t *testing.T) {
		repository := &fakeBoarRepository{}
		service := boarapplication.NewDeleteBoarService(repository)

		if err := service.Execute(context.Background(), boarID); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deletedID != boarID {
			t.Fatalf("deleted id = %v, want %v", repository.deletedID, boarID)
		}
	})

	t.Run("propagates a missing boar", func(t *testing.T) {
		repository := &fakeBoarRepository{deleteErr: ports.ErrBoarNotFound}
		service := boarapplication.NewDeleteBoarService(repository)

		if err := service.Execute(context.Background(), boarID); !errors.Is(err, ports.ErrBoarNotFound) {
			t.Fatalf("error = %v, want ErrBoarNotFound", err)
		}
	})

	t.Run("propagates a boar in use", func(t *testing.T) {
		repository := &fakeBoarRepository{deleteErr: ports.ErrBoarInUse}
		service := boarapplication.NewDeleteBoarService(repository)

		if err := service.Execute(context.Background(), boarID); !errors.Is(err, ports.ErrBoarInUse) {
			t.Fatalf("error = %v, want ErrBoarInUse", err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeBoarRepository{deleteErr: expected}
		service := boarapplication.NewDeleteBoarService(repository)

		if err := service.Execute(context.Background(), boarID); !errors.Is(err, expected) {
			t.Fatalf("error = %v, want %v", err, expected)
		}
	})
}
