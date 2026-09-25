package tests

import (
	"context"

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

	createErr      error
	created        *abortiondomain.Abortion
	createdSow     *sowdomain.Sow
	createdService *servicedomain.Service

	listAbortions []*abortiondomain.Abortion
	listErr       error
	listFilter    ports.AbortionFilter
}

func (f *fakeAbortionRepository) GetSow(_ context.Context, _ uuid.UUID) (*sowdomain.Sow, error) {
	return f.sow, f.sowErr
}

func (f *fakeAbortionRepository) GetLastService(_ context.Context, _ uuid.UUID) (*servicedomain.Service, error) {
	return f.service, f.serviceErr
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

func (f *fakeAbortionRepository) List(_ context.Context, filter ports.AbortionFilter) ([]*abortiondomain.Abortion, error) {
	f.listFilter = filter
	return f.listAbortions, f.listErr
}
