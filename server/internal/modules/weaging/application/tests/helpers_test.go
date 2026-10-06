package tests

import (
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	sowdomain "server/internal/modules/sow/domain"
	weagingapplication "server/internal/modules/weaging/application"
)

var (
	farrowDay    = time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	lastEventDay = time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
	weagingDay   = time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)
	nowReference = time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
)

func testSow(id uuid.UUID, state sowdomain.State) *sowdomain.Sow {
	return &sowdomain.Sow{ID: id, Code: "C-001", Active: true, State: state}
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

func repositoryWith(sowID uuid.UUID, currentPiglets int) *fakeWeagingRepository {
	return &fakeWeagingRepository{
		sows: map[uuid.UUID]*sowdomain.Sow{
			sowID: testSow(sowID, sowdomain.StateLactating),
		},
		farrowings: map[uuid.UUID]*farrowingdomain.Farrowing{
			sowID: testFarrowing(uuid.New(), sowID, currentPiglets),
		},
		lastEventDate: lastEventDay,
	}
}

func validCommand(sowID, createdBy uuid.UUID) weagingapplication.CreateWeagingCommand {
	destination := "Nave 2"
	note := "Destete completo de la camada"
	return weagingapplication.CreateWeagingCommand{
		SowID:       sowID,
		WeagingDate: weagingDay,
		Quantity:    8,
		Destination: &destination,
		Note:        &note,
		CreatedBy:   createdBy,
	}
}
