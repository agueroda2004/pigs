package application

import (
	"context"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

// DefaultBoarPageSize is the fixed number of boars returned per page.
const DefaultBoarPageSize = 10

// BoarPage is a page of boars together with its pagination metadata.
type BoarPage struct {
	Items      []*boardomain.Boar
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListBoarsService struct {
	repository ports.BoarRepository
}

// NewListBoarsService builds a list-boars use case with its repository.
// It returns a service ready to return boars matching the given filter.
func NewListBoarsService(repository ports.BoarRepository) *ListBoarsService {
	return &ListBoarsService{repository: repository}
}

// Execute returns one page of boars matching the filter, ordered by code.
// Pages start at one and always contain DefaultBoarPageSize boars.
func (s *ListBoarsService) Execute(ctx context.Context, filter ports.BoarFilter, page int) (BoarPage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultBoarPageSize
	items, total, err := s.repository.List(ctx, filter, DefaultBoarPageSize, offset)
	if err != nil {
		return BoarPage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultBoarPageSize - 1) / DefaultBoarPageSize
	}

	return BoarPage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultBoarPageSize,
		TotalPages: totalPages,
	}, nil
}
