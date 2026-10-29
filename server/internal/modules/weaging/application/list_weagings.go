package application

import (
	"context"

	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

// DefaultWeagingPageSize is the fixed number of weagings returned per page.
const DefaultWeagingPageSize = 10

// WeagingPage is a page of weagings together with its pagination metadata.
type WeagingPage struct {
	Items      []*weagingdomain.Weaging
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListWeagingsService struct {
	repository ports.WeagingRepository
}

// NewListWeagingsService builds a list-weagings use case with its repository.
// It returns a service ready to return weagings matching the given filter.
func NewListWeagingsService(repository ports.WeagingRepository) *ListWeagingsService {
	return &ListWeagingsService{repository: repository}
}

// Execute returns one page of weagings matching the filter.
// Pages start at one and always contain DefaultWeagingPageSize weagings.
func (s *ListWeagingsService) Execute(ctx context.Context, filter ports.WeagingFilter, page int) (WeagingPage, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * DefaultWeagingPageSize
	items, total, err := s.repository.List(ctx, filter, DefaultWeagingPageSize, offset)
	if err != nil {
		return WeagingPage{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + DefaultWeagingPageSize - 1) / DefaultWeagingPageSize
	}

	return WeagingPage{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   DefaultWeagingPageSize,
		TotalPages: totalPages,
	}, nil
}
