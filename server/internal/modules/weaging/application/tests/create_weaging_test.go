package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	sowdomain "server/internal/modules/sow/domain"
	weagingapplication "server/internal/modules/weaging/application"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

func TestCreateWeagingExecute(t *testing.T) {
	sowID := uuid.New()
	actorID := uuid.New()

	t.Run("registers the weaging, zeroes the balance and weans the sow", func(t *testing.T) {
		repository := repositoryWith(sowID, 8)
		useCase := weagingapplication.NewCreateWeagingService(repository, func() time.Time { return nowReference })

		weaging, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if weaging.SowID != sowID || weaging.Quantity != 8 {
			t.Fatalf("unexpected weaging: %#v", weaging)
		}
		if weaging.FarrowingID != repository.farrowings[sowID].ID {
			t.Fatalf("unexpected farrowing: %#v", weaging.FarrowingID)
		}
		if weaging.CreatedBy != actorID || !weaging.CreatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", weaging)
		}
		if repository.created != weaging {
			t.Fatalf("weaging was not persisted")
		}
		if repository.createdFarrowing.CurrentPiglets != 0 {
			t.Fatalf("current piglets = %d, want 0", repository.createdFarrowing.CurrentPiglets)
		}
		if repository.createdSow.State != sowdomain.StateWeaned {
			t.Fatalf("sow state = %q, want Destetada", repository.createdSow.State)
		}
	})

	t.Run("rejects a sow that is not lactating", func(t *testing.T) {
		repository := repositoryWith(sowID, 8)
		repository.sows[sowID] = testSow(sowID, sowdomain.StateWeaned)
		useCase := weagingapplication.NewCreateWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, weagingapplication.ErrSowNotLactating) {
			t.Fatalf("error = %v, want ErrSowNotLactating", err)
		}
	})

	t.Run("rejects a quantity that does not match the balance", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		useCase := weagingapplication.NewCreateWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, farrowingdomain.ErrWeagingQuantityMismatch) {
			t.Fatalf("error = %v, want ErrWeagingQuantityMismatch", err)
		}
		if repository.created != nil {
			t.Fatalf("weaging should not be persisted")
		}
	})

	t.Run("rejects a weaging date before the last event", func(t *testing.T) {
		repository := repositoryWith(sowID, 8)
		useCase := weagingapplication.NewCreateWeagingService(repository, func() time.Time { return nowReference })
		command := validCommand(sowID, actorID)
		command.WeagingDate = lastEventDay

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, weagingdomain.ErrWeagingDateBeforeEvents) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeEvents", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := repositoryWith(sowID, 8)
		repository.sowErr = ports.ErrSowNotFound
		useCase := weagingapplication.NewCreateWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository := repositoryWith(sowID, 8)
		repository.farrowErr = ports.ErrFarrowingNotFound
		useCase := weagingapplication.NewCreateWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates the latest event date error", func(t *testing.T) {
		repository := repositoryWith(sowID, 8)
		repository.lastEventErr = errors.New("boom")
		useCase := weagingapplication.NewCreateWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.lastEventErr) {
			t.Fatalf("error = %v, want the event date error", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith(sowID, 8)
		repository.createErr = errors.New("boom")
		useCase := weagingapplication.NewCreateWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
