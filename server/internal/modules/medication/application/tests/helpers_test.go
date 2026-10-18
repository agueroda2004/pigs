package tests

import (
	"time"

	"github.com/google/uuid"

	medicationdomain "server/internal/modules/medication/domain"
)

func testMedication(id uuid.UUID) *medicationdomain.Medication {
	return &medicationdomain.Medication{
		ID: id, Name: "Old Name", Active: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

func testMedicationListItem(id uuid.UUID) medicationdomain.MedicationListItem {
	return medicationdomain.MedicationListItem{ID: id, Name: "Old Name", Active: true}
}
