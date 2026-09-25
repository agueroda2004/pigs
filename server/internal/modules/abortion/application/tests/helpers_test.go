package tests

import (
	"time"

	"github.com/google/uuid"

	abortionapplication "server/internal/modules/abortion/application"
	abortiondomain "server/internal/modules/abortion/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	mountDayOne  = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	mountDayTwo  = time.Date(2026, time.January, 11, 0, 0, 0, 0, time.UTC)
	abortionDay  = time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
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

func validCommand(sowID, createdBy uuid.UUID) abortionapplication.CreateAbortionCommand {
	note := "Aborto espontáneo"
	return abortionapplication.CreateAbortionCommand{
		SowID:        sowID,
		AbortionDate: abortionDay,
		Cause:        abortiondomain.CauseInfectious,
		Note:         &note,
		CreatedBy:    createdBy,
	}
}
