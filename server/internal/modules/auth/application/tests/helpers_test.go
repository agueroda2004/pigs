package tests

import (
	"time"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
)

func testUser(id uuid.UUID) *userdomain.User {
	return &userdomain.User{
		ID: id, Name: "Ana", Username: "ana", Password: "old-hash",
		Role: userdomain.RoleUser, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}
