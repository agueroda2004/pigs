package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
	sowdomain "server/internal/modules/sow/domain"
)

func TestCreatePartialWeagingExecute(t *testing.T) {
	sowID := uuid.New()
	actorID := uuid.New()

	t.Run("registers a normal partial weaging and weans the sow", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		weaging, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if weaging.SowID != sowID || weaging.Quantity != 2 || weaging.Type != partialweagingdomain.TypeNormal {
			t.Fatalf("unexpected weaging: %#v", weaging)
		}
		if weaging.SowCode != "C-001" {
			t.Fatalf("sow code = %q, want C-001", weaging.SowCode)
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

	t.Run("flags the farrowing as nurse and keeps the sow lactating for Nodriza", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
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
		repository := repositoryWith(sowID, 2)
		repository.sows[sowID] = testSow(sowID, sowdomain.StatePregnant)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, partialweagingapplication.ErrSowNotLactating) {
			t.Fatalf("error = %v, want ErrSowNotLactating", err)
		}
	})

	t.Run("rejects a quantity that differs from the current piglets", func(t *testing.T) {
		repository := repositoryWith(sowID, 1)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, partialweagingdomain.ErrQuantityMismatch) {
			t.Fatalf("error = %v, want ErrQuantityMismatch", err)
		}
		if repository.created != nil {
			t.Fatalf("weaging should not be persisted")
		}
	})

	t.Run("rejects a quantity above the maximum", func(t *testing.T) {
		repository := repositoryWith(sowID, 31)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })
		command := validCommand(sowID, actorID)
		command.Quantity = 31

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, partialweagingdomain.ErrQuantityExceedsMax) {
			t.Fatalf("error = %v, want ErrQuantityExceedsMax", err)
		}
	})

	t.Run("rejects a weight above the maximum", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })
		command := validCommand(sowID, actorID)
		weight := 500.0
		command.TotalWeight = &weight

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, partialweagingdomain.ErrTotalWeightExceedsMax) {
			t.Fatalf("error = %v, want ErrTotalWeightExceedsMax", err)
		}
	})

	t.Run("accepts a weaging date on the last partial weaging day", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.lastPartialWeagingDate = weagingDay
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("rejects a weaging date before the last partial weaging", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.lastPartialWeagingDate = weagingDay.AddDate(0, 0, 1)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateBeforePartial) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforePartial", err)
		}
	})

	t.Run("propagates the last partial weaging date error", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.lastPartialWeagingErr = errors.New("boom")
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.lastPartialWeagingErr) {
			t.Fatalf("error = %v, want the partial weaging date error", err)
		}
	})

	t.Run("accepts a weaging date on the last event day", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.lastEventDate = weagingDay
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("rejects a weaging date before the last event", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.lastEventDate = weagingDay.AddDate(0, 0, 1)
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, partialweagingdomain.ErrWeagingDateBeforeEvents) {
			t.Fatalf("error = %v, want ErrWeagingDateBeforeEvents", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.sowErr = ports.ErrSowNotFound
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.farrowErr = ports.ErrFarrowingNotFound
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates the latest event date error", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.lastEventErr = errors.New("boom")
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.lastEventErr) {
			t.Fatalf("error = %v, want the event date error", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith(sowID, 2)
		repository.createErr = errors.New("boom")
		useCase := partialweagingapplication.NewCreatePartialWeagingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
