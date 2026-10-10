package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	boarapplication "server/internal/modules/boar/application"
	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

func TestListBoars(t *testing.T) {
	t.Run("returns the first page ordered by code", func(t *testing.T) {
		repository := &fakeBoarRepository{listBoars: []*boardomain.Boar{
			testBoar(uuid.New()),
			testBoar(uuid.New()),
		}}
		service := boarapplication.NewListBoarsService(repository)

		result, err := service.Execute(context.Background(), ports.BoarFilter{}, 1)

		if err != nil || len(result.Items) != 2 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
		if result.Page != 1 || result.PageSize != boarapplication.DefaultBoarPageSize {
			t.Fatalf("unexpected pagination: %#v", result)
		}
		if result.Total != 2 || result.TotalPages != 1 {
			t.Fatalf("unexpected totals: %#v", result)
		}
		if repository.listLimit != boarapplication.DefaultBoarPageSize || repository.listOffset != 0 {
			t.Fatalf("unexpected limit/offset: %d/%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("forwards the filter and computes the offset for later pages", func(t *testing.T) {
		repository := &fakeBoarRepository{listTotal: 25}
		service := boarapplication.NewListBoarsService(repository)
		code := "B-001"
		active := true
		origin := boardomain.OriginExternal
		breedID := uuid.New()

		result, err := service.Execute(context.Background(), ports.BoarFilter{
			Code:    &code,
			BreedID: &breedID,
			Origin:  &origin,
			Active:  &active,
		}, 3)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		filter := repository.listFilter
		if filter.Code == nil || *filter.Code != code {
			t.Fatalf("unexpected code: %#v", filter.Code)
		}
		if filter.BreedID == nil || *filter.BreedID != breedID {
			t.Fatalf("unexpected breed id: %#v", filter.BreedID)
		}
		if filter.Origin == nil || *filter.Origin != origin {
			t.Fatalf("unexpected origin: %#v", filter.Origin)
		}
		if filter.Active == nil || *filter.Active != active {
			t.Fatalf("unexpected active: %#v", filter.Active)
		}
		if repository.listOffset != 2*boarapplication.DefaultBoarPageSize {
			t.Fatalf("unexpected offset: %d", repository.listOffset)
		}
		if result.Page != 3 || result.Total != 25 || result.TotalPages != 3 {
			t.Fatalf("unexpected pagination: %#v", result)
		}
	})

	t.Run("clamps a non-positive page to one", func(t *testing.T) {
		repository := &fakeBoarRepository{}
		service := boarapplication.NewListBoarsService(repository)

		result, err := service.Execute(context.Background(), ports.BoarFilter{}, 0)

		if err != nil || result.Page != 1 || repository.listOffset != 0 {
			t.Fatalf("unexpected result: page=%d offset=%d err=%v", result.Page, repository.listOffset, err)
		}
	})

	t.Run("reports zero total pages when there are no boars", func(t *testing.T) {
		repository := &fakeBoarRepository{listBoars: []*boardomain.Boar{}}
		service := boarapplication.NewListBoarsService(repository)

		result, err := service.Execute(context.Background(), ports.BoarFilter{}, 1)

		if err != nil || result.Total != 0 || result.TotalPages != 0 {
			t.Fatalf("unexpected result: %#v err=%v", result, err)
		}
	})

	t.Run("propagates the list error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeBoarRepository{listErr: unexpected}
		service := boarapplication.NewListBoarsService(repository)

		_, err := service.Execute(context.Background(), ports.BoarFilter{}, 1)

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
