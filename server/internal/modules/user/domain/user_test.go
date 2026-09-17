package user

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserUpdateByAdmin(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, time.March, 4, 5, 6, 7, 0, time.UTC)

	t.Run("updates all fields", func(t *testing.T) {
		user := &User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: RoleUser}
		name, username, password, role := " New Name ", " new-user ", "new-hash", RoleAdmin

		err := user.UpdateByAdmin(&name, &username, &password, &role, "admin-1", now)

		if err != nil {
			t.Fatalf("UpdateByAdmin() error = %v", err)
		}
		if user.Name != "New Name" || user.Username != "new-user" || user.Password != password || user.Role != RoleAdmin || user.UpdatedBy != "admin-1" || !user.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected user: %#v", user)
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		user := &User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: RoleAdmin}
		name := "Updated"

		err := user.UpdateByAdmin(&name, nil, nil, nil, "admin-1", now)

		if err != nil || user.Name != name || user.Username != "old" || user.Password != "old-hash" || user.Role != RoleAdmin {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects invalid input without mutating the user", func(t *testing.T) {
		user := &User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: RoleUser}
		name := "Valid"
		invalidRole := Role("Unknown")

		err := user.UpdateByAdmin(&name, nil, nil, &invalidRole, "admin-1", now)

		if !errors.Is(err, ErrInvalidRole) || user.Name != "Old" || user.Role != RoleUser {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})
}
