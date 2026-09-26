package tests

import (
	"context"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

type fakeSowRemovalRepository struct {
	sow         *sowdomain.Sow
	sowErr      error
	service     *servicedomain.Service
	serviceErr  error
	abortion    *abortiondomain.Abortion
	abortionErr error

	createErr      error
	created        *sowremovaldomain.SowRemoval
	createdSow     *sowdomain.Sow
	createdService *servicedomain.Service

	listRemovals []*sowremovaldomain.SowRemoval
	listErr      error
	listFilter   ports.SowRemovalFilter
}

func (f *fakeSowRemovalRepository) GetSow(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.sow, f.sowErr
}

func (f *fakeSowRemovalRepository) GetLastService(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.service, f.serviceErr
}

func (f *fakeSowRemovalRepository) GetLastAbortion(_ context.Context, _ uuid.UUID) (*abortiondomain.Abortion, error) {
	return f.abortion, f.abortionErr
}

func (f *fakeSowRemovalRepository) Create(
	_ context.Context,
	removal *sowremovaldomain.SowRemoval,
	sow *sowdomain.Sow,
	service *servicedomain.Service,
) error {
	f.created = removal
	f.createdSow = sow
	f.createdService = service
	return f.createErr
}

func (f *fakeSowRemovalRepository) List(_ context.Context, filter ports.SowRemovalFilter) ([]*sowremovaldomain.SowRemoval, error) {
	f.listFilter = filter
	return f.listRemovals, f.listErr
}
