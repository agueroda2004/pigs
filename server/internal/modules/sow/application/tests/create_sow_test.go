package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
)

func TestCreateSow(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
	clock := func() time.Time { return now }
	createdBy := uuid.New()
	breedID := uuid.New()
	entry := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)

	command := func() sowapplication.CreateSowCommand {
		return sowapplication.CreateSowCommand{
			Code:      "  C-001  ",
			EntryDate: entry,
			Origin:    sowdomain.OriginOwn,
			Parity:    3,
			BreedID:   breedID,
			CreatedBy: createdBy,
		}
	}

	t.Run("creates an alive sow with trimmed code", func(t *testing.T) {
		repository := &fakeSowRepository{}
		service := sowapplication.NewCreateSowService(repository, clock)

		sow, err := service.Execute(context.Background(), command())

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if sow.Code != "C-001" || sow.State != sowdomain.StateAlive || !sow.Active {
			t.Fatalf("unexpected sow: %#v", sow)
		}
		if sow.Origin != sowdomain.OriginOwn || sow.Parity != 3 {
			t.Fatalf("unexpected sow: %#v", sow)
		}
		if sow.CreatedBy != createdBy || sow.BreedID != breedID || !sow.CreatedAt.Equal(now) {
			t.Fatalf("unexpected sow: %#v", sow)
		}
		if repository.created != sow {
			t.Fatalf("repository did not persist the created sow")
		}
	})

	t.Run("creates a sow with an external origin", func(t *testing.T) {
		repository := &fakeSowRepository{}
		service := sowapplication.NewCreateSowService(repository, clock)
		external := command()
		external.Origin = sowdomain.OriginExternal

		sow, err := service.Execute(context.Background(), external)

		if err != nil || sow.Origin != sowdomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v sow=%#v", err, sow)
		}
	})

	t.Run("rejects a duplicated code", func(t *testing.T) {
		repository := &fakeSowRepository{exists: true}
		service := sowapplication.NewCreateSowService(repository, clock)

		_, err := service.Execute(context.Background(), command())

		if !errors.Is(err, sowapplication.ErrSowCodeAlreadyExists) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%#v", err, repository.created)
		}
	})

	t.Run("propagates the exists check error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeSowRepository{existsErr: unexpected}
		service := sowapplication.NewCreateSowService(repository, clock)

		_, err := service.Execute(context.Background(), command())

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})

	t.Run("rejects an invalid command before persisting", func(t *testing.T) {
		repository := &fakeSowRepository{}
		service := sowapplication.NewCreateSowService(repository, clock)
		invalid := command()
		invalid.EntryDate = time.Time{}

		_, err := service.Execute(context.Background(), invalid)

		if !errors.Is(err, sowdomain.ErrInvalidEntryDate) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%#v", err, repository.created)
		}
	})

	t.Run("rejects a negative parity", func(t *testing.T) {
		repository := &fakeSowRepository{}
		service := sowapplication.NewCreateSowService(repository, clock)
		invalid := command()
		invalid.Parity = -1

		_, err := service.Execute(context.Background(), invalid)

		if !errors.Is(err, sowdomain.ErrInvalidParity) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%#v", err, repository.created)
		}
	})

	t.Run("propagates the create error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeSowRepository{createErr: unexpected}
		service := sowapplication.NewCreateSowService(repository, clock)

		_, err := service.Execute(context.Background(), command())

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
