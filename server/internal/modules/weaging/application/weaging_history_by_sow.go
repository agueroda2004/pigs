package application

import (
	"context"

	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

// WeagingHistoryBySowService returns the full weaging history of a sow code.
type WeagingHistoryBySowService struct {
	repository ports.WeagingRepository
}

// NewWeagingHistoryBySowService creates the history query with its repository.
func NewWeagingHistoryBySowService(repository ports.WeagingRepository) *WeagingHistoryBySowService {
	return &WeagingHistoryBySowService{repository: repository}
}

// Execute returns every full weaging history item for the requested sow code.
// It returns an empty slice when none exist and propagates repository errors.
func (s *WeagingHistoryBySowService) Execute(ctx context.Context, sowCode string) ([]*weagingdomain.WeagingHistoryBySow, error) {
	items, err := s.repository.HistoryBySow(ctx, sowCode)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*weagingdomain.WeagingHistoryBySow, 0)
	}
	return items, nil
}
