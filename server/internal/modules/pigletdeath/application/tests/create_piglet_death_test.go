package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
	sowdomain "server/internal/modules/sow/domain"
)

func TestCreatePigletDeathExecute(t *testing.T) {
	sowID := uuid.New()
	farrowingID := uuid.New()
	operatorID := uuid.New()
	actorID := uuid.New()

	repositoryWith := func() *fakePigletDeathRepository {
		return &fakePigletDeathRepository{
			sow:       testSow(sowID, sowdomain.StateLactating),
			farrowing: testFarrowing(farrowingID, sowID, 10),
			operator:  testOperator(operatorID),
		}
	}

	t.Run("registers the death and reduces the current piglets", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewCreatePigletDeathService(repository, func() time.Time { return nowReference })

		death, err := useCase.Execute(context.Background(), validCommand(sowID, operatorID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if death.FarrowingID != farrowingID || death.SowID != sowID || death.Quantity != 2 {
			t.Fatalf("unexpected death: %#v", death)
		}
		if death.OperatorName != "Operador" {
			t.Fatalf("operator name = %q, want Operador", death.OperatorName)
		}
		if death.Weight == nil || *death.Weight != 2.5 {
			t.Fatalf("weight = %v, want 2.5", death.Weight)
		}
		if death.CreatedBy != actorID || !death.CreatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", death)
		}
		if repository.created != death {
			t.Fatalf("death was not persisted")
		}
		if repository.createdFarrowing.CurrentPiglets != 8 {
			t.Fatalf("current piglets = %d, want 8", repository.createdFarrowing.CurrentPiglets)
		}
	})

	t.Run("rejects a sow that is not lactating", func(t *testing.T) {
		repository := repositoryWith()
		repository.sow = testSow(sowID, sowdomain.StatePregnant)
		useCase := pigletdeathapplication.NewCreatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, operatorID, actorID))

		if !errors.Is(err, pigletdeathapplication.ErrSowNotLactating) {
			t.Fatalf("error = %v, want ErrSowNotLactating", err)
		}
	})

	t.Run("rejects a quantity greater than the balance", func(t *testing.T) {
		repository := repositoryWith()
		repository.farrowing = testFarrowing(farrowingID, sowID, 1)
		useCase := pigletdeathapplication.NewCreatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, operatorID, actorID))

		if !errors.Is(err, farrowingdomain.ErrInsufficientPiglets) {
			t.Fatalf("error = %v, want ErrInsufficientPiglets", err)
		}
	})

	t.Run("rejects a death date before the farrowing", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewCreatePigletDeathService(repository, func() time.Time { return nowReference })
		command := validCommand(sowID, operatorID, actorID)
		command.DeathDate = farrowDay

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, pigletdeathdomain.ErrDeathDateBeforeFarrowing) {
			t.Fatalf("error = %v, want ErrDeathDateBeforeFarrowing", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := &fakePigletDeathRepository{sowErr: ports.ErrSowNotFound}
		useCase := pigletdeathapplication.NewCreatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, operatorID, actorID))

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository := repositoryWith()
		repository.farrowing = nil
		repository.farrowingErr = ports.ErrFarrowingNotFound
		useCase := pigletdeathapplication.NewCreatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, operatorID, actorID))

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates a missing operator", func(t *testing.T) {
		repository := repositoryWith()
		repository.operatorErr = ports.ErrOperatorNotFound
		useCase := pigletdeathapplication.NewCreatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, operatorID, actorID))

		if !errors.Is(err, ports.ErrOperatorNotFound) {
			t.Fatalf("error = %v, want ErrOperatorNotFound", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.createErr = errors.New("boom")
		useCase := pigletdeathapplication.NewCreatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, operatorID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
