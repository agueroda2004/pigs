package tests

import (
	"context"
	"errors"
	"testing"

	farrowingapplication "server/internal/modules/farrowing/application"
	farrowingdomain "server/internal/modules/farrowing/domain"
)

func TestFarrowingHistoryBySowExecute(t *testing.T) {
	t.Run("returns the history with derived durations", func(t *testing.T) {
		start := "22:00"
		end := "02:30"
		items := []*farrowingdomain.FarrowingHistoryBySow{
			{FarrowDate: "2026-04-20", StartTime: &start, EndTime: &end},
		}
		repository := &fakeFarrowingRepository{historyItems: items}
		useCase := farrowingapplication.NewFarrowingHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.historyCode != "C-001" {
			t.Fatalf("sow code = %q, want C-001", repository.historyCode)
		}
		if len(result) != 1 || result[0].Duration == nil || *result[0].Duration != "04:30" {
			t.Fatalf("unexpected result: %#v", result)
		}
	})

	t.Run("returns an empty slice when there is no history", func(t *testing.T) {
		repository := &fakeFarrowingRepository{}
		useCase := farrowingapplication.NewFarrowingHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result == nil || len(result) != 0 {
			t.Fatalf("result = %#v, want an empty slice", result)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakeFarrowingRepository{historyErr: errors.New("boom")}
		useCase := farrowingapplication.NewFarrowingHistoryBySowService(repository)

		_, err := useCase.Execute(context.Background(), "C-001")

		if !errors.Is(err, repository.historyErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
