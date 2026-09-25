package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

type UpdateUserByAdminCommand struct {
	Name      *string
	Username  *string
	Password  *string
	Role      *userdomain.Role
	Active    *bool
	UpdatedBy string
}

type UpdateUserByAdminService struct {
	repository ports.UserRepository
	hasher     ports.PasswordHasher
	clock      func() time.Time
}

// NewUpdateUserByAdminService builds an admin update use case with its dependencies.
// It returns a service ready to execute UpdateUserByAdminCommand values.
func NewUpdateUserByAdminService(
	repository ports.UserRepository,
	hasher ports.PasswordHasher,
	clock func() time.Time,
) *UpdateUserByAdminService {
	return &UpdateUserByAdminService{
		repository: repository,
		hasher:     hasher,
		clock:      clock,
	}
}

// Execute updates any user's fields on behalf of an administrator.
// It hashes the password only when provided and persists the updated user.
func (s *UpdateUserByAdminService) Execute(
	ctx context.Context,
	userID uuid.UUID,
	command UpdateUserByAdminCommand,
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

	if err := currentUser.UpdateByAdmin(
		command.Name,
		command.Username,
		passwordHash,
		command.Role,
		command.Active,
		command.UpdatedBy,
		s.clock(),
	); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, currentUser); err != nil {
		return nil, err
	}
	return currentUser, nil
}
