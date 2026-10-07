package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

func TestListSows(t *testing.T) {
	t.Run("returns the first page with its metadata", func(t *testing.T) {
		repository := &fakeSowRepository{
			listSows:  []*sowdomain.Sow{testSow(uuid.New()), testSow(uuid.New())},
			listTotal: 42,
		}
		service := sowapplication.NewListSowsService(repository)

		result, err := service.Execute(context.Background(), ports.SowFilter{}, 1)

		if err != nil || len(result.Items) != 2 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
		if result.Total != 42 || result.Page != 1 || result.PageSize != 20 || result.TotalPages != 3 {
			t.Fatalf("unexpected metadata: %#v", result)
		}
		if repository.listLimit != 20 || repository.listOffset != 0 {
			t.Fatalf("unexpected pagination: limit=%d offset=%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("computes the offset for a later page", func(t *testing.T) {
		repository := &fakeSowRepository{listSows: []*sowdomain.Sow{}, listTotal: 0}
		service := sowapplication.NewListSowsService(repository)

		result, err := service.Execute(context.Background(), ports.SowFilter{}, 3)

		if err != nil || result.Page != 3 || result.TotalPages != 0 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
		if repository.listLimit != 20 || repository.listOffset != 40 {
			t.Fatalf("unexpected pagination: limit=%d offset=%d", repository.listLimit, repository.listOffset)
		}
	})

	t.Run("defaults an invalid page to the first one", func(t *testing.T) {
		repository := &fakeSowRepository{listSows: []*sowdomain.Sow{}}
		service := sowapplication.NewListSowsService(repository)

		result, err := service.Execute(context.Background(), ports.SowFilter{}, 0)

		if err != nil || result.Page != 1 || repository.listOffset != 0 {
			t.Fatalf("unexpected result: err=%v result=%#v offset=%d", err, result, repository.listOffset)
		}
	})

	t.Run("forwards the filter to the repository", func(t *testing.T) {
		repository := &fakeSowRepository{}
		service := sowapplication.NewListSowsService(repository)
		code := "C-001"
		active := true
		origin := sowdomain.OriginExternal
		state := sowdomain.StatePregnant
		breedID := uuid.New()

		_, err := service.Execute(context.Background(), ports.SowFilter{
			Code:    &code,
			BreedID: &breedID,
			Origin:  &origin,
			Active:  &active,
			State:   &state,
		}, 1)

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
		if filter.State == nil || *filter.State != state {
			t.Fatalf("unexpected state: %#v", filter.State)
		}
	})

	t.Run("propagates the list error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeSowRepository{listErr: unexpected}
		service := sowapplication.NewListSowsService(repository)

		_, err := service.Execute(context.Background(), ports.SowFilter{}, 1)

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
