package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	medicationapplication "server/internal/modules/medication/application"
	"server/internal/modules/medication/ports"
)

func TestDeleteMedicationServiceExecute(t *testing.T) {
	medicationID := uuid.New()

	t.Run("deletes the medication by its identifier", func(t *testing.T) {
		repository := &fakeMedicationRepository{}
		service := medicationapplication.NewDeleteMedicationService(repository)

		err := service.Execute(context.Background(), medicationID)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !repository.deleteCalled || repository.deletedID != medicationID {
			t.Fatalf("repository did not receive the medication id: %#v", repository)
		}
	})

	t.Run("returns not found when the medication does not exist", func(t *testing.T) {
		repository := &fakeMedicationRepository{deleteErr: ports.ErrMedicationNotFound}
		service := medicationapplication.NewDeleteMedicationService(repository)

		err := service.Execute(context.Background(), medicationID)

		if !errors.Is(err, ports.ErrMedicationNotFound) {
			t.Fatalf("Execute() error = %v, want ErrMedicationNotFound", err)
		}
	})

	t.Run("returns in use when the medication has related records", func(t *testing.T) {
		repository := &fakeMedicationRepository{deleteErr: ports.ErrMedicationInUse}
		service := medicationapplication.NewDeleteMedicationService(repository)

		err := service.Execute(context.Background(), medicationID)

		if !errors.Is(err, ports.ErrMedicationInUse) {
			t.Fatalf("Execute() error = %v, want ErrMedicationInUse", err)
		}
	})
}
