package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	sowremovalapplication "server/internal/modules/sowremoval/application"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

func TestListSowRemovalsExecute(t *testing.T) {
	t.Run("returns the removals from the repository", func(t *testing.T) {
		expected := []*sowremovaldomain.SowRemoval{{ID: uuid.New()}}
		repository := &fakeSowRemovalRepository{listRemovals: expected}
		useCase := sowremovalapplication.NewListSowRemovalsService(repository)

		removals, err := useCase.Execute(context.Background(), ports.SowRemovalFilter{})

		if err != nil || len(removals) != 1 || removals[0] != expected[0] {
			t.Fatalf("unexpected result: err=%v removals=%#v", err, removals)
		}
	})

	t.Run("forwards the filter to the repository", func(t *testing.T) {
		sowID := uuid.New()
		repository := &fakeSowRemovalRepository{}
		useCase := sowremovalapplication.NewListSowRemovalsService(repository)

		_, err := useCase.Execute(context.Background(), ports.SowRemovalFilter{SowID: &sowID})

		if err != nil || repository.listFilter.SowID == nil || *repository.listFilter.SowID != sowID {
			t.Fatalf("unexpected filter: %#v err=%v", repository.listFilter, err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakeSowRemovalRepository{listErr: errors.New("boom")}
		useCase := sowremovalapplication.NewListSowRemovalsService(repository)

		_, err := useCase.Execute(context.Background(), ports.SowRemovalFilter{})

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
