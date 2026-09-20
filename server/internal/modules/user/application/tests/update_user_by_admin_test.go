package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	userapplication "server/internal/modules/user/application"
	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

func TestUpdateUserByAdminServiceExecute(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, time.March, 4, 5, 6, 7, 0, time.UTC)

	t.Run("updates all fields and hashes password", func(t *testing.T) {
		user := testUser(userID)
		name, username, password, role := "New Name", "new-user", "new-secret", userdomain.RoleAdmin
		repository := &fakeUserRepository{getUser: user}
		hasher := &fakePasswordHasher{hash: "new-hash"}
		service := userapplication.NewUpdateUserByAdminService(repository, hasher, func() time.Time { return now })

		updated, err := service.Execute(context.Background(), userID, userapplication.UpdateUserByAdminCommand{
			Name: &name, Username: &username, Password: &password, Role: &role, UpdatedBy: "admin-1",
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated != user || user.Name != name || user.Username != username || user.Password != "new-hash" || user.Role != role || user.UpdatedBy != "admin-1" || !user.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected user: %#v", user)
		}
		if hasher.password != password || repository.updated != user {
			t.Fatalf("dependencies received unexpected values")
		}
	})

	t.Run("does not hash when password is omitted", func(t *testing.T) {
		user := testUser(userID)
		username := "new-user"
		hasher := &fakePasswordHasher{hash: "unused"}
		repository := &fakeUserRepository{getUser: user}
		service := userapplication.NewUpdateUserByAdminService(repository, hasher, time.Now)

		_, err := service.Execute(context.Background(), userID, userapplication.UpdateUserByAdminCommand{Username: &username, UpdatedBy: "admin-1"})

		if err != nil || hasher.password != "" || user.Password != "old-hash" {
			t.Fatalf("unexpected result: err=%v hashInput=%q password=%q", err, hasher.password, user.Password)
		}
	})

	t.Run("propagates dependency and validation errors", func(t *testing.T) {
		getErr := errors.New("get failed")
		repository := &fakeUserRepository{getErr: getErr}
		service := userapplication.NewUpdateUserByAdminService(repository, &fakePasswordHasher{}, time.Now)
		_, err := service.Execute(context.Background(), userID, userapplication.UpdateUserByAdminCommand{UpdatedBy: "admin-1"})
		if !errors.Is(err, getErr) {
			t.Fatalf("get error = %v", err)
		}

		hashErr := errors.New("hash failed")
		password := "secret"
		repository = &fakeUserRepository{getUser: testUser(userID)}
		service = userapplication.NewUpdateUserByAdminService(repository, &fakePasswordHasher{err: hashErr}, time.Now)
		_, err = service.Execute(context.Background(), userID, userapplication.UpdateUserByAdminCommand{Password: &password, UpdatedBy: "admin-1"})
		if !errors.Is(err, hashErr) || repository.updated != nil {
			t.Fatalf("hash error = %v updated=%v", err, repository.updated)
		}

		repository = &fakeUserRepository{getUser: testUser(userID)}
		service = userapplication.NewUpdateUserByAdminService(repository, &fakePasswordHasher{}, time.Now)
		_, err = service.Execute(context.Background(), userID, userapplication.UpdateUserByAdminCommand{UpdatedBy: "admin-1"})
		if !errors.Is(err, userdomain.ErrInvalidUpdate) || repository.updated != nil {
			t.Fatalf("empty command error = %v", err)
		}

		username := "taken"
		repository = &fakeUserRepository{getUser: testUser(userID), updateErr: ports.ErrUsernameAlreadyUsed}
		service = userapplication.NewUpdateUserByAdminService(repository, &fakePasswordHasher{}, time.Now)
		_, err = service.Execute(context.Background(), userID, userapplication.UpdateUserByAdminCommand{Username: &username, UpdatedBy: "admin-1"})
		if !errors.Is(err, ports.ErrUsernameAlreadyUsed) {
			t.Fatalf("repository error = %v", err)
		}
	})
}
