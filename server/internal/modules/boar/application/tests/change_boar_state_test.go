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

func TestChangeBoarState(t *testing.T) {
	now := time.Date(2026, time.April, 1, 2, 3, 4, 0, time.UTC)
	clock := func() time.Time { return now }
	updatedBy := uuid.New()
	boarID := uuid.New()

	t.Run("changes the state through the repository", func(t *testing.T) {
		repository := &fakeBoarRepository{getBoar: testBoar(boarID)}
		service := boarapplication.NewChangeBoarStateService(repository, clock)

		boar, err := service.Execute(context.Background(), boarID, boarapplication.ChangeBoarStateCommand{
			State:     boardomain.StateDead,
			UpdatedBy: updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if boar.State != boardomain.StateDead || repository.updatedState != boar || repository.updated != nil {
			t.Fatalf("unexpected result: boar=%#v repository=%#v", boar, repository)
		}
		if boar.UpdatedBy != updatedBy || !boar.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected audit fields: %#v", boar)
		}
	})

	t.Run("propagates the get error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeBoarRepository{getErr: unexpected}
		service := boarapplication.NewChangeBoarStateService(repository, clock)

		_, err := service.Execute(context.Background(), boarID, boarapplication.ChangeBoarStateCommand{
			State:     boardomain.StateDead,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) || repository.updatedState != nil {
			t.Fatalf("unexpected result: err=%v updatedState=%#v", err, repository.updatedState)
		}
	})

	t.Run("rejects an invalid state", func(t *testing.T) {
		repository := &fakeBoarRepository{getBoar: testBoar(boarID)}
		service := boarapplication.NewChangeBoarStateService(repository, clock)

		_, err := service.Execute(context.Background(), boarID, boarapplication.ChangeBoarStateCommand{
			State:     boardomain.State("Unknown"),
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, boardomain.ErrInvalidState) || repository.updatedState != nil {
			t.Fatalf("unexpected result: err=%v updatedState=%#v", err, repository.updatedState)
		}
	})

	t.Run("propagates the update state error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeBoarRepository{getBoar: testBoar(boarID), updateStateErr: unexpected}
		service := boarapplication.NewChangeBoarStateService(repository, clock)

		_, err := service.Execute(context.Background(), boarID, boarapplication.ChangeBoarStateCommand{
			State:     boardomain.StateDead,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
