package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

var ErrBreedNameAlreadyExists = ports.ErrBreedNameAlreadyUsed

type CreateBreedCommand struct {
	Name      string
	CreatedBy uuid.UUID
}

type CreateBreedService struct {
	repository ports.BreedRepository
	clock      func() time.Time
}

// NewCreateBreedService builds a create-breed use case with its repository and clock.
// It returns a service ready to execute CreateBreedCommand values.
func NewCreateBreedService(repository ports.BreedRepository, clock func() time.Time) *CreateBreedService {
	return &CreateBreedService{repository: repository, clock: clock}
}

// Execute creates a breed after ensuring its name is free.
// It returns ErrBreedNameAlreadyExists when the name is already taken.
func (s *CreateBreedService) Execute(ctx context.Context, command CreateBreedCommand) (*breeddomain.Breed, error) {
	name := strings.TrimSpace(command.Name)
	exists, err := s.repository.ExistsByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrBreedNameAlreadyExists
	}

	newBreed, err := breeddomain.NewBreed(uuid.New(), name, command.CreatedBy, s.clock())
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newBreed); err != nil {
		return nil, err
	}
	return newBreed, nil
}
