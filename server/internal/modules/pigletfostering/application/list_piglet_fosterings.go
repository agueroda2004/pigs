package application

import (
	"context"

	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
)

// DefaultPigletFosteringPageSize is the fixed number of fosterings returned per page.
const DefaultPigletFosteringPageSize = 10

// PigletFosteringPage is a page of fosterings together with its pagination metadata.
type PigletFosteringPage struct {
	Items      []*pigletfosteringdomain.PigletFostering
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListPigletFosteringsService struct {
	repository ports.PigletFosteringRepository
}

// NewListPigletFosteringsService builds a list-fosterings use case with its repository.
// It returns a service ready to return fosterings matching the given filter.
func NewListPigletFosteringsService(repository ports.PigletFosteringRepository) *ListPigletFosteringsService {
	return &ListPigletFosteringsService{repository: repository}
}

// Execute returns one page of fosterings matching the filter.
// Pages start at one and always contain DefaultPigletFosteringPageSize fosterings.
func (s *ListPigletFosteringsService) Execute(ctx context.Context, filter ports.PigletFosteringFilter, page int) (PigletFosteringPage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultPigletFosteringPageSize
	items, total, err := s.repository.List(ctx, filter, DefaultPigletFosteringPageSize, offset)
	if err != nil {
		return PigletFosteringPage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultPigletFosteringPageSize - 1) / DefaultPigletFosteringPageSize
	}

	return PigletFosteringPage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultPigletFosteringPageSize,
		TotalPages: totalPages,
	}, nil
}
