package application

import (
	"context"

	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
)

// DefaultPartialWeagingPageSize is the fixed number of partial weagings returned per page.
const DefaultPartialWeagingPageSize = 10

// PartialWeagingPage is a page of partial weagings together with its pagination metadata.
type PartialWeagingPage struct {
	Items      []*partialweagingdomain.PartialWeaging
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListPartialWeagingsService struct {
	repository ports.PartialWeagingRepository
}

// NewListPartialWeagingsService builds a list-partial-weagings use case with its repository.
// It returns a service ready to return partial weagings matching the given filter.
func NewListPartialWeagingsService(repository ports.PartialWeagingRepository) *ListPartialWeagingsService {
	return &ListPartialWeagingsService{repository: repository}
}

// Execute returns one page of partial weagings matching the filter.
// Pages start at one and always contain DefaultPartialWeagingPageSize records.
func (s *ListPartialWeagingsService) Execute(ctx context.Context, filter ports.PartialWeagingFilter, page int) (PartialWeagingPage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultPartialWeagingPageSize
	items, total, err := s.repository.List(ctx, filter, DefaultPartialWeagingPageSize, offset)
	if err != nil {
		return PartialWeagingPage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultPartialWeagingPageSize - 1) / DefaultPartialWeagingPageSize
	}

	return PartialWeagingPage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultPartialWeagingPageSize,
		TotalPages: totalPages,
	}, nil
}
