package application

import (
	"context"

	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
)

// ServicesHistoryBySowService returns every service associated with a sow code.
type ServicesHistoryBySowService struct {
	repository ports.ServiceRepository
}

// NewServicesHistoryBySowService creates the internal history query with its repository.
func NewServicesHistoryBySowService(repository ports.ServiceRepository) *ServicesHistoryBySowService {
	return &ServicesHistoryBySowService{repository: repository}
}

// Execute returns all service-history items for the requested sow code.
func (s *ServicesHistoryBySowService) Execute(ctx context.Context, sowCode string) ([]*servicedomain.ServiceHistoryBySow, error) {
	items, err := s.repository.HistoryBySow(ctx, sowCode)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*servicedomain.ServiceHistoryBySow, 0)
	}
	return items, nil
}