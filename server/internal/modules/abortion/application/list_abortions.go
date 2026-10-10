package application

import (
	"context"

	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
)

// DefaultAbortionPageSize is the fixed number of abortions returned per page.
const DefaultAbortionPageSize = 10

// AbortionPage is a page of abortions together with its pagination metadata.
type AbortionPage struct {
	Items      []*abortiondomain.Abortion
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListAbortionsService struct {
	repository ports.AbortionRepository
}

// NewListAbortionsService builds a list-abortions use case with its repository.
// It returns a service ready to return abortions matching the given filter.
func NewListAbortionsService(repository ports.AbortionRepository) *ListAbortionsService {
	return &ListAbortionsService{repository: repository}
}

// Execute returns one page of abortions matching the filter, ordered by date.
// Pages start at one and always contain DefaultAbortionPageSize abortions.
func (s *ListAbortionsService) Execute(ctx context.Context, filter ports.AbortionFilter, page int) (AbortionPage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultAbortionPageSize
	items, total, err := s.repository.List(ctx, filter, DefaultAbortionPageSize, offset)
	if err != nil {
		return AbortionPage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultAbortionPageSize - 1) / DefaultAbortionPageSize
	}

	return AbortionPage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultAbortionPageSize,
		TotalPages: totalPages,
	}, nil
}
