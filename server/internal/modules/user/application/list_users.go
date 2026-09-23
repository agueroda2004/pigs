package application

import (
	"context"

	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

type ListUsersService struct {
	repository ports.UserRepository
}

// NewListUsersService builds a list-users use case with its repository.
// It returns a service ready to return every registered user.
func NewListUsersService(repository ports.UserRepository) *ListUsersService {
	return &ListUsersService{repository: repository}
}

// Execute returns every registered user without pagination or filtering.
// It is intended for administrative listings of the small active user base.
func (s *ListUsersService) Execute(ctx context.Context) ([]*userdomain.User, error) {
	return s.repository.List(ctx)
}
