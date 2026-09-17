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
	UpdatedBy string
}

type UpdateUserByAdminService struct {
	repository ports.UserRepository
	hasher     ports.PasswordHasher
	clock      func() time.Time
}

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
