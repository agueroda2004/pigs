package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

var ErrUsernameAlreadyExists = ports.ErrUsernameAlreadyUsed

type CreateUserCommand struct {
	Name      string
	Username  string
	Password  string
	Role      userdomain.Role
	CreatedBy string
}

type CreateUserService struct {
	repository ports.UserRepository
	hasher     ports.PasswordHasher
	clock      func() time.Time
}

func NewCreateUserService(
	repository ports.UserRepository,
	hasher ports.PasswordHasher,
	clock func() time.Time,
) *CreateUserService {
	return &CreateUserService{
		repository: repository,
		hasher:     hasher,
		clock:      clock,
	}
}

func (s *CreateUserService) Execute(ctx context.Context, command CreateUserCommand) (*userdomain.User, error) {
	username := strings.TrimSpace(command.Username)
	exists, err := s.repository.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	passwordHash, err := s.hasher.Hash(command.Password)
	if err != nil {
		return nil, err
	}

	createdAt := s.clock()
	newUser, err := userdomain.NewUser(
		uuid.New(),
		command.Name,
		username,
		passwordHash,
		command.Role,
		command.CreatedBy,
		createdAt,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}
