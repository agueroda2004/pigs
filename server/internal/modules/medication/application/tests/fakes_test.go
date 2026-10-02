package tests

import (
	"context"

	"github.com/google/uuid"

	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
)

type fakeMedicationRepository struct {
	exists            bool
	existsErr         error
	getMedication     *medicationdomain.Medication
	getErr            error
	listMedications   []*medicationdomain.Medication
	listErr           error
	listFilter        ports.MedicationFilter
	listOptions       []medicationdomain.MedicationOption
	listOptionsErr    error
	listOptionsActive *bool
	createErr         error
	updateErr         error
	created           *medicationdomain.Medication
	updated           *medicationdomain.Medication
}

func (f *fakeMedicationRepository) Create(_ context.Context, medication *medicationdomain.Medication) error {
	f.created = medication
	return f.createErr
}

func (f *fakeMedicationRepository) GetByID(_ context.Context, _ uuid.UUID) (*medicationdomain.Medication, error) {
	return f.getMedication, f.getErr
}

func (f *fakeMedicationRepository) ExistsByName(_ context.Context, _ string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeMedicationRepository) List(_ context.Context, filter ports.MedicationFilter) ([]*medicationdomain.Medication, error) {
	f.listFilter = filter
	return f.listMedications, f.listErr
}

func (f *fakeMedicationRepository) ListOptions(_ context.Context, active *bool) ([]medicationdomain.MedicationOption, error) {
	f.listOptionsActive = active
	return f.listOptions, f.listOptionsErr
}

func (f *fakeMedicationRepository) Update(_ context.Context, medication *medicationdomain.Medication) error {
	f.updated = medication
	return f.updateErr
}
