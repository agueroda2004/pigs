package tests

import (
	"context"
	"time"

	"github.com/google/uuid"

	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

type fakeAbortionRepository struct {
	sow        *sowdomain.Sow
	sowErr     error
	service    *servicedomain.Service
	serviceErr error

	getAbortion    *abortiondomain.Abortion
	getAbortionErr error

	createErr      error
	created        *abortiondomain.Abortion
	createdSow     *sowdomain.Sow
	createdService *servicedomain.Service

	updated   *abortiondomain.Abortion
	updateErr error

	hasFutureEvents      bool
	hasFutureEventsErr   error
	hasFutureEventsAfter time.Time

	deleted        *abortiondomain.Abortion
	deletedSow     *sowdomain.Sow
	deletedService *servicedomain.Service
	deleteErr      error

	listAbortions []*abortiondomain.Abortion
	listTotal     int
	listErr       error
	listFilter    ports.AbortionFilter
	listLimit     int
	listOffset    int
}

func (f *fakeAbortionRepository) GetSow(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.sow, f.sowErr
}

func (f *fakeAbortionRepository) GetByID(_ context.Context, _ uuid.UUID) (*abortiondomain.Abortion, error) {
	return f.getAbortion, f.getAbortionErr
}

func (f *fakeAbortionRepository) GetService(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.service, f.serviceErr
}

func (f *fakeAbortionRepository) GetLastService(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.service, f.serviceErr
}

func (f *fakeAbortionRepository) Update(_ context.Context, abortion *abortiondomain.Abortion) error {
	f.updated = abortion
	return f.updateErr
}

func (f *fakeAbortionRepository) HasFutureEvents(_ context.Context, _ uuid.UUID, after time.Time) (bool, error) {
	f.hasFutureEventsAfter = after
	return f.hasFutureEvents, f.hasFutureEventsErr
}

func (f *fakeAbortionRepository) Delete(_ context.Context, abortion *abortiondomain.Abortion, sow *sowdomain.Sow, service *servicedomain.Service) error {
	f.deleted = abortion
	f.deletedSow = sow
	f.deletedService = service
	return f.deleteErr
}

func (f *fakeAbortionRepository) Create(
	_ context.Context,
	abortion *abortiondomain.Abortion,
	sow *sowdomain.Sow,
	service *servicedomain.Service,
) error {
	f.created = abortion
	f.createdSow = sow
	f.createdService = service
	return f.createErr
}

func (f *fakeAbortionRepository) List(_ context.Context, filter ports.AbortionFilter, limit, offset int) ([]*abortiondomain.Abortion, int, error) {
	f.listFilter = filter
	f.listLimit = limit
	f.listOffset = offset
	return f.listAbortions, f.listTotal, f.listErr
}
