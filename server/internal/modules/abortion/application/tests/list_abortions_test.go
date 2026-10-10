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
	t.Run("returns the first page with its metadata", func(t *testing.T) {
		expected := []*abortiondomain.Abortion{{ID: uuid.New()}, {ID: uuid.New()}}
		repository := &fakeAbortionRepository{listAbortions: expected, listTotal: 25}
		useCase := abortionapplication.NewListAbortionsService(repository)

		result, err := useCase.Execute(context.Background(), ports.AbortionFilter{}, 1)

		if err != nil || len(result.Items) != 2 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
		if result.Total != 25 || result.Page != 1 || result.PageSize != 10 || result.TotalPages != 3 {
			t.Fatalf("unexpected metadata: %#v", result)
		}
		if repository.listLimit != 10 || repository.listOffset != 0 {
			t.Fatalf("unexpected pagination: limit=%d offset=%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("computes the offset for a later page", func(t *testing.T) {
		repository := &fakeAbortionRepository{listAbortions: []*abortiondomain.Abortion{}, listTotal: 0}
		useCase := abortionapplication.NewListAbortionsService(repository)

		result, err := useCase.Execute(context.Background(), ports.AbortionFilter{}, 3)

		if err != nil || result.Page != 3 || result.TotalPages != 0 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
		if repository.listLimit != 10 || repository.listOffset != 20 {
			t.Fatalf("unexpected pagination: limit=%d offset=%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("defaults an invalid page to the first one", func(t *testing.T) {
		repository := &fakeAbortionRepository{listAbortions: []*abortiondomain.Abortion{}}
		useCase := abortionapplication.NewListAbortionsService(repository)

		result, err := useCase.Execute(context.Background(), ports.AbortionFilter{}, 0)

		if err != nil || result.Page != 1 || repository.listOffset != 0 {
			t.Fatalf("unexpected result: err=%v result=%#v offset=%d", err, result, repository.listOffset)
		}
	})

	t.Run("forwards the filter to the repository", func(t *testing.T) {
		sowCode := "C-001"
		repository := &fakeAbortionRepository{}
		useCase := abortionapplication.NewListAbortionsService(repository)

		_, err := useCase.Execute(context.Background(), ports.AbortionFilter{SowCode: &sowCode}, 1)

		if err != nil || repository.listFilter.SowCode == nil || *repository.listFilter.SowCode != sowCode {
			t.Fatalf("unexpected filter: %#v err=%v", repository.listFilter, err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakeAbortionRepository{listErr: errors.New("boom")}
		useCase := abortionapplication.NewListAbortionsService(repository)

		_, err := useCase.Execute(context.Background(), ports.AbortionFilter{}, 1)

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
