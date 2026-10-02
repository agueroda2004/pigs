package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	medicationdomain "server/internal/modules/medication/domain"
)

func TestNewMedication(t *testing.T) {
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	medicationID := uuid.New()
	createdBy := uuid.New()

	t.Run("creates a valid active medication", func(t *testing.T) {
		medication, err := medicationdomain.NewMedication(medicationID, "  Ivermectina  ", createdBy, now)

		if err != nil {
			t.Fatalf("NewMedication() error = %v", err)
		}
		if medication.ID != medicationID || medication.Name != "Ivermectina" || !medication.Active {
			t.Fatalf("unexpected medication: %#v", medication)
		}
		if medication.CreatedBy != createdBy || medication.UpdatedBy != createdBy {
			t.Fatalf("unexpected audit fields: %#v", medication)
		}
		if !medication.CreatedAt.Equal(now) || !medication.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected timestamps: %#v", medication)
		}
	})

	t.Run("rejects nil id", func(t *testing.T) {
		_, err := medicationdomain.NewMedication(uuid.Nil, "Ivermectina", createdBy, now)
		if !errors.Is(err, medicationdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})

	t.Run("rejects invalid name", func(t *testing.T) {
		_, err := medicationdomain.NewMedication(medicationID, "  ", createdBy, now)
		if !errors.Is(err, medicationdomain.ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}

		_, err = medicationdomain.NewMedication(medicationID, strings.Repeat("a", 51), createdBy, now)
		if !errors.Is(err, medicationdomain.ErrInvalidName) {
			t.Fatalf("error = %v, want ErrInvalidName", err)
		}
	})

	t.Run("accepts name at exactly max length", func(t *testing.T) {
		name := strings.Repeat("a", 50)
		medication, err := medicationdomain.NewMedication(medicationID, name, createdBy, now)
		if err != nil || medication.Name != name {
			t.Fatalf("unexpected result: err=%v medication=%#v", err, medication)
		}
	})

	t.Run("rejects nil created by", func(t *testing.T) {
		_, err := medicationdomain.NewMedication(medicationID, "Ivermectina", uuid.Nil, now)
		if !errors.Is(err, medicationdomain.ErrInvalidCreatedBy) {
			t.Fatalf("error = %v, want ErrInvalidCreatedBy", err)
		}
	})
}

func TestMedicationUpdate(t *testing.T) {
	medicationID := uuid.New()
	updatedBy := uuid.New()
	now := time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)

	t.Run("updates name and active", func(t *testing.T) {
		medication := &medicationdomain.Medication{ID: medicationID, Name: "Old", Active: true}
		name := "  New Name  "
		active := false

		err := medication.Update(&name, &active, updatedBy, now)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if medication.Name != "New Name" || medication.Active || medication.UpdatedBy != updatedBy || !medication.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected medication: %#v", medication)
		}
	})

	t.Run("keeps omitted fields unchanged", func(t *testing.T) {
		medication := &medicationdomain.Medication{ID: medicationID, Name: "Old", Active: true}
		name := "Updated"

		err := medication.Update(&name, nil, updatedBy, now)

		if err != nil || medication.Name != name || !medication.Active {
			t.Fatalf("unexpected result: err=%v medication=%#v", err, medication)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		medication := &medicationdomain.Medication{ID: medicationID, Name: "Old", Active: true}

		err := medication.Update(nil, nil, updatedBy, now)

		if !errors.Is(err, medicationdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects nil updated by", func(t *testing.T) {
		medication := &medicationdomain.Medication{ID: medicationID, Name: "Old", Active: true}
		name := "New Name"

		err := medication.Update(&name, nil, uuid.Nil, now)

		if !errors.Is(err, medicationdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects invalid name without mutating the medication", func(t *testing.T) {
		medication := &medicationdomain.Medication{ID: medicationID, Name: "Old", Active: true}
		name := "  "
		active := false

		err := medication.Update(&name, &active, updatedBy, now)

		if !errors.Is(err, medicationdomain.ErrInvalidName) || medication.Name != "Old" || !medication.Active {
			t.Fatalf("unexpected result: err=%v medication=%#v", err, medication)
		}
	})

	t.Run("rejects nil receiver", func(t *testing.T) {
		var medication *medicationdomain.Medication
		name := "New Name"

		err := medication.Update(&name, nil, updatedBy, now)

		if !errors.Is(err, medicationdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}
