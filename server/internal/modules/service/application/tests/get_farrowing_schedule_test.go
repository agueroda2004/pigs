package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	serviceapplication "server/internal/modules/service/application"
	servicedomain "server/internal/modules/service/domain"
)

func TestGetFarrowingSchedule(t *testing.T) {
	t.Run("forwards the reference date taken from the clock", func(t *testing.T) {
		now := time.Date(2026, time.June, 15, 23, 30, 0, 0, time.UTC)
		repository := &fakeServiceRepository{}
		service := serviceapplication.NewGetFarrowingScheduleService(repository, func() time.Time { return now })

		if _, err := service.Execute(context.Background()); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !repository.scheduleCalled {
			t.Fatal("expected the repository to be called")
		}
		if !repository.scheduleReference.Equal(now) {
			t.Fatalf("reference = %v, want %v", repository.scheduleReference, now)
		}
	})

	t.Run("returns the repository schedule", func(t *testing.T) {
		average := 110.5
		averageDate := time.Date(2026, time.June, 22, 0, 0, 0, 0, time.UTC)
		expected := servicedomain.FarrowingSchedule{
			Upcoming: []servicedomain.FarrowingScheduleEntry{{
				ServiceID:                    uuid.New(),
				SowID:                        uuid.New(),
				SowCode:                      "C-001",
				ExpectedFarrowingDate:        time.Date(2026, time.June, 20, 0, 0, 0, 0, time.UTC),
				AverageGestationDays:         &average,
				AverageExpectedFarrowingDate: &averageDate,
			}},
			Overdue: []servicedomain.FarrowingScheduleEntry{},
		}
		repository := &fakeServiceRepository{schedule: &expected}
		service := serviceapplication.NewGetFarrowingScheduleService(repository, time.Now)

		result, err := service.Execute(context.Background())

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result != &expected {
			t.Fatalf("result = %#v, want %#v", result, &expected)
		}
		if len(result.Upcoming) != 1 || result.Upcoming[0].SowCode != "C-001" {
			t.Fatalf("unexpected upcoming entries: %#v", result.Upcoming)
		}
		if result.Upcoming[0].AverageExpectedFarrowingDate == nil ||
			!result.Upcoming[0].AverageExpectedFarrowingDate.Equal(averageDate) {
			t.Fatalf("expected average date %v, got %#v", averageDate, result.Upcoming[0].AverageExpectedFarrowingDate)
		}
		if result.Overdue == nil {
			t.Fatal("expected a non-nil overdue group")
		}
	})

	t.Run("propagates the repository error", func(t *testing.T) {
		unexpected := errors.New("unexpected")
		repository := &fakeServiceRepository{scheduleErr: unexpected}
		service := serviceapplication.NewGetFarrowingScheduleService(repository, time.Now)

		result, err := service.Execute(context.Background())

		if !errors.Is(err, unexpected) {
			t.Fatalf("error = %v, want %v", err, unexpected)
		}
		if result != nil {
			t.Fatalf("result = %#v, want nil", result)
		}
	})
}
