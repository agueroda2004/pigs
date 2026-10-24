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

func TestDeletePigletDeathExecute(t *testing.T) {
	deathID := uuid.New()
	farrowingID := uuid.New()
	actorID := uuid.New()

	repositoryWith := func() *fakePigletDeathRepository {
		return &fakePigletDeathRepository{
			deathByID:     testPigletDeath(deathID, farrowingID, 2),
			farrowingByID: testFarrowing(farrowingID, uuid.New(), 8),
		}
	}

	t.Run("deletes the death and restores the current piglets", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewDeletePigletDeathService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), pigletdeathapplication.DeletePigletDeathCommand{
			ID:        deathID,
			DeletedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.farrowingByID.CurrentPiglets != 10 {
			t.Fatalf("current piglets = %d, want 10", repository.farrowingByID.CurrentPiglets)
		}
		if repository.deleted != repository.deathByID || repository.deletedFarrowing != repository.farrowingByID {
			t.Fatalf("death was not deleted")
		}
	})

	t.Run("blocks the deletion when a weaging happened after the death", func(t *testing.T) {
		repository := repositoryWith()
		repository.hasWeagingAfter = true
		useCase := pigletdeathapplication.NewDeletePigletDeathService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), pigletdeathapplication.DeletePigletDeathCommand{
			ID:        deathID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, pigletdeathdomain.ErrHasLaterWeaging) {
			t.Fatalf("error = %v, want ErrHasLaterWeaging", err)
		}
		if repository.deleted != nil {
			t.Fatalf("death should not be deleted")
		}
	})

	t.Run("blocks the deletion when a partial weaging happened after the death", func(t *testing.T) {
		repository := repositoryWith()
		repository.hasPartialWeagingAfter = true
		useCase := pigletdeathapplication.NewDeletePigletDeathService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), pigletdeathapplication.DeletePigletDeathCommand{
			ID:        deathID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, pigletdeathdomain.ErrHasLaterPartialWeaging) {
			t.Fatalf("error = %v, want ErrHasLaterPartialWeaging", err)
		}
		if repository.deleted != nil {
			t.Fatalf("death should not be deleted")
		}
	})

	t.Run("propagates a missing death", func(t *testing.T) {
		repository := &fakePigletDeathRepository{deathByIDErr: ports.ErrPigletDeathNotFound}
		useCase := pigletdeathapplication.NewDeletePigletDeathService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), pigletdeathapplication.DeletePigletDeathCommand{
			ID:        deathID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrPigletDeathNotFound) {
			t.Fatalf("error = %v, want ErrPigletDeathNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository := repositoryWith()
		repository.farrowingByID = nil
		repository.farrowingIDErr = ports.ErrFarrowingNotFound
		useCase := pigletdeathapplication.NewDeletePigletDeathService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), pigletdeathapplication.DeletePigletDeathCommand{
			ID:        deathID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates the weaging lookup error", func(t *testing.T) {
		repository := repositoryWith()
		repository.hasWeagingAfterErr = errors.New("boom")
		useCase := pigletdeathapplication.NewDeletePigletDeathService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), pigletdeathapplication.DeletePigletDeathCommand{
			ID:        deathID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.hasWeagingAfterErr) {
			t.Fatalf("error = %v, want the weaging lookup error", err)
		}
	})

	t.Run("propagates the partial weaging lookup error", func(t *testing.T) {
		repository := repositoryWith()
		repository.hasPartialWeagingErr = errors.New("boom")
		useCase := pigletdeathapplication.NewDeletePigletDeathService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), pigletdeathapplication.DeletePigletDeathCommand{
			ID:        deathID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.hasPartialWeagingErr) {
			t.Fatalf("error = %v, want the partial weaging lookup error", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.deleteErr = errors.New("boom")
		useCase := pigletdeathapplication.NewDeletePigletDeathService(repository, func() time.Time { return nowReference })

		err := useCase.Execute(context.Background(), pigletdeathapplication.DeletePigletDeathCommand{
			ID:        deathID,
			DeletedBy: actorID,
		})

		if !errors.Is(err, repository.deleteErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
