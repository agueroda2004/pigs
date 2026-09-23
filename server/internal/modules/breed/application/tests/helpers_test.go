package tests

import (
	"time"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
)

func testBreed(id uuid.UUID) *breeddomain.Breed {
	return &breeddomain.Breed{
		ID: id, Name: "Old Name", Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}
