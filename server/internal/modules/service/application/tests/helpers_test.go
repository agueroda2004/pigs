package tests

import (
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	operatordomain "server/internal/modules/operator/domain"
	serviceapplication "server/internal/modules/service/application"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	mountDayOne  = time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	mountDayNext = time.Date(2026, time.January, 11, 0, 0, 0, 0, time.UTC)
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

func testBoar(id uuid.UUID) *boardomain.Boar {
	return &boardomain.Boar{
		ID:     id,
		Code:   "V-001",
		Active: true,
		State:  boardomain.StateAlive,
		Origin: boardomain.OriginOwn,
	}
}

func testOperator(id uuid.UUID) *operatordomain.Operator {
	return &operatordomain.Operator{
		ID:     id,
		Name:   "Operador Uno",
		Active: true,
	}
}

func mountCommand(boarID, operatorID uuid.UUID, date time.Time) serviceapplication.CreateMountCommand {
	return serviceapplication.CreateMountCommand{
		BoarID:     boarID,
		OperatorID: operatorID,
		MountDate:  date,
	}
}

func validCommand(sowID, boarID, operatorID, createdBy uuid.UUID) serviceapplication.CreateServiceCommand {
	location := "Nave 1"
	note := "Servicio programado"
	return serviceapplication.CreateServiceCommand{
		SowID:     sowID,
		Note:      &note,
		Location:  &location,
		Mounts:    []serviceapplication.CreateMountCommand{mountCommand(boarID, operatorID, mountDayOne)},
		CreatedBy: createdBy,
	}
}
