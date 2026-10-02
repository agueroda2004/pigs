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

func TestCreateMedicationServiceExecute(t *testing.T) {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	createdBy := uuid.New()

	t.Run("creates a medication with a normalized name", func(t *testing.T) {
		repository := &fakeMedicationRepository{}
		service := medicationapplication.NewCreateMedicationService(repository, func() time.Time { return createdAt })

		medication, err := service.Execute(context.Background(), medicationapplication.CreateMedicationCommand{
			Name:      "  Ivermectina  ",
			CreatedBy: createdBy,
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if medication.ID == uuid.Nil || medication.Name != "Ivermectina" || !medication.Active {
			t.Fatalf("unexpected medication: %#v", medication)
		}
		if medication.CreatedBy != createdBy || medication.UpdatedBy != createdBy || !medication.CreatedAt.Equal(createdAt) {
			t.Fatalf("unexpected audit fields: %#v", medication)
		}
		if repository.created != medication {
			t.Fatalf("repository did not receive the created medication")
		}
	})

	t.Run("returns duplicate name error", func(t *testing.T) {
		repository := &fakeMedicationRepository{exists: true}
		service := medicationapplication.NewCreateMedicationService(repository, time.Now)

		_, err := service.Execute(context.Background(), medicationapplication.CreateMedicationCommand{Name: "Ivermectina", CreatedBy: createdBy})

		if !errors.Is(err, ports.ErrMedicationNameAlreadyUsed) || repository.created != nil {
			t.Fatalf("unexpected result: err=%v created=%v", err, repository.created)
		}
	})

	t.Run("propagates repository existence error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeMedicationRepository{existsErr: expected}
		service := medicationapplication.NewCreateMedicationService(repository, time.Now)

		_, err := service.Execute(context.Background(), medicationapplication.CreateMedicationCommand{Name: "Ivermectina", CreatedBy: createdBy})

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})

	t.Run("propagates validation and create errors", func(t *testing.T) {
		repository := &fakeMedicationRepository{createErr: errors.New("create failed")}
		service := medicationapplication.NewCreateMedicationService(repository, time.Now)

		_, err := service.Execute(context.Background(), medicationapplication.CreateMedicationCommand{Name: "  ", CreatedBy: createdBy})
		if !errors.Is(err, medicationdomain.ErrInvalidName) || repository.created != nil {
			t.Fatalf("unexpected validation result: %v", err)
		}

		_, err = service.Execute(context.Background(), medicationapplication.CreateMedicationCommand{Name: "Ivermectina", CreatedBy: createdBy})
		if !errors.Is(err, repository.createErr) {
			t.Fatalf("unexpected create result: %v", err)
		}
	})
}
