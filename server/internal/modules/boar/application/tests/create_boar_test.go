package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	boarapplication "server/internal/modules/boar/application"
	boardomain "server/internal/modules/boar/domain"
)

func TestCreateBoar(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
	clock := func() time.Time { return now }
	createdBy := uuid.New()
	breedID := uuid.New()
	entry := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)

	command := func() boarapplication.CreateBoarCommand {
		return boarapplication.CreateBoarCommand{
			Code:      "  B-001  ",
			EntryDate: entry,
			Origin:    boardomain.OriginOwn,
			BreedID:   breedID,
			CreatedBy: createdBy,
		}
	}

	t.Run("creates an alive boar with trimmed code", func(t *testing.T) {
		repository := &fakeBoarRepository{}
		service := boarapplication.NewCreateBoarService(repository, clock)

		boar, err := service.Execute(context.Background(), command())

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if boar.Code != "B-001" || boar.State != boardomain.StateAlive || !boar.Active {
			t.Fatalf("unexpected boar: %#v", boar)
		}
		if boar.Origin != boardomain.OriginOwn {
			t.Fatalf("unexpected origin: %#v", boar.Origin)
		}
		if boar.CreatedBy != createdBy || boar.BreedID != breedID || !boar.CreatedAt.Equal(now) {
			t.Fatalf("unexpected boar: %#v", boar)
		}
		if repository.created != boar {
			t.Fatalf("repository did not persist the created boar")
		}
	})

	t.Run("creates a boar with an external origin", func(t *testing.T) {
		repository := &fakeBoarRepository{}
		service := boarapplication.NewCreateBoarService(repository, clock)
		external := command()
		external.Origin = boardomain.OriginExternal

		boar, err := service.Execute(context.Background(), external)

		if err != nil || boar.Origin != boardomain.OriginExternal {
			t.Fatalf("unexpected result: err=%v boar=%#v", err, boar)
		}
	})

	t.Run("rejects a duplicated code", func(t *testing.T) {
		repository := &fakeBoarRepository{exists: true}
		service := boarapplication.NewCreateBoarService(repository, clock)

		_, err := service.Execute(context.Background(), command())

		if !errors.Is(err, boarapplication.ErrBoarCodeAlreadyExists) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%#v", err, repository.created)
		}
	})

	t.Run("propagates the exists check error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeBoarRepository{existsErr: unexpected}
		service := boarapplication.NewCreateBoarService(repository, clock)

		_, err := service.Execute(context.Background(), command())

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})

	t.Run("rejects an invalid command before persisting", func(t *testing.T) {
		repository := &fakeBoarRepository{}
		service := boarapplication.NewCreateBoarService(repository, clock)
		invalid := command()
		invalid.EntryDate = time.Time{}

		_, err := service.Execute(context.Background(), invalid)

		if !errors.Is(err, boardomain.ErrInvalidEntryDate) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%#v", err, repository.created)
		}
	})

	t.Run("propagates the create error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeBoarRepository{createErr: unexpected}
		service := boarapplication.NewCreateBoarService(repository, clock)

		_, err := service.Execute(context.Background(), command())

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
