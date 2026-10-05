package tests

import (
	"context"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
	sowdomain "server/internal/modules/sow/domain"
)

type fakePigletFosteringRepository struct {
	sows       map[uuid.UUID]*sowdomain.Sow
	farrowings map[uuid.UUID]*farrowingdomain.Farrowing
	sowErr     error
	farrowErr  error

	createErr       error
	created         *pigletfosteringdomain.PigletFostering
	createdDonor    *farrowingdomain.Farrowing
	createdReceiver *farrowingdomain.Farrowing

	listFosterings []*pigletfosteringdomain.PigletFostering
	listErr        error
	listFilter     ports.PigletFosteringFilter
}

func (f *fakePigletFosteringRepository) GetSow(_ context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
	if f.sowErr != nil {
		return nil, f.sowErr
	}
	return f.sows[id], nil
}

func (f *fakePigletFosteringRepository) GetLastFarrowing(_ context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error) {
	if f.farrowErr != nil {
		return nil, f.farrowErr
	}
	return f.farrowings[sowID], nil
}

func (f *fakePigletFosteringRepository) Create(
	_ context.Context,
	fostering *pigletfosteringdomain.PigletFostering,
	donor *farrowingdomain.Farrowing,
	receiver *farrowingdomain.Farrowing,
) error {
	f.created = fostering
	f.createdDonor = donor
	f.createdReceiver = receiver
	return f.createErr
}

func (f *fakePigletFosteringRepository) List(_ context.Context, filter ports.PigletFosteringFilter) ([]*pigletfosteringdomain.PigletFostering, error) {
	f.listFilter = filter
	return f.listFosterings, f.listErr
}
