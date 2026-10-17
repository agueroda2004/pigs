package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	sowdomain "server/internal/modules/sow/domain"
)

// serviceReference builds the date bounds a service's mounts must respect.
// It picks the reference event from the given sow state: the last weaning when
// the sow is weaned, the last abortion when it is aborted, the previous service
// when it is gestating and only the entry date when it is alive. It also returns
// the previous service when one is loaded, so callers can transition it.
// The excluded service is skipped when looking for the previous service; pass
// uuid.Nil to include every service.
func serviceReference(
	ctx context.Context,
	repository ports.ServiceRepository,
	sow *sowdomain.Sow,
	state sowdomain.State,
	excludeID uuid.UUID,
) (*servicedomain.Service, servicedomain.Reference, error) {
	reference := servicedomain.Reference{EntryDate: sow.EntryDate}

	switch state {
	case sowdomain.StateWeaned:
		lastWeaning, err := repository.GetLastWeaningDate(ctx, sow.ID)
		if err != nil {
			return nil, servicedomain.Reference{}, err
		}
		if lastWeaning != nil {
			reference.LastWeaningDate = *lastWeaning
		}
	case sowdomain.StateAborted:
		lastAbortion, err := repository.GetLastAbortionDate(ctx, sow.ID)
		if err != nil {
			return nil, servicedomain.Reference{}, err
		}
		if lastAbortion != nil {
			reference.LastAbortionDate = *lastAbortion
		}
	case sowdomain.StatePregnant:
		previous, err := findPreviousService(ctx, repository, sow.ID, excludeID)
		if err != nil {
			return nil, servicedomain.Reference{}, err
		}
		if previous != nil {
			reference.PreviousMountDate = latestMountDate(previous)
		}
		return previous, reference, nil
	}

	return nil, reference, nil
}

// findPreviousService loads the most recent service of a sow, skipping the
// excluded service when its identifier is not nil. It returns nil when the sow
// has no service.
func findPreviousService(ctx context.Context, repository ports.ServiceRepository, sowID uuid.UUID, excludeID uuid.UUID) (*servicedomain.Service, error) {
	var (
		previous *servicedomain.Service
		err      error
	)
	if excludeID == uuid.Nil {
		previous, err = repository.GetLastService(ctx, sowID)
	} else {
		previous, err = repository.GetPreviousService(ctx, sowID, excludeID)
	}
	if err != nil && !errors.Is(err, ports.ErrServiceNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, nil
	}
	return previous, nil
}

// latestMountDate returns the most recent mount date of a service.
// It returns the zero time when the service has no mounts.
func latestMountDate(service *servicedomain.Service) time.Time {
	var latest time.Time
	for _, mount := range service.Mounts {
		if mount.MountDate.After(latest) {
			latest = mount.MountDate
		}
	}
	return latest
}
