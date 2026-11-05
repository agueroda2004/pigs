package tests

import (
	"testing"

	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	sowdomain "server/internal/modules/sow/domain"
	weagingdomain "server/internal/modules/weaging/domain"
)

// TestCalculateSowHistoryAverages covers a sow with one farrowing and one piglet death.
func TestCalculateSowHistoryAverages(t *testing.T) {
	gestation := 114
	farrowings := []*farrowingdomain.FarrowingHistoryBySow{
		{
			FarrowDate:      "2026-04-01",
			Duration:        ptrString("04:30"),
			LiveBorn:        10,
			Stillborn:       1,
			Mummified:       1,
			LitterWeight:    ptrFloat(15.5),
			StillbornWeight: ptrFloat(2.5),
			IsManipulated:   true,
			GestationDays:   &gestation,
		},
	}
	weight := 2.5
	deaths := []*pigletdeathdomain.PigletDeathHistoryBySow{
		{Quantity: 3, Weight: &weight},
	}
	weaningAge := 21
	weagings := []*weagingdomain.WeagingHistoryBySow{
		{Quantity: 5, TotalWeight: ptrFloat(55.5), WeaningAgeDays: &weaningAge},
	}
	partial := []*partialweagingdomain.PartialWeagingHistoryBySow{}

	averages := sowdomain.CalculateSowHistoryAverages(farrowings, deaths, weagings, partial)

	if averages.TotalFarrowings != 1 || averages.ManipulatedFarrowings != 1 {
		t.Fatalf("unexpected farrowing counts: %+v", averages)
	}
	if averages.AverageLiveBorn == nil || *averages.AverageLiveBorn != 10 {
		t.Fatalf("average live born = %v", averages.AverageLiveBorn)
	}
	if averages.AverageFarrowingDuration == nil || *averages.AverageFarrowingDuration != "04:30" {
		t.Fatalf("duration = %v", averages.AverageFarrowingDuration)
	}
	if averages.AverageGestationDays == nil || *averages.AverageGestationDays != 114 {
		t.Fatalf("gestation = %v", averages.AverageGestationDays)
	}
	if averages.AverageWeaningAgeDays == nil || *averages.AverageWeaningAgeDays != 21 {
		t.Fatalf("weaning age = %v", averages.AverageWeaningAgeDays)
	}
}

// TestCalculateSowHistoryAveragesEmpty covers the zero-farrowing case.
func TestCalculateSowHistoryAveragesEmpty(t *testing.T) {
	averages := sowdomain.CalculateSowHistoryAverages(nil, nil, nil, nil)
	if averages.TotalFarrowings != 0 {
		t.Fatalf("expected zero farrowings")
	}
	if averages.AverageLiveBorn != nil || averages.AverageFarrowingDuration != nil {
		t.Fatalf("expected nil averages")
	}
	if averages.ManipulatedText != "Sin partos registrados" {
		t.Fatalf("unexpected text: %s", averages.ManipulatedText)
	}
}

func ptrFloat(value float64) *float64 { return &value }
func ptrString(value string) *string  { return &value }
