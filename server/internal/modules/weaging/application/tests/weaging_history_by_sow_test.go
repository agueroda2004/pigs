package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	weagingapplication "server/internal/modules/weaging/application"
	weagingdomain "server/internal/modules/weaging/domain"
)

func TestWeagingHistoryBySowExecute(t *testing.T) {
	t.Run("returns every full weaging for the requested sow", func(t *testing.T) {
		weight := 55.5
		note := "Destete completo"
		items := []*weagingdomain.WeagingHistoryBySow{
			{
				WeagingDate: time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC),
				Quantity:    5,
				TotalWeight: &weight,
				Note:        &note,
			},
		}
		repository := &fakeWeagingRepository{historyItems: items}
		useCase := weagingapplication.NewWeagingHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(result) != 1 || result[0] != items[0] {
			t.Fatalf("unexpected result: %#v", result)
		}
		if repository.historyCode != "C-001" {
			t.Fatalf("history code = %q, want C-001", repository.historyCode)
		}
	})

	t.Run("returns an empty slice when there are no records", func(t *testing.T) {
		repository := &fakeWeagingRepository{historyItems: nil}
		useCase := weagingapplication.NewWeagingHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result == nil {
			t.Fatal("result must not be nil")
		}
		if len(result) != 0 {
			t.Fatalf("unexpected result: %#v", result)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		repository := &fakeWeagingRepository{historyErr: errors.New("boom")}
		useCase := weagingapplication.NewWeagingHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if !errors.Is(err, repository.historyErr) {
			t.Fatalf("error = %v, want the repository error", err)
		}
		if result != nil {
			t.Fatalf("result = %#v, want nil", result)
		}
	})
}
