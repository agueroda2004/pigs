package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	userapplication "server/internal/modules/user/application"
	userdomain "server/internal/modules/user/domain"
)

func TestListUsersServiceExecute(t *testing.T) {
	t.Run("returns every user from the repository", func(t *testing.T) {
		users := []*userdomain.User{
			testUser(uuid.New()),
			testUser(uuid.New()),
		}
		repository := &fakeUserRepository{listUsers: users}
		service := userapplication.NewListUsersService(repository)

		result, err := service.Execute(context.Background())

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(users) || result[0] != users[0] || result[1] != users[1] {
			t.Fatalf("unexpected users: %#v", result)
		}
	})

	t.Run("returns an empty list when there are no users", func(t *testing.T) {
		repository := &fakeUserRepository{listUsers: []*userdomain.User{}}
		service := userapplication.NewListUsersService(repository)

		result, err := service.Execute(context.Background())

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: users=%#v err=%v", result, err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeUserRepository{listErr: expected}
		service := userapplication.NewListUsersService(repository)

		_, err := service.Execute(context.Background())

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
