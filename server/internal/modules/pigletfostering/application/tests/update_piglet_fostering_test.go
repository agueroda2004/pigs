package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	pigletfosteringapplication "server/internal/modules/pigletfostering/application"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

func TestUpdatePigletFosteringExecute(t *testing.T) {
	actorID := uuid.New()

	t.Run("updates the note without touching the date, quantity or balances", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		note := "camada ajustada"
		storedDate := fostering.MovementDate

		updated, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Note: &note},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated != fostering || repository.updated != fostering {
			t.Fatalf("fostering was not persisted")
		}
		if !fostering.MovementDate.Equal(storedDate) || fostering.Quantity != 2 {
			t.Fatalf("date or quantity changed: %#v", fostering)
		}
		if fostering.Note == nil || *fostering.Note != note {
			t.Fatalf("unexpected note: %#v", fostering.Note)
		}
		if repository.farrowings[fostering.DonorSowID].CurrentPiglets != 8 ||
			repository.farrowings[fostering.ReceiverSowID].CurrentPiglets != 8 {
			t.Fatalf("balances should not change on a note update")
		}
		if repository.partialWeagingCalls != 0 {
			t.Fatalf("partial weaging calls = %d, want 0", repository.partialWeagingCalls)
		}
	})

	t.Run("clears the note with an empty value", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		empty := ""

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Note: &empty},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if fostering.Note != nil {
			t.Fatalf("note = %v, want nil", fostering.Note)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		long := string(make([]rune, 501))

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Note: &long},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
		if repository.updated != nil {
			t.Fatalf("fostering should not be persisted")
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			UpdatedBy: actorID,
		})

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		note := "camada ajustada"

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params: pigletfosteringdomain.UpdatePigletFosteringParams{Note: &note},
		})

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("propagates a missing fostering", func(t *testing.T) {
		repository, _ := repositoryWithFostering(8, 8, 2)
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		note := "camada ajustada"

		_, err := useCase.Execute(context.Background(), uuid.New(), pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Note: &note},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrPigletFosteringNotFound) {
			t.Fatalf("error = %v, want ErrPigletFosteringNotFound", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository, fostering := repositoryWithFostering(8, 8, 2)
		repository.updateErr = errors.New("boom")
		useCase := pigletfosteringapplication.NewUpdatePigletFosteringService(repository, func() time.Time { return nowReference })
		note := "camada ajustada"

		_, err := useCase.Execute(context.Background(), fostering.ID, pigletfosteringapplication.UpdatePigletFosteringCommand{
			Params:    pigletfosteringdomain.UpdatePigletFosteringParams{Note: &note},
			UpdatedBy: actorID,
		})

		if !errors.Is(err, repository.updateErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
