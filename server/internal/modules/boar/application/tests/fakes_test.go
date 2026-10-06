package tests

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

type fakeBoarRepository struct {
	exists             bool
	existsErr          error
	getBoar            *boardomain.Boar
	getErr             error
	earliestEventDate  time.Time
	earliestEventErr   error
	listBoars          []*boardomain.Boar
	listErr            error
	listFilter         ports.BoarFilter
	listDropdown       []boardomain.BoarDropdown
	listDropdownErr    error
	listDropdownActive *bool
	listDropdownState  *boardomain.State
	createErr          error
	updateErr          error
	updateStateErr     error
	deleteErr          error
	deletedID          uuid.UUID
	created            *boardomain.Boar
	updated            *boardomain.Boar
	updatedState       *boardomain.Boar
}

func (f *fakeBoarRepository) Create(_ context.Context, boar *boardomain.Boar) error {
	f.created = boar
	return f.createErr
}

func (f *fakeBoarRepository) GetByID(_ context.Context, _ uuid.UUID) (*boardomain.Boar, error) {
	return f.getBoar, f.getErr
}

func (f *fakeBoarRepository) GetEarliestEventDate(_ context.Context, _ uuid.UUID) (time.Time, error) {
	return f.earliestEventDate, f.earliestEventErr
}

func (f *fakeBoarRepository) ExistsByCode(_ context.Context, _ string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeBoarRepository) List(_ context.Context, filter ports.BoarFilter) ([]*boardomain.Boar, error) {
	f.listFilter = filter
	return f.listBoars, f.listErr
}

func (f *fakeBoarRepository) ListDropdown(_ context.Context, active *bool, state *boardomain.State) ([]boardomain.BoarDropdown, error) {
	f.listDropdownActive = active
	f.listDropdownState = state
	return f.listDropdown, f.listDropdownErr
}

func (f *fakeBoarRepository) Update(_ context.Context, boar *boardomain.Boar) error {
	f.updated = boar
	return f.updateErr
}

func (f *fakeBoarRepository) UpdateState(_ context.Context, boar *boardomain.Boar) error {
	f.updatedState = boar
	return f.updateStateErr
}

func (f *fakeBoarRepository) Delete(_ context.Context, id uuid.UUID) error {
	f.deletedID = id
	return f.deleteErr
}
