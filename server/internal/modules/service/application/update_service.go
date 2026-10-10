package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	boardomain "server/internal/modules/boar/domain"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
)

// UpdateMountCommand carries the fields required to update a single mount.
// The identifier selects the mount; the rest of the fields replace its values.
type UpdateMountCommand struct {
	ID         uuid.UUID
	BoarID     uuid.UUID
	OperatorID uuid.UUID
	MountDate  time.Time
	Type       servicedomain.MountType
	Note       *string
}

// UpdateServiceCommand carries the editable fields of a service update.
// A nil location or note is left unchanged; the mounts are edited through the
// create, update and delete lists.
type UpdateServiceCommand struct {
	Location       *string
	Note           *string
	CreateMounts   []CreateMountCommand
	UpdateMounts   []UpdateMountCommand
	DeleteMountIDs []uuid.UUID
	UpdatedBy      uuid.UUID
}

type UpdateServiceService struct {
	repository ports.ServiceRepository
	clock      func() time.Time
}

// NewUpdateServiceService builds an update-service use case with its repository
// and clock, ready to execute UpdateServiceCommand values.
func NewUpdateServiceService(repository ports.ServiceRepository, clock func() time.Time) *UpdateServiceService {
	return &UpdateServiceService{repository: repository, clock: clock}
}

// Execute applies the provided fields and mount operations to a confirmed service.
// It reloads the reference dates, validates the boars and operators of every
// created or updated mount, then persists the service and its mounts atomically.
func (s *UpdateServiceService) Execute(ctx context.Context, id uuid.UUID, command UpdateServiceCommand) error {
	service, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := service.EnsureEditable(); err != nil {
		return err
	}

	sow, err := s.repository.GetSow(ctx, service.SowID)
	if err != nil {
		return err
	}

	reference := servicedomain.Reference{EntryDate: sow.EntryDate}
	previous, err := s.repository.GetPreviousService(ctx, service.SowID, service.ID)
	if err != nil && !errors.Is(err, ports.ErrServiceNotFound) {
		return err
	}
	if err == nil && previous != nil {
		reference.PreviousMountDate = latestMountDate(previous)
	}

	lastAbortion, err := s.repository.GetLastAbortionDate(ctx, service.SowID)
	if err != nil {
		return err
	}
	if lastAbortion != nil {
		reference.LastAbortionDate = *lastAbortion
	}

	createParams := make([]servicedomain.NewMountParams, 0, len(command.CreateMounts))
	for _, create := range command.CreateMounts {
		if err := s.validateMountReferences(ctx, create.BoarID, create.OperatorID, create.MountDate); err != nil {
			return err
		}
		createParams = append(createParams, servicedomain.NewMountParams{
			ID:         uuid.New(),
			BoarID:     create.BoarID,
			OperatorID: create.OperatorID,
			MountDate:  create.MountDate,
			Type:       create.Type,
			Note:       create.Note,
		})
	}

	updateParams := make([]servicedomain.UpdateMountParams, 0, len(command.UpdateMounts))
	for _, update := range command.UpdateMounts {
		if err := s.validateMountReferences(ctx, update.BoarID, update.OperatorID, update.MountDate); err != nil {
			return err
		}
		updateParams = append(updateParams, servicedomain.UpdateMountParams{
			ID:         update.ID,
			BoarID:     update.BoarID,
			OperatorID: update.OperatorID,
			MountDate:  update.MountDate,
			Type:       update.Type,
			Note:       update.Note,
		})
	}

	if err := service.Update(servicedomain.UpdateServiceParams{
		Location:       command.Location,
		Note:           command.Note,
		CreateMounts:   createParams,
		UpdateMounts:   updateParams,
		DeleteMountIDs: command.DeleteMountIDs,
	}, reference, command.UpdatedBy, s.clock()); err != nil {
		return err
	}

	return s.repository.Update(ctx, service)
}

// validateMountReferences checks that the mount's boar is alive and active, that
// the mount date is not before the boar entry date and that the operator is active.
func (s *UpdateServiceService) validateMountReferences(ctx context.Context, boarID, operatorID uuid.UUID, mountDate time.Time) error {
	boar, err := s.repository.GetBoar(ctx, boarID)
	if err != nil {
		return err
	}
	if !boar.Active || boar.State != boardomain.StateAlive {
		return ErrBoarNotEligible
	}
	if truncateToDay(mountDate).Before(truncateToDay(boar.EntryDate)) {
		return ErrMountBeforeBoarEntry
	}

	operator, err := s.repository.GetOperator(ctx, operatorID)
	if err != nil {
		return err
	}
	if !operator.Active {
		return ErrOperatorNotAvailable
	}
	return nil
}
