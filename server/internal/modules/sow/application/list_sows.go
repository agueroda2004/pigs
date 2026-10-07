package application

import (
	"context"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

// DefaultSowPageSize is the fixed number of sows returned per page.
const DefaultSowPageSize = 20

// SowPage is a page of sows together with its pagination metadata.
type SowPage struct {
	Items      []*sowdomain.Sow
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListSowsService struct {
	repository ports.SowRepository
}

// NewListSowsService builds a list-sows use case with its repository.
// It returns a service ready to return sows matching the given filter.
func NewListSowsService(repository ports.SowRepository) *ListSowsService {
	return &ListSowsService{repository: repository}
}

// Execute returns one page of sows matching the filter, ordered by code.
// Pages start at one and always contain DefaultSowPageSize sows.
func (s *ListSowsService) Execute(ctx context.Context, filter ports.SowFilter, page int) (SowPage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultSowPageSize
	items, total, err := s.repository.List(ctx, filter, DefaultSowPageSize, offset)
	if err != nil {
		return SowPage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultSowPageSize - 1) / DefaultSowPageSize
	}

	return SowPage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultSowPageSize,
		TotalPages: totalPages,
	}, nil
}
