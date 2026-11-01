package tests

import (
	"testing"

	farrowingdomain "server/internal/modules/farrowing/domain"
)

func durationStringPtr(value string) *string {
	return &value
}

func TestFormatDuration(t *testing.T) {
	t.Run("returns the difference between two times", func(t *testing.T) {
		result := farrowingdomain.FormatDuration(durationStringPtr("08:00"), durationStringPtr("10:30"))

		if result == nil || *result != "02:30" {
			t.Fatalf("duration = %v, want 02:30", result)
		}
	})

	t.Run("adds a day when the farrowing crosses midnight", func(t *testing.T) {
		result := farrowingdomain.FormatDuration(durationStringPtr("22:00"), durationStringPtr("02:30"))

		if result == nil || *result != "04:30" {
			t.Fatalf("duration = %v, want 04:30", result)
		}
	})

	t.Run("returns zero when both times are equal", func(t *testing.T) {
		result := farrowingdomain.FormatDuration(durationStringPtr("09:00"), durationStringPtr("09:00"))

		if result == nil || *result != "00:00" {
			t.Fatalf("duration = %v, want 00:00", result)
		}
	})

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		result := farrowingdomain.FormatDuration(durationStringPtr(" 08:00 "), durationStringPtr(" 10:00 "))

		if result == nil || *result != "02:00" {
			t.Fatalf("duration = %v, want 02:00", result)
		}
	})

	t.Run("returns nil when a time is missing", func(t *testing.T) {
		for _, testCase := range []struct {
			name  string
			start *string
			end   *string
		}{
			{name: "missing start", end: durationStringPtr("10:00")},
			{name: "missing end", start: durationStringPtr("10:00")},
			{name: "both missing"},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				if result := farrowingdomain.FormatDuration(testCase.start, testCase.end); result != nil {
					t.Fatalf("duration = %v, want nil", result)
				}
			})
		}
	})

	t.Run("returns nil when a time is invalid", func(t *testing.T) {
		for _, testCase := range []struct {
			name  string
			start *string
			end   *string
		}{
			{name: "empty start", start: durationStringPtr(""), end: durationStringPtr("10:00")},
			{name: "invalid end", start: durationStringPtr("10:00"), end: durationStringPtr("25:00")},
			{name: "malformed start", start: durationStringPtr("8"), end: durationStringPtr("10:00")},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				if result := farrowingdomain.FormatDuration(testCase.start, testCase.end); result != nil {
					t.Fatalf("duration = %v, want nil", result)
				}
			})
		}
	})
}
