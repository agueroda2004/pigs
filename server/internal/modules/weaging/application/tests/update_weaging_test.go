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

func updateRepositoryWithWeaging() (*fakeWeagingRepository, *weagingdomain.Weaging) {
	repository, weaging := repositoryWithWeaging(8)
	weight := 120.5
	destination := "Nave 2"
	note := "Destete completo de la camada"
	weaging.TotalWeight = &weight
	weaging.Destination = &destination
	weaging.Note = &note
	return repository, weaging
}

func TestUpdateWeagingService(t *testing.T) {
	ctx := context.Background()
	actor := uuid.New()

	t.Run("updates the editable fields", func(t *testing.T) {
		repository, weaging := updateRepositoryWithWeaging()
		weight := 130.5
		destination := "  Nave 3  "
		note := "  Nota actualizada  "
		useCase := weagingapplication.NewUpdateWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, weagingapplication.UpdateWeagingCommand{
			TotalWeight: &weight,
			Destination: &destination,
			Note:        &note,
			UpdatedBy:   actor,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.updated != weaging {
			t.Fatalf("unexpected repository update: %#v", repository.updated)
		}
		if weaging.TotalWeight == nil || *weaging.TotalWeight != 130.5 {
			t.Fatalf("unexpected weight: %#v", weaging.TotalWeight)
		}
		if weaging.Destination == nil || *weaging.Destination != "Nave 3" {
			t.Fatalf("unexpected destination: %#v", weaging.Destination)
		}
		if weaging.Note == nil || *weaging.Note != "Nota actualizada" {
			t.Fatalf("unexpected note: %#v", weaging.Note)
		}
		if weaging.UpdatedBy != actor || !weaging.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit fields: %#v", weaging)
		}
	})

	t.Run("clears the weight", func(t *testing.T) {
		repository, weaging := updateRepositoryWithWeaging()
		useCase := weagingapplication.NewUpdateWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, weagingapplication.UpdateWeagingCommand{
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

	t.Run("propagates a missing weaging", func(t *testing.T) {
		repository, weaging := updateRepositoryWithWeaging()
		repository.weagingByIDErr = ports.ErrWeagingNotFound
		useCase := weagingapplication.NewUpdateWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, weagingapplication.UpdateWeagingCommand{
			ClearTotalWeight: true,
			UpdatedBy:        actor,
		})

		if !errors.Is(err, ports.ErrWeagingNotFound) {
			t.Fatalf("error = %v, want ErrWeagingNotFound", err)
		}
		if repository.updated != nil {
			t.Fatalf("weaging should not be updated")
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		repository, weaging := updateRepositoryWithWeaging()
		useCase := weagingapplication.NewUpdateWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, weagingapplication.UpdateWeagingCommand{
			UpdatedBy: actor,
		})

		if !errors.Is(err, weagingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
		if repository.updated != nil {
			t.Fatalf("weaging should not be updated")
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		repository, weaging := updateRepositoryWithWeaging()
		useCase := weagingapplication.NewUpdateWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, weagingapplication.UpdateWeagingCommand{
			ClearTotalWeight: true,
		})

		if !errors.Is(err, weagingdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
		if repository.updated != nil {
			t.Fatalf("weaging should not be updated")
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository, weaging := updateRepositoryWithWeaging()
		repository.updateErr = errors.New("boom")
		useCase := weagingapplication.NewUpdateWeagingService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(ctx, weaging.ID, weagingapplication.UpdateWeagingCommand{
			ClearTotalWeight: true,
			UpdatedBy:        actor,
		})

		if !errors.Is(err, repository.updateErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
