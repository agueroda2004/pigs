package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	farrowingapplication "server/internal/modules/farrowing/application"
	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
	medicationdomain "server/internal/modules/medication/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

func TestCreateFarrowingExecute(t *testing.T) {
	sowID := uuid.New()
	serviceID := uuid.New()
	actorID := uuid.New()

	repositoryWith := func() *fakeFarrowingRepository {
		return &fakeFarrowingRepository{
			sow:         testSow(sowID, sowdomain.StatePregnant),
			service:     testService(serviceID, sowID, servicedomain.StateConfirmed, testMount(serviceID, 1, mountDayOne)),
			operators:   map[uuid.UUID]*operatordomain.Operator{},
			medications: map[uuid.UUID]*medicationdomain.Medication{},
		}
	}

	t.Run("registers the farrowing and moves the sow and service", func(t *testing.T) {
		repository := repositoryWith()
		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })

		farrowing, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if farrowing.SowID != sowID || farrowing.ServiceID != serviceID {
			t.Fatalf("unexpected references: %#v", farrowing)
		}
		if farrowing.LiveBorn != 10 || farrowing.Stillborn != 1 {
			t.Fatalf("unexpected counts: %#v", farrowing)
		}
		if farrowing.CurrentPiglets != 10 {
			t.Fatalf("current piglets = %d, want 10", farrowing.CurrentPiglets)
		}
		if farrowing.CreatedBy != actorID || !farrowing.CreatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit: %#v", farrowing)
		}
		if repository.created != farrowing {
			t.Fatalf("farrowing was not persisted")
		}
		if repository.createdSow.State != sowdomain.StateLactating {
			t.Fatalf("sow state = %v, want %v", repository.createdSow.State, sowdomain.StateLactating)
		}
		if repository.createdService.State != servicedomain.StateFinished {
			t.Fatalf("service state = %v, want %v", repository.createdService.State, servicedomain.StateFinished)
		}
	})

	t.Run("resolves operators, medications and its applying operator", func(t *testing.T) {
		repository := repositoryWith()
		operatorID := uuid.New()
		medicationID := uuid.New()
		repository.operators[operatorID] = testOperator(operatorID)
		repository.medications[medicationID] = testMedication(medicationID)

		command := validCommand(sowID, actorID)
		command.Operators = []farrowingapplication.CreateFarrowingOperatorInput{{OperatorID: operatorID}}
		command.Medications = []farrowingapplication.CreateFarrowingMedicationInput{
			{MedicationID: medicationID, Dose: 2, AppliedBy: operatorID},
		}

		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })
		farrowing, err := useCase.Execute(context.Background(), command)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(farrowing.Operators) != 1 || farrowing.Operators[0].OperatorID != operatorID {
			t.Fatalf("unexpected operators: %#v", farrowing.Operators)
		}
		if len(farrowing.Medications) != 1 || farrowing.Medications[0].MedicationID != medicationID {
			t.Fatalf("unexpected medications: %#v", farrowing.Medications)
		}
	})

	t.Run("rejects a missing operator", func(t *testing.T) {
		repository := repositoryWith()
		command := validCommand(sowID, actorID)
		command.Operators = []farrowingapplication.CreateFarrowingOperatorInput{{OperatorID: uuid.New()}}

		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })
		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, ports.ErrOperatorNotFound) {
			t.Fatalf("error = %v, want ErrOperatorNotFound", err)
		}
	})

	t.Run("rejects a missing medication", func(t *testing.T) {
		repository := repositoryWith()
		command := validCommand(sowID, actorID)
		command.Medications = []farrowingapplication.CreateFarrowingMedicationInput{
			{MedicationID: uuid.New(), Dose: 1, AppliedBy: uuid.New()},
		}

		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })
		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, ports.ErrMedicationNotFound) {
			t.Fatalf("error = %v, want ErrMedicationNotFound", err)
		}
	})

	t.Run("uses the latest mount date for the rule", func(t *testing.T) {
		repository := repositoryWith()
		repository.service.Mounts = []*servicedomain.Mount{
			testMount(serviceID, 1, mountDayOne),
			testMount(serviceID, 2, mountDayTwo),
		}
		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })

		command := validCommand(sowID, actorID)
		command.FarrowDate = mountDayTwo

		_, err := useCase.Execute(context.Background(), command)

		if !errors.Is(err, farrowingdomain.ErrFarrowDateBeforeMount) {
			t.Fatalf("error = %v, want ErrFarrowDateBeforeMount", err)
		}
	})

	t.Run("propagates a missing sow", func(t *testing.T) {
		repository := &fakeFarrowingRepository{sowErr: ports.ErrSowNotFound}
		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrSowNotFound) {
			t.Fatalf("error = %v, want ErrSowNotFound", err)
		}
	})

	t.Run("rejects a sow that is not gestating", func(t *testing.T) {
		repository := repositoryWith()
		repository.sow = testSow(sowID, sowdomain.StateWeaned)
		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, farrowingapplication.ErrSowNotGestating) {
			t.Fatalf("error = %v, want ErrSowNotGestating", err)
		}
	})

	t.Run("propagates a missing last service", func(t *testing.T) {
		repository := repositoryWith()
		repository.serviceErr = ports.ErrServiceNotFound
		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, ports.ErrServiceNotFound) {
			t.Fatalf("error = %v, want ErrServiceNotFound", err)
		}
	})

	t.Run("rejects a last service that is not confirmed", func(t *testing.T) {
		repository := repositoryWith()
		repository.service.State = servicedomain.StateFailed
		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, farrowingapplication.ErrServiceNotConfirmed) {
			t.Fatalf("error = %v, want ErrServiceNotConfirmed", err)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := repositoryWith()
		repository.createErr = errors.New("boom")
		useCase := farrowingapplication.NewCreateFarrowingService(repository, func() time.Time { return nowReference })

		_, err := useCase.Execute(context.Background(), validCommand(sowID, actorID))

		if !errors.Is(err, repository.createErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
