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

// deleteWeaging builds the stored partial weaging used by the delete scenarios.
func deleteWeaging(farrowingID, sowID uuid.UUID, weagingType partialweagingdomain.Type, quantity int) *partialweagingdomain.PartialWeaging {
	return &partialweagingdomain.PartialWeaging{
		ID:          uuid.New(),
		FarrowingID: farrowingID,
		SowID:       sowID,
		WeagingDate: weagingDay,
		Quantity:    quantity,
		Type:        weagingType,
		CreatedAt:   weagingDay,
		UpdatedAt:   weagingDay,
		CreatedBy:   uuid.New(),
		UpdatedBy:   uuid.New(),
	}
}

// deleteRepository builds a fake repository holding a farrowing and a lactating sow.
func deleteRepository(
	weaging *partialweagingdomain.PartialWeaging,
	currentPiglets int,
	sowState sowdomain.State,
) *fakePartialWeagingRepository {
	farrowing := testFarrowing(weaging.FarrowingID, weaging.SowID, currentPiglets)
	return &fakePartialWeagingRepository{
		sows: map[uuid.UUID]*sowdomain.Sow{
			weaging.SowID: testSow(weaging.SowID, sowState),
		},
		weagingByID:   weaging,
		farrowingByID: farrowing,
	}
}

