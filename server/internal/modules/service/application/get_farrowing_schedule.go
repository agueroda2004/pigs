package application

import (
	"context"
	"time"

	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
)

type GetFarrowingScheduleService struct {
	repository ports.ServiceRepository
	clock      func() time.Time
}

// NewGetFarrowingScheduleService builds the schedule use case with its repository and clock.
// It returns a read-only service ready to query confirmed services that require attention.
func NewGetFarrowingScheduleService(repository ports.ServiceRepository, clock func() time.Time) *GetFarrowingScheduleService {
	return &GetFarrowingScheduleService{repository: repository, clock: clock}
}

// Execute returns the confirmed services due to farrow from today through the
// next seven days plus the overdue ones, grouped by the repository read model.
func (s *GetFarrowingScheduleService) Execute(ctx context.Context) (*servicedomain.FarrowingSchedule, error) {
	return s.repository.GetFarrowingSchedule(ctx, s.clock())
}
