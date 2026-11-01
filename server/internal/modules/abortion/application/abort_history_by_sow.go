package application

import (
	"context"

	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
)

// AbortHistoryBySowService returns every abortion associated with a sow code.
type AbortHistoryBySowService struct {
	repository ports.AbortionRepository
}

// NewAbortHistoryBySowService creates the internal history query with its repository.
func NewAbortHistoryBySowService(repository ports.AbortionRepository) *AbortHistoryBySowService {
	return &AbortHistoryBySowService{repository: repository}
}

// Execute returns all abortion-history items for the requested sow code.
func (s *AbortHistoryBySowService) Execute(ctx context.Context, sowCode string) ([]abortiondomain.AbortionHistoryBySow, error) {
	abortions, err := s.repository.HistoryBySow(ctx, sowCode)
	if err != nil {
		return nil, err
	}
	items := make([]abortiondomain.AbortionHistoryBySow, 0, len(abortions))
	for _, abortion := range abortions {
		items = append(items, abortiondomain.AbortionHistoryBySow{
			SowCode:      abortion.SowCode,
			AbortionDate: abortion.AbortionDate.UTC().Format("2006-01-02"),
			Cause:        string(abortion.Cause),
			Note:         abortion.Note,
		})
	}
	return items, nil
}
