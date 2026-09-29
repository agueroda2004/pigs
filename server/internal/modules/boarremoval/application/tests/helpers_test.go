package tests

import (
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	boarremovalapplication "server/internal/modules/boarremoval/application"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
)

var (
	entryDay     = time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	mountDay     = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	removalDay   = time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
	nowReference = time.Date(2026, time.February, 1, 12, 0, 0, 0, time.UTC)
)

func testBoar(id uuid.UUID, state boardomain.State) *boardomain.Boar {
	return &boardomain.Boar{
		ID:        id,
		Code:      "B-001",
		Active:    true,
		EntryDate: entryDay,
		State:     state,
		Origin:    boardomain.OriginOwn,
		BreedID:   uuid.New(),
	}
}

func validCommand(boarID, createdBy uuid.UUID) boarremovalapplication.CreateBoarRemovalCommand {
	note := "Baja por enfermedad"
	return boarremovalapplication.CreateBoarRemovalCommand{
		BoarID:      boarID,
		RemovalDate: removalDay,
		Type:        boarremovaldomain.TypeDeath,
		Reason:      boarremovaldomain.ReasonDisease,
		Note:        &note,
		CreatedBy:   createdBy,
	}
}
