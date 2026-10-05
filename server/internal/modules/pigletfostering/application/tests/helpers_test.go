package tests

import (
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	donorFarrowDay    = time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	receiverFarrowDay = time.Date(2026, time.April, 22, 0, 0, 0, 0, time.UTC)
	movementDay       = time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)
	nowReference      = time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
)

func testSow(id uuid.UUID, state sowdomain.State) *sowdomain.Sow {
	return &sowdomain.Sow{ID: id, Code: "C-001", Active: true, State: state}
}

func testFarrowing(id, sowID uuid.UUID, farrowDate time.Time, currentPiglets int) *farrowingdomain.Farrowing {
	return &farrowingdomain.Farrowing{
		ID:             id,
		SowID:          sowID,
		FarrowDate:     farrowDate,
		LiveBorn:       10,
		CurrentPiglets: currentPiglets,
	}
}

func repositoryWith(donorSowID, receiverSowID uuid.UUID, donorPiglets, receiverPiglets int) *fakePigletFosteringRepository {
	return &fakePigletFosteringRepository{
		sows: map[uuid.UUID]*sowdomain.Sow{
			donorSowID:    testSow(donorSowID, sowdomain.StateLactating),
			receiverSowID: testSow(receiverSowID, sowdomain.StateLactating),
		},
		farrowings: map[uuid.UUID]*farrowingdomain.Farrowing{
			donorSowID:    testFarrowing(uuid.New(), donorSowID, donorFarrowDay, donorPiglets),
			receiverSowID: testFarrowing(uuid.New(), receiverSowID, receiverFarrowDay, receiverPiglets),
		},
	}
}

func validCommand(donorSowID, receiverSowID, createdBy uuid.UUID) pigletfosteringapplication.CreatePigletFosteringCommand {
	note := "Traslado por camada numerosa"
	return pigletfosteringapplication.CreatePigletFosteringCommand{
		DonorSowID:    donorSowID,
		ReceiverSowID: receiverSowID,
		MovementDate:  movementDay,
		Quantity:      2,
		Note:          &note,
		CreatedBy:     createdBy,
	}
}
