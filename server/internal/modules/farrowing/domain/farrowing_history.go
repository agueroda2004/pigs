package farrowing

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FarrowingHistoryBySow is the read model returned for each farrowing in a sow history.
// It carries the farrowing result plus the operator and medication details shown in the timeline.
type FarrowingHistoryBySow struct {
	ID              uuid.UUID
	FarrowDate      string
	StartTime       *string
	EndTime         *string
	Duration        *string
	Location        *string
	LiveBorn        int
	Mummified       int
	Stillborn       int
	LitterWeight    *float64
	StillbornWeight *float64
	IsManipulated   bool
	Note            *string
	IsNurse         bool
	Operators       []FarrowingHistoryOperator
	Medications     []FarrowingHistoryMedication
}

// FarrowingHistoryOperator is the operator name shown for a farrowing history event.
// It only carries the display name because the history response does not expose ids.
type FarrowingHistoryOperator struct {
	OperatorName string
}

// FarrowingHistoryMedication is the medication detail shown for a farrowing history event.
// It records the medication name, its dose and the name of the operator that applied it.
type FarrowingHistoryMedication struct {
	MedicationName string
	Dose           float64
	AppliedByName  string
}

// FormatDuration returns the HH:MM duration between two optional HH:MM times.
// It returns nil when either time is missing or invalid and adds 24 hours when
// the end time is earlier than the start time, so overnight farrowings are valid.
func FormatDuration(start, end *string) *string {
	startMinutes, ok := parseTimeMinutes(start)
	if !ok {
		return nil
	}
	endMinutes, ok := parseTimeMinutes(end)
	if !ok {
		return nil
	}
	if endMinutes < startMinutes {
		endMinutes += 24 * 60
	}

	total := endMinutes - startMinutes
	formatted := fmt.Sprintf("%02d:%02d", total/60, total%60)
	return &formatted
}

// parseTimeMinutes parses an optional HH:MM time into minutes since midnight.
// It returns false when the value is nil, empty or does not match the HH:MM format.
func parseTimeMinutes(value *string) (int, bool) {
	if value == nil {
		return 0, false
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" || !timeLayout.MatchString(trimmed) {
		return 0, false
	}
	parsed, err := time.Parse("15:04", trimmed)
	if err != nil {
		return 0, false
	}
	return parsed.Hour()*60 + parsed.Minute(), true
}
