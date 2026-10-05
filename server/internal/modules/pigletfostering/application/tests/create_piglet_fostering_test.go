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
	sowdomain "server/internal/modules/sow/domain"
)

func TestCreatePigletFosteringExecute(t *testing.T) {
	donorSowID := uuid.New()
	receiverSowID := uuid.New()
	actorID := uuid.New()

	t.Run("registers the fostering and moves the piglets", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 10, 6)
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })

		fostering, err := useCase.Execute(context.Background(), validCommand(donorSowID, receiverSowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if fostering.DonorSowID != donorSowID || fostering.ReceiverSowID != receiverSowID || fostering.Quantity != 2 {
			t.Fatalf("unexpected fostering: %#v", fostering)
		}
		if fostering.DonorFarrowingID != repository.farrowings[donorSowID].ID {
			t.Fatalf("unexpected donor farrowing: %#v", fostering.DonorFarrowingID)
		}
		if fostering.ReceiverFarrowingID != repository.farrowings[receiverSowID].ID {
			t.Fatalf("unexpected receiver farrowing: %#v", fostering.ReceiverFarrowingID)
		}
		if fostering.CreatedBy != actorID || !fostering.CreatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", fostering)
		}
		if repository.created != fostering {
			t.Fatalf("fostering was not persisted")
		}
		if repository.createdDonor.CurrentPiglets != 8 {
			t.Fatalf("donor piglets = %d, want 8", repository.createdDonor.CurrentPiglets)
		}
		if repository.createdReceiver.CurrentPiglets != 8 {
			t.Fatalf("receiver piglets = %d, want 8", repository.createdReceiver.CurrentPiglets)
		}
	})

	t.Run("rejects a donor that is not lactating", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 10, 6)
		repository.sows[donorSowID] = testSow(donorSowID, sowdomain.StatePregnant)
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(donorSowID, receiverSowID, actorID))

		if !errors.Is(err, pigletfosteringapplication.ErrSowNotLactating) {
			t.Fatalf("error = %v, want ErrSowNotLactating", err)
		}
	})

	t.Run("rejects a receiver that is not lactating", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 10, 6)
		repository.sows[receiverSowID] = testSow(receiverSowID, sowdomain.StatePregnant)
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(donorSowID, receiverSowID, actorID))

		if !errors.Is(err, pigletfosteringapplication.ErrSowNotLactating) {
			t.Fatalf("error = %v, want ErrSowNotLactating", err)
		}
	})

	t.Run("rejects a quantity greater than the donor balance", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 1, 6)
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(donorSowID, receiverSowID, actorID))

		if !errors.Is(err, farrowingdomain.ErrInsufficientPiglets) {
			t.Fatalf("error = %v, want ErrInsufficientPiglets", err)
		}
		if repository.created != nil {
			t.Fatalf("fostering should not be persisted")
		}
	})

	t.Run("rejects the same sow on both sides", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 10, 6)
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })
		command := validCommand(donorSowID, donorSowID, actorID)

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, pigletfosteringdomain.ErrSameFarrowing) {
			t.Fatalf("error = %v, want ErrSameFarrowing", err)
		}
	})

	t.Run("rejects a movement date before the farrowings", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 10, 6)
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })
		command := validCommand(donorSowID, receiverSowID, actorID)
		command.MovementDate = donorFarrowDay

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, pigletfosteringdomain.ErrMovementDateBeforeFarrowing) {
			t.Fatalf("error = %v, want ErrMovementDateBeforeFarrowing", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 10, 6)
		repository.sowErr = ports.ErrSowNotFound
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(donorSowID, receiverSowID, actorID))

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 10, 6)
		repository.farrowErr = ports.ErrFarrowingNotFound
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(donorSowID, receiverSowID, actorID))

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith(donorSowID, receiverSowID, 10, 6)
		repository.createErr = errors.New("boom")
		useCase := pigletfosteringapplication.NewCreatePigletFosteringService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(donorSowID, receiverSowID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
