package tests

import (
	"context"
	"time"

	"github.com/google/uuid"

	farrowingdomain "server/internal/modules/farrowing/domain"
	operatordomain "server/internal/modules/operator/domain"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
	sowdomain "server/internal/modules/sow/domain"
)

type fakePigletDeathRepository struct {
	sow          *sowdomain.Sow
	sowErr       error
	farrowing    *farrowingdomain.Farrowing
	farrowingErr error
	operator     *operatordomain.Operator
	operatorErr  error

	lastPartialWeagingDate    time.Time
	lastPartialWeagingDateErr error
	lastPartialWeagingCalls   int

	deathByID      *pigletdeathdomain.PigletDeath
	deathByIDErr   error
	farrowingByID  *farrowingdomain.Farrowing
	farrowingIDErr error

	hasWeagingAfter        bool
	hasWeagingAfterErr     error
	hasPartialWeagingAfter bool
	hasPartialWeagingErr   error

	deleted          *pigletdeathdomain.PigletDeath
	deletedFarrowing *farrowingdomain.Farrowing
	deleteErr        error

	updated   *pigletdeathdomain.PigletDeath
	updateErr error

	createErr        error
	created          *pigletdeathdomain.PigletDeath
	createdFarrowing *farrowingdomain.Farrowing

	listDeaths []*pigletdeathdomain.PigletDeath
	listTotal  int
	listErr    error
	listFilter ports.PigletDeathFilter
	listLimit  int
	listOffset int
}

func (f *fakePigletDeathRepository) GetSow(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.sow, f.sowErr
}

func (f *fakePigletDeathRepository) GetLastFarrowing(_ context.Context, _ uuid.UUID) (*farrowingdomain.Farrowing, error) {
	return f.farrowing, f.farrowingErr
}

func (f *fakePigletDeathRepository) GetLastPartialWeagingDate(_ context.Context, _ uuid.UUID) (time.Time, error) {
	f.lastPartialWeagingCalls++
	return f.lastPartialWeagingDate, f.lastPartialWeagingDateErr
}

func (f *fakePigletDeathRepository) GetOperator(_ context.Context, _ uuid.UUID) (*operatordomain.Operator, error) {
	return f.operator, f.operatorErr
}

func (f *fakePigletDeathRepository) Create(
	_ context.Context,
	death *pigletdeathdomain.PigletDeath,
	farrowing *farrowingdomain.Farrowing,
) error {
	f.created = death
	f.createdFarrowing = farrowing
	return f.createErr
}

func (f *fakePigletDeathRepository) GetByID(_ context.Context, _ uuid.UUID) (*pigletdeathdomain.PigletDeath, error) {
	return f.deathByID, f.deathByIDErr
}

func (f *fakePigletDeathRepository) GetFarrowingByID(_ context.Context, _ uuid.UUID) (*farrowingdomain.Farrowing, error) {
	return f.farrowingByID, f.farrowingIDErr
}

func (f *fakePigletDeathRepository) HasWeagingAfter(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return f.hasWeagingAfter, f.hasWeagingAfterErr
}

func (f *fakePigletDeathRepository) HasPartialWeagingAfter(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return f.hasPartialWeagingAfter, f.hasPartialWeagingErr
}

func (f *fakePigletDeathRepository) Update(_ context.Context, death *pigletdeathdomain.PigletDeath) error {
	f.updated = death
	return f.updateErr
}

func (f *fakePigletDeathRepository) Delete(
	_ context.Context,
	death *pigletdeathdomain.PigletDeath,
	farrowing *farrowingdomain.Farrowing,
) error {
	f.deleted = death
	f.deletedFarrowing = farrowing
	return f.deleteErr
}

func (f *fakePigletDeathRepository) List(
	_ context.Context,
	filter ports.PigletDeathFilter,
	limit int,
	offset int,
) ([]*pigletdeathdomain.PigletDeath, int, error) {
	f.listFilter = filter
	f.listLimit = limit
	f.listOffset = offset
	return f.listDeaths, f.listTotal, f.listErr
}
