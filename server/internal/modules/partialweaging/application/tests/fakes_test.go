package tests

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
	sowdomain "server/internal/modules/sow/domain"
)

type fakePartialWeagingRepository struct {
	sows       map[uuid.UUID]*sowdomain.Sow
	farrowings map[uuid.UUID]*farrowingdomain.Farrowing
	sowErr     error
	farrowErr  error

	lastEventDate time.Time
	lastEventErr  error

	lastPartialWeagingDate time.Time
	lastPartialWeagingErr  error

	updateErr error
	updated   *partialweagingdomain.PartialWeaging

	createErr        error
	created          *partialweagingdomain.PartialWeaging
	createdFarrowing *farrowingdomain.Farrowing
	createdSow       *sowdomain.Sow

	listWeagings []*partialweagingdomain.PartialWeaging
	listTotal    int
	listErr      error
	listFilter   ports.PartialWeagingFilter
	listLimit    int
	listOffset   int

	weagingByID    *partialweagingdomain.PartialWeaging
	weagingByIDErr error

	farrowingByID    *farrowingdomain.Farrowing
	farrowingByIDErr error

	hasFutureEvents    bool
	hasFutureEventsErr error

	hasLaterNurseEvents    bool
	hasLaterNurseEventsErr error

	previousNurseDate    time.Time
	previousNurseDateErr error

	deleteErr        error
	deleted          *partialweagingdomain.PartialWeaging
	deletedFarrowing *farrowingdomain.Farrowing
	deletedSow       *sowdomain.Sow
}

func (f *fakePartialWeagingRepository) GetSow(_ context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
	if f.sowErr != nil {
		return nil, f.sowErr
	}
	return f.sows[id], nil
}

func (f *fakePartialWeagingRepository) GetLastFarrowing(_ context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error) {
	if f.farrowErr != nil {
		return nil, f.farrowErr
	}
	return f.farrowings[sowID], nil
}

func (f *fakePartialWeagingRepository) GetLatestEventDate(_ context.Context, _ uuid.UUID) (time.Time, error) {
	return f.lastEventDate, f.lastEventErr
}

func (f *fakePartialWeagingRepository) GetLastPartialWeagingDate(_ context.Context, _ uuid.UUID) (time.Time, error) {
	return f.lastPartialWeagingDate, f.lastPartialWeagingErr
}

func (f *fakePartialWeagingRepository) Update(
	_ context.Context,
	weaging *partialweagingdomain.PartialWeaging,
) error {
	f.updated = weaging
	return f.updateErr
}

func (f *fakePartialWeagingRepository) Create(
	_ context.Context,
	weaging *partialweagingdomain.PartialWeaging,
	farrowing *farrowingdomain.Farrowing,
	sow *sowdomain.Sow,
) error {
	f.created = weaging
	f.createdFarrowing = farrowing
	f.createdSow = sow
	return f.createErr
}

func (f *fakePartialWeagingRepository) List(
	_ context.Context,
	filter ports.PartialWeagingFilter,
	limit int,
	offset int,
) ([]*partialweagingdomain.PartialWeaging, int, error) {
	f.listFilter = filter
	f.listLimit = limit
	f.listOffset = offset
	return f.listWeagings, f.listTotal, f.listErr
}

func (f *fakePartialWeagingRepository) GetByID(_ context.Context, _ uuid.UUID) (*partialweagingdomain.PartialWeaging, error) {
	return f.weagingByID, f.weagingByIDErr
}

func (f *fakePartialWeagingRepository) GetFarrowingByID(_ context.Context, _ uuid.UUID) (*farrowingdomain.Farrowing, error) {
	return f.farrowingByID, f.farrowingByIDErr
}

func (f *fakePartialWeagingRepository) HasFutureEvents(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return f.hasFutureEvents, f.hasFutureEventsErr
}

func (f *fakePartialWeagingRepository) HasLaterNurseEvents(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return f.hasLaterNurseEvents, f.hasLaterNurseEventsErr
}

func (f *fakePartialWeagingRepository) GetPreviousNurseDate(_ context.Context, _ uuid.UUID, _ time.Time) (time.Time, error) {
	return f.previousNurseDate, f.previousNurseDateErr
}

func (f *fakePartialWeagingRepository) Delete(
	_ context.Context,
	weaging *partialweagingdomain.PartialWeaging,
	farrowing *farrowingdomain.Farrowing,
	sow *sowdomain.Sow,
) error {
	f.deleted = weaging
	f.deletedFarrowing = farrowing
	f.deletedSow = sow
	return f.deleteErr
}
