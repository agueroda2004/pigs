package application

import (
	"context"

	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
)

// FarrowingHistoryBySowService returns the farrowing history of a sow code.
type FarrowingHistoryBySowService struct {
	repository ports.FarrowingRepository
}

// NewFarrowingHistoryBySowService creates the history query with its repository.
func NewFarrowingHistoryBySowService(repository ports.FarrowingRepository) *FarrowingHistoryBySowService {
	return &FarrowingHistoryBySowService{repository: repository}
}

// Execute returns every farrowing history item for the requested sow code.
// It derives the duration of each item and returns an empty slice when none exist.
func (s *FarrowingHistoryBySowService) Execute(ctx context.Context, sowCode string) ([]*farrowingdomain.FarrowingHistoryBySow, error) {
	items, err := s.repository.HistoryBySow(ctx, sowCode)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*farrowingdomain.FarrowingHistoryBySow, 0)
	}
	for _, item := range items {
		item.Duration = farrowingdomain.FormatDuration(item.StartTime, item.EndTime)
	}
	return items, nil
}
