package application

import (
	"context"

	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

// PigletFosteringHistoryBySowService returns the fostering history of a sow code.
type PigletFosteringHistoryBySowService struct {
	repository ports.PigletFosteringRepository
}

// NewPigletFosteringHistoryBySowService creates the history query with its repository.
func NewPigletFosteringHistoryBySowService(repository ports.PigletFosteringRepository) *PigletFosteringHistoryBySowService {
	return &PigletFosteringHistoryBySowService{repository: repository}
}

// Execute returns every fostering history item for the requested sow code.
// It returns an empty slice when none exist and propagates repository errors.
func (s *PigletFosteringHistoryBySowService) Execute(ctx context.Context, sowCode string) ([]*pigletfosteringdomain.PigletFosteringHistoryBySow, error) {
	items, err := s.repository.HistoryBySow(ctx, sowCode)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*pigletfosteringdomain.PigletFosteringHistoryBySow, 0)
	}
	return items, nil
}
