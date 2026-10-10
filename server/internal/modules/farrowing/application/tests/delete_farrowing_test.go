package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	farrowingapplication "server/internal/modules/farrowing/application"
	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

func deleteRepository(farrowing *farrowingdomain.Farrowing) *fakeFarrowingRepository {
	return &fakeFarrowingRepository{
		farrowing: farrowing,
		sow:       testSow(farrowing.SowID, sowdomain.StateLactating),
		serviceByID: testService(
			farrowing.ServiceID,
			farrowing.SowID,
			servicedomain.StateFinished,
		),
	}
}

func TestDeleteFarrowingExecute(t *testing.T) {
	actor := uuid.New()
	clock := func() time.Time { return nowReference }

	t.Run("deletes the farrowing and restores the sow and service", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := deleteRepository(farrowing)
		useCase := farrowingapplication.NewDeleteFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowingapplication.DeleteFarrowingCommand{
			ID:        farrowing.ID,
			DeletedBy: actor,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deleted != farrowing || repository.deletedSow != repository.sow ||
			repository.deletedService != repository.serviceByID {
			t.Fatalf("repository delete was not called with the expected rows")
		}
		if repository.sow.State != sowdomain.StatePregnant {
			t.Fatalf("sow state = %v, want Gestando", repository.sow.State)
		}
		if repository.sow.Parity != 1 {
			t.Fatalf("sow parity = %d, want 1", repository.sow.Parity)
		}
		if repository.serviceByID.State != servicedomain.StateConfirmed {
			t.Fatalf("service state = %v, want Confirmado", repository.serviceByID.State)
		}
		if !repository.sow.UpdatedAt.Equal(nowReference) || repository.sow.UpdatedBy != actor {
			t.Fatalf("unexpected sow audit fields: %#v", repository.sow)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository := &fakeFarrowingRepository{farrowingErr: ports.ErrFarrowingNotFound}
		useCase := farrowingapplication.NewDeleteFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowingapplication.DeleteFarrowingCommand{
			ID:        uuid.New(),
			DeletedBy: actor,
		})

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := deleteRepository(farrowing)
		repository.sow = nil
		repository.sowErr = ports.ErrSowNotFound
		useCase := farrowingapplication.NewDeleteFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowingapplication.DeleteFarrowingCommand{
			ID:        farrowing.ID,
			DeletedBy: actor,
		})

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a missing service", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := deleteRepository(farrowing)
		repository.serviceByID = nil
		repository.serviceByIDErr = ports.ErrServiceNotFound
		useCase := farrowingapplication.NewDeleteFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowingapplication.DeleteFarrowingCommand{
			ID:        farrowing.ID,
			DeletedBy: actor,
		})

		if !errors.Is(err, ports.ErrServiceNotFound) {
			t.Fatalf("error = %v, want ErrServiceNotFound", err)
		}
	})

	t.Run("rejects an empty sow parity", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := deleteRepository(farrowing)
		repository.sow.Parity = 0
		useCase := farrowingapplication.NewDeleteFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowingapplication.DeleteFarrowingCommand{
			ID:        farrowing.ID,
			DeletedBy: actor,
		})

		if !errors.Is(err, sowdomain.ErrInvalidParity) {
			t.Fatalf("error = %v, want ErrInvalidParity", err)
		}
	})

	t.Run("propagates the linked-records error from the repository", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := deleteRepository(farrowing)
		repository.deleteErr = ports.ErrFarrowingHasRelatedRecords
		useCase := farrowingapplication.NewDeleteFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowingapplication.DeleteFarrowingCommand{
			ID:        farrowing.ID,
			DeletedBy: actor,
		})

		if !errors.Is(err, ports.ErrFarrowingHasRelatedRecords) {
			t.Fatalf("error = %v, want ErrFarrowingHasRelatedRecords", err)
		}
	})
}
