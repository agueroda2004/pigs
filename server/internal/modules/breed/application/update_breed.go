package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

type UpdateBreedCommand struct {
	Name      *string
	Active    *bool
	UpdatedBy uuid.UUID
}

type UpdateBreedService struct {
	repository ports.BreedRepository
	clock      func() time.Time
}

// NewUpdateBreedService builds an update-breed use case with its repository and clock.
// It returns a service ready to execute UpdateBreedCommand values.
func NewUpdateBreedService(repository ports.BreedRepository, clock func() time.Time) *UpdateBreedService {
	return &UpdateBreedService{repository: repository, clock: clock}
}

// Execute applies the provided fields to an existing breed by its identifier.
// It loads the breed, validates the change and persists the updated entity.
func (s *UpdateBreedService) Execute(
	ctx context.Context,
	breedID uuid.UUID,
	command UpdateBreedCommand,
) (*breeddomain.Breed, error) {
	currentBreed, err := s.repository.GetByID(ctx, breedID)
	if err != nil {
		return nil, err
	}

	if err := s.checkNameAvailable(ctx, breedID, currentBreed.Name, command.Name); err != nil {
		return nil, err
	}

	if err := currentBreed.Update(command.Name, command.Active, command.UpdatedBy, s.clock()); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, currentBreed); err != nil {
		return nil, err
	}
	return currentBreed, nil
}

// checkNameAvailable verifies that a new breed name is not used by another breed.
// It skips the check when the name is omitted or unchanged and returns
// ports.ErrBreedNameAlreadyUsed when the name is taken.
func (s *UpdateBreedService) checkNameAvailable(ctx context.Context, breedID uuid.UUID, currentName string, name *string) error {
	if name == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*name)
	if trimmed == currentName {
		return nil
	}
	exists, err := s.repository.ExistsByNameExcludingID(ctx, trimmed, breedID)
	if err != nil {
		return err
	}
	if exists {
		return ports.ErrBreedNameAlreadyUsed
	}
	return nil
}
