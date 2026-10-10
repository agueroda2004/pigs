package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	serviceapplication "server/internal/modules/service/application"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	sowdomain "server/internal/modules/sow/domain"
)

func updatableService(sowID uuid.UUID) *servicedomain.Service {
	serviceID := uuid.New()
	return &servicedomain.Service{
		ID:        serviceID,
		SowID:     sowID,
		State:     servicedomain.StateConfirmed,
		LastState: string(sowdomain.StateAlive),
		Mounts: []*servicedomain.Mount{
			{
				ID:          uuid.New(),
				ServiceID:   serviceID,
				BoarID:      uuid.New(),
				OperatorID:  uuid.New(),
				MountNumber: 1,
				MountDate:   mountDayOne,
				Type:        servicedomain.MountTypeArtificial,
			},
			{
				ID:          uuid.New(),
				ServiceID:   serviceID,
				BoarID:      uuid.New(),
				OperatorID:  uuid.New(),
				MountNumber: 2,
				MountDate:   mountDayNext,
				Type:        servicedomain.MountTypeArtificial,
			},
		},
	}
}

func TestUpdateService(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
	clock := func() time.Time { return now }

	newRepository := func() (*fakeServiceRepository, *servicedomain.Service) {
		sowID := uuid.New()
		service := updatableService(sowID)
		return &fakeServiceRepository{
			getService: service,
			sow:        testSow(sowID, sowdomain.StatePregnant),
			boar:       testBoar(uuid.New()),
			operator:   testOperator(uuid.New()),
		}, service
	}

	t.Run("applies the mount operations and persists", func(t *testing.T) {
		repository, service := newRepository()
		useCase := serviceapplication.NewUpdateServiceService(repository, clock)
		boarID := uuid.New()
		operatorID := uuid.New()

		err := useCase.Execute(context.Background(), service.ID, serviceapplication.UpdateServiceCommand{
			CreateMounts:   []serviceapplication.CreateMountCommand{mountCommand(boarID, operatorID, mountDayNext)},
			UpdateMounts:   []serviceapplication.UpdateMountCommand{{ID: service.Mounts[0].ID, BoarID: boarID, OperatorID: operatorID, MountDate: mountDayOne}},
			DeleteMountIDs: []uuid.UUID{service.Mounts[1].ID},
			UpdatedBy:      uuid.New(),
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.updated != service {
			t.Fatalf("repository did not persist the service")
		}
		if len(service.Mounts) != 2 {
			t.Fatalf("mounts = %d, want 2", len(service.Mounts))
		}
	})

	t.Run("blocks a non-confirmed service", func(t *testing.T) {
		repository, service := newRepository()
		service.State = servicedomain.StateFinished
		useCase := serviceapplication.NewUpdateServiceService(repository, clock)

		err := useCase.Execute(context.Background(), service.ID, serviceapplication.UpdateServiceCommand{
			UpdatedBy: uuid.New(),
		})

		if !errors.Is(err, servicedomain.ErrServiceNotEditable) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("propagates the service lookup error", func(t *testing.T) {
		repository := &fakeServiceRepository{getErr: ports.ErrServiceNotFound}
		useCase := serviceapplication.NewUpdateServiceService(repository, clock)

		err := useCase.Execute(context.Background(), uuid.New(), serviceapplication.UpdateServiceCommand{UpdatedBy: uuid.New()})

		if !errors.Is(err, ports.ErrServiceNotFound) {
			t.Fatalf("error = %v, want ErrServiceNotFound", err)
		}
	})

	t.Run("rejects an ineligible boar", func(t *testing.T) {
		repository, service := newRepository()
		repository.boar.Active = false
		useCase := serviceapplication.NewUpdateServiceService(repository, clock)

		err := useCase.Execute(context.Background(), service.ID, serviceapplication.UpdateServiceCommand{
			CreateMounts: []serviceapplication.CreateMountCommand{mountCommand(uuid.New(), uuid.New(), mountDayNext)},
			UpdatedBy:    uuid.New(),
		})

		if !errors.Is(err, serviceapplication.ErrBoarNotEligible) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("rejects an unavailable operator", func(t *testing.T) {
		repository, service := newRepository()
		repository.operator.Active = false
		useCase := serviceapplication.NewUpdateServiceService(repository, clock)

		err := useCase.Execute(context.Background(), service.ID, serviceapplication.UpdateServiceCommand{
			CreateMounts: []serviceapplication.CreateMountCommand{mountCommand(uuid.New(), uuid.New(), mountDayNext)},
			UpdatedBy:    uuid.New(),
		})

		if !errors.Is(err, serviceapplication.ErrOperatorNotAvailable) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("rejects a mount before the previous service", func(t *testing.T) {
		repository, service := newRepository()
		repository.previousService = &servicedomain.Service{
			ID:     uuid.New(),
			SowID:  service.SowID,
			State:  servicedomain.StateFinished,
			Mounts: []*servicedomain.Mount{{ID: uuid.New(), MountDate: mountDayNext}},
		}
		useCase := serviceapplication.NewUpdateServiceService(repository, clock)

		err := useCase.Execute(context.Background(), service.ID, serviceapplication.UpdateServiceCommand{
			CreateMounts: []serviceapplication.CreateMountCommand{mountCommand(uuid.New(), uuid.New(), mountDayOne)},
			UpdatedBy:    uuid.New(),
		})

		if !errors.Is(err, servicedomain.ErrMountBeforePreviousService) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("propagates the update error", func(t *testing.T) {
		repository, service := newRepository()
		unexpected := errors.New("unexpected")
		repository.updateErr = unexpected
		useCase := serviceapplication.NewUpdateServiceService(repository, clock)

		err := useCase.Execute(context.Background(), service.ID, serviceapplication.UpdateServiceCommand{
			UpdatedBy: uuid.New(),
		})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
