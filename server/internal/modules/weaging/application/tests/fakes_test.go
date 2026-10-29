package tests

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	sowdomain "server/internal/modules/sow/domain"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

type fakeWeagingRepository struct {
	sows       map[uuid.UUID]*sowdomain.Sow
	farrowings map[uuid.UUID]*farrowingdomain.Farrowing
	sowErr     error
	farrowErr  error

	farrowingsByID    map[uuid.UUID]*farrowingdomain.Farrowing
	farrowingByIDErr  error
	weagings          map[uuid.UUID]*weagingdomain.Weaging
	weagingByIDErr    error
	hasFutureEvents   bool
	futureEventsErr   error
	futureEventsSowID uuid.UUID
	futureEventsAfter time.Time

	lastEventDate time.Time
	lastEventErr  error

	createErr        error
	created          *weagingdomain.Weaging
	createdFarrowing *farrowingdomain.Farrowing
	createdSow       *sowdomain.Sow

	updateErr error
	updated   *weagingdomain.Weaging

	deleteErr        error
	deleted          *weagingdomain.Weaging
	deletedFarrowing *farrowingdomain.Farrowing
	deletedSow       *sowdomain.Sow

	listWeagings []*weagingdomain.Weaging
	listTotal    int
	listErr      error
	listFilter   ports.WeagingFilter
	listLimit    int
	listOffset   int
}

func (f *fakeWeagingRepository) GetSow(_ context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
	if f.sowErr != nil {
		return nil, f.sowErr
	}
	return f.sows[id], nil
}

func (f *fakeWeagingRepository) GetLastFarrowing(_ context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error) {
	if f.farrowErr != nil {
		return nil, f.farrowErr
	}
	return f.farrowings[sowID], nil
}

func (f *fakeWeagingRepository) GetFarrowingByID(_ context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error) {
	if f.farrowingByIDErr != nil {
		return nil, f.farrowingByIDErr
	}
	return f.farrowingsByID[id], nil
}

func (f *fakeWeagingRepository) GetByID(_ context.Context, id uuid.UUID) (*weagingdomain.Weaging, error) {
	if f.weagingByIDErr != nil {
		return nil, f.weagingByIDErr
	}
	return f.weagings[id], nil
}

func (f *fakeWeagingRepository) HasFutureEvents(_ context.Context, sowID uuid.UUID, after time.Time) (bool, error) {
	f.futureEventsSowID = sowID
	f.futureEventsAfter = after
	return f.hasFutureEvents, f.futureEventsErr
}

func (f *fakeWeagingRepository) GetLatestEventDate(_ context.Context, _ uuid.UUID) (time.Time, error) {
	return f.lastEventDate, f.lastEventErr
}

func (f *fakeWeagingRepository) Create(
	_ context.Context,
	weaging *weagingdomain.Weaging,
	farrowing *farrowingdomain.Farrowing,
	sow *sowdomain.Sow,
) error {
	f.created = weaging
	f.createdFarrowing = farrowing
	f.createdSow = sow
	return f.createErr
}

func (f *fakeWeagingRepository) Update(_ context.Context, weaging *weagingdomain.Weaging) error {
	f.updated = weaging
	return f.updateErr
}

func (f *fakeWeagingRepository) Delete(
	_ context.Context,
	weaging *weagingdomain.Weaging,
	farrowing *farrowingdomain.Farrowing,
	sow *sowdomain.Sow,
) error {
	f.deleted = weaging
	f.deletedFarrowing = farrowing
	f.deletedSow = sow
	return f.deleteErr
}

func (f *fakeWeagingRepository) List(_ context.Context, filter ports.WeagingFilter, limit, offset int) ([]*weagingdomain.Weaging, int, error) {
	f.listFilter = filter
	f.listLimit = limit
	f.listOffset = offset
	return f.listWeagings, f.listTotal, f.listErr
}
