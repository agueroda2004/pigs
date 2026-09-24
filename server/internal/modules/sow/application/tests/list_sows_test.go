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
	t.Run("returns every registered sow", func(t *testing.T) {
		repository := &fakeSowRepository{listSows: []*sowdomain.Sow{
			testSow(uuid.New()),
			testSow(uuid.New()),
		}}
		service := sowapplication.NewListSowsService(repository)

		result, err := service.Execute(context.Background(), ports.SowFilter{})

		if err != nil || len(result) != 2 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
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
		})

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

		_, err := service.Execute(context.Background(), ports.SowFilter{})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
