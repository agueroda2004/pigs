package tests

import (
	"context"
	"time"

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

	farrowing       *farrowingdomain.Farrowing
	farrowingErr    error
	serviceByID     *servicedomain.Service
	serviceByIDErr  error
	earliestDate    *time.Time
	earliestDateErr error

	operators     map[uuid.UUID]*operatordomain.Operator
	operatorErr   error
	medications   map[uuid.UUID]*medicationdomain.Medication
	medicationErr error

	createErr      error
	created        *farrowingdomain.Farrowing
	createdSow     *sowdomain.Sow
	createdService *servicedomain.Service

	updateErr           error
	updated             *farrowingdomain.Farrowing
	replacedOperators   bool
	replacedMedications bool

	deleteErr      error
	deleted        *farrowingdomain.Farrowing
	deletedSow     *sowdomain.Sow
	deletedService *servicedomain.Service

	listFarrowings []*farrowingdomain.Farrowing
	listTotal      int
	listErr        error
	listFilter     ports.FarrowingFilter
	listLimit      int
	listOffset     int

	historyItems []*farrowingdomain.FarrowingHistoryBySow
	historyErr   error
	historyCode  string
}

func (f *fakeFarrowingRepository) GetSow(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.sow, f.sowErr
}

func (f *fakeFarrowingRepository) GetLastService(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.service, f.serviceErr
}

func (f *fakeFarrowingRepository) GetByID(_ context.Context, _ uuid.UUID) (*farrowingdomain.Farrowing, error) {
	return f.farrowing, f.farrowingErr
}

func (f *fakeFarrowingRepository) GetService(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.serviceByID, f.serviceByIDErr
}

func (f *fakeFarrowingRepository) GetEarliestRelatedDate(_ context.Context, _ uuid.UUID) (*time.Time, error) {
	return f.earliestDate, f.earliestDateErr
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

func (f *fakeFarrowingRepository) Update(_ context.Context, farrowing *farrowingdomain.Farrowing, replaceOperators, replaceMedications bool) error {
	f.updated = farrowing
	f.replacedOperators = replaceOperators
	f.replacedMedications = replaceMedications
	return f.updateErr
}

func (f *fakeFarrowingRepository) Delete(_ context.Context, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow, service *servicedomain.Service) error {
	f.deleted = farrowing
	f.deletedSow = sow
	f.deletedService = service
	return f.deleteErr
}

func (f *fakeFarrowingRepository) List(_ context.Context, filter ports.FarrowingFilter, limit, offset int) ([]*farrowingdomain.Farrowing, int, error) {
	f.listFilter = filter
	f.listLimit = limit
	f.listOffset = offset
	total := f.listTotal
	if total == 0 {
		total = len(f.listFarrowings)
	}
	return f.listFarrowings, total, f.listErr
}

func (f *fakeFarrowingRepository) HistoryBySow(_ context.Context, sowCode string) ([]*farrowingdomain.FarrowingHistoryBySow, error) {
	f.historyCode = sowCode
	return f.historyItems, f.historyErr
}
