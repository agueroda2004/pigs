package tests

import (
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
)

func testSow(id uuid.UUID) *sowdomain.Sow {
	return &sowdomain.Sow{
		ID:        id,
		Code:      "C-001",
		Active:    true,
		EntryDate: time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
		State:     sowdomain.StateAlive,
		Origin:    sowdomain.OriginOwn,
		Parity:    2,
		BreedID:   uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
