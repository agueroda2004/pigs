package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	boarremovalapplication "server/internal/modules/boarremoval/application"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
)

func TestListBoarRemovalsExecute(t *testing.T) {
	t.Run("returns the removals from the repository", func(t *testing.T) {
		expected := []*boarremovaldomain.BoarRemoval{{ID: uuid.New()}}
		repository := &fakeBoarRemovalRepository{listRemovals: expected}
		useCase := boarremovalapplication.NewListBoarRemovalsService(repository)

		removals, err := useCase.Execute(context.Background(), ports.BoarRemovalFilter{})

		if err != nil || len(removals) != 1 || removals[0] != expected[0] {
			t.Fatalf("unexpected result: err=%v removals=%#v", err, removals)
		}
	})

	t.Run("forwards the filter to the repository", func(t *testing.T) {
		boarID := uuid.New()
		repository := &fakeBoarRemovalRepository{}
		useCase := boarremovalapplication.NewListBoarRemovalsService(repository)

		_, err := useCase.Execute(context.Background(), ports.BoarRemovalFilter{BoarID: &boarID})

		if err != nil || repository.listFilter.BoarID == nil || *repository.listFilter.BoarID != boarID {
			t.Fatalf("unexpected filter: %#v err=%v", repository.listFilter, err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakeBoarRemovalRepository{listErr: errors.New("boom")}
		useCase := boarremovalapplication.NewListBoarRemovalsService(repository)

		_, err := useCase.Execute(context.Background(), ports.BoarRemovalFilter{})

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
