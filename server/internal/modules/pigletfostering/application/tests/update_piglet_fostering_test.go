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

func TestUpdatePigletFosteringExecute(t *testing.T) {
	actorID := uuid.New()
	laterDay := movementDay.AddDate(0, 0, 1)

	t.Run("updates the fostering fields without touching the balances", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		note := "camada ajustada"

		updated, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params: pigletfosteringdomain.UpdatePigletFosteringParams{
				MovementDate: &laterDay,
				Note:         &note,
			},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated != fostering || repository.updated != fostering {
			t.Fatalf("fostering was not persisted")
		}
		if !fostering.MovementDate.Equal(laterDay) || fostering.Note == nil || *fostering.Note != note {
			t.Fatalf("unexpected fostering: %#v", fostering)
		}
		if repository.farrowings[fostering.DonorSowID].CurrentPiglets != 8 ||
			repository.farrowings[fostering.ReceiverSowID].CurrentPiglets != 8 {
			t.Fatalf("balances should not change when the quantity does")
		}
	})

	t.Run("rebalances when the quantity increases", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 5

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.farrowings[fostering.DonorSowID].CurrentPiglets != 5 {
			t.Fatalf("donor piglets = %d, want 5", repository.farrowings[fostering.DonorSowID].CurrentPiglets)
		}
		if repository.farrowings[fostering.ReceiverSowID].CurrentPiglets != 11 {
			t.Fatalf("receiver piglets = %d, want 11", repository.farrowings[fostering.ReceiverSowID].CurrentPiglets)
		}
	})

	t.Run("rebalances when the quantity decreases", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 1

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.farrowings[fostering.DonorSowID].CurrentPiglets != 9 {
			t.Fatalf("donor piglets = %d, want 9", repository.farrowings[fostering.DonorSowID].CurrentPiglets)
		}
		if repository.farrowings[fostering.ReceiverSowID].CurrentPiglets != 7 {
			t.Fatalf("receiver piglets = %d, want 7", repository.farrowings[fostering.ReceiverSowID].CurrentPiglets)
		}
	})

	t.Run("rejects an increase above the donor balance", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(1, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 5

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, farrowingdomain.ErrInsufficientPiglets) {
			t.Fatalf("error = %v, want ErrInsufficientPiglets", err)
		}
		if repository.updated != nil {
			t.Fatalf("fostering should not be persisted")
		}
	})

	t.Run("rejects a decrease above the receiver balance", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 0, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 1

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, farrowingdomain.ErrInsufficientPiglets) {
			t.Fatalf("error = %v, want ErrInsufficientPiglets", err)
		}
	})

	t.Run("accepts a changed date on the nurse partial weaging day", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		donor := repository.farrowings[fostering.DonorSowID]
		donor.IsNurse = true
		repository.partialWeagingDates[donor.ID] = nurseWeagingDay
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		changed := nurseWeagingDay

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{MovementDate: &changed},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.updated == nil {
			t.Fatalf("fostering was not persisted")
		}
	})

	t.Run("rejects a changed date before the nurse partial weaging", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		donor := repository.farrowings[fostering.DonorSowID]
		donor.IsNurse = true
		repository.partialWeagingDates[donor.ID] = nurseWeagingDay
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		changed := nurseWeagingDay.AddDate(0, 0, -1)

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{MovementDate: &changed},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, pigletfosteringdomain.ErrMovementDateBeforePartialWeaging) {
			t.Fatalf("error = %v, want ErrMovementDateBeforePartialWeaging", err)
		}
		if repository.updated != nil {
			t.Fatalf("fostering should not be persisted")
		}
	})

	t.Run("does not query partial weagings for non-nurse sows", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 3

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.partialWeagingCalls != 0 {
			t.Fatalf("partial weaging calls = %d, want 0", repository.partialWeagingCalls)
		}
	})

	t.Run("propagates a missing fostering", func(t *testing.T) {
		repository, _ := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 3

		_, err := useCase.Execute(context.Background(), uuid.New(), pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrPigletFosteringNotFound) {
			t.Fatalf("error = %v, want ErrPigletFosteringNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.farrowingByIDErr = ports.ErrFarrowingNotFound
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 3

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates a missing partial weaging lookup error", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.farrowings[fostering.DonorSowID].IsNurse = true
		repository.partialWeagingErr = errors.New("boom")
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 3

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, repository.partialWeagingErr) {
			t.Fatalf("error = %v, want the lookup error", err)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			UpdatedBy: actorID,
		})

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.updateErr = errors.New("boom")
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		quantity := 3

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Quantity: &quantity},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, repository.updateErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
