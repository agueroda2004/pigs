package tests

import (
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	sowdomain "server/internal/modules/sow/domain"
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

func repositoryWith(sowID uuid.UUID, currentPiglets int) *fakePartialWeagingRepository {
	return &fakePartialWeagingRepository{
		sows: map[uuid.UUID]*sowdomain.Sow{
			sowID: testSow(sowID, sowdomain.StateLactating),
		},
		farrowings: map[uuid.UUID]*farrowingdomain.Farrowing{
			sowID: testFarrowing(uuid.New(), sowID, currentPiglets),
		},
		lastEventDate: lastEventDay,
	}
}

func validCommand(sowID, createdBy uuid.UUID) partialweagingapplication.CreatePartialWeagingCommand {
	note := "Destete parcial de la camada"
	return partialweagingapplication.CreatePartialWeagingCommand{
		SowID:       sowID,
		WeagingDate: weagingDay,
		Quantity:    2,
		Type:        partialweagingdomain.TypeNormal,
		Note:        &note,
		CreatedBy:   createdBy,
	}
}
