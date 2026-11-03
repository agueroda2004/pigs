package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	pigletdeathapplication "server/internal/modules/pigletdeath/application"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
)

func TestPigletDeathHistoryBySowExecute(t *testing.T) {
	t.Run("returns the history of the requested sow", func(t *testing.T) {
		note := "Muerte durante la noche"
		weight := 2.5
		items := []*pigletdeathdomain.PigletDeathHistoryBySow{
			{
				OperatorName: "Operador",
				DeathDate:    time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
				Quantity:     2,
				Weight:       &weight,
				Cause:        pigletdeathdomain.CauseCrushed,
				Turn:         pigletdeathdomain.TurnMorning,
				Note:         &note,
			},
		}
		repository := &fakePigletDeathRepository{historyItems: items}
		useCase := pigletdeathapplication.NewPigletDeathHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if repository.historyCode != "C-001" {
			t.Fatalf("sow code = %q, want C-001", repository.historyCode)
		}
		if len(result) != 1 || result[0].OperatorName != "Operador" || result[0].Quantity != 2 {
			t.Fatalf("unexpected result: %#v", result)
		}
		if result[0].Weight == nil || *result[0].Weight != weight {
			t.Fatalf("weight = %v, want %v", result[0].Weight, weight)
		}
	})

	t.Run("returns an empty slice when there is no history", func(t *testing.T) {
		repository := &fakePigletDeathRepository{}
		useCase := pigletdeathapplication.NewPigletDeathHistoryBySowService(repository)

		result, err := useCase.Execute(context.Background(), "C-001")

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result == nil || len(result) != 0 {
			t.Fatalf("result = %#v, want an empty slice", result)
		}
	})

	t.Run("propagates a persistence error", func(t *testing.T) {
		repository := &fakePigletDeathRepository{historyErr: errors.New("boom")}
		useCase := pigletdeathapplication.NewPigletDeathHistoryBySowService(repository)

		_, err := useCase.Execute(context.Background(), "C-001")

		if !errors.Is(err, repository.historyErr) {
			t.Fatalf("error = %v, want the persistence error", err)
		}
	})
}
