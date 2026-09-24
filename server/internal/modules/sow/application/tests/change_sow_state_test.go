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

func TestChangeSowState(t *testing.T) {
	now := time.Date(2026, time.April, 1, 2, 3, 4, 0, time.UTC)
	clock := func() time.Time { return now }
	updatedBy := uuid.New()
	sowID := uuid.New()

	t.Run("changes the state through the repository", func(t *testing.T) {
		repository := &fakeSowRepository{getSow: testSow(sowID)}
		service := sowapplication.NewChangeSowStateService(repository, clock)

		sow, err := service.Execute(context.Background(), sowID, sowapplication.ChangeSowStateCommand{
			State:     sowdomain.StatePregnant,
			UpdatedBy: updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if sow.State != sowdomain.StatePregnant || repository.updatedState != sow || repository.updated != nil {
			t.Fatalf("unexpected result: sow=%#v repository=%#v", sow, repository)
		}
		if sow.UpdatedBy != updatedBy || !sow.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected audit fields: %#v", sow)
		}
	})

	t.Run("propagates the get error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeSowRepository{getErr: unexpected}
		service := sowapplication.NewChangeSowStateService(repository, clock)

		_, err := service.Execute(context.Background(), sowID, sowapplication.ChangeSowStateCommand{
			State:     sowdomain.StateDead,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) || repository.updatedState != nil {
			t.Fatalf("unexpected result: err=%v updatedState=%#v", err, repository.updatedState)
		}
	})

	t.Run("rejects an invalid state", func(t *testing.T) {
		repository := &fakeSowRepository{getSow: testSow(sowID)}
		service := sowapplication.NewChangeSowStateService(repository, clock)

		_, err := service.Execute(context.Background(), sowID, sowapplication.ChangeSowStateCommand{
			State:     sowdomain.State("Unknown"),
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, sowdomain.ErrInvalidState) || repository.updatedState != nil {
			t.Fatalf("unexpected result: err=%v updatedState=%#v", err, repository.updatedState)
		}
	})

	t.Run("propagates the update state error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeSowRepository{getSow: testSow(sowID), updateStateErr: unexpected}
		service := sowapplication.NewChangeSowStateService(repository, clock)

		_, err := service.Execute(context.Background(), sowID, sowapplication.ChangeSowStateCommand{
			State:     sowdomain.StateDead,
			UpdatedBy: updatedBy,
		})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
