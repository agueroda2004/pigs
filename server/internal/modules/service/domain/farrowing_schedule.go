package service

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleUpcomingDays is the number of days after today included in the
// upcoming group. The upcoming window is inclusive of today and this offset.
const ScheduleUpcomingDays = 7

// FarrowingScheduleEntry is a read-model row for a confirmed service due to
// farrow. AverageGestationDays and AverageExpectedFarrowingDate are nil when the
// sow has no completed history.
type FarrowingScheduleEntry struct {
	ServiceID                    uuid.UUID
	SowID                        uuid.UUID
	SowCode                      string
	ExpectedFarrowingDate        time.Time
	AverageGestationDays         *float64
	AverageExpectedFarrowingDate *time.Time
}

// FarrowingSchedule groups the confirmed services that require attention.
// Upcoming covers the next seven days and overdue covers earlier dates.
type FarrowingSchedule struct {
	Upcoming []FarrowingScheduleEntry
	Overdue  []FarrowingScheduleEntry
}

// FarrowingScheduleRow pairs a schedule entry with the SQL date classification.
// It is used to assemble the final read model without reinterpreting dates.
type FarrowingScheduleRow struct {
	Entry   FarrowingScheduleEntry
	Overdue bool
}

// ScheduleDateRange returns the calendar dates used by the schedule query.
// The upcoming range is inclusive of today and today plus ScheduleUpcomingDays.
func ScheduleDateRange(reference time.Time) (today time.Time, upcomingEnd time.Time) {
	today = time.Date(reference.Year(), reference.Month(), reference.Day(), 0, 0, 0, 0, time.UTC)
	return today, today.AddDate(0, 0, ScheduleUpcomingDays)
}

// BuildFarrowingSchedule splits query rows into upcoming and overdue entries.
// It preserves the query order and returns empty slices instead of nil slices.
func BuildFarrowingSchedule(rows []FarrowingScheduleRow) FarrowingSchedule {
	schedule := FarrowingSchedule{
		Upcoming: make([]FarrowingScheduleEntry, 0),
		Overdue:  make([]FarrowingScheduleEntry, 0),
	}

	for _, row := range rows {
		if row.Overdue {
			schedule.Overdue = append(schedule.Overdue, row.Entry)
			continue
		}
		schedule.Upcoming = append(schedule.Upcoming, row.Entry)
	}

	return schedule
}
