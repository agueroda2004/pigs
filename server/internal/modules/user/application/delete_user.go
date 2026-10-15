package application

import (
	"context"

	"github.com/google/uuid"

	"server/internal/modules/user/ports"
)

type DeleteUserService struct {
	repository ports.UserRepository
}

// NewDeleteUserService builds a delete-user use case with its repository.
// It returns a service ready to execute user deletions.
func NewDeleteUserService(repository ports.UserRepository) *DeleteUserService {
	return &DeleteUserService{repository: repository}
}

// Execute deletes a user by its identifier.
// It returns ErrUserNotFound when the user does not exist and ErrUserInUse
// when the user still has records linked to it.
func (s *DeleteUserService) Execute(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
