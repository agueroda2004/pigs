package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	farrowingapplication "server/internal/modules/farrowing/application"
	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
)

func TestListFarrowingsExecute(t *testing.T) {
	t.Run("returns the first page matching the filter", func(t *testing.T) {
		sowCode := "C-001"
		expected := []*farrowingdomain.Farrowing{{ID: uuid.New(), SowCode: sowCode}}
		repository := &fakeFarrowingRepository{listFarrowings: expected}
		useCase := farrowingapplication.NewListFarrowingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.FarrowingFilter{SowCode: &sowCode}, 1)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(page.Items) != 1 || page.Items[0].SowCode != sowCode {
			t.Fatalf("unexpected items: %#v", page.Items)
		}
		if repository.listFilter.SowCode == nil || *repository.listFilter.SowCode != sowCode {
			t.Fatalf("unexpected filter: %#v", repository.listFilter)
		}
		if page.Page != 1 || page.PageSize != farrowingapplication.DefaultFarrowingPageSize {
			t.Fatalf("unexpected pagination: %#v", page)
		}
		if repository.listLimit != farrowingapplication.DefaultFarrowingPageSize || repository.listOffset != 0 {
			t.Fatalf("unexpected limit/offset: %d/%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("computes the offset and total pages for later pages", func(t *testing.T) {
		repository := &fakeFarrowingRepository{listTotal: 25}
		useCase := farrowingapplication.NewListFarrowingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.FarrowingFilter{}, 3)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.listOffset != 2*farrowingapplication.DefaultFarrowingPageSize {
			t.Fatalf("unexpected offset: %d", repository.listOffset)
		}
		if page.Page != 3 || page.Total != 25 || page.TotalPages != 3 {
			t.Fatalf("unexpected pagination: %#v", page)
		}
	})

	t.Run("clamps a non-positive page to one", func(t *testing.T) {
		repository := &fakeFarrowingRepository{}
		useCase := farrowingapplication.NewListFarrowingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.FarrowingFilter{}, 0)

		if err != nil || page.Page != 1 || repository.listOffset != 0 {
			t.Fatalf("unexpected result: page=%d offset=%d err=%v", page.Page, repository.listOffset, err)
		}
	})

	t.Run("returns an empty page when there are no farrowings", func(t *testing.T) {
		repository := &fakeFarrowingRepository{listFarrowings: []*farrowingdomain.Farrowing{}}
		useCase := farrowingapplication.NewListFarrowingsService(repository)

		page, err := useCase.Execute(context.Background(), ports.FarrowingFilter{}, 1)

		if err != nil || len(page.Items) != 0 || page.Total != 0 || page.TotalPages != 0 {
			t.Fatalf("unexpected page: %#v err=%v", page, err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakeFarrowingRepository{listErr: errors.New("boom")}
		useCase := farrowingapplication.NewListFarrowingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.FarrowingFilter{}, 1)

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
