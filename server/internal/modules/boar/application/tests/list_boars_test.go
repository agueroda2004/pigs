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
	t.Run("returns every registered boar", func(t *testing.T) {
		repository := &fakeBoarRepository{listBoars: []*boardomain.Boar{
			testBoar(uuid.New()),
			testBoar(uuid.New()),
		}}
		service := boarapplication.NewListBoarsService(repository)

		result, err := service.Execute(context.Background(), ports.BoarFilter{})

		if err != nil || len(result) != 2 {
			t.Fatalf("unexpected result: err=%v result=%#v", err, result)
		}
	})

	t.Run("forwards the filter to the repository", func(t *testing.T) {
		repository := &fakeBoarRepository{}
		service := boarapplication.NewListBoarsService(repository)
		code := "B-001"
		active := true
		origin := boardomain.OriginExternal
		breedID := uuid.New()

		_, err := service.Execute(context.Background(), ports.BoarFilter{
			Code:    &code,
			BreedID: &breedID,
			Origin:  &origin,
			Active:  &active,
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
	})

	t.Run("propagates the list error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeBoarRepository{listErr: unexpected}
		service := boarapplication.NewListBoarsService(repository)

		_, err := service.Execute(context.Background(), ports.BoarFilter{})

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
