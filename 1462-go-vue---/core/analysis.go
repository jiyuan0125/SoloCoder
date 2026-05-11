package core

import (
	"fmt"
	"sort"
	"time"

	"energymanagement/common"
)

func (s *Service) GetPointConsumption(pointID string, start, end time.Time, granularity common.TimeGranularity) []common.EnergyConsumptionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := s.data[pointID]
	incrementByTime := s.aggregateByTime(data, start, end, granularity)
	return s.fillMissingTimeSlots(incrementByTime, start, end, granularity)
}

func (s *Service) GetAreaConsumption(area string, start, end time.Time, granularity common.TimeGranularity) []common.EnergyConsumptionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var allData []*common.EnergyData
	for _, p := range s.points {
		if p.Area == area {
			allData = append(allData, s.data[p.ID]...)
		}
	}

	incrementByTime := s.aggregateByTime(allData, start, end, granularity)
	return s.fillMissingTimeSlots(incrementByTime, start, end, granularity)
}

func (s *Service) GetTotalConsumption(start, end time.Time) map[common.EnergyType]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[common.EnergyType]float64)
	for _, p := range s.points {
		for _, d := range s.data[p.ID] {
			if (d.Timestamp.After(start) || d.Timestamp.Equal(start)) &&
				(d.Timestamp.Before(end) || d.Timestamp.Equal(end)) {
				result[p.EnergyType] += d.Increment
			}
		}
	}
	return result
}

func (s *Service) GetAreaComparison(start, end time.Time) []common.AreaConsumptionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	areaTotal := make(map[string]float64)
	var grandTotal float64

	for _, p := range s.points {
		for _, d := range s.data[p.ID] {
			if (d.Timestamp.After(start) || d.Timestamp.Equal(start)) &&
				(d.Timestamp.Before(end) || d.Timestamp.Equal(end)) {
				areaTotal[p.Area] += d.Increment
				grandTotal += d.Increment
			}
		}
	}

	result := make([]common.AreaConsumptionItem, 0, len(areaTotal))
	for area, total := range areaTotal {
		var ratio float64
		if grandTotal > 0 {
			ratio = total / grandTotal
		}
		result = append(result, common.AreaConsumptionItem{
			Area:  area,
			Value: total,
			Ratio: ratio,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})

	return result
}

func (s *Service) GetOverviewMetrics(refTime time.Time) common.OverviewMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currentStart := time.Date(refTime.Year(), refTime.Month(), 1, 0, 0, 0, 0, refTime.Location())
	currentEnd := currentStart.AddDate(0, 1, 0).Add(-time.Second)

	prevMonthStart := currentStart.AddDate(0, -1, 0)
	prevMonthEnd := currentStart.Add(-time.Second)

	prevYearStart := currentStart.AddDate(-1, 0, 0)
	prevYearEnd := prevMonthEnd.AddDate(-11, 0, 0)

	byEnergyType := s.calculateConsumptionByType(currentStart, currentEnd)
	prevMonthByType := s.calculateConsumptionByType(prevMonthStart, prevMonthEnd)
	prevYearByType := s.calculateConsumptionByType(prevYearStart, prevYearEnd)

	yoyChange := make(map[common.EnergyType]float64)
	yoyIsNew := make(map[common.EnergyType]bool)
	momChange := make(map[common.EnergyType]float64)
	momIsNew := make(map[common.EnergyType]bool)

	for et, current := range byEnergyType {
		momChange[et], momIsNew[et] = s.calculateChangeRate(current, prevMonthByType[et])
		yoyChange[et], yoyIsNew[et] = s.calculateChangeRate(current, prevYearByType[et])
	}

	return common.OverviewMetrics{
		ByEnergyType: byEnergyType,
		YoYChange:    yoyChange,
		YoYIsNew:     yoyIsNew,
		MoMChange:    momChange,
		MoMIsNew:     momIsNew,
	}
}

func (s *Service) aggregateByTime(data []*common.EnergyData, start, end time.Time, granularity common.TimeGranularity) map[string]float64 {
	result := make(map[string]float64)
	for _, d := range data {
		if (d.Timestamp.After(start) || d.Timestamp.Equal(start)) &&
			(d.Timestamp.Before(end) || d.Timestamp.Equal(end)) {
			key := s.getTimeKey(d.Timestamp, granularity)
			result[key] += d.Increment
		}
	}
	return result
}

func (s *Service) getTimeKey(t time.Time, granularity common.TimeGranularity) string {
	switch granularity {
	case common.GranularityHour:
		return t.Format("2006-01-02 15:00")
	case common.GranularityDay:
		return t.Format("2006-01-02")
	case common.GranularityMonth:
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02")
	}
}

func (s *Service) fillMissingTimeSlots(data map[string]float64, start, end time.Time, granularity common.TimeGranularity) []common.EnergyConsumptionItem {
	result := make([]common.EnergyConsumptionItem, 0)
	var step time.Duration

	switch granularity {
	case common.GranularityHour:
		step = time.Hour
	case common.GranularityDay:
		step = 24 * time.Hour
	case common.GranularityMonth:
		current := start
		for current.Before(end) || current.Equal(end) {
			key := current.Format("2006-01")
			value, hasData := data[key]
			result = append(result, common.EnergyConsumptionItem{
				Time:    key,
				Value:   value,
				HasData: hasData,
			})
			nextMonth := current.AddDate(0, 1, 0)
			current = nextMonth
		}
		return result
	default:
		step = 24 * time.Hour
	}

	current := start
	for current.Before(end) || current.Equal(end) {
		key := s.getTimeKey(current, granularity)
		value, hasData := data[key]
		result = append(result, common.EnergyConsumptionItem{
			Time:    key,
			Value:   value,
			HasData: hasData,
		})
		current = current.Add(step)
	}

	return result
}

