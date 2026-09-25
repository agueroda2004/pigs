package application

import (
	"context"

	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
)

type ListServicesService struct {
	repository ports.ServiceRepository
}

// NewListServicesService builds a list-services use case with its repository.
// It returns a service ready to return services matching the given filter.
func NewListServicesService(repository ports.ServiceRepository) *ListServicesService {
	return &ListServicesService{repository: repository}
}

// Execute returns the services matching the filter, each one with its mounts.
// A zero-value filter returns every registered service.
func (s *ListServicesService) Execute(ctx context.Context, filter ports.ServiceFilter) ([]*servicedomain.Service, error) {
	return s.repository.List(ctx, filter)
}
