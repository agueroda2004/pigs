package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	sowapplication "server/internal/modules/sow/application"
	"server/internal/modules/sow/ports"
)

func TestDeleteSowServiceExecute(t *testing.T) {
	sowID := uuid.New()

	t.Run("deletes the sow by its identifier", func(t *testing.T) {
		repository := &fakeSowRepository{}
		service := sowapplication.NewDeleteSowService(repository)

		if err := service.Execute(context.Background(), sowID); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deletedID != sowID {
			t.Fatalf("deleted id = %v, want %v", repository.deletedID, sowID)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := &fakeSowRepository{deleteErr: ports.ErrSowNotFound}
		service := sowapplication.NewDeleteSowService(repository)

		if err := service.Execute(context.Background(), sowID); !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a sow in use", func(t *testing.T) {
		repository := &fakeSowRepository{deleteErr: ports.ErrSowInUse}
		service := sowapplication.NewDeleteSowService(repository)

		if err := service.Execute(context.Background(), sowID); !errors.Is(err, ports.ErrSowInUse) {
			t.Fatalf("error = %v, want ErrSowInUse", err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeSowRepository{deleteErr: expected}
		service := sowapplication.NewDeleteSowService(repository)

		if err := service.Execute(context.Background(), sowID); !errors.Is(err, expected) {
			t.Fatalf("error = %v, want %v", err, expected)
		}
	})
}
