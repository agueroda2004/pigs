package tests

import (
	"time"

	"github.com/google/uuid"

	farrowingapplication "server/internal/modules/farrowing/application"
	medicationdomain "server/internal/modules/medication/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	mountDayOne  = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	mountDayTwo  = time.Date(2026, time.January, 11, 0, 0, 0, 0, time.UTC)
	farrowDay    = time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	nowReference = time.Date(2026, time.April, 20, 12, 0, 0, 0, time.UTC)
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

func testOperator(id uuid.UUID) *operatordomain.Operator {
	return &operatordomain.Operator{ID: id, Name: "Operador", Active: true}
}

func testMedication(id uuid.UUID) *medicationdomain.Medication {
	return &medicationdomain.Medication{ID: id, Name: "Medicamento", Active: true}
}

func validCommand(sowID, createdBy uuid.UUID) farrowingapplication.CreateFarrowingCommand {
	location := "Corral 3"
	note := "Parto sin complicaciones"
	return farrowingapplication.CreateFarrowingCommand{
		SowID:      sowID,
		FarrowDate: farrowDay,
		Location:   &location,
		LiveBorn:   10,
		Stillborn:  1,
		Mummified:  0,
		Note:       &note,
		CreatedBy:  createdBy,
	}
}
