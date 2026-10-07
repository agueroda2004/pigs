package tests

import (
	"context"
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

type fakeSowRepository struct {
	exists          bool
	existsErr       error
	getSow          *sowdomain.Sow
	getErr          error
	listSows        []*sowdomain.Sow
	listTotal       int
	listErr         error
	listFilter      ports.SowFilter
	listLimit       int
	listOffset      int
	listDropdown    []sowdomain.SowDropdown
	listDropdownErr error
	listActive      *bool
	listStates      []sowdomain.State
	lastServiceDate *time.Time
	lastServiceErr  error
	createErr       error
	updateErr       error
	updateStateErr  error
	created         *sowdomain.Sow
	updated         *sowdomain.Sow
	updatedState    *sowdomain.Sow
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

func (f *fakeSowRepository) List(_ context.Context, filter ports.SowFilter, limit, offset int) ([]*sowdomain.Sow, int, error) {
	f.listFilter = filter
	f.listLimit = limit
	f.listOffset = offset
	return f.listSows, f.listTotal, f.listErr
}

func (f *fakeSowRepository) ListDropdown(_ context.Context, active *bool, states []sowdomain.State) ([]sowdomain.SowDropdown, error) {
	f.listActive = active
	f.listStates = states
	return f.listDropdown, f.listDropdownErr
}

func (f *fakeSowRepository) LastServiceDate(_ context.Context, _ uuid.UUID) (*time.Time, error) {
	return f.lastServiceDate, f.lastServiceErr
}

func (f *fakeSowRepository) Update(_ context.Context, sow *sowdomain.Sow) error {
	f.updated = sow
	return f.updateErr
}

func (f *fakeSowRepository) UpdateState(_ context.Context, sow *sowdomain.Sow) error {
	f.updatedState = sow
	return f.updateStateErr
}
