package application

import (
	"context"

	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
)

// DefaultServicePageSize is the fixed number of services returned per page.
const DefaultServicePageSize = 10

// ServicePage is a page of services together with its pagination metadata.
type ServicePage struct {
	Items      []*servicedomain.Service
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListServicesService struct {
	repository ports.ServiceRepository
}

// NewListServicesService builds a list-services use case with its repository.
// It returns a service ready to return services matching the given filter.
func NewListServicesService(repository ports.ServiceRepository) *ListServicesService {
	return &ListServicesService{repository: repository}
}

// Execute returns one page of services matching the filter, ordered by creation date.
// Pages start at one and always contain DefaultServicePageSize services.
func (s *ListServicesService) Execute(ctx context.Context, filter ports.ServiceFilter, page int) (ServicePage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultServicePageSize
	items, total, err := s.repository.List(ctx, filter, DefaultServicePageSize, offset)
	if err != nil {
		return ServicePage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultServicePageSize - 1) / DefaultServicePageSize
	}

	return ServicePage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultServicePageSize,
		TotalPages: totalPages,
	}, nil
}
