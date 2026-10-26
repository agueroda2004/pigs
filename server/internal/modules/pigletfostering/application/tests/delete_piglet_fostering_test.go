package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

func TestDeletePigletFosteringExecute(t *testing.T) {
	actorID := uuid.New()
	laterDay := movementDay.AddDate(0, 0, 1)

	t.Run("deletes the fostering and moves the piglets back", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deleted != fostering {
			t.Fatalf("fostering was not deleted")
		}
		donor := repository.farrowings[fostering.DonorSowID]
		receiver := repository.farrowings[fostering.ReceiverSowID]
		if donor.CurrentPiglets != 10 {
			t.Fatalf("donor piglets = %d, want 10", donor.CurrentPiglets)
		}
		if receiver.CurrentPiglets != 6 {
			t.Fatalf("receiver piglets = %d, want 6", receiver.CurrentPiglets)
		}
		if donor.UpdatedBy != actorID || !donor.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected donor audit fields: %#v", donor)
		}
		if receiver.UpdatedBy != actorID || !receiver.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected receiver audit fields: %#v", receiver)
		}
	})

	t.Run("blocks the deletion when the donor has a later partial weaging", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.partialWeagingDates[fostering.DonorFarrowingID] = laterDay
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if !errors.Is(err, pigletfosteringdomain.ErrHasLaterPartialWeaging) {
			t.Fatalf("error = %v, want ErrHasLaterPartialWeaging", err)
		}
		if repository.deleted != nil {
			t.Fatalf("fostering should not be deleted")
		}
		if repository.farrowings[fostering.DonorSowID].CurrentPiglets != 8 {
			t.Fatalf("balances should not change")
		}
	})

	t.Run("blocks the deletion when the receiver has a later weaging", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.weagingDates[fostering.ReceiverFarrowingID] = laterDay
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if !errors.Is(err, pigletfosteringdomain.ErrHasLaterWeaging) {
			t.Fatalf("error = %v, want ErrHasLaterWeaging", err)
		}
		if repository.deleted != nil {
			t.Fatalf("fostering should not be deleted")
		}
	})

	t.Run("allows weagings registered on or before the movement date", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.weagingDates[fostering.DonorFarrowingID] = movementDay.AddDate(0, 0, -1)
		repository.partialWeagingDates[fostering.ReceiverFarrowingID] = movementDay
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deleted == nil {
			t.Fatalf("fostering should be deleted")
		}
	})

	t.Run("rejects the deletion when the receiver cannot give the piglets back", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 1, 2)
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if !errors.Is(err, farrowingdomain.ErrInsufficientPiglets) {
			t.Fatalf("error = %v, want ErrInsufficientPiglets", err)
		}
		if repository.deleted != nil {
			t.Fatalf("fostering should not be deleted")
		}
		if repository.farrowings[fostering.DonorSowID].CurrentPiglets != 8 {
			t.Fatalf("donor balance should not change")
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{})

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidDeletedBy) {
			t.Fatalf("error = %v, want ErrInvalidDeletedBy", err)
		}
		if repository.deleted != nil {
			t.Fatalf("fostering should not be deleted")
		}
	})

	t.Run("propagates a missing fostering", func(t *testing.T) {
		repository, _ := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), uuid.New(), pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrPigletFosteringNotFound) {
			t.Fatalf("error = %v, want ErrPigletFosteringNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.farrowingByIDErr = ports.ErrFarrowingNotFound
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates a weaging lookup error", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.weagingErr = errors.New("boom")
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.weagingErr) {
			t.Fatalf("error = %v, want the weaging lookup error", err)
		}
	})

	t.Run("propagates a partial weaging lookup error", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.partialWeagingErr = errors.New("boom")
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.partialWeagingErr) {
			t.Fatalf("error = %v, want the partial weaging lookup error", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.deleteErr = errors.New("boom")
		useCase := pigletfosteringapplication.NewDeletePigletFosteringService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.DeletePigletFosteringCommand{
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.deleteErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
