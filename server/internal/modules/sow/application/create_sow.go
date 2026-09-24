package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

var ErrSowCodeAlreadyExists = ports.ErrSowCodeAlreadyUsed

type CreateSowCommand struct {
	Code      string
	Location  *string
	EntryDate time.Time
	BirthDate *time.Time
	Note      *string
	Origin    sowdomain.Origin
	Parity    int
	BreedID   uuid.UUID
	CreatedBy uuid.UUID
}

type CreateSowService struct {
	repository ports.SowRepository
	clock      func() time.Time
}

// NewCreateSowService builds a create-sow use case with its repository and clock.
// It returns a service ready to execute CreateSowCommand values.
func NewCreateSowService(repository ports.SowRepository, clock func() time.Time) *CreateSowService {
	return &CreateSowService{repository: repository, clock: clock}
}

// Execute creates a sow after ensuring its code is free.
// Every new sow starts as StateAlive and returns ErrSowCodeAlreadyExists when
// the code is already taken.
func (s *CreateSowService) Execute(ctx context.Context, command CreateSowCommand) (*sowdomain.Sow, error) {
	code := strings.TrimSpace(command.Code)
	exists, err := s.repository.ExistsByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSowCodeAlreadyExists
	}

	newSow, err := sowdomain.NewSow(sowdomain.NewSowParams{
		ID:        uuid.New(),
		Code:      code,
		Location:  command.Location,
		EntryDate: command.EntryDate,
		BirthDate: command.BirthDate,
		Note:      command.Note,
		Origin:    command.Origin,
		Parity:    command.Parity,
		BreedID:   command.BreedID,
		CreatedBy: command.CreatedBy,
	}, s.clock())
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newSow); err != nil {
		return nil, err
	}
	return newSow, nil
}
