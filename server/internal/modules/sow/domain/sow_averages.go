package sow

import (
	"fmt"
	"math"

	farrowingdomain "server/internal/modules/farrowing/domain"
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	weagingdomain "server/internal/modules/weaging/domain"
)

// SowHistoryAverages holds the 13 reproductive and operational averages calculated
// from a sow's history. Nullable metrics are nil when their source data is missing.
type SowHistoryAverages struct {
	AverageLiveBorn                   *float64
	AverageStillborn                  *float64
	AverageMummified                  *float64
	AverageLiveBornWeight             *float64
	AverageLitterWeight               *float64
	AverageStillbornWeight            *float64
	MortalityRatePercentage           *float64
	AverageDeathsPerFarrowing         *float64
	AverageDeadPigletWeight           *float64
	AverageWeanedPiglets              *float64
	AverageWeanedPigletWeight         *float64
	AverageGestationDays              *float64
	AverageWeaningAgeDays             *float64
	ManipulatedFarrowings             int
	TotalFarrowings                   int
	ManipulatedText                   string
	AverageFarrowingDuration          *string
	AverageFarrowingDurationFormatted *string
}

// CalculateSowHistoryAverages computes the reproductive averages from a sow's
// farrowings, piglet deaths, full weagings and partial weagings. It is a pure
// function and returns nil metrics instead of dividing by zero.
func CalculateSowHistoryAverages(
	farrowings []*farrowingdomain.FarrowingHistoryBySow,
	pigletDeaths []*pigletdeathdomain.PigletDeathHistoryBySow,
	weagings []*weagingdomain.WeagingHistoryBySow,
	partialWeagings []*partialweagingdomain.PartialWeagingHistoryBySow,
) SowHistoryAverages {
	result := SowHistoryAverages{}
	totalFarrowings := len(farrowings)
	result.TotalFarrowings = totalFarrowings

	manipulated := 0
	for _, farrowing := range farrowings {
		if farrowing.IsManipulated {
			manipulated++
		}
	}
	result.ManipulatedFarrowings = manipulated
	if totalFarrowings == 0 {
		result.ManipulatedText = "Sin partos registrados"
	} else {
		result.ManipulatedText = fmt.Sprintf("Esta cerda ha sido manipulada %d partos de %d totales", manipulated, totalFarrowings)
	}

	if totalFarrowings > 0 {
		var liveBornSum, stillbornSum, mummifiedSum int
		for _, farrowing := range farrowings {
			liveBornSum += farrowing.LiveBorn
			stillbornSum += farrowing.Stillborn
			mummifiedSum += farrowing.Mummified
		}
		liveBornAvg := float64(liveBornSum) / float64(totalFarrowings)
		stillbornAvg := float64(stillbornSum) / float64(totalFarrowings)
		mummifiedAvg := float64(mummifiedSum) / float64(totalFarrowings)
		liveBornRounded := roundTo(liveBornAvg, 1)
		stillbornRounded := roundTo(stillbornAvg, 1)
		mummifiedRounded := roundTo(mummifiedAvg, 1)
		result.AverageLiveBorn = &liveBornRounded
		result.AverageStillborn = &stillbornRounded
		result.AverageMummified = &mummifiedRounded
	}

	// Average piglet weight and average litter weight from farrowings with valid litter weight.
	var litterWeightSum float64
	litterWeightCount := 0
	var liveBornForWeight int
	for _, farrowing := range farrowings {
		if farrowing.LitterWeight == nil || *farrowing.LitterWeight <= 0 {
			continue
		}
		if farrowing.LiveBorn <= 0 {
			continue
		}
		litterWeightSum += *farrowing.LitterWeight
		litterWeightCount++
		liveBornForWeight += farrowing.LiveBorn
	}
	if litterWeightCount > 0 {
		avgPigletWeight := litterWeightSum / float64(liveBornForWeight)
		avgLitterWeight := litterWeightSum / float64(litterWeightCount)
		avgPigletWeight = roundTo(avgPigletWeight, 2)
		avgLitterWeight = roundTo(avgLitterWeight, 2)
		result.AverageLiveBornWeight = &avgPigletWeight
		result.AverageLitterWeight = &avgLitterWeight
	}

	// Average stillborn piglet weight.
	var stillbornWeightSum float64
	var stillbornForWeight int
	stillbornWeightCount := 0
	for _, farrowing := range farrowings {
		if farrowing.StillbornWeight == nil || *farrowing.StillbornWeight <= 0 {
			continue
		}
		if farrowing.Stillborn <= 0 {
			continue
		}
		stillbornWeightSum += *farrowing.StillbornWeight
		stillbornForWeight += farrowing.Stillborn
		stillbornWeightCount++
	}
	if stillbornWeightCount > 0 && stillbornForWeight > 0 {
		avgStillbornWeight := stillbornWeightSum / float64(stillbornForWeight)
		avgStillbornWeight = roundTo(avgStillbornWeight, 2)
		result.AverageStillbornWeight = &avgStillbornWeight
	}

	// Mortality rate percentage and average deaths per farrowing.
	var totalDeaths int
	var totalLiveBorn int
	for _, death := range pigletDeaths {
		totalDeaths += death.Quantity
	}
	for _, farrowing := range farrowings {
		totalLiveBorn += farrowing.LiveBorn
	}
	if totalLiveBorn > 0 {
		mortality := (float64(totalDeaths) / float64(totalLiveBorn)) * 100
		mortality = roundTo(mortality, 1)
		result.MortalityRatePercentage = &mortality
	}
	if totalFarrowings > 0 {
		avgDeathsPerFarrowing := roundTo(float64(totalDeaths)/float64(totalFarrowings), 1)
		result.AverageDeathsPerFarrowing = &avgDeathsPerFarrowing
	}

	// Average dead piglet weight.
	var deadWeightSum float64
	var deadQuantitySum int
	hasDeadWeight := false
	for _, death := range pigletDeaths {
		if death.Weight == nil || *death.Weight <= 0 {
			continue
		}
		deadWeightSum += *death.Weight
		deadQuantitySum += death.Quantity
		hasDeadWeight = true
	}
	if hasDeadWeight && deadQuantitySum > 0 {
		avgDeadWeight := deadWeightSum / float64(deadQuantitySum)
		avgDeadWeight = roundTo(avgDeadWeight, 2)
		result.AverageDeadPigletWeight = &avgDeadWeight
	}

	// Average weaned piglets.
	if totalFarrowings > 0 {
		var weanedSum int
		for _, weaging := range weagings {
			weanedSum += weaging.Quantity
		}
		for _, partial := range partialWeagings {
			weanedSum += partial.Quantity
		}
		avgWeaned := float64(weanedSum) / float64(totalFarrowings)
		avgWeaned = roundTo(avgWeaned, 1)
		result.AverageWeanedPiglets = &avgWeaned
	}

	// Average weaned piglet weight.
	var weanedWeightSum float64
	var weanedQuantitySum int
	hasWeanedWeight := false
	for _, weaging := range weagings {
		if weaging.TotalWeight == nil || *weaging.TotalWeight <= 0 {
			continue
		}
		weanedWeightSum += *weaging.TotalWeight
		weanedQuantitySum += weaging.Quantity
		hasWeanedWeight = true
	}
	for _, partial := range partialWeagings {
		if partial.TotalWeight == nil || *partial.TotalWeight <= 0 {
			continue
		}
		weanedWeightSum += *partial.TotalWeight
		weanedQuantitySum += partial.Quantity
		hasWeanedWeight = true
	}
	if hasWeanedWeight && weanedQuantitySum > 0 {
		avgWeanedWeight := weanedWeightSum / float64(weanedQuantitySum)
		avgWeanedWeight = roundTo(avgWeanedWeight, 2)
		result.AverageWeanedPigletWeight = &avgWeanedWeight
	}

	// Average gestation days.
	var gestationSum int
	gestationCount := 0
	for _, farrowing := range farrowings {
		if farrowing.GestationDays == nil {
			continue
		}
		gestationSum += *farrowing.GestationDays
		gestationCount++
	}
	if gestationCount > 0 {
		avgGestation := float64(gestationSum) / float64(gestationCount)
		avgGestation = roundTo(avgGestation, 1)
		result.AverageGestationDays = &avgGestation
	}

	// Average weaning age days.
	var weaningAgeSum int
	weaningAgeCount := 0
	for _, weaging := range weagings {
		if weaging.WeaningAgeDays == nil {
			continue
		}
		weaningAgeSum += *weaging.WeaningAgeDays
		weaningAgeCount++
	}
	for _, partial := range partialWeagings {
		if partial.WeaningAgeDays == nil {
			continue
		}
		weaningAgeSum += *partial.WeaningAgeDays
		weaningAgeCount++
	}
	if weaningAgeCount > 0 {
		avgWeaningAge := float64(weaningAgeSum) / float64(weaningAgeCount)
		avgWeaningAge = roundTo(avgWeaningAge, 1)
		result.AverageWeaningAgeDays = &avgWeaningAge
	}

	// Average farrowing duration.
	var durationMinutesSum int
	durationCount := 0
	for _, farrowing := range farrowings {
		if farrowing.Duration == nil {
			continue
		}
		minutes, ok := parseDurationMinutes(*farrowing.Duration)
		if !ok {
			continue
		}
		durationMinutesSum += minutes
		durationCount++
	}
	if durationCount > 0 {
		avgMinutes := int(math.Round(float64(durationMinutesSum) / float64(durationCount)))
		formatted := fmt.Sprintf("%02d:%02d", avgMinutes/60, avgMinutes%60)
		result.AverageFarrowingDuration = &formatted
		hours := avgMinutes / 60
		mins := avgMinutes % 60
		formattedHuman := fmt.Sprintf("%d h %d min", hours, mins)
		result.AverageFarrowingDurationFormatted = &formattedHuman
	}

	return result
}

// roundTo rounds a float to the given number of decimals.
func roundTo(value float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}

// parseDurationMinutes parses an HH:MM duration into total minutes.
func parseDurationMinutes(value string) (int, bool) {
	var hours, minutes int
	if _, err := fmt.Sscanf(value, "%d:%d", &hours, &minutes); err != nil {
		return 0, false
	}
	if hours < 0 || minutes < 0 || minutes > 59 {
		return 0, false
	}
	return hours*60 + minutes, true
}
