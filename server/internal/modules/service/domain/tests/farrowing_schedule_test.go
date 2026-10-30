package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"

	servicedomain "server/internal/modules/service/domain"
)

func scheduleEntry(serviceID uuid.UUID, sowCode string, expected time.Time, average *float64) servicedomain.FarrowingScheduleEntry {
	return servicedomain.FarrowingScheduleEntry{
		ServiceID:             serviceID,
		SowID:                 uuid.New(),
		SowCode:               sowCode,
		ExpectedFarrowingDate: expected,
		AverageGestationDays:  average,
	}
}

func TestScheduleDateRange(t *testing.T) {
	t.Run("uses the reference calendar date and includes seven days", func(t *testing.T) {
		reference := time.Date(2026, time.June, 15, 23, 30, 0, 0, time.FixedZone("UTC-5", -5*60*60))

		today, upcomingEnd := servicedomain.ScheduleDateRange(reference)

		wantToday := time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)
		wantEnd := time.Date(2026, time.June, 22, 0, 0, 0, 0, time.UTC)
		if !today.Equal(wantToday) {
			t.Fatalf("today = %v, want %v", today, wantToday)
		}
		if !upcomingEnd.Equal(wantEnd) {
			t.Fatalf("upcoming end = %v, want %v", upcomingEnd, wantEnd)
		}
	})
}

func TestBuildFarrowingSchedule(t *testing.T) {
	t.Run("returns non-nil empty groups without rows", func(t *testing.T) {
		schedule := servicedomain.BuildFarrowingSchedule(nil)

		if schedule.Upcoming == nil || schedule.Overdue == nil {
			t.Fatalf("expected non-nil groups, got %#v", schedule)
		}
		if len(schedule.Upcoming) != 0 || len(schedule.Overdue) != 0 {
			t.Fatalf("expected empty groups, got %#v", schedule)
		}
	})

	t.Run("splits rows and preserves their order", func(t *testing.T) {
		average := 110.5
		averageDate := time.Date(2026, time.June, 17, 0, 0, 0, 0, time.UTC)
		firstUpcoming := scheduleEntry(uuid.New(), "A-001", time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC), &average)
		firstUpcoming.AverageExpectedFarrowingDate = &averageDate
		firstOverdue := scheduleEntry(uuid.New(), "B-001", time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC), nil)
		secondUpcoming := scheduleEntry(uuid.New(), "A-002", time.Date(2026, time.June, 22, 0, 0, 0, 0, time.UTC), nil)
		secondOverdue := scheduleEntry(uuid.New(), "C-001", time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), nil)

		schedule := servicedomain.BuildFarrowingSchedule([]servicedomain.FarrowingScheduleRow{
			{Entry: firstUpcoming},
			{Entry: firstOverdue, Overdue: true},
			{Entry: secondUpcoming},
			{Entry: secondOverdue, Overdue: true},
		})

		if len(schedule.Upcoming) != 2 || schedule.Upcoming[0].SowCode != "A-001" || schedule.Upcoming[1].SowCode != "A-002" {
			t.Fatalf("unexpected upcoming entries: %#v", schedule.Upcoming)
		}
		if len(schedule.Overdue) != 2 || schedule.Overdue[0].SowCode != "B-001" || schedule.Overdue[1].SowCode != "C-001" {
			t.Fatalf("unexpected overdue entries: %#v", schedule.Overdue)
		}
		if schedule.Upcoming[0].AverageGestationDays == nil || *schedule.Upcoming[0].AverageGestationDays != average {
			t.Fatalf("expected average %v, got %#v", average, schedule.Upcoming[0].AverageGestationDays)
		}
		if schedule.Upcoming[1].AverageGestationDays != nil {
			t.Fatalf("expected nil average, got %#v", schedule.Upcoming[1].AverageGestationDays)
		}
		if schedule.Upcoming[0].AverageExpectedFarrowingDate == nil ||
			!schedule.Upcoming[0].AverageExpectedFarrowingDate.Equal(averageDate) {
			t.Fatalf("expected average date %v, got %#v", averageDate, schedule.Upcoming[0].AverageExpectedFarrowingDate)
		}
		if schedule.Overdue[0].AverageExpectedFarrowingDate != nil {
			t.Fatalf("expected nil average date, got %#v", schedule.Overdue[0].AverageExpectedFarrowingDate)
		}
	})
}
