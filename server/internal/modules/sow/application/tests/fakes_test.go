package tests

import (
	"context"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

type fakeSowRepository struct {
	exists         bool
	existsErr      error
	getSow         *sowdomain.Sow
	getErr         error
	listSows       []*sowdomain.Sow
	listErr        error
	listFilter     ports.SowFilter
	listOptions    []sowdomain.SowOption
	listOptionsErr error
	listActive     *bool
	createErr      error
	updateErr      error
	updateStateErr error
	created        *sowdomain.Sow
	updated        *sowdomain.Sow
	updatedState   *sowdomain.Sow
}

func (f *fakeSowRepository) Create(_ context.Context, sow *sowdomain.Sow) error {
	f.created = sow
	return f.createErr
}

func (f *fakeSowRepository) GetByID(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.getSow, f.getErr
}

func (f *fakeSowRepository) ExistsByCode(_ context.Context, _ string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeSowRepository) List(_ context.Context, filter ports.SowFilter) ([]*sowdomain.Sow, error) {
	f.listFilter = filter
	return f.listSows, f.listErr
}

func (f *fakeSowRepository) ListOptions(_ context.Context, active *bool) ([]sowdomain.SowOption, error) {
	f.listActive = active
	return f.listOptions, f.listOptionsErr
}

func (f *fakeSowRepository) Update(_ context.Context, sow *sowdomain.Sow) error {
	f.updated = sow
	return f.updateErr
}

func (f *fakeSowRepository) UpdateState(_ context.Context, sow *sowdomain.Sow) error {
	f.updatedState = sow
	return f.updateStateErr
}
