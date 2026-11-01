package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

// ErrSowInactive indicates the sow is no longer part of the active herd.
var ErrSowInactive = errors.New("Esta cerda ya no se encuentra en la granja")

// SowHistoryBySowService builds the internal history-ready sow read model.
type SowHistoryBySowService struct {
	repository ports.SowRepository
	clock      func() time.Time
}

// NewSowHistoryBySowService creates the internal history query with its repository and clock.
func NewSowHistoryBySowService(repository ports.SowRepository, clock func() time.Time) *SowHistoryBySowService {
	return &SowHistoryBySowService{repository: repository, clock: clock}
}

// Execute returns the history-ready sow data for the requested code.
func (s *SowHistoryBySowService) Execute(ctx context.Context, code string) (*sowdomain.SowHistoryBySow, error) {
	history, err := s.repository.HistoryBySow(ctx, code)
	if err != nil {
		return nil, err
	}
	history.Age = s.age(history.EntryDate)
	return history, nil
}

// age calculates the calendar difference between entryDate and today.
func (s *SowHistoryBySowService) age(entryDate string) string {
	parsed, err := time.Parse("2006-01-02", entryDate)
	if err != nil {
		return "0 días"
	}
	entryDay := time.Date(parsed.UTC().Year(), parsed.UTC().Month(), parsed.UTC().Day(), 0, 0, 0, 0, time.UTC)
	today := s.today()
	if !entryDay.Before(today) {
		return "0 días"
	}

	years := today.Year() - entryDay.Year()
	months := int(today.Month()) - int(entryDay.Month())
	days := today.Day() - entryDay.Day()

	if days < 0 {
		previousMonth := today.AddDate(0, -1, 0)
		days += daysInMonth(previousMonth.Year(), previousMonth.Month())
		months--
	}
	if months < 0 {
		months += 12
		years--
	}

	parts := make([]string, 0, 3)
	if years > 0 {
		parts = append(parts, pluralize(years, "año", "años"))
	}
	if months > 0 {
		parts = append(parts, pluralize(months, "mes", "meses"))
	}
	if days > 0 || len(parts) == 0 {
		parts = append(parts, pluralize(days, "día", "días"))
	}
	if len(parts) == 0 {
		parts = append(parts, "0 días")
	}
	return strings.Join(parts, ", ")
}

func (s *SowHistoryBySowService) today() time.Time {
	now := s.clock().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func pluralize(value int, singular, plural string) string {
	if value == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", value, plural)
}
