package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	medicationapplication "server/internal/modules/medication/application"
	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
)

func TestUpdateMedicationServiceExecute(t *testing.T) {
	medicationID := uuid.New()
	updatedBy := uuid.New()
	now := time.Date(2026, time.March, 4, 5, 6, 7, 0, time.UTC)

	t.Run("updates name and active", func(t *testing.T) {
		medication := testMedication(medicationID)
		name := "  New Name  "
		active := false
		repository := &fakeMedicationRepository{getMedication: medication}
		service := medicationapplication.NewUpdateMedicationService(repository, func() time.Time { return now })

		updated, err := service.Execute(context.Background(), medicationID, medicationapplication.UpdateMedicationCommand{
			Name: &name, Active: &active, UpdatedBy: updatedBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if updated != medication || medication.Name != "New Name" || medication.Active || medication.UpdatedBy != updatedBy || !medication.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected medication: %#v", medication)
		}
		if repository.updated != medication {
			t.Fatalf("repository did not receive the updated medication")
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		medication := testMedication(medicationID)
		active := false
		repository := &fakeMedicationRepository{getMedication: medication}
		service := medicationapplication.NewUpdateMedicationService(repository, time.Now)

		_, err := service.Execute(context.Background(), medicationID, medicationapplication.UpdateMedicationCommand{Active: &active, UpdatedBy: updatedBy})

		if err != nil || medication.Active || medication.Name != "Old Name" {
			t.Fatalf("unexpected result: err=%v medication=%#v", err, medication)
		}
	})

	t.Run("propagates dependency and validation errors", func(t *testing.T) {
		getErr := errors.New("get failed")
		repository := &fakeMedicationRepository{getErr: getErr}
		service := medicationapplication.NewUpdateMedicationService(repository, time.Now)
		_, err := service.Execute(context.Background(), medicationID, medicationapplication.UpdateMedicationCommand{UpdatedBy: updatedBy})
		if !errors.Is(err, getErr) {
			t.Fatalf("get error = %v", err)
		}

		repository = &fakeMedicationRepository{getMedication: testMedication(medicationID)}
		service = medicationapplication.NewUpdateMedicationService(repository, time.Now)
		_, err = service.Execute(context.Background(), medicationID, medicationapplication.UpdateMedicationCommand{UpdatedBy: updatedBy})
		if !errors.Is(err, medicationdomain.ErrInvalidUpdate) || repository.updated != nil {
			t.Fatalf("empty command error = %v", err)
		}

		name := "New Name"
		repository = &fakeMedicationRepository{getMedication: testMedication(medicationID), updateErr: ports.ErrMedicationNameAlreadyUsed}
		service = medicationapplication.NewUpdateMedicationService(repository, time.Now)
		_, err = service.Execute(context.Background(), medicationID, medicationapplication.UpdateMedicationCommand{Name: &name, UpdatedBy: updatedBy})
		if !errors.Is(err, ports.ErrMedicationNameAlreadyUsed) {
			t.Fatalf("repository error = %v", err)
		}
	})
}
