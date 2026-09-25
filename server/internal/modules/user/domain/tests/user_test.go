package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
)

func TestNewUser(t *testing.T) {
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	userID := uuid.New()

	t.Run("creates a valid user", func(t *testing.T) {
		user, err := userdomain.NewUser(userID, "  Ana  ", "  ana  ", "hash", userdomain.RoleAdmin, "admin-1", now)

		if err != nil {
			t.Fatalf("NewUser() error = %v", err)
		}
		if user.ID != userID || user.Name != "Ana" || user.Username != "ana" || user.Password != "hash" || user.Role != userdomain.RoleAdmin || !user.Active || user.CreatedBy != "admin-1" {
			t.Fatalf("unexpected user: %#v", user)
		}
		if !user.CreatedAt.Equal(now) || !user.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", user)
		}
	})

	t.Run("defaults role to User when empty", func(t *testing.T) {
		user, err := userdomain.NewUser(userID, "Ana", "ana", "hash", "", "admin-1", now)

		if err != nil || user.Role != userdomain.RoleUser {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		_, err := userdomain.NewUser(uuid.Nil, "Ana", "ana", "hash", "", "admin-1", now)
		if !errors.Is(err, userdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid name", func(t *testing.T) {
		_, err := userdomain.NewUser(userID, " ", "ana", "hash", "", "admin-1", now)
		if !errors.Is(err, userdomain.ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}

		_, err = userdomain.NewUser(userID, strings.Repeat("a", 51), "ana", "hash", "", "admin-1", now)
		if !errors.Is(err, userdomain.ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}
	})

	t.Run("accepts name at exactly max length", func(t *testing.T) {
		name := strings.Repeat("a", 50)
		user, err := userdomain.NewUser(userID, name, "ana", "hash", "", "admin-1", now)
		if err != nil || user.Name != name {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects invalid username", func(t *testing.T) {
		_, err := userdomain.NewUser(userID, "Ana", " ", "hash", "", "admin-1", now)
		if !errors.Is(err, userdomain.ErrInvalidUsername) {
			t.Fatalf("error = %v, want ErrInvalidUsername", err)
		}

		_, err = userdomain.NewUser(userID, "Ana", strings.Repeat("a", 51), "hash", "", "admin-1", now)
		if !errors.Is(err, userdomain.ErrInvalidUsername) {
			t.Fatalf("error = %v, want ErrInvalidUsername", err)
		}
	})

	t.Run("accepts username at exactly max length", func(t *testing.T) {
		username := strings.Repeat("a", 50)
		user, err := userdomain.NewUser(userID, "Ana", username, "hash", "", "admin-1", now)
		if err != nil || user.Username != username {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects empty password hash", func(t *testing.T) {
		_, err := userdomain.NewUser(userID, "Ana", "ana", "   ", "", "admin-1", now)
		if !errors.Is(err, userdomain.ErrInvalidPassword) {
			t.Fatalf("error = %v, want ErrInvalidPassword", err)
		}
	})

	t.Run("rejects invalid role", func(t *testing.T) {
		_, err := userdomain.NewUser(userID, "Ana", "ana", "hash", userdomain.Role("Unknown"), "admin-1", now)
		if !errors.Is(err, userdomain.ErrInvalidRole) {
			t.Fatalf("error = %v, want ErrInvalidRole", err)
		}
	})
}

func TestUserUpdateOwnProfile(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)

	t.Run("updates name and password", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Password: "old-hash"}
		name, password := " New Name ", "new-hash"

		err := user.UpdateOwnProfile(&name, &password, "actor-1", now)

		if err != nil {
			t.Fatalf("UpdateOwnProfile() error = %v", err)
		}
		if user.Name != "New Name" || user.Password != "new-hash" || user.UpdatedBy != "actor-1" || !user.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected user: %#v", user)
		}
	})

	t.Run("updates only the provided fields", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Password: "old-hash"}
		name := "New Name"

		err := user.UpdateOwnProfile(&name, nil, "actor-1", now)

		if err != nil || user.Name != "New Name" || user.Password != "old-hash" {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Password: "old-hash"}

		err := user.UpdateOwnProfile(nil, nil, "actor-1", now)

		if !errors.Is(err, userdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects empty updated by", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Password: "old-hash"}
		name := "New Name"

		err := user.UpdateOwnProfile(&name, nil, " ", now)

		if !errors.Is(err, userdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects nil user", func(t *testing.T) {
		var user *userdomain.User
		name := "New Name"

		err := user.UpdateOwnProfile(&name, nil, "actor-1", now)

		if !errors.Is(err, userdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}

func TestUserUpdateByAdmin(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, time.March, 4, 5, 6, 7, 0, time.UTC)

	t.Run("updates all fields", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: userdomain.RoleUser}
		name, username, password, role := " New Name ", " new-user ", "new-hash", userdomain.RoleAdmin

		active := false

		err := user.UpdateByAdmin(&name, &username, &password, &role, &active, "admin-1", now)

		if err != nil {
			t.Fatalf("UpdateByAdmin() error = %v", err)
		}
		if user.Name != "New Name" || user.Username != "new-user" || user.Password != password || user.Role != userdomain.RoleAdmin || user.Active || user.UpdatedBy != "admin-1" || !user.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected user: %#v", user)
		}
	})

	t.Run("updates only the active flag", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: userdomain.RoleUser, Active: true}
		active := false

		err := user.UpdateByAdmin(nil, nil, nil, nil, &active, "admin-1", now)

		if err != nil || user.Active || user.Name != "Old" || user.UpdatedBy != "admin-1" || !user.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: userdomain.RoleAdmin, Active: true}
		name := "Updated"

		err := user.UpdateByAdmin(&name, nil, nil, nil, nil, "admin-1", now)

		if err != nil || user.Name != name || user.Username != "old" || user.Password != "old-hash" || user.Role != userdomain.RoleAdmin || !user.Active {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects invalid input without mutating the user", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: userdomain.RoleUser, Active: true}
		name := "Valid"
		invalidRole := userdomain.Role("Unknown")

		err := user.UpdateByAdmin(&name, nil, nil, &invalidRole, nil, "admin-1", now)

		if !errors.Is(err, userdomain.ErrInvalidRole) || user.Name != "Old" || user.Role != userdomain.RoleUser || !user.Active {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		user := &userdomain.User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: userdomain.RoleUser}

		err := user.UpdateByAdmin(nil, nil, nil, nil, nil, "admin-1", now)

		if !errors.Is(err, userdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})
}
