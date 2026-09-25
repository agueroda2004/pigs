package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	serviceapplication "server/internal/modules/service/application"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	sowdomain "server/internal/modules/sow/domain"
)

func TestCreateService(t *testing.T) {
	now := time.Date(2026, time.February, 1, 3, 4, 5, 0, time.UTC)
	clock := func() time.Time { return now }

	newRepository := func() (*fakeServiceRepository, uuid.UUID, uuid.UUID, uuid.UUID) {
		sowID := uuid.New()
		boarID := uuid.New()
		operatorID := uuid.New()
		return &fakeServiceRepository{
			sow:      testSow(sowID, sowdomain.StateAlive),
			boar:     testBoar(boarID),
			operator: testOperator(operatorID),
		}, sowID, boarID, operatorID
	}

	t.Run("registers a service and moves the sow to pregnant", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		service := serviceapplication.NewCreateServiceService(repository, clock)
		createdBy := uuid.New()

		result, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, createdBy))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.SowID != sowID || result.State != servicedomain.StateConfirmed {
			t.Fatalf("unexpected service: %#v", result)
		}
		if len(result.Mounts) != 1 || result.Mounts[0].BoarID != boarID || result.Mounts[0].OperatorID != operatorID {
			t.Fatalf("unexpected mounts: %#v", result.Mounts)
		}
		if result.Mounts[0].MountNumber != 1 || result.Mounts[0].Type != servicedomain.MountTypeArtificial {
			t.Fatalf("unexpected mount: %#v", result.Mounts[0])
		}
		wantFarrowing := mountDayOne.AddDate(0, 0, 114)
		if result.ExpectedFarrowingDate == nil || !result.ExpectedFarrowingDate.Equal(wantFarrowing) {
			t.Fatalf("expected farrowing = %#v, want %v", result.ExpectedFarrowingDate, wantFarrowing)
		}
		if result.CreatedBy != createdBy {
			t.Fatalf("unexpected creator: %#v", result)
		}
		if repository.sow.State != sowdomain.StatePregnant {
			t.Fatalf("sow state = %v, want %v", repository.sow.State, sowdomain.StatePregnant)
		}
		if repository.sow.UpdatedBy != createdBy || !repository.sow.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected sow audit: %#v", repository.sow)
		}
		if repository.created != result || repository.createdSow != repository.sow {
			t.Fatalf("repository did not persist the service and sow together")
		}
	})

	t.Run("supports up to three mounts", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		service := serviceapplication.NewCreateServiceService(repository, clock)
		command := validCommand(sowID, boarID, operatorID, uuid.New())
		command.Mounts = append(command.Mounts,
			mountCommand(boarID, operatorID, mountDayOne),
			mountCommand(boarID, operatorID, mountDayNext),
		)

		result, err := service.Execute(context.Background(), command)

		if err != nil || len(result.Mounts) != 3 {
			t.Fatalf("unexpected result: err=%v service=%#v", err, result)
		}
		for index, mount := range result.Mounts {
			if mount.MountNumber != index+1 {
				t.Fatalf("mount[%d] number = %d", index, mount.MountNumber)
			}
		}
	})

	t.Run("rejects a sow that is not eligible", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		repository.sow = testSow(sowID, sowdomain.StateDead)
		service := serviceapplication.NewCreateServiceService(repository, clock)

		_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))

		if !errors.Is(err, serviceapplication.ErrSowNotEligible) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%#v", err, repository.created)
		}
	})

	t.Run("accepts every eligible sow state", func(t *testing.T) {
		for _, state := range []sowdomain.State{
			sowdomain.StateAlive,
			sowdomain.StateWeaned,
			sowdomain.StateAborted,
			sowdomain.StatePregnant,
		} {
			repository, sowID, boarID, operatorID := newRepository()
			repository.sow = testSow(sowID, state)
			service := serviceapplication.NewCreateServiceService(repository, clock)

			_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))
			if err != nil {
				t.Fatalf("state=%v err=%v", state, err)
			}
		}
	})

	t.Run("propagates the sow lookup error", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		unexpected := errors.New("unexpected")
		repository.sowErr = unexpected
		service := serviceapplication.NewCreateServiceService(repository, clock)

		_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})

	t.Run("propagates the boar lookup error", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		repository.boarErr = ports.ErrBoarNotFound
		service := serviceapplication.NewCreateServiceService(repository, clock)

		_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))

		if !errors.Is(err, ports.ErrBoarNotFound) {
			t.Fatalf("error = %v, want ErrBoarNotFound", err)
		}
	})

	t.Run("rejects an inactive or non-alive boar", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		repository.boar.Active = false
		service := serviceapplication.NewCreateServiceService(repository, clock)

		_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))
		if !errors.Is(err, serviceapplication.ErrBoarNotEligible) {
			t.Fatalf("error = %v, want ErrBoarNotEligible", err)
		}

		repository, sowID, boarID, operatorID = newRepository()
		repository.boar.State = boardomain.StateDead
		service = serviceapplication.NewCreateServiceService(repository, clock)

		_, err = service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))
		if !errors.Is(err, serviceapplication.ErrBoarNotEligible) {
			t.Fatalf("error = %v, want ErrBoarNotEligible", err)
		}
	})

	t.Run("rejects an inactive operator", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		repository.operator.Active = false
		service := serviceapplication.NewCreateServiceService(repository, clock)

		_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))

		if !errors.Is(err, serviceapplication.ErrOperatorNotAvailable) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%#v", err, repository.created)
		}
	})

	t.Run("propagates the operator lookup error", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		repository.operatorErr = ports.ErrOperatorNotFound
		service := serviceapplication.NewCreateServiceService(repository, clock)

		_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))

		if !errors.Is(err, ports.ErrOperatorNotFound) {
			t.Fatalf("error = %v, want ErrOperatorNotFound", err)
		}
	})

	t.Run("rejects a domain validation error before persisting", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		service := serviceapplication.NewCreateServiceService(repository, clock)
		future := now.AddDate(0, 0, 1)

		_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		repository, sowID, boarID, operatorID = newRepository()
		command := validCommand(sowID, boarID, operatorID, uuid.New())
		command.Mounts = []serviceapplication.CreateMountCommand{mountCommand(boarID, operatorID, future)}
		service = serviceapplication.NewCreateServiceService(repository, clock)

		_, err = service.Execute(context.Background(), command)
		if !errors.Is(err, servicedomain.ErrMountDateInFuture) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%#v", err, repository.created)
		}
	})

	t.Run("propagates the create error", func(t *testing.T) {
		repository, sowID, boarID, operatorID := newRepository()
		unexpected := errors.New("unexpected")
		repository.createErr = unexpected
		service := serviceapplication.NewCreateServiceService(repository, clock)

		_, err := service.Execute(context.Background(), validCommand(sowID, boarID, operatorID, uuid.New()))

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
	})
}
