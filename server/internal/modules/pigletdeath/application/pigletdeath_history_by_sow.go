package application

import (
	"context"

	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
)

// PigletDeathHistoryBySowService returns the piglet death history of a sow code.
type PigletDeathHistoryBySowService struct {
	repository ports.PigletDeathRepository
}

// NewPigletDeathHistoryBySowService creates the history query with its repository.
func NewPigletDeathHistoryBySowService(repository ports.PigletDeathRepository) *PigletDeathHistoryBySowService {
	return &PigletDeathHistoryBySowService{repository: repository}
}

// Execute returns every piglet death history item for the requested sow code.
// It returns an empty slice when none exist and propagates repository errors.
func (s *PigletDeathHistoryBySowService) Execute(ctx context.Context, sowCode string) ([]*pigletdeathdomain.PigletDeathHistoryBySow, error) {
	items, err := s.repository.HistoryBySow(ctx, sowCode)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*pigletdeathdomain.PigletDeathHistoryBySow, 0)
	}
	return items, nil
}
