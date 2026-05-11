package benchmark

import (
	"github.com/example/benchcompare/pkg/api"
	"math"
)

const (
	changedThreshold    = 5.0
	significantThreshold = 20.0
)

func Compare(baselineText, currentText string) api.CompareResponse {
	baselineRaw := ParseOutput(baselineText)
	currentRaw := ParseOutput(currentText)
	
	baseline := AggregateBenchmarks(baselineRaw)
	current := AggregateBenchmarks(currentRaw)
	
	var items []api.ComparisonItem
	
	allNames := make(map[string]bool)
	for name := range baseline {
		allNames[name] = true
	}
	for name := range current {
		allNames[name] = true
	}
	
	for name := range allNames {
		baseResult, hasBase := baseline[name]
		currentResult, hasCurrent := current[name]
		
		item := api.ComparisonItem{
			Name: name,
		}
		
		if hasBase && hasCurrent {
			baseCopy := baseResult
			currentCopy := currentResult
			item.Baseline = &baseCopy
			item.Current = &currentCopy
			
			item.NsChangePercent = calculateChangePercent(baseResult.NsPerOp, currentResult.NsPerOp)
			item.MemChangePercent = 0.0
			if baseResult.BytesPerOp > 0 && currentResult.BytesPerOp > 0 {
				item.MemChangePercent = calculateChangePercent(float64(baseResult.BytesPerOp), float64(currentResult.BytesPerOp))
			}
			
			item.Status = determineStatus(item.NsChangePercent, item.MemChangePercent, 
				baseResult.Uncertain || currentResult.Uncertain)
		} else if hasCurrent {
			currentCopy := currentResult
			item.Current = &currentCopy
			item.Status = api.StatusAdded
		} else {
			baseCopy := baseResult
			item.Baseline = &baseCopy
			item.Status = api.StatusRemoved
		}
		
		items = append(items, item)
	}
	
	return api.CompareResponse{Items: items}
}

func calculateChangePercent(old, new float64) float64 {
	if old == 0 {
		return 0
	}
	return ((new - old) / old) * 100.0
}

func determineStatus(nsChange, memChange float64, uncertain bool) api.ChangeStatus {
	absNsChange := math.Abs(nsChange)
	absMemChange := math.Abs(memChange)
	maxChange := math.Max(absNsChange, absMemChange)
	
	if maxChange >= significantThreshold {
		if nsChange > 0 || memChange > 0 {
			return api.StatusRegression
		}
		return api.StatusImprovement
	}
	
	if maxChange >= changedThreshold {
		return api.StatusChanged
	}
	
	return api.StatusUnchanged
}
