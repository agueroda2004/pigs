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

// repositoryWithWeaging builds a repository holding one weaged farrowing.
// The farrowing balance is zero because the weaging already consumed it.
func repositoryWithWeaging(quantity int) (*fakeWeagingRepository, *weagingdomain.Weaging) {
	sowID := uuid.New()
	farrowing := testFarrowing(uuid.New(), sowID, 0)
	weaging := &weagingdomain.Weaging{
		ID:          uuid.New(),
		FarrowingID: farrowing.ID,
		SowID:       sowID,
		WeagingDate: weagingDay,
		Quantity:    quantity,
	}

	return &fakeWeagingRepository{
		sows: map[uuid.UUID]*sowdomain.Sow{
			sowID: testSow(sowID, sowdomain.StateWeaned),
		},
		farrowings: map[uuid.UUID]*farrowingdomain.Farrowing{
			sowID: farrowing,
		},
		farrowingsByID: map[uuid.UUID]*farrowingdomain.Farrowing{
			farrowing.ID: farrowing,
		},
		weagings: map[uuid.UUID]*weagingdomain.Weaging{
			weaging.ID: weaging,
		},
	}, weaging
}

func TestDeleteWeagingExecute(t *testing.T) {
	actorID := uuid.New()

	t.Run("restores the farrowing balance and the sow state", func(t *testing.T) {
		repository, weaging := repositoryWithWeaging(8)
		useCase := weagingapplication.NewDeleteWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), weagingapplication.DeleteWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deleted != weaging {
			t.Fatalf("weaging was not deleted")
		}
		if repository.deletedFarrowing.CurrentPiglets != 8 {
			t.Fatalf("current piglets = %d, want 8", repository.deletedFarrowing.CurrentPiglets)
		}
		if repository.deletedSow.State != sowdomain.StateLactating {
			t.Fatalf("sow state = %q, want Lactando", repository.deletedSow.State)
		}
		if !repository.deletedFarrowing.UpdatedAt.Equal(nowReference) ||
			repository.deletedFarrowing.UpdatedBy != actorID {
			t.Fatalf("unexpected farrowing audit fields: %#v", repository.deletedFarrowing)
		}
		if repository.futureEventsSowID != weaging.SowID ||
			!repository.futureEventsAfter.Equal(weaging.WeagingDate) {
			t.Fatalf("unexpected future-events check: %v %v", repository.futureEventsSowID, repository.futureEventsAfter)
		}
	})

	t.Run("blocks the delete when the sow has a future mount or removal", func(t *testing.T) {
		repository, weaging := repositoryWithWeaging(8)
		repository.hasFutureEvents = true
		useCase := weagingapplication.NewDeleteWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), weagingapplication.DeleteWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, weagingdomain.ErrHasFutureEvents) {
			t.Fatalf("error = %v, want ErrHasFutureEvents", err)
		}
		if repository.deleted != nil {
			t.Fatalf("weaging should not be deleted")
		}
		if repository.sows[weaging.SowID].State != sowdomain.StateWeaned {
			t.Fatalf("sow state changed to %q", repository.sows[weaging.SowID].State)
		}
		if repository.farrowingsByID[weaging.FarrowingID].CurrentPiglets != 0 {
			t.Fatalf("farrowing balance should not change")
		}
	})

	t.Run("propagates a missing weaging", func(t *testing.T) {
		repository, weaging := repositoryWithWeaging(8)
		repository.weagingByIDErr = ports.ErrWeagingNotFound
		useCase := weagingapplication.NewDeleteWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), weagingapplication.DeleteWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrWeagingNotFound) {
			t.Fatalf("error = %v, want ErrWeagingNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository, weaging := repositoryWithWeaging(8)
		repository.farrowingByIDErr = ports.ErrFarrowingNotFound
		useCase := weagingapplication.NewDeleteWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), weagingapplication.DeleteWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates a future-events lookup error", func(t *testing.T) {
		repository, weaging := repositoryWithWeaging(8)
		repository.futureEventsErr = errors.New("boom")
		useCase := weagingapplication.NewDeleteWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), weagingapplication.DeleteWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.futureEventsErr) {
			t.Fatalf("error = %v, want the lookup error", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository, weaging := repositoryWithWeaging(8)
		repository.sowErr = ports.ErrSowNotFound
		useCase := weagingapplication.NewDeleteWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), weagingapplication.DeleteWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		repository, weaging := repositoryWithWeaging(8)
		useCase := weagingapplication.NewDeleteWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), weagingapplication.DeleteWeagingCommand{
			ID: weaging.ID,
		})

		if !errors.Is(err, sowdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
		if repository.deleted != nil {
			t.Fatalf("weaging should not be deleted")
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository, weaging := repositoryWithWeaging(8)
		repository.deleteErr = errors.New("boom")
		useCase := weagingapplication.NewDeleteWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), weagingapplication.DeleteWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.deleteErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
