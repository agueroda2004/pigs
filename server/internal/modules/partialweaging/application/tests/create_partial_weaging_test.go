package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
	sowdomain "server/internal/modules/sow/domain"
)

func TestCreatePartialWeagingExecute(t *testing.T) {
	sowID := uuid.New()
	actorID := uuid.New()

	t.Run("registers a normal partial weaging and weans the sow", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		weaging, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if weaging.SowID != sowID || weaging.Quantity != 2 || weaging.Type != partialweagingdomain.TypeNormal {
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
		if repository.createdFarrowing.CurrentPiglets != 8 {
			t.Fatalf("current piglets = %d, want 8", repository.createdFarrowing.CurrentPiglets)
		}
		if repository.createdSow.State != sowdomain.StateWeaned {
			t.Fatalf("sow state = %q, want Destetada", repository.createdSow.State)
		}
	})

	t.Run("flags the farrowing as nurse and keeps the sow lactating for Nodriza", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })
		command := validCommand(sowID, actorID)
		command.Type = partialweagingdomain.TypeNodriza

		_, err := useCase.Execute(context.Background(), command)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !repository.createdFarrowing.IsNurse || repository.createdFarrowing.NurseStartDate == nil {
			t.Fatalf("farrowing was not flagged as nurse: %#v", repository.createdFarrowing)
		}
		if !repository.createdFarrowing.NurseStartDate.Equal(weagingDay) {
			t.Fatalf("nurse start date = %v, want %v", repository.createdFarrowing.NurseStartDate, weagingDay)
		}
		if repository.createdSow.State != sowdomain.StateLactating {
			t.Fatalf("sow state = %q, want Lactando", repository.createdSow.State)
		}
	})

	t.Run("rejects a sow that is not lactating", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		repository.sows[sowID] = testSow(sowID, sowdomain.StatePregnant)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, partialweagingapplication.ErrSowNotLactating) {
			t.Fatalf("error = %v, want ErrSowNotLactating", err)
		}
	})

	t.Run("rejects a quantity greater than the balance", func(t *testing.T) {
		repository := repositoryWith(sowID, 1)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, farrowingdomain.ErrInsufficientPiglets) {
			t.Fatalf("error = %v, want ErrInsufficientPiglets", err)
		}
		if repository.created != nil {
			t.Fatalf("weaging should not be persisted")
		}
	})

	t.Run("rejects a weaging date before the last event", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })
		command := validCommand(sowID, actorID)
		command.WeagingDate = lastEventDay

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeEvents) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeEvents", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		repository.sowErr = ports.ErrSowNotFound
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		repository.farrowErr = ports.ErrFarrowingNotFound
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates the latest event date error", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		repository.lastEventErr = errors.New("boom")
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.lastEventErr) {
			t.Fatalf("error = %v, want the event date error", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith(sowID, 10)
		repository.createErr = errors.New("boom")
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
