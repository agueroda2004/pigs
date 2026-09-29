package tests

import (
	"context"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
)

type fakeBoarRemovalRepository struct {
	removalByID    *boarremovaldomain.BoarRemoval
	removalByIDErr error

	boar    *boardomain.Boar
	boarErr error

	lastMount    time.Time
	lastMountErr error

	createErr   error
	created     *boarremovaldomain.BoarRemoval
	createdBoar *boardomain.Boar

	updateErr   error
	updated     *boarremovaldomain.BoarRemoval
	updatedBoar *boardomain.Boar

	deleteErr   error
	deleted     *boarremovaldomain.BoarRemoval
	deletedBoar *boardomain.Boar

	listRemovals []*boarremovaldomain.BoarRemoval
	listErr      error
	listFilter   ports.BoarRemovalFilter
}

func (f *fakeBoarRemovalRepository) GetByID(_ context.Context, _ uuid.UUID) (*boarremovaldomain.BoarRemoval, error) {
	return f.removalByID, f.removalByIDErr
}

func (f *fakeBoarRemovalRepository) GetBoar(_ context.Context, _ uuid.UUID) (*boardomain.Boar, error) {
	return f.boar, f.boarErr
}

func (f *fakeBoarRemovalRepository) GetLastMountDate(_ context.Context, _ uuid.UUID) (time.Time, error) {
	return f.lastMount, f.lastMountErr
}

func (f *fakeBoarRemovalRepository) Create(_ context.Context, removal *boarremovaldomain.BoarRemoval, boar *boardomain.Boar) error {
	f.created = removal
	f.createdBoar = boar
	return f.createErr
}

func (f *fakeBoarRemovalRepository) Update(_ context.Context, removal *boarremovaldomain.BoarRemoval, boar *boardomain.Boar) error {
	f.updated = removal
	f.updatedBoar = boar
	return f.updateErr
}

func (f *fakeBoarRemovalRepository) Delete(_ context.Context, removal *boarremovaldomain.BoarRemoval, boar *boardomain.Boar) error {
	f.deleted = removal
	f.deletedBoar = boar
	return f.deleteErr
}

func (f *fakeBoarRemovalRepository) List(_ context.Context, filter ports.BoarRemovalFilter) ([]*boarremovaldomain.BoarRemoval, error) {
	f.listFilter = filter
	return f.listRemovals, f.listErr
}
