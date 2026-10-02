package ports

import (
	"context"

	"github.com/google/uuid"

	medicationdomain "server/internal/modules/medication/domain"
)

// MedicationFilter holds the optional criteria used to filter the medication list.
// A nil field means the criterion is not applied and every filter combines with AND.
type MedicationFilter struct {
	Name   *string
	Active *bool
}

type MedicationRepository interface {
	Create(ctx context.Context, medication *medicationdomain.Medication) error
	GetByID(ctx context.Context, id uuid.UUID) (*medicationdomain.Medication, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	List(ctx context.Context, filter MedicationFilter) ([]*medicationdomain.Medication, error)
	ListOptions(ctx context.Context, active *bool) ([]medicationdomain.MedicationOption, error)
	Update(ctx context.Context, medication *medicationdomain.Medication) error
}
