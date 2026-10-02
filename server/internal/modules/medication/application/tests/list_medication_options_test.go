package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	medicationapplication "server/internal/modules/medication/application"
	medicationdomain "server/internal/modules/medication/domain"
)

func TestListMedicationOptionsServiceExecute(t *testing.T) {
	t.Run("returns the active medication options when active is true", func(t *testing.T) {
		options := []medicationdomain.MedicationOption{
			{ID: uuid.New(), Name: "Ivermectina"},
			{ID: uuid.New(), Name: "Penicilina"},
		}
		repository := &fakeMedicationRepository{listOptions: options}
		service := medicationapplication.NewListMedicationOptionsService(repository)
		active := true

		result, err := service.Execute(context.Background(), &active)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != len(options) || result[0] != options[0] || result[1] != options[1] {
			t.Fatalf("unexpected options: %#v", result)
		}
		if repository.listOptionsActive == nil || *repository.listOptionsActive != active {
			t.Fatalf("unexpected active filter: %#v", repository.listOptionsActive)
		}
	})

	t.Run("forwards a nil active filter to list every medication", func(t *testing.T) {
		repository := &fakeMedicationRepository{listOptions: []medicationdomain.MedicationOption{}}
		service := medicationapplication.NewListMedicationOptionsService(repository)

		result, err := service.Execute(context.Background(), nil)

		if err != nil || len(result) != 0 {
			t.Fatalf("unexpected result: options=%#v err=%v", result, err)
		}
		if repository.listOptionsActive != nil {
			t.Fatalf("unexpected active filter: %#v", repository.listOptionsActive)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expected := errors.New("database unavailable")
		repository := &fakeMedicationRepository{listOptionsErr: expected}
		service := medicationapplication.NewListMedicationOptionsService(repository)

		_, err := service.Execute(context.Background(), nil)

		if !errors.Is(err, expected) {
			t.Fatalf("Execute() error = %v, want %v", err, expected)
		}
	})
}
