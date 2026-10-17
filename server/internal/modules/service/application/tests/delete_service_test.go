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

func confirmedService(id, sowID uuid.UUID) *servicedomain.Service {
	return &servicedomain.Service{
		ID:        id,
		SowID:     sowID,
		State:     servicedomain.StateConfirmed,
		LastState: string(sowdomain.StateAlive),
	}
}

func TestDeleteService(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
	clock := func() time.Time { return now }

	t.Run("deletes a confirmed service and restores the sow state", func(t *testing.T) {
		serviceID := uuid.New()
		sowID := uuid.New()
		deletedBy := uuid.New()
		service := confirmedService(serviceID, sowID)
		repository := &fakeServiceRepository{
			getService: service,
			sow:        testSow(sowID, sowdomain.StatePregnant),
		}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        serviceID,
			DeletedBy: deletedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.sow.State != sowdomain.StateAlive {
			t.Fatalf("sow state = %v, want %v", repository.sow.State, sowdomain.StateAlive)
		}
		if repository.sow.UpdatedBy != deletedBy || !repository.sow.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected sow audit: %#v", repository.sow)
		}
		if repository.deleted != service || repository.deletedSow != repository.sow {
			t.Fatalf("repository did not delete the service and sow together")
		}
	})

	t.Run("reopens the previous failed service", func(t *testing.T) {
		serviceID := uuid.New()
		sowID := uuid.New()
		deletedBy := uuid.New()
		service := confirmedService(serviceID, sowID)
		previous := &servicedomain.Service{ID: uuid.New(), SowID: sowID, State: servicedomain.StateFailed}
		repository := &fakeServiceRepository{
			getService:      service,
			sow:             testSow(sowID, sowdomain.StatePregnant),
			previousService: previous,
		}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        serviceID,
			DeletedBy: deletedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if previous.State != servicedomain.StateConfirmed {
			t.Fatalf("previous state = %v, want %v", previous.State, servicedomain.StateConfirmed)
		}
		if previous.UpdatedBy != deletedBy || !previous.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected previous audit: %#v", previous)
		}
		if repository.deletedPrevious != previous {
			t.Fatalf("repository did not persist the previous service")
		}
	})

	t.Run("leaves a previous service that is not failed unchanged", func(t *testing.T) {
		sowID := uuid.New()
		service := confirmedService(uuid.New(), sowID)
		previous := &servicedomain.Service{ID: uuid.New(), SowID: sowID, State: servicedomain.StateFinished}
		repository := &fakeServiceRepository{
			getService:      service,
			sow:             testSow(sowID, sowdomain.StatePregnant),
			previousService: previous,
		}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        service.ID,
			DeletedBy: uuid.New(),
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if previous.State != servicedomain.StateFinished || repository.deletedPrevious != nil {
			t.Fatalf("unexpected result: state=%v previous=%#v", previous.State, repository.deletedPrevious)
		}
	})

	t.Run("propagates the previous service lookup error", func(t *testing.T) {
		service := confirmedService(uuid.New(), uuid.New())
		unexpected := errors.New("unexpected")
		repository := &fakeServiceRepository{
			getService:         service,
			sow:                testSow(service.SowID, sowdomain.StatePregnant),
			previousServiceErr: unexpected,
		}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        service.ID,
			DeletedBy: uuid.New(),
		})

		if !errors.Is(err, unexpected) || repository.deleted != nil {
			t.Fatalf("unexpected result: err=%v deleted=%#v", err, repository.deleted)
		}
	})

	t.Run("blocks a service that is not confirmed", func(t *testing.T) {
		service := confirmedService(uuid.New(), uuid.New())
		service.State = servicedomain.StateFinished
		repository := &fakeServiceRepository{getService: service}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        service.ID,
			DeletedBy: uuid.New(),
		})

		if !errors.Is(err, servicedomain.ErrServiceNotDeletable) || repository.deleted != nil {
			t.Fatalf("unexpected result: err=%v deleted=%#v", err, repository.deleted)
		}
	})

	t.Run("propagates the service lookup error", func(t *testing.T) {
		repository := &fakeServiceRepository{getErr: ports.ErrServiceNotFound}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        uuid.New(),
			DeletedBy: uuid.New(),
		})

		if !errors.Is(err, ports.ErrServiceNotFound) {
			t.Fatalf("error = %v, want ErrServiceNotFound", err)
		}
	})

	t.Run("propagates an invalid stored last state", func(t *testing.T) {
		service := confirmedService(uuid.New(), uuid.New())
		service.LastState = "Desconocido"
		repository := &fakeServiceRepository{getService: service}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        service.ID,
			DeletedBy: uuid.New(),
		})

		if !errors.Is(err, sowdomain.ErrInvalidState) || repository.deleted != nil {
			t.Fatalf("unexpected result: err=%v deleted=%#v", err, repository.deleted)
		}
	})

	t.Run("propagates the sow lookup error", func(t *testing.T) {
		service := confirmedService(uuid.New(), uuid.New())
		repository := &fakeServiceRepository{getService: service, sowErr: ports.ErrSowNotFound}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        service.ID,
			DeletedBy: uuid.New(),
		})

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates the delete error", func(t *testing.T) {
		service := confirmedService(uuid.New(), uuid.New())
		unexpected := errors.New("unexpected")
		repository := &fakeServiceRepository{
			getService: service,
			sow:        testSow(service.SowID, sowdomain.StatePregnant),
			deleteErr:  unexpected,
		}
		useCase := serviceapplication.NewDeleteServiceService(repository, clock)

		err := useCase.Execute(context.Background(), serviceapplication.DeleteServiceCommand{
			ID:        service.ID,
			DeletedBy: uuid.New(),
		})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
