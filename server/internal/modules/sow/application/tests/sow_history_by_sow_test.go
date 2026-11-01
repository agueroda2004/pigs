package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
)

func TestSowHistoryBySowExecute(t *testing.T) {
	t.Run("returns the history with its computed age", func(t *testing.T) {
		history := &sowdomain.SowHistoryBySow{Code: "C-001", EntryDate: "2026-01-10"}
		repository := &fakeSowRepository{historyBySow: history}
		clock := func() time.Time { return time.Date(2026, time.February, 15, 12, 0, 0, 0, time.UTC) }
		useCase := sowapplication.NewSowHistoryBySowService(repository, clock)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.historyBySowCode != "C-001" {
			t.Fatalf("code = %q, want C-001", repository.historyBySowCode)
		}
		if result.Age != "1 mes, 5 días" {
			t.Fatalf("age = %q, want 1 mes, 5 días", result.Age)
		}
	})

	t.Run("returns zero days when the entry date is invalid", func(t *testing.T) {
		history := &sowdomain.SowHistoryBySow{Code: "C-001", EntryDate: "not-a-date"}
		repository := &fakeSowRepository{historyBySow: history}
		clock := func() time.Time { return time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC) }
		useCase := sowapplication.NewSowHistoryBySowService(repository, clock)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Age != "0 días" {
			t.Fatalf("age = %q, want 0 días", result.Age)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakeSowRepository{historyBySowErr: errors.New("boom")}
		clock := func() time.Time { return time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC) }
		useCase := sowapplication.NewSowHistoryBySowService(repository, clock)

		_, err := useCase.Execute(context.Background(), "C-001")

		if !errors.Is(err, repository.historyBySowErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
