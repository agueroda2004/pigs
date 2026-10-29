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

func TestUpdatePigletDeathExecute(t *testing.T) {
	deathID := uuid.New()
	farrowingID := uuid.New()
	operatorID := uuid.New()
	actorID := uuid.New()

	weight := 3.5
	note := "nota actualizada"

	repositoryWith := func() *fakePigletDeathRepository {
		return &fakePigletDeathRepository{
			deathByID: testPigletDeath(deathID, farrowingID, 2),
			operator:  testOperator(operatorID),
		}
	}

	paramsWith := func() pigletdeathdomain.UpdatePigletDeathParams {
		return pigletdeathdomain.UpdatePigletDeathParams{
			OperatorID: &operatorID,
			Weight:     &weight,
			Note:       &note,
		}
	}

	t.Run("updates the editable fields and resolves the operator", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })

		death, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    paramsWith(),
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if death.OperatorID != operatorID || death.OperatorName != "Operador" {
			t.Fatalf("unexpected operator: %#v", death)
		}
		if death.Weight == nil || *death.Weight != 3.5 || death.Note == nil || *death.Note != note {
			t.Fatalf("unexpected fields: %#v", death)
		}
		if death.UpdatedBy != actorID || !death.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", death)
		}
		if repository.updated != death {
			t.Fatalf("death was not persisted")
		}
	})

	t.Run("keeps the weight when it is omitted", func(t *testing.T) {
		repository := repositoryWith()
		repository.deathByID.Weight = &weight
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })

		death, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    pigletdeathdomain.UpdatePigletDeathParams{Note: &note},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if death.Weight == nil || *death.Weight != 3.5 {
			t.Fatalf("weight = %v, want 3.5", death.Weight)
		}
	})

	t.Run("clears the weight", func(t *testing.T) {
		repository := repositoryWith()
		repository.deathByID.Weight = &weight
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })

		death, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    pigletdeathdomain.UpdatePigletDeathParams{ClearWeight: true},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if death.Weight != nil {
			t.Fatalf("weight = %v, want nil", death.Weight)
		}
		if repository.updated != death {
			t.Fatalf("death was not persisted")
		}
	})

	t.Run("updates the cause to Desaparecido", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })
		cause := pigletdeathdomain.CauseMissing

		death, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    pigletdeathdomain.UpdatePigletDeathParams{Cause: &cause},
			UpdatedBy: actorID,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if death.Cause != pigletdeathdomain.CauseMissing {
			t.Fatalf("cause = %v, want Desaparecido", death.Cause)
		}
		if repository.updated != death {
			t.Fatalf("death was not persisted")
		}
	})

	t.Run("rejects a changed death date", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })
		params := paramsWith()
		otherDate := deathDay.AddDate(0, 0, 1)
		params.DeathDate = &otherDate

		_, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    params,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, pigletdeathdomain.ErrImmutableDeathField) {
			t.Fatalf("error = %v, want ErrImmutableDeathField", err)
		}
	})

	t.Run("rejects a changed quantity", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })
		params := paramsWith()
		otherQuantity := 4
		params.Quantity = &otherQuantity

		_, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    params,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, pigletdeathdomain.ErrImmutableDeathField) {
			t.Fatalf("error = %v, want ErrImmutableDeathField", err)
		}
	})

	t.Run("rejects a changed sow", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })
		params := paramsWith()
		otherSow := uuid.New()
		params.SowID = &otherSow

		_, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    params,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, pigletdeathdomain.ErrImmutableDeathField) {
			t.Fatalf("error = %v, want ErrImmutableDeathField", err)
		}
	})

	t.Run("propagates a missing death", func(t *testing.T) {
		repository := &fakePigletDeathRepository{deathByIDErr: ports.ErrPigletDeathNotFound}
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    paramsWith(),
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrPigletDeathNotFound) {
			t.Fatalf("error = %v, want ErrPigletDeathNotFound", err)
		}
	})

	t.Run("propagates a missing operator", func(t *testing.T) {
		repository := repositoryWith()
		repository.operatorErr = ports.ErrOperatorNotFound
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    paramsWith(),
			UpdatedBy: actorID,
		})

		if !errors.Is(err, ports.ErrOperatorNotFound) {
			t.Fatalf("error = %v, want ErrOperatorNotFound", err)
		}
	})

	t.Run("propagates a domain validation error", func(t *testing.T) {
		repository := repositoryWith()
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })
		params := paramsWith()
		tooHeavy := 20.01
		params.Weight = &tooHeavy

		_, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    params,
			UpdatedBy: actorID,
		})

		if !errors.Is(err, pigletdeathdomain.ErrWeightTooHigh) {
			t.Fatalf("error = %v, want ErrWeightTooHigh", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.updateErr = errors.New("boom")
		useCase := pigletdeathapplication.NewUpdatePigletDeathService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), deathID, pigletdeathapplication.UpdatePigletDeathCommand{
			Params:    paramsWith(),
			UpdatedBy: actorID,
		})

		if !errors.Is(err, repository.updateErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
