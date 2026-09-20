package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	userapplication "server/internal/modules/user/application"
	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

func TestCreateUserServiceExecute(t *testing.T) {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	t.Run("creates a user with normalized username and hashed password", func(t *testing.T) {
		repository := &fakeUserRepository{}
		hasher := &fakePasswordHasher{hash: "hashed-password"}
		service := userapplication.NewCreateUserService(repository, hasher, func() time.Time { return createdAt })

		user, err := service.Execute(context.Background(), userapplication.CreateUserCommand{
			Name:      " Ana ",
			Username:  "  ana  ",
			Password:  "secret",
			CreatedBy: "admin",
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if user.Username != "ana" || user.Password != "hashed-password" {
			t.Fatalf("unexpected user: %#v", user)
		}
		if user.Role != userdomain.RoleUser || !user.CreatedAt.Equal(createdAt) {
			t.Fatalf("unexpected defaults: %#v", user)
		}
		if repository.created != user || hasher.password != "secret" {
			t.Fatalf("dependencies did not receive expected values")
		}
	})

	t.Run("returns duplicate username error", func(t *testing.T) {
		repository := &fakeUserRepository{exists: true}
		hasher := &fakePasswordHasher{hash: "unused"}
		service := userapplication.NewCreateUserService(repository, hasher, time.Now)

		_, err := service.Execute(context.Background(), userapplication.CreateUserCommand{Username: "ana", Password: "secret", Name: "Ana"})

		if !errors.Is(err, ports.ErrUsernameAlreadyUsed) || hasher.password != "" || repository.created != nil {
			t.Fatalf("unexpected result: err=%v, hashInput=%q, created=%v", err, hasher.password, repository.created)
		}
	})

	t.Run("propagates repository existence error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeUserRepository{existsErr: expected}
		service := userapplication.NewCreateUserService(repository, &fakePasswordHasher{}, time.Now)

		_, err := service.Execute(context.Background(), userapplication.CreateUserCommand{Username: "ana"})

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})

	t.Run("propagates hasher error", func(t *testing.T) {
		expected := errors.New("hash failed")
		repository := &fakeUserRepository{}
		hasher := &fakePasswordHasher{err: expected}
		service := userapplication.NewCreateUserService(repository, hasher, time.Now)

		_, err := service.Execute(context.Background(), userapplication.CreateUserCommand{Name: "Ana", Username: "ana", Password: "secret"})

		if !errors.Is(err, expected) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v, created=%v", err, repository.created)
		}
	})

	t.Run("propagates validation and create errors", func(t *testing.T) {
		repository := &fakeUserRepository{createErr: errors.New("create failed")}
		service := userapplication.NewCreateUserService(repository, &fakePasswordHasher{hash: "hash"}, time.Now)

		_, err := service.Execute(context.Background(), userapplication.CreateUserCommand{Name: "", Username: "ana", Password: "secret"})
		if !errors.Is(err, userdomain.ErrInvalidName) || repository.created != nil {
			t.Fatalf("unexpected validation result: %v", err)
		}

		_, err = service.Execute(context.Background(), userapplication.CreateUserCommand{Name: "Ana", Username: "ana", Password: "secret"})
		if !errors.Is(err, repository.createErr) {
			t.Fatalf("unexpected create result: %v", err)
		}
	})
}
