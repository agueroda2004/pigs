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
)

func updateRepository(farrowing *farrowingdomain.Farrowing) *fakeFarrowingRepository {
	return &fakeFarrowingRepository{farrowing: farrowing}
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

	t.Run("rejects changes to the immovable fields", func(t *testing.T) {
		for _, test := range []struct {
			name    string
			command func(*farrowingapplication.UpdateFarrowingCommand)
		}{
			{"farrow date", func(c *farrowingapplication.UpdateFarrowingCommand) {
				value := farrowDay.AddDate(0, 0, 1)
				c.FarrowDate = &value
			}},
			{"live born", func(c *farrowingapplication.UpdateFarrowingCommand) {
				value := 11
				c.LiveBorn = &value
			}},
			{"stillborn", func(c *farrowingapplication.UpdateFarrowingCommand) {
				value := 2
				c.Stillborn = &value
			}},
			{"mummified", func(c *farrowingapplication.UpdateFarrowingCommand) {
				value := 1
				c.Mummified = &value
			}},
		} {
			t.Run(test.name, func(t *testing.T) {
				farrowing := testFarrowing()
				repository := updateRepository(farrowing)
				useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)
				command := farrowingapplication.UpdateFarrowingCommand{UpdatedBy: actor}
				test.command(&command)

				err := useCase.Execute(context.Background(), farrowing.ID, command)

				if !errors.Is(err, farrowingdomain.ErrImmutableFarrowField) {
					t.Fatalf("error = %v, want ErrImmutableFarrowField", err)
				}
				if repository.updated != nil {
					t.Fatalf("the farrowing must not be persisted")
				}
			})
		}
	})

	t.Run("rejects weights greater than 30", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)
		weight := 30.5

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			LitterWeight: &weight,
			UpdatedBy:    actor,
		})

		if !errors.Is(err, farrowingdomain.ErrLitterWeightTooHigh) {
			t.Fatalf("error = %v, want ErrLitterWeightTooHigh", err)
		}
		if repository.updated != nil {
			t.Fatalf("the farrowing must not be persisted")
		}
	})

	t.Run("rejects a medication dose greater than 30", func(t *testing.T) {
		farrowing := testFarrowing()
		repository := updateRepository(farrowing)
		operatorID := uuid.New()
		medicationID := uuid.New()
		repository.operators = map[uuid.UUID]*operatordomain.Operator{
			operatorID: testOperator(operatorID),
		}
		repository.medications = map[uuid.UUID]*medicationdomain.Medication{
			medicationID: testMedication(medicationID),
		}
		useCase := farrowingapplication.NewUpdateFarrowingService(repository, clock)

		err := useCase.Execute(context.Background(), farrowing.ID, farrowingapplication.UpdateFarrowingCommand{
			UpdateMedications: true,
			Medications: []farrowingapplication.CreateFarrowingMedicationInput{
				{MedicationID: medicationID, Dose: 31, AppliedBy: operatorID},
			},
			UpdatedBy: actor,
		})

		if !errors.Is(err, farrowingdomain.ErrMedicationDoseTooHigh) {
			t.Fatalf("error = %v, want ErrMedicationDoseTooHigh", err)
		}
		if repository.updated != nil {
			t.Fatalf("the farrowing must not be persisted")
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
