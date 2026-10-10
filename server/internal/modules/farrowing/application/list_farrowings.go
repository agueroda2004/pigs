package application

import (
	"context"

	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
)

// DefaultFarrowingPageSize is the fixed number of farrowings returned per page.
const DefaultFarrowingPageSize = 10

// FarrowingPage is a page of farrowings together with its pagination metadata.
type FarrowingPage struct {
	Items      []*farrowingdomain.Farrowing
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListFarrowingsService struct {
	repository ports.FarrowingRepository
}

// NewListFarrowingsService builds a list-farrowings use case with its repository.
// It returns a service ready to return farrowings matching the given filter.
func NewListFarrowingsService(repository ports.FarrowingRepository) *ListFarrowingsService {
	return &ListFarrowingsService{repository: repository}
}

// Execute returns one page of farrowings matching the filter.
// Pages start at one and always contain DefaultFarrowingPageSize farrowings.
func (s *ListFarrowingsService) Execute(ctx context.Context, filter ports.FarrowingFilter, page int) (FarrowingPage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultFarrowingPageSize
	items, total, err := s.repository.List(ctx, filter, DefaultFarrowingPageSize, offset)
	if err != nil {
		return FarrowingPage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultFarrowingPageSize - 1) / DefaultFarrowingPageSize
	}

	return FarrowingPage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultFarrowingPageSize,
		TotalPages: totalPages,
	}, nil
}
