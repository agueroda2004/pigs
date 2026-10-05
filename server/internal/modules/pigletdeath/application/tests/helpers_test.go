package tests

import (
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	operatordomain "server/internal/modules/operator/domain"
	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	farrowDay    = time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	deathDay     = time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)
	nowReference = time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
)

func testSow(id uuid.UUID, state sowdomain.State) *sowdomain.Sow {
	return &sowdomain.Sow{
		ID:        id,
		Code:      "C-001",
		Active:    true,
		EntryDate: time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
		State:     state,
		Origin:    sowdomain.OriginOwn,
		Parity:    2,
		BreedID:   uuid.New(),
	}
}

func testFarrowing(id, sowID uuid.UUID, currentPiglets int) *farrowingdomain.Farrowing {
	return &farrowingdomain.Farrowing{
		ID:             id,
		SowID:          sowID,
		FarrowDate:     farrowDay,
		LiveBorn:       10,
		CurrentPiglets: currentPiglets,
	}
}

func testOperator(id uuid.UUID) *operatordomain.Operator {
	return &operatordomain.Operator{ID: id, Name: "Operador", Active: true}
}

func validCommand(sowID, operatorID, createdBy uuid.UUID) pigletdeathapplication.CreatePigletDeathCommand {
	note := "Muerte durante la noche"
	weight := 2.5
	return pigletdeathapplication.CreatePigletDeathCommand{
		SowID:      sowID,
		OperatorID: operatorID,
		DeathDate:  deathDay,
		Quantity:   2,
		Weight:     &weight,
		Cause:      pigletdeathdomain.CauseCrushed,
		Turn:       pigletdeathdomain.TurnMorning,
		Note:       &note,
		CreatedBy:  createdBy,
	}
}
