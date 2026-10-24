package application

import (
	"context"

	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
)

// DefaultPigletDeathPageSize is the fixed number of piglet deaths returned per page.
const DefaultPigletDeathPageSize = 10

// PigletDeathPage is a page of piglet deaths together with its pagination metadata.
type PigletDeathPage struct {
	Items      []*pigletdeathdomain.PigletDeath
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListPigletDeathsService struct {
	repository ports.PigletDeathRepository
}

// NewListPigletDeathsService builds a list-piglet-deaths use case with its repository.
// It returns a service ready to return piglet deaths matching the given filter.
func NewListPigletDeathsService(repository ports.PigletDeathRepository) *ListPigletDeathsService {
	return &ListPigletDeathsService{repository: repository}
}

// Execute returns one page of piglet deaths matching the filter.
// Pages start at one and always contain DefaultPigletDeathPageSize deaths.
func (s *ListPigletDeathsService) Execute(ctx context.Context, filter ports.PigletDeathFilter, page int) (PigletDeathPage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultPigletDeathPageSize
	items, total, err := s.repository.List(ctx, filter, DefaultPigletDeathPageSize, offset)
	if err != nil {
		return PigletDeathPage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultPigletDeathPageSize - 1) / DefaultPigletDeathPageSize
	}

	return PigletDeathPage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultPigletDeathPageSize,
		TotalPages: totalPages,
	}, nil
}
