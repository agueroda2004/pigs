package application

import (
	"context"

	"github.com/google/uuid"

	"server/internal/modules/boar/ports"
)

type DeleteBoarService struct {
	repository ports.BoarRepository
}

// NewDeleteBoarService builds a delete-boar use case with its repository.
// It returns a service ready to execute boar deletions.
func NewDeleteBoarService(repository ports.BoarRepository) *DeleteBoarService {
	return &DeleteBoarService{repository: repository}
}

// Execute deletes a boar by its identifier.
// It returns ErrBoarNotFound when the boar does not exist and ErrBoarInUse
// when the boar still has mounts or removals linked to it.
func (s *DeleteBoarService) Execute(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
