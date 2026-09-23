package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

var ErrBoarCodeAlreadyExists = ports.ErrBoarCodeAlreadyUsed

type CreateBoarCommand struct {
	Code      string
	Location  *string
	EntryDate time.Time
	BirthDate *time.Time
	Note      *string
	Origin    boardomain.Origin
	BreedID   uuid.UUID
	CreatedBy uuid.UUID
}

type CreateBoarService struct {
	repository ports.BoarRepository
	clock      func() time.Time
}

// NewCreateBoarService builds a create-boar use case with its repository and clock.
// It returns a service ready to execute CreateBoarCommand values.
func NewCreateBoarService(repository ports.BoarRepository, clock func() time.Time) *CreateBoarService {
	return &CreateBoarService{repository: repository, clock: clock}
}

// Execute creates a boar after ensuring its code is free.
// Every new boar starts as StateAlive and returns ErrBoarCodeAlreadyExists when
// the code is already taken.
func (s *CreateBoarService) Execute(ctx context.Context, command CreateBoarCommand) (*boardomain.Boar, error) {
	code := strings.TrimSpace(command.Code)
	exists, err := s.repository.ExistsByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrBoarCodeAlreadyExists
	}

	newBoar, err := boardomain.NewBoar(boardomain.NewBoarParams{
		ID:        uuid.New(),
		Code:      code,
		Location:  command.Location,
		EntryDate: command.EntryDate,
		BirthDate: command.BirthDate,
		Note:      command.Note,
		Origin:    command.Origin,
		BreedID:   command.BreedID,
		CreatedBy: command.CreatedBy,
	}, s.clock())
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, newBoar); err != nil {
		return nil, err
	}
	return newBoar, nil
}
