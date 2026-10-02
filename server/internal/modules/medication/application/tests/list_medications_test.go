package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	medicationapplication "server/internal/modules/medication/application"
	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
)

func TestListMedicationsServiceExecute(t *testing.T) {
	t.Run("returns the medications matching the filter", func(t *testing.T) {
		medications := []*medicationdomain.Medication{
			testMedication(uuid.New()),
			testMedication(uuid.New()),
		}
		name := "Iver"
		active := true
		repository := &fakeMedicationRepository{listMedications: medications}
		service := medicationapplication.NewListMedicationsService(repository)

		result, err := service.Execute(context.Background(), ports.MedicationFilter{Name: &name, Active: &active})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(medications) || result[0] != medications[0] || result[1] != medications[1] {
			t.Fatalf("unexpected medications: %#v", result)
		}
		if repository.listFilter.Name == nil || *repository.listFilter.Name != name {
			t.Fatalf("unexpected name filter: %#v", repository.listFilter)
		}
		if repository.listFilter.Active == nil || *repository.listFilter.Active != active {
			t.Fatalf("unexpected active filter: %#v", repository.listFilter)
		}
	})

	t.Run("forwards a zero-value filter", func(t *testing.T) {
		repository := &fakeMedicationRepository{listMedications: []*medicationdomain.Medication{}}
		service := medicationapplication.NewListMedicationsService(repository)

		result, err := service.Execute(context.Background(), ports.MedicationFilter{})

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: medications=%#v err=%v", result, err)
		}
		if repository.listFilter.Name != nil || repository.listFilter.Active != nil {
			t.Fatalf("unexpected filter: %#v", repository.listFilter)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeMedicationRepository{listErr: expected}
		service := medicationapplication.NewListMedicationsService(repository)

		_, err := service.Execute(context.Background(), ports.MedicationFilter{})

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
