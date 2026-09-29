package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	sowremovalapplication "server/internal/modules/sowremoval/application"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

func testRemoval(id, sowID uuid.UUID, state sowdomain.State, removalType sowremovaldomain.Type) *sowremovaldomain.SowRemoval {
	return &sowremovaldomain.SowRemoval{
		ID:          id,
		SowID:       sowID,
		RemovalDate: removalDay,
		Type:        removalType,
		Reason:      sowremovaldomain.ReasonDisease,
		LastState:   string(state),
		CreatedAt:   nowReference,
		UpdatedAt:   nowReference,
		CreatedBy:   uuid.New(),
		UpdatedBy:   uuid.New(),
	}
}

func TestUpdateSowRemovalExecute(t *testing.T) {
	removalID := uuid.New()
	sowID := uuid.New()
	serviceID := uuid.New()
	actorID := uuid.New()
	updateNow := nowReference.Add(time.Hour)

	t.Run("updates the removal fields and records the actor", func(t *testing.T) {
		removal := testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath)
		repository := &fakeSowRemovalRepository{removalByID: removal}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		reason := sowremovaldomain.ReasonOther
		updated, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			Reason:    &reason,
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated.Reason != sowremovaldomain.ReasonOther {
			t.Fatalf("reason = %q", updated.Reason)
		}
		if !updated.UpdatedAt.Equal(updateNow) || updated.UpdatedBy != actorID {
			t.Fatalf("unexpected audit: %#v", updated)
		}
		if repository.updated != removal {
			t.Fatalf("removal was not persisted")
		}
		if repository.updatedSow != nil {
			t.Fatalf("sow must not change, got %#v", repository.updatedSow)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
		}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		_, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			UpdatedBy: actorID,
		})

		if !errors.Is(err, sowremovaldomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
		if repository.updated != nil {
			t.Fatalf("removal must not be persisted")
		}
	})

	t.Run("rejects a removal date before the sow entry date", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
			sow:         testSow(sowID, sowdomain.StateDead),
		}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		beforeEntry := time.Date(2025, time.November, 30, 0, 0, 0, 0, time.UTC)
		_, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			RemovalDate: &beforeEntry,
			UpdatedBy:   actorID,
		})

		if !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeEntry) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeEntry", err)
		}
	})

	t.Run("validates the removal date against the last mount", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StatePregnant, sowremovaldomain.TypeDeath),
			sow:         testSow(sowID, sowdomain.StateDead),
			service:     testService(serviceID, sowID, servicedomain.StateFailed, testMount(serviceID, 1, mountDayOne)),
		}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		date := mountDayOne
		_, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			RemovalDate: &date,
			UpdatedBy:   actorID,
		})

		if !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeService) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeService", err)
		}
	})

	t.Run("validates the removal date against the last abortion", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAborted, sowremovaldomain.TypeDeath),
			sow:         testSow(sowID, sowdomain.StateDead),
			abortion:    testAbortion(uuid.New(), sowID, removalDay),
		}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		date := removalDay
		_, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			RemovalDate: &date,
			UpdatedBy:   actorID,
		})

		if !errors.Is(err, sowremovaldomain.ErrRemovalDateBeforeAbortion) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeAbortion", err)
		}
	})

	t.Run("changes the sow state when the type changes", func(t *testing.T) {
		removal := testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath)
		repository := &fakeSowRemovalRepository{
			removalByID: removal,
			sow:         testSow(sowID, sowdomain.StateDead),
		}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		discard := sowremovaldomain.TypeDiscard
		updated, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			Type:      &discard,
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated.Type != sowremovaldomain.TypeDiscard {
			t.Fatalf("type = %q", updated.Type)
		}
		if repository.updatedSow == nil || repository.updatedSow.State != sowdomain.StateDiscarded {
			t.Fatalf("sow state = %#v", repository.updatedSow)
		}
		if repository.updatedSow.UpdatedBy != actorID {
			t.Fatalf("sow updatedBy = %v", repository.updatedSow.UpdatedBy)
		}
	})

	t.Run("keeps the sow when the type matches its state", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
			sow:         testSow(sowID, sowdomain.StateDead),
		}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		death := sowremovaldomain.TypeDeath
		_, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			Type:      &death,
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.updatedSow != nil {
			t.Fatalf("sow must not change, got %#v", repository.updatedSow)
		}
	})

	t.Run("propagates a missing removal", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{removalByIDErr: ports.ErrSowRemovalNotFound}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		reason := sowremovaldomain.ReasonOther
		_, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			Reason:    &reason,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrSowRemovalNotFound) {
			t.Fatalf("error = %v, want ErrSowRemovalNotFound", err)
		}
	})

	t.Run("propagates a missing sow on a date change", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
			sowErr:      ports.ErrSowNotFound,
		}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		date := removalDay
		_, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			RemovalDate: &date,
			UpdatedBy:   actorID,
		})

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{
			removalByID: testRemoval(removalID, sowID, sowdomain.StateAlive, sowremovaldomain.TypeDeath),
			updateErr:   errors.New("boom"),
		}
		useCase := sowremovalapplication.NewUpdateSowRemovalService(repository, func() time.Time { return updateNow })

		reason := sowremovaldomain.ReasonOther
		_, err := useCase.Execute(context.Background(), removalID, sowremovalapplication.UpdateSowRemovalCommand{
			Reason:    &reason,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, repository.updateErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
