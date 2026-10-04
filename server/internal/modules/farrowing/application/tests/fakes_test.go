package tests

import (
	"context"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
	medicationdomain "server/internal/modules/medication/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

type fakeFarrowingRepository struct {
	sow        *sowdomain.Sow
	sowErr     error
	service    *servicedomain.Service
	serviceErr error

	operators     map[uuid.UUID]*operatordomain.Operator
	operatorErr   error
	medications   map[uuid.UUID]*medicationdomain.Medication
	medicationErr error

	createErr      error
	created        *farrowingdomain.Farrowing
	createdSow     *sowdomain.Sow
	createdService *servicedomain.Service

	listFarrowings []*farrowingdomain.Farrowing
	listErr        error
	listFilter     ports.FarrowingFilter
}

func (f *fakeFarrowingRepository) GetSow(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.sow, f.sowErr
}

func (f *fakeFarrowingRepository) GetLastService(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.service, f.serviceErr
}

func (f *fakeFarrowingRepository) GetOperator(_ context.Context, id uuid.UUID) (*operatordomain.Operator, error) {
	if f.operatorErr != nil {
		return nil, f.operatorErr
	}
	if operator, ok := f.operators[id]; ok {
		return operator, nil
	}
	return nil, ports.ErrOperatorNotFound
}

func (f *fakeFarrowingRepository) GetMedication(_ context.Context, id uuid.UUID) (*medicationdomain.Medication, error) {
	if f.medicationErr != nil {
		return nil, f.medicationErr
	}
	if medication, ok := f.medications[id]; ok {
		return medication, nil
	}
	return nil, ports.ErrMedicationNotFound
}

func (f *fakeFarrowingRepository) Create(
	_ context.Context,
	farrowing *farrowingdomain.Farrowing,
	sow *sowdomain.Sow,
	service *servicedomain.Service,
) error {
	f.created = farrowing
	f.createdSow = sow
	f.createdService = service
	return f.createErr
}

func (f *fakeFarrowingRepository) List(_ context.Context, filter ports.FarrowingFilter) ([]*farrowingdomain.Farrowing, error) {
	f.listFilter = filter
	return f.listFarrowings, f.listErr
}
