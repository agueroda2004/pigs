package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
)

func TestListPigletDeathsExecute(t *testing.T) {
	t.Run("returns the first page with its metadata", func(t *testing.T) {
		expected := []*pigletdeathdomain.PigletDeath{{ID: uuid.New()}}
		repository := &fakePigletDeathRepository{listDeaths: expected, listTotal: 25}
		useCase := pigletdeathapplication.NewListPigletDeathsService(repository)

		page, err := useCase.Execute(context.Background(), ports.PigletDeathFilter{}, 1)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(page.Items) != 1 || page.Items[0] != expected[0] {
			t.Fatalf("unexpected items: %#v", page.Items)
		}
		if page.Total != 25 || page.Page != 1 || page.PageSize != 10 || page.TotalPages != 3 {
			t.Fatalf("unexpected page: %#v", page)
		}
		if repository.listLimit != 10 || repository.listOffset != 0 {
			t.Fatalf("unexpected pagination: limit=%d offset=%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("computes the offset for a later page", func(t *testing.T) {
		repository := &fakePigletDeathRepository{listTotal: 25}
		useCase := pigletdeathapplication.NewListPigletDeathsService(repository)

		page, err := useCase.Execute(context.Background(), ports.PigletDeathFilter{}, 3)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if page.Page != 3 || repository.listOffset != 20 {
			t.Fatalf("unexpected page: %#v offset=%d", page, repository.listOffset)
		}
	})

	t.Run("falls back to the first page when the page is below one", func(t *testing.T) {
		repository := &fakePigletDeathRepository{}
		useCase := pigletdeathapplication.NewListPigletDeathsService(repository)

		page, err := useCase.Execute(context.Background(), ports.PigletDeathFilter{}, 0)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if page.Page != 1 || page.TotalPages != 0 || repository.listOffset != 0 {
			t.Fatalf("unexpected page: %#v", page)
		}
	})

	t.Run("forwards the sow code and date filters to the repository", func(t *testing.T) {
		sowCode := "C-001"
		from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)
		repository := &fakePigletDeathRepository{}
		useCase := pigletdeathapplication.NewListPigletDeathsService(repository)

		_, err := useCase.Execute(
			context.Background(),
			ports.PigletDeathFilter{SowCode: &sowCode, FromDate: &from, ToDate: &to},
			1,
		)

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
		repository := &fakePigletDeathRepository{listErr: errors.New("boom")}
		useCase := pigletdeathapplication.NewListPigletDeathsService(repository)

		_, err := useCase.Execute(context.Background(), ports.PigletDeathFilter{}, 1)

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
	})
}
