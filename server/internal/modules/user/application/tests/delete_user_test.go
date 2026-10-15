package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	userapplication "server/internal/modules/user/application"
	"server/internal/modules/user/ports"
)

func TestDeleteUserServiceExecute(t *testing.T) {
	userID := uuid.New()

	t.Run("deletes the user by its identifier", func(t *testing.T) {
		repository := &fakeUserRepository{}
		service := userapplication.NewDeleteUserService(repository)

		if err := service.Execute(context.Background(), userID); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deletedID != userID {
			t.Fatalf("deleted id = %v, want %v", repository.deletedID, userID)
		}
	})

	t.Run("propagates a missing user", func(t *testing.T) {
		repository := &fakeUserRepository{deleteErr: ports.ErrUserNotFound}
		service := userapplication.NewDeleteUserService(repository)

		if err := service.Execute(context.Background(), userID); !errors.Is(err, ports.ErrUserNotFound) {
			t.Fatalf("error = %v, want ErrUserNotFound", err)
		}
	})

	t.Run("propagates a user with related records", func(t *testing.T) {
		repository := &fakeUserRepository{deleteErr: ports.ErrUserInUse}
		service := userapplication.NewDeleteUserService(repository)

		if err := service.Execute(context.Background(), userID); !errors.Is(err, ports.ErrUserInUse) {
			t.Fatalf("error = %v, want ErrUserInUse", err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeUserRepository{deleteErr: expected}
		service := userapplication.NewDeleteUserService(repository)

		if err := service.Execute(context.Background(), userID); !errors.Is(err, expected) {
			t.Fatalf("error = %v, want %v", err, expected)
		}
	})
}
