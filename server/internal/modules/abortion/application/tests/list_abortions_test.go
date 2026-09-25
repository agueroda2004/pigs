package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	abortionapplication "server/internal/modules/abortion/application"
	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
)

func TestListAbortionsExecute(t *testing.T) {
	t.Run("returns the abortions from the repository", func(t *testing.T) {
		expected := []*abortiondomain.Abortion{{ID: uuid.New()}}
		repository := &fakeAbortionRepository{listAbortions: expected}
		useCase := abortionapplication.NewListAbortionsService(repository)

		abortions, err := useCase.Execute(context.Background(), ports.AbortionFilter{})

		if err != nil || len(abortions) != 1 || abortions[0] != expected[0] {
			t.Fatalf("unexpected result: err=%v abortions=%#v", err, abortions)
		}
	})

	t.Run("forwards the filter to the repository", func(t *testing.T) {
		sowID := uuid.New()
		repository := &fakeAbortionRepository{}
		useCase := abortionapplication.NewListAbortionsService(repository)

		_, err := useCase.Execute(context.Background(), ports.AbortionFilter{SowID: &sowID})

		if err != nil || repository.listFilter.SowID == nil || *repository.listFilter.SowID != sowID {
			t.Fatalf("unexpected filter: %#v err=%v", repository.listFilter, err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakeAbortionRepository{listErr: errors.New("boom")}
		useCase := abortionapplication.NewListAbortionsService(repository)

		_, err := useCase.Execute(context.Background(), ports.AbortionFilter{})

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
