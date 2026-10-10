package application

import (
	"context"

	"github.com/google/uuid"

	"server/internal/modules/sow/ports"
)

type DeleteSowService struct {
	repository ports.SowRepository
}

// NewDeleteSowService builds a delete-sow use case with its repository.
// It returns a service ready to execute sow deletions.
func NewDeleteSowService(repository ports.SowRepository) *DeleteSowService {
	return &DeleteSowService{repository: repository}
}

// Execute deletes a sow by its identifier.
// It returns ErrSowNotFound when the sow does not exist and ErrSowInUse when the
// sow still has linked records such as services, abortions or farrowings.
func (s *DeleteSowService) Execute(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
