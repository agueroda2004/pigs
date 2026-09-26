package tests

import (
	"time"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	sowremovalapplication "server/internal/modules/sowremoval/application"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
)

var (
	mountDayOne  = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	abortionDay  = time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	removalDay   = time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
	nowReference = time.Date(2026, time.February, 1, 12, 0, 0, 0, time.UTC)
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

func testService(id, sowID uuid.UUID, state servicedomain.State, mounts ...*servicedomain.Mount) *servicedomain.Service {
	return &servicedomain.Service{
		ID:     id,
		SowID:  sowID,
		State:  state,
		Mounts: mounts,
	}
}

func testMount(serviceID uuid.UUID, number int, date time.Time) *servicedomain.Mount {
	return &servicedomain.Mount{
		ID:          uuid.New(),
		ServiceID:   serviceID,
		BoarID:      uuid.New(),
		OperatorID:  uuid.New(),
		MountNumber: number,
		MountDate:   date,
		Type:        servicedomain.MountTypeNatural,
	}
}

func testAbortion(id, sowID uuid.UUID, date time.Time) *abortiondomain.Abortion {
	return &abortiondomain.Abortion{
		ID:           id,
		SowID:        sowID,
		ServiceID:    uuid.New(),
		AbortionDate: date,
		Cause:        abortiondomain.CauseInfectious,
	}
}

func validCommand(sowID, createdBy uuid.UUID) sowremovalapplication.CreateSowRemovalCommand {
	note := "Baja por enfermedad"
	return sowremovalapplication.CreateSowRemovalCommand{
		SowID:       sowID,
		RemovalDate: removalDay,
		Type:        sowremovaldomain.TypeDeath,
		Reason:      sowremovaldomain.ReasonDisease,
		Note:        &note,
		CreatedBy:   createdBy,
	}
}
