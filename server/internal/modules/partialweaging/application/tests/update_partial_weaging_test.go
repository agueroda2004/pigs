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

func updateTestWeaging(weagingType partialweagingdomain.Type) *partialweagingdomain.PartialWeaging {
	weight := 40.0
	note := "Destete parcial"
	actor := uuid.New()
	return &partialweagingdomain.PartialWeaging{
		ID:          uuid.New(),
		FarrowingID: uuid.New(),
		SowID:       uuid.New(),
		WeagingDate: time.Date(2026, time.April, 26, 0, 0, 0, 0, time.UTC),
		Quantity:    4,
		TotalWeight: &weight,
		Type:        weagingType,
		Note:        &note,
		CreatedAt:   nowReference,
		UpdatedAt:   nowReference,
		CreatedBy:   actor,
		UpdatedBy:   actor,
	}
}

func TestUpdatePartialWeagingService(t *testing.T) {
	ctx := context.Background()
	actor := uuid.New()

	t.Run("updates weight and note", func(t *testing.T) {
		weaging := updateTestWeaging(partialweagingdomain.TypeNormal)
		repository := &fakePartialWeagingRepository{weagingByID: weaging}
		weight := 55.5
		note := "  Nota actualizada  "
		useCase := partialweagingapplication.NewUpdatePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, partialweagingapplication.UpdatePartialWeagingCommand{
			TotalWeight: &weight,
			Note:        &note,
			UpdatedBy:   actor,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.updated != weaging {
			t.Fatalf("unexpected repository update: %#v", repository)
		}
		if weaging.TotalWeight == nil || *weaging.TotalWeight != 55.5 || weaging.Note == nil || *weaging.Note != "Nota actualizada" {
			t.Fatalf("unexpected fields: %#v", weaging)
		}
		if !weaging.UpdatedAt.Equal(nowReference) || weaging.UpdatedBy != actor {
			t.Fatalf("unexpected audit fields: %#v", weaging)
		}
		if !weaging.WeagingDate.Equal(time.Date(2026, time.April, 26, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("date should not change: %#v", weaging.WeagingDate)
		}
	})

	t.Run("clears the weight", func(t *testing.T) {
		weaging := updateTestWeaging(partialweagingdomain.TypeNormal)
		repository := &fakePartialWeagingRepository{weagingByID: weaging}
		useCase := partialweagingapplication.NewUpdatePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, partialweagingapplication.UpdatePartialWeagingCommand{
			ClearTotalWeight: true,
			UpdatedBy:        actor,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if weaging.TotalWeight != nil {
			t.Fatalf("weight = %v, want nil", weaging.TotalWeight)
		}
	})

	t.Run("clears the note with an empty value", func(t *testing.T) {
		weaging := updateTestWeaging(partialweagingdomain.TypeNormal)
		repository := &fakePartialWeagingRepository{weagingByID: weaging}
		empty := "   "
		useCase := partialweagingapplication.NewUpdatePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, partialweagingapplication.UpdatePartialWeagingCommand{
			Note:      &empty,
			UpdatedBy: actor,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if weaging.Note != nil {
			t.Fatalf("note = %v, want nil", weaging.Note)
		}
	})

	t.Run("propagates the lookup error", func(t *testing.T) {
		repository := &fakePartialWeagingRepository{weagingByIDErr: ports.ErrPartialWeagingNotFound}
		useCase := partialweagingapplication.NewUpdatePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, uuid.New(), partialweagingapplication.UpdatePartialWeagingCommand{UpdatedBy: actor})

		if !errors.Is(err, ports.ErrPartialWeagingNotFound) {
			t.Fatalf("error = %v, want ErrPartialWeagingNotFound", err)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		weaging := updateTestWeaging(partialweagingdomain.TypeNormal)
		repository := &fakePartialWeagingRepository{weagingByID: weaging}
		useCase := partialweagingapplication.NewUpdatePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, partialweagingapplication.UpdatePartialWeagingCommand{UpdatedBy: actor})

		if !errors.Is(err, partialweagingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
		if repository.updated != nil {
			t.Fatalf("nothing should be persisted")
		}
	})

	t.Run("propagates the update error", func(t *testing.T) {
		weaging := updateTestWeaging(partialweagingdomain.TypeNormal)
		repository := &fakePartialWeagingRepository{weagingByID: weaging, updateErr: ports.ErrPartialWeagingNotFound}
		useCase := partialweagingapplication.NewUpdatePartialWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, partialweagingapplication.UpdatePartialWeagingCommand{
			ClearTotalWeight: true,
			UpdatedBy:        actor,
		})

		if !errors.Is(err, ports.ErrPartialWeagingNotFound) {
			t.Fatalf("error = %v, want ErrPartialWeagingNotFound", err)
		}
	})
}
