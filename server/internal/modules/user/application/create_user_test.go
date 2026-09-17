package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

func TestCreateUserServiceExecute(t *testing.T) {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	t.Run("creates a user with normalized username and hashed password", func(t *testing.T) {
		repository := &fakeUserRepository{}
		hasher := &fakePasswordHasher{hash: "hashed-password"}
		service := NewCreateUserService(repository, hasher, func() time.Time { return createdAt })

		user, err := service.Execute(context.Background(), CreateUserCommand{
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
		service := NewCreateUserService(repository, hasher, time.Now)

		_, err := service.Execute(context.Background(), CreateUserCommand{Username: "ana", Password: "secret", Name: "Ana"})

		if !errors.Is(err, ports.ErrUsernameAlreadyUsed) || hasher.password != "" || repository.created != nil {
			t.Fatalf("unexpected result: err=%v, hashInput=%q, created=%v", err, hasher.password, repository.created)
		}
	})

	t.Run("propagates repository existence error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeUserRepository{existsErr: expected}
		service := NewCreateUserService(repository, &fakePasswordHasher{}, time.Now)

		_, err := service.Execute(context.Background(), CreateUserCommand{Username: "ana"})

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})

	t.Run("propagates hasher error", func(t *testing.T) {
		expected := errors.New("hash failed")
		repository := &fakeUserRepository{}
		hasher := &fakePasswordHasher{err: expected}
		service := NewCreateUserService(repository, hasher, time.Now)

		_, err := service.Execute(context.Background(), CreateUserCommand{Name: "Ana", Username: "ana", Password: "secret"})

		if !errors.Is(err, expected) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v, created=%v", err, repository.created)
		}
	})

	t.Run("propagates validation and create errors", func(t *testing.T) {
		repository := &fakeUserRepository{createErr: errors.New("create failed")}
		service := NewCreateUserService(repository, &fakePasswordHasher{hash: "hash"}, time.Now)

		_, err := service.Execute(context.Background(), CreateUserCommand{Name: "", Username: "ana", Password: "secret"})
		if !errors.Is(err, userdomain.ErrInvalidName) || repository.created != nil {
			t.Fatalf("unexpected validation result: %v", err)
		}

		_, err = service.Execute(context.Background(), CreateUserCommand{Name: "Ana", Username: "ana", Password: "secret"})
		if !errors.Is(err, repository.createErr) {
			t.Fatalf("unexpected create result: %v", err)
		}
	})
}

type fakeUserRepository struct {
	exists    bool
	existsErr error
	getUser   *userdomain.User
	getErr    error
	createErr error
	updateErr error
	created   *userdomain.User
	updated   *userdomain.User
}

func (f *fakeUserRepository) Create(_ context.Context, user *userdomain.User) error {
	f.created = user
	return f.createErr
}

func (f *fakeUserRepository) GetByID(_ context.Context, _ uuid.UUID) (*userdomain.User, error) {
	return f.getUser, f.getErr
}

func (f *fakeUserRepository) ExistsByUsername(_ context.Context, _ string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeUserRepository) Update(_ context.Context, user *userdomain.User) error {
	f.updated = user
	return f.updateErr
}

type fakePasswordHasher struct {
	hash     string
	err      error
	password string
}

func (f *fakePasswordHasher) Hash(password string) (string, error) {
	f.password = password
	return f.hash, f.err
}
