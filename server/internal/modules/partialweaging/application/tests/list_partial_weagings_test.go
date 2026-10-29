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
	t.Run("returns the first page with its pagination metadata", func(t *testing.T) {
		expected := []*partialweagingdomain.PartialWeaging{{ID: uuid.New()}}
		repository := &fakePartialWeagingRepository{listWeagings: expected, listTotal: 25}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{}, 1)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(page.Items) != 1 || page.Items[0] != expected[0] {
			t.Fatalf("unexpected items: %#v", page.Items)
		}
		if page.Total != 25 || page.Page != 1 || page.PageSize != 10 || page.TotalPages != 3 {
			t.Fatalf("unexpected pagination: %#v", page)
		}
		if repository.listLimit != 10 || repository.listOffset != 0 {
			t.Fatalf("unexpected limit or offset: %d %d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("calculates the offset for a later page", func(t *testing.T) {
		repository := &fakePartialWeagingRepository{listTotal: 25}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{}, 3)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if page.Page != 3 || page.TotalPages != 3 {
			t.Fatalf("unexpected pagination: %#v", page)
		}
		if repository.listLimit != 10 || repository.listOffset != 20 {
			t.Fatalf("unexpected limit or offset: %d %d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("normalizes a non-positive page to the first page", func(t *testing.T) {
		repository := &fakePartialWeagingRepository{}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{}, 0)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if page.Page != 1 || repository.listOffset != 0 {
			t.Fatalf("unexpected pagination: %#v offset=%d", page, repository.listOffset)
		}
	})

	t.Run("returns zero total pages when there are no matches", func(t *testing.T) {
		repository := &fakePartialWeagingRepository{}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{}, 1)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if page.Total != 0 || page.TotalPages != 0 {
			t.Fatalf("unexpected pagination: %#v", page)
		}
	})

	t.Run("forwards the sow code and date filters to the repository", func(t *testing.T) {
		sowCode := "C-001"
		from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)
		repository := &fakePartialWeagingRepository{}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{
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
		repository := &fakePartialWeagingRepository{listErr: errors.New("boom")}
		useCase := partialweagingapplication.NewListPartialWeagingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PartialWeagingFilter{}, 1)

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
