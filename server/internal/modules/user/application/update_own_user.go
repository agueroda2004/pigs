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
