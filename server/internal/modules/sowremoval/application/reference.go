package application

import (
	"context"
	"errors"
	"time"

	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

// removalReference builds the date bounds a removal must respect for a sow.
// It picks the reference event from the given last state: the last weaning when
// the sow was weaned, the last abortion when it was aborted, the last service's
// mount date when it was gestating and only the entry date when it was alive.
// It also returns the service loaded for the gestating state so callers can
// fail or restore it; the service is nil for every other state.
func removalReference(
	ctx context.Context,
	repository ports.SowRemovalRepository,
	sow *sowdomain.Sow,
	lastState sowdomain.State,
) (sowremovaldomain.Reference, *servicedomain.Service, error) {
	reference := sowremovaldomain.Reference{EntryDate: sow.EntryDate}

	switch lastState {
	case sowdomain.StateWeaned:
		lastWeaning, err := repository.GetLastWeaningDate(ctx, sow.ID)
		if err != nil {
			return sowremovaldomain.Reference{}, nil, err
		}
		if lastWeaning != nil {
			reference.LastWeaningDate = *lastWeaning
		}
	case sowdomain.StateAborted:
		lastAbortion, err := repository.GetLastAbortion(ctx, sow.ID)
		if err != nil && !errors.Is(err, ports.ErrAbortionNotFound) {
			return sowremovaldomain.Reference{}, nil, err
		}
		if err == nil && lastAbortion != nil {
			reference.LastAbortionDate = lastAbortion.AbortionDate
		}
	case sowdomain.StatePregnant:
		service, err := repository.GetLastService(ctx, sow.ID)
		if err != nil && !errors.Is(err, ports.ErrServiceNotFound) {
			return sowremovaldomain.Reference{}, nil, err
		}
		if err != nil || service == nil {
			return reference, nil, nil
		}
		reference.LastMountDate = latestMountDate(service)
		return reference, service, nil
	}

	return reference, nil, nil
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