func (s *Service) calculateConsumptionByType(start, end time.Time) map[common.EnergyType]float64 {
	result := make(map[common.EnergyType]float64)
	for _, p := range s.points {
		for _, d := range s.data[p.ID] {
			if (d.Timestamp.After(start) || d.Timestamp.Equal(start)) &&
				(d.Timestamp.Before(end) || d.Timestamp.Equal(end)) {
				result[p.EnergyType] += d.Increment
			}
		}
	}
	return result
}

func (s *Service) calculateChangeRate(current, previous float64) (float64, bool) {
	if previous == 0 {
		return 0, true
	}
	return ((current - previous) / previous) * 100, false
}

func (s *Service) GetSuggestions() []*common.Suggestion {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.Suggestion, 0)
	for _, sugs := range s.suggestions {
		result = append(result, sugs...)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

func (s *Service) checkAndGenerateSuggestions(pointID string, timestamp time.Time) {
	point := s.points[pointID]
	if point == nil {
		return
	}

	currentMonth := timestamp.Format("2006-01")
	prevMonth := timestamp.AddDate(0, -1, 0).Format("2006-01")

	s.checkUsageGrowthSuggestion(point, currentMonth, prevMonth)
	s.checkHighConsumptionSuggestion(point, currentMonth)
}

func (s *Service) checkUsageGrowthSuggestion(point *common.Point, currentMonth, prevMonth string) {
	if point.EnergyType != common.EnergyTypeElectricity {
		return
	}

	if s.hasSuggestion(point.ID, currentMonth, common.SuggestionTypeUsageGrowth) {
		return
	}

	currentUsage := s.getAreaMonthUsage(point.Area, currentMonth)
	prevUsage := s.getAreaMonthUsage(point.Area, prevMonth)

	if prevUsage > 0 && currentUsage > prevUsage*1.2 {
		s.addSuggestion(&common.Suggestion{
			ID:        generateID(point.ID + currentMonth + "growth"),
			PointID:   point.ID,
			Area:      point.Area,
			Type:      common.SuggestionTypeUsageGrowth,
			Message:   fmt.Sprintf("区域 %s 本月用电量较上月增长超过20%%，建议检查节能措施", point.Area),
			Month:     currentMonth,
			CreatedAt: time.Now(),
		})
	}
}

func (s *Service) checkHighConsumptionSuggestion(point *common.Point, currentMonth string) {
	if point.AreaSize <= 0 {
		return
	}

	if s.hasSuggestion(point.ID, currentMonth, common.SuggestionTypeHighConsumption) {
		return
	}

	pointUsage := s.getPointMonthUsage(point.ID, currentMonth)
	if pointUsage <= 0 {
		return
	}

	pointUnitConsumption := pointUsage / point.AreaSize

	avgUnitConsumption := s.getAvgUnitConsumption(point.EnergyType, currentMonth)
	if avgUnitConsumption <= 0 {
		return
	}

	if pointUnitConsumption > avgUnitConsumption*1.3 {
		s.addSuggestion(&common.Suggestion{
			ID:        generateID(point.ID + currentMonth + "high"),
			PointID:   point.ID,
			Area:      point.Area,
			Type:      common.SuggestionTypeHighConsumption,
			Message:   fmt.Sprintf("点位 %s 单位面积能耗高于同类型平均水平30%%以上，建议排查", point.Name),
			Month:     currentMonth,
			CreatedAt: time.Now(),
		})
	}
}

func (s *Service) hasSuggestion(pointID, month string, sugType common.SuggestionType) bool {
	for _, s := range s.suggestions[pointID] {
		if s.Month == month && s.Type == sugType {
			return true
		}
	}
	return false
}

func (s *Service) addSuggestion(sug *common.Suggestion) {
	if s.suggestions[sug.PointID] == nil {
		s.suggestions[sug.PointID] = make([]*common.Suggestion, 0)
	}
	s.suggestions[sug.PointID] = append(s.suggestions[sug.PointID], sug)
}

func (s *Service) getAreaMonthUsage(area, month string) float64 {
	var total float64
	for _, p := range s.points {
		if p.Area == area && p.EnergyType == common.EnergyTypeElectricity {
			total += s.getPointMonthUsage(p.ID, month)
		}
	}
	return total
}

func (s *Service) getPointMonthUsage(pointID, month string) float64 {
	var total float64
	for _, d := range s.data[pointID] {
		if d.Timestamp.Format("2006-01") == month {
			total += d.Increment
		}
	}
	return total
}

func (s *Service) getAvgUnitConsumption(energyType common.EnergyType, month string) float64 {
	var totalUsage float64
	var totalArea float64
	var count int

	for _, p := range s.points {
		if p.EnergyType == energyType && p.AreaSize > 0 {
			usage := s.getPointMonthUsage(p.ID, month)
			if usage > 0 {
				totalUsage += usage
				totalArea += p.AreaSize
				count++
			}
		}
	}

	if count == 0 || totalArea <= 0 {
		return 0
	}
	return totalUsage / totalArea
}
