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
	t.Run("returns one page of weagings with its metadata", func(t *testing.T) {
		expected := []*weagingdomain.Weaging{{ID: uuid.New()}}
		repository := &fakeWeagingRepository{listWeagings: expected, listTotal: 12}
		useCase := weagingapplication.NewListWeagingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.WeagingFilter{}, 1)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(page.Items) != 1 || page.Items[0] != expected[0] {
			t.Fatalf("unexpected items: %#v", page.Items)
		}
		if page.Total != 12 || page.Page != 1 || page.PageSize != 10 || page.TotalPages != 2 {
			t.Fatalf("unexpected metadata: %#v", page)
		}
		if repository.listLimit != 10 || repository.listOffset != 0 {
			t.Fatalf("limit=%d offset=%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("computes the offset from the requested page", func(t *testing.T) {
		repository := &fakeWeagingRepository{listTotal: 30}
		useCase := weagingapplication.NewListWeagingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.WeagingFilter{}, 3)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listOffset != 20 || page.Page != 3 || page.TotalPages != 3 {
			t.Fatalf("unexpected pagination: page=%#v limit=%d offset=%d", page, repository.listLimit, repository.listOffset)
		}
	})

	t.Run("defaults a non-positive page to one", func(t *testing.T) {
		repository := &fakeWeagingRepository{}
		useCase := weagingapplication.NewListWeagingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.WeagingFilter{}, 0)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if page.Page != 1 || repository.listOffset != 0 {
			t.Fatalf("unexpected page: %#v offset=%d", page, repository.listOffset)
		}
	})

	t.Run("returns zero total pages when there are no matches", func(t *testing.T) {
		repository := &fakeWeagingRepository{}
		useCase := weagingapplication.NewListWeagingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.WeagingFilter{}, 1)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if page.Total != 0 || page.TotalPages != 0 {
			t.Fatalf("unexpected metadata: %#v", page)
		}
	})

	t.Run("forwards the sow code and date filters to the repository", func(t *testing.T) {
		sowCode := "C-001"
		from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)
		repository := &fakeWeagingRepository{}
		useCase := weagingapplication.NewListWeagingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.WeagingFilter{
			SowCode:  &sowCode,
			FromDate: &from,
			ToDate:   &to,
		}, 1)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listFilter.SowCode == nil || *repository.listFilter.SowCode != sowCode {
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

		_, err := useCase.Execute(context.Background(), ports.WeagingFilter{}, 1)

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
