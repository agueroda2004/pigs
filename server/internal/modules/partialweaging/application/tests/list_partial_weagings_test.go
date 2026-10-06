package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	partialweagingapplication "server/internal/modules/partialweaging/application"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
)

func TestListPartialWeagingsExecute(t *testing.T) {
	t.Run("returns the partial weagings from the repository", func(t *testing.T) {
		expected := []*partialweagingdomain.PartialWeaging{{ID: uuid.New()}}
		repository := &fakePartialWeagingRepository{listWeagings: expected}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		weagings, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{})

		if err != nil || len(weagings) != 1 || weagings[0] != expected[0] {
			t.Fatalf("unexpected result: err=%v weagings=%#v", err, weagings)
		}
	})

	t.Run("forwards the sow and date filters to the repository", func(t *testing.T) {
		sowID := uuid.New()
		from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)
		repository := &fakePartialWeagingRepository{}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{
			SowID:    &sowID,
			FromDate: &from,
			ToDate:   &to,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listFilter.SowID == nil || *repository.listFilter.SowID != sowID {
			t.Fatalf("unexpected sow filter: %#v", repository.listFilter)
		}
		if repository.listFilter.FromDate == nil || !repository.listFilter.FromDate.Equal(from) {
			t.Fatalf("unexpected from filter: %#v", repository.listFilter.FromDate)
		}
		if repository.listFilter.ToDate == nil || !repository.listFilter.ToDate.Equal(to) {
			t.Fatalf("unexpected to filter: %#v", repository.listFilter.ToDate)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakePartialWeagingRepository{listErr: errors.New("boom")}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{})

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
