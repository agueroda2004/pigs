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

	createErr        error
	created          *partialweagingdomain.PartialWeaging
	createdFarrowing *farrowingdomain.Farrowing
	createdSow       *sowdomain.Sow

	listWeagings []*partialweagingdomain.PartialWeaging
	listErr      error
	listFilter   ports.PartialWeagingFilter
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

func (f *fakePartialWeagingRepository) List(_ context.Context, filter ports.PartialWeagingFilter) ([]*partialweagingdomain.PartialWeaging, error) {
	f.listFilter = filter
	return f.listWeagings, f.listErr
}