func TestDeletePartialWeagingExecute(t *testing.T) {
	actorID := uuid.New()

	t.Run("deletes a normal weaging, restores the piglets and the lactating sow", func(t *testing.T) {
		farrowingID := uuid.New()
		sowID := uuid.New()
		weaging := deleteWeaging(farrowingID, sowID, partialweagingdomain.TypeNormal, 4)
		repository := deleteRepository(weaging, 0, sowdomain.StateWeaned)
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deleted != weaging {
			t.Fatalf("weaging was not deleted")
		}
		if repository.deletedFarrowing.CurrentPiglets != 4 {
			t.Fatalf("current piglets = %d, want 4", repository.deletedFarrowing.CurrentPiglets)
		}
		if repository.deletedSow == nil || repository.deletedSow.State != sowdomain.StateLactating {
			t.Fatalf("sow was not restored to Lactando: %#v", repository.deletedSow)
		}
		if !repository.deletedFarrowing.UpdatedAt.Equal(nowReference) {
			t.Fatalf("farrowing updated at = %v", repository.deletedFarrowing.UpdatedAt)
		}
	})

	t.Run("blocks a normal weaging with later mounts or removals", func(t *testing.T) {
		weaging := deleteWeaging(uuid.New(), uuid.New(), partialweagingdomain.TypeNormal, 4)
		repository := deleteRepository(weaging, 0, sowdomain.StateWeaned)
		repository.hasFutureEvents = true
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, partialweagingdomain.ErrHasFutureEvents) {
			t.Fatalf("error = %v, want ErrHasFutureEvents", err)
		}
		if repository.deleted != nil {
			t.Fatalf("weaging should not be deleted")
		}
		if repository.deletedFarrowing != nil {
			t.Fatalf("farrowing should not change")
		}
	})

	t.Run("deletes a Nodriza weaging and clears the nurse state", func(t *testing.T) {
		farrowingID := uuid.New()
		weaging := deleteWeaging(farrowingID, uuid.New(), partialweagingdomain.TypeNodriza, 3)
		repository := deleteRepository(weaging, 0, sowdomain.StateLactating)
		repository.farrowingByID.IsNurse = true
		repository.farrowingByID.NurseStartDate = &weagingDay
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.deletedFarrowing.IsNurse || repository.deletedFarrowing.NurseStartDate != nil {
			t.Fatalf("nurse state was not cleared: %#v", repository.deletedFarrowing)
		}
		if repository.deletedFarrowing.CurrentPiglets != 3 {
			t.Fatalf("current piglets = %d, want 3", repository.deletedFarrowing.CurrentPiglets)
		}
		if repository.deletedSow != nil {
			t.Fatalf("the sow should be left untouched for a Nodriza")
		}
	})

	t.Run("keeps the nurse state when an earlier Nodriza weaging exists", func(t *testing.T) {
		farrowingID := uuid.New()
		weaging := deleteWeaging(farrowingID, uuid.New(), partialweagingdomain.TypeNodriza, 2)
		repository := deleteRepository(weaging, 0, sowdomain.StateLactating)
		repository.farrowingByID.IsNurse = true
		repository.farrowingByID.NurseStartDate = &weagingDay
		previous := weagingDay.AddDate(0, 0, -2)
		repository.previousNurseDate = previous
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !repository.deletedFarrowing.IsNurse || repository.deletedFarrowing.NurseStartDate == nil {
			t.Fatalf("nurse state should stay active: %#v", repository.deletedFarrowing)
		}
		if !repository.deletedFarrowing.NurseStartDate.Equal(previous) {
			t.Fatalf("nurse start date = %v, want %v", repository.deletedFarrowing.NurseStartDate, previous)
		}
	})

	t.Run("blocks a Nodriza weaging with later farrowing events", func(t *testing.T) {
		weaging := deleteWeaging(uuid.New(), uuid.New(), partialweagingdomain.TypeNodriza, 2)
		repository := deleteRepository(weaging, 0, sowdomain.StateLactating)
		repository.hasLaterNurseEvents = true
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, partialweagingdomain.ErrHasLaterNurseEvents) {
			t.Fatalf("error = %v, want ErrHasLaterNurseEvents", err)
		}
		if repository.deleted != nil {
			t.Fatalf("weaging should not be deleted")
		}
	})

	t.Run("propagates a missing weaging", func(t *testing.T) {
		repository := &fakePartialWeagingRepository{weagingByIDErr: ports.ErrPartialWeagingNotFound}
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        uuid.New(),
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrPartialWeagingNotFound) {
			t.Fatalf("error = %v, want ErrPartialWeagingNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		weaging := deleteWeaging(uuid.New(), uuid.New(), partialweagingdomain.TypeNormal, 2)
		repository := deleteRepository(weaging, 0, sowdomain.StateWeaned)
		repository.farrowingByIDErr = ports.ErrFarrowingNotFound
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates the future events query error", func(t *testing.T) {
		weaging := deleteWeaging(uuid.New(), uuid.New(), partialweagingdomain.TypeNormal, 2)
		repository := deleteRepository(weaging, 0, sowdomain.StateWeaned)
		repository.hasFutureEventsErr = errors.New("boom")
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.hasFutureEventsErr) {
			t.Fatalf("error = %v, want the future events error", err)
		}
	})

	t.Run("propagates the later nurse events query error", func(t *testing.T) {
		weaging := deleteWeaging(uuid.New(), uuid.New(), partialweagingdomain.TypeNodriza, 2)
		repository := deleteRepository(weaging, 0, sowdomain.StateLactating)
		repository.hasLaterNurseEventsErr = errors.New("boom")
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.hasLaterNurseEventsErr) {
			t.Fatalf("error = %v, want the nurse events error", err)
		}
	})

	t.Run("propagates the previous nurse date query error", func(t *testing.T) {
		weaging := deleteWeaging(uuid.New(), uuid.New(), partialweagingdomain.TypeNodriza, 2)
		repository := deleteRepository(weaging, 0, sowdomain.StateLactating)
		repository.previousNurseDateErr = errors.New("boom")
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.previousNurseDateErr) {
			t.Fatalf("error = %v, want the previous nurse date error", err)
		}
	})

	t.Run("propagates the persistence error", func(t *testing.T) {
		weaging := deleteWeaging(uuid.New(), uuid.New(), partialweagingdomain.TypeNormal, 2)
		repository := deleteRepository(weaging, 0, sowdomain.StateWeaned)
		repository.deleteErr = errors.New("boom")
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.deleteErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		weaging := deleteWeaging(uuid.New(), uuid.New(), partialweagingdomain.TypeNormal, 2)
		repository := deleteRepository(weaging, 0, sowdomain.StateWeaned)
		useCase := partialweagingapplication.NewDeletePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), partialweagingapplication.DeletePartialWeagingCommand{
			ID:        weaging.ID,
			DeletedBy: uuid.Nil,
		})

		if !errors.Is(err, sowdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})
}
