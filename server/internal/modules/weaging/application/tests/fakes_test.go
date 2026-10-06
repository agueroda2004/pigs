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

	lastEventDate time.Time
	lastEventErr  error

	createErr        error
	created          *weagingdomain.Weaging
	createdFarrowing *farrowingdomain.Farrowing
	createdSow       *sowdomain.Sow

	listWeagings []*weagingdomain.Weaging
	listErr      error
	listFilter   ports.WeagingFilter
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

func (f *fakeWeagingRepository) List(_ context.Context, filter ports.WeagingFilter) ([]*weagingdomain.Weaging, error) {
	f.listFilter = filter
	return f.listWeagings, f.listErr
}
