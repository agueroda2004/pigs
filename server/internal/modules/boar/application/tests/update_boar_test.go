package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	boarapplication "server/internal/modules/boar/application"
	boardomain "server/internal/modules/boar/domain"
)

func TestUpdateBoar(t *testing.T) {
	now := time.Date(2026, time.March, 5, 6, 7, 8, 0, time.UTC)
	clock := func() time.Time { return now }
	updatedBy := uuid.New()
	boarID := uuid.New()

	t.Run("updates provided fields and keeps the state", func(t *testing.T) {
		repository := &fakeBoarRepository{getBoar: testBoar(boarID)}
		service := boarapplication.NewUpdateBoarService(repository, clock)
		code := "  B-002  "
		active := false

		boar, err := service.Execute(context.Background(), boarID, boarapplication.UpdateBoarCommand{
			Code:      &code,
			Active:    &active,
			UpdatedBy: updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if boar.Code != "B-002" || boar.Active || boar.State != boardomain.StateAlive {
			t.Fatalf("unexpected boar: %#v", boar)
		}
		if boar.UpdatedBy != updatedBy || !boar.UpdatedAt.Equal(now) || repository.updated != boar {
			t.Fatalf("unexpected boar: %#v", boar)
		}
	})

	t.Run("clears nullable fields", func(t *testing.T) {
		boar := testBoar(boarID)
		location := "Corral A"
		note := "Nota"
		birth := boar.EntryDate.Add(-24 * time.Hour)
		boar.Location = &location
		boar.Note = &note
		boar.BirthDate = &birth
		repository := &fakeBoarRepository{getBoar: boar}
		service := boarapplication.NewUpdateBoarService(repository, clock)
		empty := ""

		result, err := service.Execute(context.Background(), boarID, boarapplication.UpdateBoarCommand{
			Location:       &empty,
			Note:           &empty,
			ClearBirthDate: true,
			UpdatedBy:      updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Location != nil || result.Note != nil || result.BirthDate != nil {
			t.Fatalf("unexpected boar: %#v", result)
		}
		if repository.updated != result {
			t.Fatalf("repository did not persist the cleared boar")
		}
	})

	t.Run("updates the origin", func(t *testing.T) {
		repository := &fakeBoarRepository{getBoar: testBoar(boarID)}
		service := boarapplication.NewUpdateBoarService(repository, clock)
		origin := boardomain.OriginExternal

		boar, err := service.Execute(context.Background(), boarID, boarapplication.UpdateBoarCommand{
			Origin:    &origin,
			UpdatedBy: updatedBy,
		})

		if err != nil || boar.Origin != boardomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects an invalid origin", func(t *testing.T) {
		repository := &fakeBoarRepository{getBoar: testBoar(boarID)}
		service := boarapplication.NewUpdateBoarService(repository, clock)
		origin := boardomain.Origin("Desconocido")

		_, err := service.Execute(context.Background(), boarID, boarapplication.UpdateBoarCommand{
			Origin:    &origin,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, boardomain.ErrInvalidOrigin) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("propagates the get error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeBoarRepository{getErr: unexpected}
		service := boarapplication.NewUpdateBoarService(repository, clock)
		code := "B-002"

		_, err := service.Execute(context.Background(), boarID, boarapplication.UpdateBoarCommand{
			Code:      &code,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		repository := &fakeBoarRepository{getBoar: testBoar(boarID)}
		service := boarapplication.NewUpdateBoarService(repository, clock)

		_, err := service.Execute(context.Background(), boarID, boarapplication.UpdateBoarCommand{
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, boardomain.ErrInvalidUpdate) || repository.updated != nil {
			t.Fatalf("unexpected result: err=%v updated=%#v", err, repository.updated)
		}
	})

	t.Run("propagates the update error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeBoarRepository{getBoar: testBoar(boarID), updateErr: unexpected}
		service := boarapplication.NewUpdateBoarService(repository, clock)
		code := "B-002"

		_, err := service.Execute(context.Background(), boarID, boarapplication.UpdateBoarCommand{
			Code:      &code,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
