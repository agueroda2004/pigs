package tests

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	sowdomain "server/internal/modules/sow/domain"
)

type fakeServiceRepository struct {
	sow         *sowdomain.Sow
	sowErr      error
	boar        *boardomain.Boar
	boarErr     error
	operator    *operatordomain.Operator
	operatorErr error

	createErr  error
	created    *servicedomain.Service
	createdSow *sowdomain.Sow

	getService *servicedomain.Service
	getErr     error

	lastService        *servicedomain.Service
	lastServiceErr     error
	previousService    *servicedomain.Service
	previousServiceErr error
	lastAbortionDate   *time.Time
	lastAbortionErr    error

	updated   *servicedomain.Service
	updateErr error

	deleted    *servicedomain.Service
	deletedSow *sowdomain.Sow
	deleteErr  error

	listServices []*servicedomain.Service
	listTotal    int
	listErr      error
	listFilter   ports.ServiceFilter
	listLimit    int
	listOffset   int
}

func (f *fakeServiceRepository) GetSow(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.sow, f.sowErr
}

func (f *fakeServiceRepository) GetBoar(_ context.Context, _ uuid.UUID) (*boardomain.Boar, error) {
	return f.boar, f.boarErr
}

func (f *fakeServiceRepository) GetOperator(_ context.Context, _ uuid.UUID) (*operatordomain.Operator, error) {
	return f.operator, f.operatorErr
}

func (f *fakeServiceRepository) GetByID(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.getService, f.getErr
}

func (f *fakeServiceRepository) GetLastService(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.lastService, f.lastServiceErr
}

func (f *fakeServiceRepository) GetPreviousService(_ context.Context, _, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.previousService, f.previousServiceErr
}

func (f *fakeServiceRepository) GetLastAbortionDate(_ context.Context, _ uuid.UUID) (*time.Time, error) {
	return f.lastAbortionDate, f.lastAbortionErr
}

func (f *fakeServiceRepository) Update(_ context.Context, service *servicedomain.Service) error {
	f.updated = service
	return f.updateErr
}

func (f *fakeServiceRepository) Create(_ context.Context, service *servicedomain.Service, sow *sowdomain.Sow) error {
	f.created = service
	f.createdSow = sow
	return f.createErr
}

func (f *fakeServiceRepository) Delete(_ context.Context, service *servicedomain.Service, sow *sowdomain.Sow) error {
	f.deleted = service
	f.deletedSow = sow
	return f.deleteErr
}

func (f *fakeServiceRepository) List(_ context.Context, filter ports.ServiceFilter, limit, offset int) ([]*servicedomain.Service, int, error) {
	f.listFilter = filter
	f.listLimit = limit
	f.listOffset = offset
	return f.listServices, f.listTotal, f.listErr
}
