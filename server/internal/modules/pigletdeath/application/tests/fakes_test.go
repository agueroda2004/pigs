package tests

import (
	"context"

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

	createErr        error
	created          *pigletdeathdomain.PigletDeath
	createdFarrowing *farrowingdomain.Farrowing

	listDeaths []*pigletdeathdomain.PigletDeath
	listErr    error
	listFilter ports.PigletDeathFilter
}

func (f *fakePigletDeathRepository) GetSow(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.sow, f.sowErr
}

func (f *fakePigletDeathRepository) GetLastFarrowing(_ context.Context, _ uuid.UUID) (*farrowingdomain.Farrowing, error) {
	return f.farrowing, f.farrowingErr
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

func (f *fakePigletDeathRepository) List(_ context.Context, filter ports.PigletDeathFilter) ([]*pigletdeathdomain.PigletDeath, error) {
	f.listFilter = filter
	return f.listDeaths, f.listErr
}
