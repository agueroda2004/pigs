package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	userapplication "server/internal/modules/user/application"
	userdomain "server/internal/modules/user/domain"
)

func TestUpdateOwnUserServiceExecute(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)

	t.Run("updates name and password", func(t *testing.T) {
		user := testUser(userID)
		name := " New Name "
		password := "new-secret"
		repository := &fakeUserRepository{getUser: user}
		hasher := &fakePasswordHasher{hash: "new-hash"}
		service := userapplication.NewUpdateOwnUserService(repository, hasher, func() time.Time { return now })

		updated, err := service.Execute(context.Background(), userID, userapplication.UpdateOwnUserCommand{Name: &name, Password: &password})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated != user || user.Name != "New Name" || user.Password != "new-hash" || user.UpdatedBy != userID.String() {
			t.Fatalf("unexpected user: %#v", user)
		}
		if !user.UpdatedAt.Equal(now) || repository.updated != user || hasher.password != password {
			t.Fatalf("update dependencies or timestamp are incorrect")
		}
	})

	t.Run("does not hash when password is omitted", func(t *testing.T) {
		user := testUser(userID)
		name := "Updated"
		hasher := &fakePasswordHasher{hash: "unused"}
		repository := &fakeUserRepository{getUser: user}
		service := userapplication.NewUpdateOwnUserService(repository, hasher, time.Now)

		_, err := service.Execute(context.Background(), userID, userapplication.UpdateOwnUserCommand{Name: &name})

		if err != nil || hasher.password != "" || user.Password != "old-hash" {
			t.Fatalf("unexpected result: err=%v, hashInput=%q, password=%q", err, hasher.password, user.Password)
		}
	})

	t.Run("propagates get, hash, validation, and update errors", func(t *testing.T) {
		getErr := errors.New("get failed")
		repository := &fakeUserRepository{getErr: getErr}
		service := userapplication.NewUpdateOwnUserService(repository, &fakePasswordHasher{}, time.Now)
		_, err := service.Execute(context.Background(), userID, userapplication.UpdateOwnUserCommand{})
		if !errors.Is(err, getErr) {
			t.Fatalf("get error = %v", err)
		}

		hashErr := errors.New("hash failed")
		user := testUser(userID)
		repository = &fakeUserRepository{getUser: user}
		service = userapplication.NewUpdateOwnUserService(repository, &fakePasswordHasher{err: hashErr}, time.Now)
		password := "new-secret"
		_, err = service.Execute(context.Background(), userID, userapplication.UpdateOwnUserCommand{Password: &password})
		if !errors.Is(err, hashErr) || repository.updated != nil {
			t.Fatalf("hash error = %v, updated=%v", err, repository.updated)
		}

		repository = &fakeUserRepository{getUser: testUser(userID)}
		service = userapplication.NewUpdateOwnUserService(repository, &fakePasswordHasher{}, time.Now)
		_, err = service.Execute(context.Background(), userID, userapplication.UpdateOwnUserCommand{})
		if !errors.Is(err, userdomain.ErrInvalidUpdate) || repository.updated != nil {
			t.Fatalf("validation error = %v", err)
		}

		updateErr := errors.New("update failed")
		repository = &fakeUserRepository{getUser: testUser(userID), updateErr: updateErr}
		name := "Updated"
		service = userapplication.NewUpdateOwnUserService(repository, &fakePasswordHasher{}, time.Now)
		_, err = service.Execute(context.Background(), userID, userapplication.UpdateOwnUserCommand{Name: &name})
		if !errors.Is(err, updateErr) {
			t.Fatalf("update error = %v", err)
		}
	})
}
