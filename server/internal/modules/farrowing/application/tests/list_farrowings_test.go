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
	t.Run("returns the farrowings matching the filter", func(t *testing.T) {
		sowID := uuid.New()
		expected := []*farrowingdomain.Farrowing{{ID: uuid.New(), SowID: sowID}}
		repository := &fakeFarrowingRepository{listFarrowings: expected}
		useCase := farrowingapplication.NewListFarrowingsService(repository)

		farrowings, err := useCase.Execute(context.Background(), ports.FarrowingFilter{SowID: &sowID})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(farrowings) != 1 || farrowings[0].SowID != sowID {
			t.Fatalf("unexpected farrowings: %#v", farrowings)
		}
		if repository.listFilter.SowID == nil || *repository.listFilter.SowID != sowID {
			t.Fatalf("unexpected filter: %#v", repository.listFilter)
		}
	})

	t.Run("returns an empty slice when there are no farrowings", func(t *testing.T) {
		repository := &fakeFarrowingRepository{listFarrowings: []*farrowingdomain.Farrowing{}}
		useCase := farrowingapplication.NewListFarrowingsService(repository)

		farrowings, err := useCase.Execute(context.Background(), ports.FarrowingFilter{})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(farrowings) != 0 {
			t.Fatalf("unexpected farrowings: %#v", farrowings)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakeFarrowingRepository{listErr: errors.New("boom")}
		useCase := farrowingapplication.NewListFarrowingsService(repository)

		_, err := useCase.Execute(context.Background(), ports.FarrowingFilter{})

		if !errors.Is(err, repository.listErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
