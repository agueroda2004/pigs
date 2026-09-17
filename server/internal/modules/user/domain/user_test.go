package user

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	userID := uuid.New()

	t.Run("creates a valid user", func(t *testing.T) {
		user, err := NewUser(userID, "  Ana  ", "  ana  ", "hash", RoleAdmin, "admin-1", now)

		if err != nil {
			t.Fatalf("NewUser() error = %v", err)
		}
		if user.ID != userID || user.Name != "Ana" || user.Username != "ana" || user.Password != "hash" || user.Role != RoleAdmin || user.CreatedBy != "admin-1" {
			t.Fatalf("unexpected user: %#v", user)
		}
		if !user.CreatedAt.Equal(now) || !user.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", user)
		}
	})

	t.Run("defaults role to User when empty", func(t *testing.T) {
		user, err := NewUser(userID, "Ana", "ana", "hash", "", "admin-1", now)

		if err != nil || user.Role != RoleUser {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		_, err := NewUser(uuid.Nil, "Ana", "ana", "hash", "", "admin-1", now)
		if !errors.Is(err, ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid name", func(t *testing.T) {
		_, err := NewUser(userID, " ", "ana", "hash", "", "admin-1", now)
		if !errors.Is(err, ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}

		_, err = NewUser(userID, strings.Repeat("a", 51), "ana", "hash", "", "admin-1", now)
		if !errors.Is(err, ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}
	})

	t.Run("rejects invalid username", func(t *testing.T) {
		_, err := NewUser(userID, "Ana", " ", "hash", "", "admin-1", now)
		if !errors.Is(err, ErrInvalidUsername) {
			t.Fatalf("error = %v, want ErrInvalidUsername", err)
		}

		_, err = NewUser(userID, "Ana", strings.Repeat("a", 51), "hash", "", "admin-1", now)
		if !errors.Is(err, ErrInvalidUsername) {
			t.Fatalf("error = %v, want ErrInvalidUsername", err)
		}
	})

	t.Run("rejects empty password hash", func(t *testing.T) {
		_, err := NewUser(userID, "Ana", "ana", "   ", "", "admin-1", now)
		if !errors.Is(err, ErrInvalidPassword) {
			t.Fatalf("error = %v, want ErrInvalidPassword", err)
		}
	})

	t.Run("rejects invalid role", func(t *testing.T) {
		_, err := NewUser(userID, "Ana", "ana", "hash", Role("Unknown"), "admin-1", now)
		if !errors.Is(err, ErrInvalidRole) {
			t.Fatalf("error = %v, want ErrInvalidRole", err)
		}
	})
}

func TestUserUpdateOwnProfile(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)

	t.Run("updates name and password", func(t *testing.T) {
		user := &User{ID: userID, Name: "Old", Password: "old-hash"}
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
		user := &User{ID: userID, Name: "Old", Password: "old-hash"}
		name := "New Name"

		err := user.UpdateOwnProfile(&name, nil, "actor-1", now)

		if err != nil || user.Name != "New Name" || user.Password != "old-hash" {
			t.Fatalf("unexpected result: err=%v user=%#v", err, user)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		user := &User{ID: userID, Name: "Old", Password: "old-hash"}

		err := user.UpdateOwnProfile(nil, nil, "actor-1", now)

		if !errors.Is(err, ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects empty updated by", func(t *testing.T) {
		user := &User{ID: userID, Name: "Old", Password: "old-hash"}
		name := "New Name"

		err := user.UpdateOwnProfile(&name, nil, " ", now)

		if !errors.Is(err, ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects nil user", func(t *testing.T) {
		var user *User
		name := "New Name"

		err := user.UpdateOwnProfile(&name, nil, "actor-1", now)

		if !errors.Is(err, ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}

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

	t.Run("rejects empty update", func(t *testing.T) {
		user := &User{ID: userID, Name: "Old", Username: "old", Password: "old-hash", Role: RoleUser}

		err := user.UpdateByAdmin(nil, nil, nil, nil, "admin-1", now)

		if !errors.Is(err, ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid name is trimmed", "  Ana  ", "Ana", false},
		{"empty name", "   ", "", true},
		{"over max length", strings.Repeat("a", 51), "", true},
		{"exactly max length", strings.Repeat("a", 50), strings.Repeat("a", 50), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateName(tt.input)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidName) {
					t.Fatalf("error = %v, want ErrInvalidName", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("validateName() = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid username is trimmed", "  ana  ", "ana", false},
		{"empty username", "   ", "", true},
		{"over max length", strings.Repeat("a", 51), "", true},
		{"exactly max length", strings.Repeat("a", 50), strings.Repeat("a", 50), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateUsername(tt.input)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidUsername) {
					t.Fatalf("error = %v, want ErrInvalidUsername", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("validateUsername() = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	if !isValidRole(RoleUser) || !isValidRole(RoleAdmin) {
		t.Fatal("expected RoleUser and RoleAdmin to be valid")
	}
	if isValidRole(Role("")) || isValidRole(Role("Unknown")) {
		t.Fatal("expected empty and unknown roles to be invalid")
	}
}
