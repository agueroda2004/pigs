package application

import (
	"context"

	"github.com/google/uuid"

	"server/internal/modules/breed/ports"
)

type DeleteBreedService struct {
	repository ports.BreedRepository
}

// NewDeleteBreedService builds a delete-breed use case with its repository.
// It returns a service ready to execute breed deletions.
func NewDeleteBreedService(repository ports.BreedRepository) *DeleteBreedService {
	return &DeleteBreedService{repository: repository}
}

// Execute deletes a breed by its identifier.
// It returns ErrBreedNotFound when the breed does not exist and ErrBreedInUse
// when the breed still has sows or boars linked to it.
func (s *DeleteBreedService) Execute(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
