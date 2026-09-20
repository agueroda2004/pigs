package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

type UpdateOwnUserCommand struct {
	Name     *string
	Password *string
}

type UpdateOwnUserService struct {
	repository ports.UserRepository
	hasher     ports.PasswordHasher
	clock      func() time.Time
}

// NewUpdateOwnUserService builds an update-own-profile use case with its dependencies.
// It returns a service ready to execute UpdateOwnUserCommand values.
func NewUpdateOwnUserService(
	repository ports.UserRepository,
	hasher ports.PasswordHasher,
	clock func() time.Time,
) *UpdateOwnUserService {
	return &UpdateOwnUserService{
		repository: repository,
		hasher:     hasher,
		clock:      clock,
	}
}

// Execute updates the caller's own name and/or password by their identifier.
// It hashes the password only when provided and persists the updated user.
func (s *UpdateOwnUserService) Execute(
	ctx context.Context,
	userID uuid.UUID,
	command UpdateOwnUserCommand,
) (*userdomain.User, error) {
	currentUser, err := s.repository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var passwordHash *string
	if command.Password != nil {
		hash, err := s.hasher.Hash(*command.Password)
		if err != nil {
			return nil, err
		}
		passwordHash = &hash
	}

	if err := currentUser.UpdateOwnProfile(
		command.Name,
		passwordHash,
		userID.String(),
		s.clock(),
	); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, currentUser); err != nil {
		return nil, err
	}
	return currentUser, nil
}
