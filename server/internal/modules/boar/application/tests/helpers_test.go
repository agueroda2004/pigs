package tests

import (
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
)

func testBoar(id uuid.UUID) *boardomain.Boar {
	return &boardomain.Boar{
		ID:        id,
		Code:      "B-001",
		Active:    true,
		EntryDate: time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
		State:     boardomain.StateAlive,
		Origin:    boardomain.OriginOwn,
		BreedID:   uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
