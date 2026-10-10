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
)

func updateRepository(farrowing *farrowingdomain.Farrowing) *fakeFarrowingRepository {
	return &fakeFarrowingRepository{
		farrowing: farrowing,
		serviceByID: testService(
			farrowing.ServiceID,
			farrowing.SowID,
			servicedomain.StateFinished,
			testMount(farrowing.ServiceID, 1, mountDayOne),
		),
	}
}

func testFarrowing() *farrowingdomain.Farrowing {
	return &farrowingdomain.Farrowing{
		ID:             uuid.New(),
		ServiceID:      uuid.New(),
		SowID:          uuid.New(),
		FarrowDate:     farrowDay,
		LiveBorn:       10,
		Stillborn:      1,
		CurrentPiglets: 10,
	}
}

func TestUpdateFarrowingExecute(t *testing.T) {
	actor := uuid.New()
	clock := func() time.Time { return nowReference }

	t.Run("updates the provided fields and persists the farrowing", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)
		location := "Corral 5"

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			Location:  &location,
			UpdatedBy: actor,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.updated != farrowing {
			t.Fatalf("repository was not updated")
		}
		if farrowing.Location == nil || *farrowing.Location != "Corral 5" {
			t.Fatalf("unexpected location: %#v", farrowing.Location)
		}
		if repository.replacedOperators || repository.replacedMedications {
			t.Fatalf("links must not be replaced: %v/%v", repository.replacedOperators, repository.replacedMedications)
		}
	})

	t.Run("replaces the operator and medication links", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		operatorID := uuid.New()
		medicationID := uuid.New()
		appliedBy := uuid.New()
		repository.operators = map[uuid.UUID]*operatordomain.Operator{
			operatorID: testOperator(operatorID),
			appliedBy:  testOperator(appliedBy),
		}
		repository.medications = map[uuid.UUID]*medicationdomain.Medication{
			medicationID: testMedication(medicationID),
		}
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			UpdateOperators:   true,
			Operators:         []farrowingapplication.CreateFarrowingOperatorInput{{OperatorID: operatorID}},
			UpdateMedications: true,
			Medications: []farrowingapplication.CreateFarrowingMedicationInput{
				{MedicationID: medicationID, Dose: 2, AppliedBy: appliedBy},
			},
			UpdatedBy: actor,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !repository.replacedOperators || !repository.replacedMedications {
			t.Fatalf("links must be replaced: %v/%v", repository.replacedOperators, repository.replacedMedications)
		}
		if len(farrowing.Operators) != 1 || len(farrowing.Medications) != 1 {
			t.Fatalf("unexpected links: %#v", farrowing)
		}
	})

	t.Run("rejects an unknown referenced operator", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			UpdateOperators: true,
			Operators:       []farrowingapplication.CreateFarrowingOperatorInput{{OperatorID: uuid.New()}},
			UpdatedBy:       actor,
		})

		if !errors.Is(err, ports.ErrOperatorNotFound) {
			t.Fatalf("error = %v, want ErrOperatorNotFound", err)
		}
	})

	t.Run("propagates a missing farrowing", func(t *testing.T) {
		repository := &fakeFarrowingRepository{farrowingErr: ports.ErrFarrowingNotFound}
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)
		location := "Corral"

		err := useCase.Execute(context.Background(), uuid.New(), farrowingapplication.UpdateFarrowingCommand{
			Location:  &location,
			UpdatedBy: actor,
		})

		if !errors.Is(err, ports.ErrFarrowingNotFound) {
			t.Fatalf("error = %v, want ErrFarrowingNotFound", err)
		}
	})

	t.Run("propagates a missing service", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := &fakeFarrowingRepository{farrowing: farrowing, serviceByIDErr: ports.ErrServiceNotFound}
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)
		location := "Corral"

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			Location:  &location,
			UpdatedBy: actor,
		})

		if !errors.Is(err, ports.ErrServiceNotFound) {
			t.Fatalf("error = %v, want ErrServiceNotFound", err)
		}
	})

	t.Run("propagates a related-records error from the repository", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		repository.earliestDateErr = errors.New("boom")
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)
		location := "Corral"

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			Location:  &location,
			UpdatedBy: actor,
		})

		if !errors.Is(err, repository.earliestDateErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})

	t.Run("locks the live born count when related records exist", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		earliest := time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)
		repository.earliestDate = &earliest
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)
		liveBorn := 12

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			LiveBorn:  &liveBorn,
			UpdatedBy: actor,
		})

		if !errors.Is(err, farrowingdomain.ErrLiveBornLocked) {
			t.Fatalf("error = %v, want ErrLiveBornLocked", err)
		}
	})

	t.Run("bounds the farrow date with the earliest related record", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		earliest := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
		repository.earliestDate = &earliest
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)
		newDate := time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC)

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			FarrowDate: &newDate,
			UpdatedBy:  actor,
		})

		if !errors.Is(err, farrowingdomain.ErrFarrowDateAfterRelatedRecords) {
			t.Fatalf("error = %v, want ErrFarrowDateAfterRelatedRecords", err)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			UpdatedBy: actor,
		})

		if !errors.Is(err, farrowingdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})
}
