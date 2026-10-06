package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	weagingapplication "server/internal/modules/weaging/application"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

func TestListWeagingsExecute(t *testing.T) {
	t.Run("returns the weagings from the repository", func(t *testing.T) {
		expected := []*weagingdomain.Weaging{{ID: uuid.New()}}
		repository := &fakeWeagingRepository{listWeagings: expected}
		useCase := weagingapplication.NewListWeagingsService(repository)

		weagings, err := useCase.Execute(context.Background(), ports.WeagingFilter{})

		if err != nil || len(weagings) != 1 || weagings[0] != expected[0] {
			t.Fatalf("unexpected result: err=%v weagings=%#v", err, weagings)
		}
	})

	t.Run("forwards the sow and date filters to the repository", func(t *testing.T) {
		sowID := uuid.New()
		from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)
		repository := &fakeWeagingRepository{}
		useCase := weagingapplication.NewListWeagingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.WeagingFilter{
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
		repository := &fakeWeagingRepository{listErr: errors.New("boom")}
		useCase := weagingapplication.NewListWeagingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.WeagingFilter{})

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
