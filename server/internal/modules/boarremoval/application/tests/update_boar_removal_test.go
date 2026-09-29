package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	boarremovalapplication "server/internal/modules/boarremoval/application"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
)

func testRemoval(id, boarID uuid.UUID, removalType boarremovaldomain.Type) *boarremovaldomain.BoarRemoval {
	return &boarremovaldomain.BoarRemoval{
		ID:          id,
		BoarID:      boarID,
		RemovalDate: removalDay,
		Type:        removalType,
		Reason:      boarremovaldomain.ReasonDisease,
		LastState:   "Vivo",
		CreatedAt:   nowReference,
		UpdatedAt:   nowReference,
		CreatedBy:   uuid.New(),
		UpdatedBy:   uuid.New(),
	}
}

func TestUpdateBoarRemovalExecute(t *testing.T) {
	removalID := uuid.New()
	boarID := uuid.New()
	actorID := uuid.New()
	updateNow := nowReference.Add(time.Hour)

	t.Run("updates the removal fields and records the actor", func(t *testing.T) {
		removal := testRemoval(removalID, boarID, boarremovaldomain.TypeDeath)
		repository := &fakeBoarRemovalRepository{removalByID: removal}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		reason := boarremovaldomain.ReasonOther
		updated, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			Reason:    &reason,
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated.Reason != boarremovaldomain.ReasonOther {
			t.Fatalf("reason = %q", updated.Reason)
		}
		if !updated.UpdatedAt.Equal(updateNow) || updated.UpdatedBy != actorID {
			t.Fatalf("unexpected audit: %#v", updated)
		}
		if repository.updated != removal {
			t.Fatalf("removal was not persisted")
		}
		if repository.updatedBoar != nil {
			t.Fatalf("boar must not change, got %#v", repository.updatedBoar)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
		}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		_, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			UpdatedBy: actorID,
		})

		if !errors.Is(err, boarremovaldomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
		if repository.updated != nil {
			t.Fatalf("removal must not be persisted")
		}
	})

	t.Run("rejects a removal date before the boar entry date", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			boar:        testBoar(boarID, boardomain.StateDead),
		}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		beforeEntry := entryDay.AddDate(0, 0, -1)
		_, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			RemovalDate: &beforeEntry,
			UpdatedBy:   actorID,
		})

		if !errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeEntry) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeEntry", err)
		}
	})

	t.Run("validates the removal date against the last mount", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			boar:        testBoar(boarID, boardomain.StateDead),
			lastMount:   mountDay,
		}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		date := mountDay.AddDate(0, 0, -1)
		_, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			RemovalDate: &date,
			UpdatedBy:   actorID,
		})

		if !errors.Is(err, boarremovaldomain.ErrRemovalDateBeforeMount) {
			t.Fatalf("error = %v, want ErrRemovalDateBeforeMount", err)
		}
	})

	t.Run("changes the boar state when the type changes", func(t *testing.T) {
		removal := testRemoval(removalID, boarID, boarremovaldomain.TypeDeath)
		repository := &fakeBoarRemovalRepository{
			removalByID: removal,
			boar:        testBoar(boarID, boardomain.StateDead),
		}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		discard := boarremovaldomain.TypeDiscard
		updated, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			Type:      &discard,
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated.Type != boarremovaldomain.TypeDiscard {
			t.Fatalf("type = %q", updated.Type)
		}
		if repository.updatedBoar == nil || repository.updatedBoar.State != boardomain.StateDiscarded {
			t.Fatalf("boar state = %#v", repository.updatedBoar)
		}
		if repository.updatedBoar.UpdatedBy != actorID {
			t.Fatalf("boar updatedBy = %v", repository.updatedBoar.UpdatedBy)
		}
	})

	t.Run("keeps the boar when the type matches its state", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			boar:        testBoar(boarID, boardomain.StateDead),
		}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		death := boarremovaldomain.TypeDeath
		_, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			Type:      &death,
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.updatedBoar != nil {
			t.Fatalf("boar must not change, got %#v", repository.updatedBoar)
		}
	})

	t.Run("propagates a missing removal", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{removalByIDErr: ports.ErrBoarRemovalNotFound}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		reason := boarremovaldomain.ReasonOther
		_, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			Reason:    &reason,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrBoarRemovalNotFound) {
			t.Fatalf("error = %v, want ErrBoarRemovalNotFound", err)
		}
	})

	t.Run("propagates a missing boar on a date change", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			boarErr:     ports.ErrBoarNotFound,
		}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		date := removalDay
		_, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			RemovalDate: &date,
			UpdatedBy:   actorID,
		})

		if !errors.Is(err, ports.ErrBoarNotFound) {
			t.Fatalf("error = %v, want ErrBoarNotFound", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{
			removalByID: testRemoval(removalID, boarID, boarremovaldomain.TypeDeath),
			updateErr:   errors.New("boom"),
		}
		useCase := boarremovalapplication.NewUpdateBoarRemovalService(repository, func() time.Time { return updateNow })

		reason := boarremovaldomain.ReasonOther
		_, err := useCase.Execute(context.Background(), removalID, boarremovalapplication.UpdateBoarRemovalCommand{
			Reason:    &reason,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, repository.updateErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
