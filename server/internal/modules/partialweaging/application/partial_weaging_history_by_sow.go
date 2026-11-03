package application

import (
	"context"

	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
)

// PartialWeagingHistoryBySowService returns the partial weaging history of a sow code.
type PartialWeagingHistoryBySowService struct {
	repository ports.PartialWeagingRepository
}

// NewPartialWeagingHistoryBySowService creates the history query with its repository.
func NewPartialWeagingHistoryBySowService(repository ports.PartialWeagingRepository) *PartialWeagingHistoryBySowService {
	return &PartialWeagingHistoryBySowService{repository: repository}
}

// Execute returns every partial weaging history item for the requested sow code.
// It returns an empty slice when none exist and propagates repository errors.
func (s *PartialWeagingHistoryBySowService) Execute(ctx context.Context, sowCode string) ([]*partialweagingdomain.PartialWeagingHistoryBySow, error) {
	items, err := s.repository.HistoryBySow(ctx, sowCode)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*partialweagingdomain.PartialWeagingHistoryBySow, 0)
	}
	return items, nil
}
