package tests

import (
	"context"
	"time"

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

	farrowingByIDErr error

	partialWeagingDates map[uuid.UUID]time.Time
	partialWeagingErr   error
	partialWeagingCalls int

	weagingDates map[uuid.UUID]time.Time
	weagingErr   error

	fosteringByID    map[uuid.UUID]*pigletfosteringdomain.PigletFostering
	fosteringByIDErr error

	createErr       error
	created         *pigletfosteringdomain.PigletFostering
	createdDonor    *farrowingdomain.Farrowing
	createdReceiver *farrowingdomain.Farrowing

	updateErr error
	updated   *pigletfosteringdomain.PigletFostering

	deleteErr       error
	deleted         *pigletfosteringdomain.PigletFostering
	deletedDonor    *farrowingdomain.Farrowing
	deletedReceiver *farrowingdomain.Farrowing

	listFosterings []*pigletfosteringdomain.PigletFostering
	listTotal      int
	listErr        error
	listFilter     ports.PigletFosteringFilter
	listLimit      int
	listOffset     int
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

func (f *fakePigletFosteringRepository) GetFarrowingByID(_ context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error) {
	if f.farrowingByIDErr != nil {
		return nil, f.farrowingByIDErr
	}
	for _, farrowing := range f.farrowings {
		if farrowing.ID == id {
			return farrowing, nil
		}
	}
	return nil, ports.ErrFarrowingNotFound
}

func (f *fakePigletFosteringRepository) GetLastPartialWeagingDate(_ context.Context, farrowingID uuid.UUID) (time.Time, error) {
	f.partialWeagingCalls++
	if f.partialWeagingErr != nil {
		return time.Time{}, f.partialWeagingErr
	}
	return f.partialWeagingDates[farrowingID], nil
}

func (f *fakePigletFosteringRepository) GetLastWeagingDate(_ context.Context, farrowingID uuid.UUID) (time.Time, error) {
	if f.weagingErr != nil {
		return time.Time{}, f.weagingErr
	}
	return f.weagingDates[farrowingID], nil
}

func (f *fakePigletFosteringRepository) GetByID(_ context.Context, id uuid.UUID) (*pigletfosteringdomain.PigletFostering, error) {
	if f.fosteringByIDErr != nil {
		return nil, f.fosteringByIDErr
	}
	fostering, ok := f.fosteringByID[id]
	if !ok {
		return nil, ports.ErrPigletFosteringNotFound
	}
	return fostering, nil
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

func (f *fakePigletFosteringRepository) Update(_ context.Context, fostering *pigletfosteringdomain.PigletFostering) error {
	f.updated = fostering
	return f.updateErr
}

func (f *fakePigletFosteringRepository) Delete(
	_ context.Context,
	fostering *pigletfosteringdomain.PigletFostering,
	donor *farrowingdomain.Farrowing,
	receiver *farrowingdomain.Farrowing,
) error {
	f.deleted = fostering
	f.deletedDonor = donor
	f.deletedReceiver = receiver
	return f.deleteErr
}

func (f *fakePigletFosteringRepository) List(
	_ context.Context,
	filter ports.PigletFosteringFilter,
	limit int,
	offset int,
) ([]*pigletfosteringdomain.PigletFostering, int, error) {
	f.listFilter = filter
	f.listLimit = limit
	f.listOffset = offset
	return f.listFosterings, f.listTotal, f.listErr
}
