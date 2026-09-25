package tests

import (
	"context"

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

	listServices []*servicedomain.Service
	listErr      error
	listFilter   ports.ServiceFilter
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

func (f *fakeServiceRepository) Create(_ context.Context, service *servicedomain.Service, sow *sowdomain.Sow) error {
	f.created = service
	f.createdSow = sow
	return f.createErr
}

func (f *fakeServiceRepository) List(_ context.Context, filter ports.ServiceFilter) ([]*servicedomain.Service, error) {
	f.listFilter = filter
	return f.listServices, f.listErr
}
