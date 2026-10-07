package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
)

func TestUpdateSow(t *testing.T) {
	now := time.Date(2026, time.March, 5, 6, 7, 8, 0, time.UTC)
	clock := func() time.Time { return now }
	updatedBy := uuid.New()
	sowID := uuid.New()

	t.Run("updates provided fields and keeps the state and active flag", func(t *testing.T) {
		repository := &fakeSowRepository{getSow: testSow(sowID)}
		service := sowapplication.NewUpdateSowService(repository, clock)
		code := "  C-002  "
		parity := 5

		sow, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			Code:      &code,
			Parity:    &parity,
			UpdatedBy: updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if sow.Code != "C-002" || !sow.Active || sow.State != sowdomain.StateAlive || sow.Parity != 5 {
			t.Fatalf("unexpected sow: %#v", sow)
		}
		if sow.UpdatedBy != updatedBy || !sow.UpdatedAt.Equal(now) || repository.updated != sow {
			t.Fatalf("unexpected sow: %#v", sow)
		}
	})

	t.Run("rejects a date change on an inactive sow", func(t *testing.T) {
		sow := testSow(sowID)
		sow.Active = false
		repository := &fakeSowRepository{getSow: sow}
		service := sowapplication.NewUpdateSowService(repository, clock)
		entry := sow.EntryDate.Add(-24 * time.Hour)

		_, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			EntryDate: &entry,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, sowdomain.ErrInactiveSowCannotUpdateDates) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("rejects a parity change when the sow is not alive", func(t *testing.T) {
		sow := testSow(sowID)
		sow.State = sowdomain.StatePregnant
		repository := &fakeSowRepository{getSow: sow}
		service := sowapplication.NewUpdateSowService(repository, clock)
		parity := 4

		_, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			Parity:    &parity,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, sowdomain.ErrParityRequiresAliveState) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("rejects an entry date after the last service", func(t *testing.T) {
		sow := testSow(sowID)
		lastService := sow.EntryDate.Add(-24 * time.Hour)
		repository := &fakeSowRepository{getSow: sow, lastServiceDate: &lastService}
		service := sowapplication.NewUpdateSowService(repository, clock)
		entry := sow.EntryDate

		_, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			EntryDate: &entry,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, sowdomain.ErrEntryDateAfterService) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("propagates the last service lookup error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeSowRepository{getSow: testSow(sowID), lastServiceErr: unexpected}
		service := sowapplication.NewUpdateSowService(repository, clock)
		entry := testSow(sowID).EntryDate

		_, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			EntryDate: &entry,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("clears nullable fields", func(t *testing.T) {
		sow := testSow(sowID)
		location := "Corral A"
		note := "Nota"
		birth := sow.EntryDate.Add(-24 * time.Hour)
		sow.Location = &location
		sow.Note = &note
		sow.BirthDate = &birth
		repository := &fakeSowRepository{getSow: sow}
		service := sowapplication.NewUpdateSowService(repository, clock)
		empty := ""

		result, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			Location:       &empty,
			Note:           &empty,
			ClearBirthDate: true,
			UpdatedBy:      updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Location != nil || result.Note != nil || result.BirthDate != nil {
			t.Fatalf("unexpected sow: %#v", result)
		}
		if repository.updated != result {
			t.Fatalf("repository did not persist the cleared sow")
		}
	})

	t.Run("updates the origin", func(t *testing.T) {
		repository := &fakeSowRepository{getSow: testSow(sowID)}
		service := sowapplication.NewUpdateSowService(repository, clock)
		origin := sowdomain.OriginExternal

		sow, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			Origin:    &origin,
			UpdatedBy: updatedBy,
		})

		if err != nil || sow.Origin != sowdomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects an invalid origin", func(t *testing.T) {
		repository := &fakeSowRepository{getSow: testSow(sowID)}
		service := sowapplication.NewUpdateSowService(repository, clock)
		origin := sowdomain.Origin("Desconocido")

		_, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			Origin:    &origin,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, sowdomain.ErrInvalidOrigin) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("propagates the get error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeSowRepository{getErr: unexpected}
		service := sowapplication.NewUpdateSowService(repository, clock)
		code := "C-002"

		_, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			Code:      &code,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		repository := &fakeSowRepository{getSow: testSow(sowID)}
		service := sowapplication.NewUpdateSowService(repository, clock)

		_, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, sowdomain.ErrInvalidUpdate) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("propagates the update error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeSowRepository{getSow: testSow(sowID), updateErr: unexpected}
		service := sowapplication.NewUpdateSowService(repository, clock)
		code := "C-002"

		_, err := service.Execute(context.Background(), sowID, sowapplication.UpdateSowCommand{
			Code:      &code,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
