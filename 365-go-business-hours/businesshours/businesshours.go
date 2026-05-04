package businesshours

import (
	"time"
)

func NewBusinessHours() *BusinessHours {
	return &BusinessHours{
		Weekly:   make(WeeklySchedule),
		Holidays: make(HolidaySet),
	}
}

func (bh *BusinessHours) AddHoliday(date time.Time) {
	key := date.Format("2006-01-02")
	bh.Holidays[key] = struct{}{}
}

func (bh *BusinessHours) IsHoliday(date time.Time) bool {
	key := date.Format("2006-01-02")
	_, exists := bh.Holidays[key]
	return exists
}

func (bh *BusinessHours) SetDailySchedule(weekday time.Weekday, schedule *DailySchedule) {
	bh.Weekly[weekday] = schedule
}

func (bh *BusinessHours) Check(t time.Time) *CheckResult {
	slot, schedule := bh.findMatchingSlot(t)
	if slot != nil && !isTimeInBreaks(t, schedule.BreakSlots) {
		return &CheckResult{
			IsOpen:       true,
			NextOpenTime: t,
			TimeToOpen:   0,
			CurrentSlot:  slot,
		}
	}

	return bh.findNextOpenTime(t)
}

func (bh *BusinessHours) findMatchingSlot(t time.Time) (*TimeSlot, *DailySchedule) {
	tDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	if !bh.IsHoliday(tDay) {
		weekday := t.Weekday()
		if schedule, exists := bh.Weekly[weekday]; exists {
			for i := range schedule.OpenSlots {
				slot := &schedule.OpenSlots[i]
				if isTimeInSlotOnDay(t, slot, tDay) {
					return slot, schedule
				}
			}
		}
	}

	prevDay := tDay.AddDate(0, 0, -1)
	if !bh.IsHoliday(prevDay) {
		prevWeekday := prevDay.Weekday()
		if schedule, exists := bh.Weekly[prevWeekday]; exists {
			for i := range schedule.OpenSlots {
				slot := &schedule.OpenSlots[i]
				if isCrossDay(slot) && isTimeInCrossDaySlot(t, slot, prevDay) {
					return slot, schedule
				}
			}
		}
	}

	return nil, nil
}

func isTimeInSlotOnDay(t time.Time, slot *TimeSlot, day time.Time) bool {
	slotStart := adjustSlotTime(day, slot.Start)
	slotEnd := adjustSlotTime(day, slot.End)

	if isCrossDay(slot) {
		return !t.Before(slotStart) || t.Before(slotEnd)
	}

	return !t.Before(slotStart) && t.Before(slotEnd)
}

func isTimeInCrossDaySlot(t time.Time, slot *TimeSlot, prevDay time.Time) bool {
	prevDayEnd := adjustSlotTime(prevDay.AddDate(0, 0, 1), slot.End)
	return t.Before(prevDayEnd)
}

func (bh *BusinessHours) CalculateHours(start, end time.Time) *HoursResult {
	result := &HoursResult{
		TotalHours: 0,
		Slots:      []TimeSlot{},
	}

	if !start.Before(end) {
		return result
	}

	cursor := start
	for cursor.Before(end) {
		dayStart := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, cursor.Location())
		dayEnd := dayStart.Add(24 * time.Hour)
		if dayEnd.After(end) {
			dayEnd = end
		}

		prevDay := dayStart.AddDate(0, 0, -1)
		if !bh.IsHoliday(prevDay) {
			prevWeekday := prevDay.Weekday()
			if schedule, exists := bh.Weekly[prevWeekday]; exists {
				for _, slot := range schedule.OpenSlots {
					if isCrossDay(&slot) {
						prevDaySlotEnd := adjustSlotTime(dayStart, slot.End)
						if prevDaySlotEnd.After(dayStart) && dayStart.Before(dayEnd) {
							effectiveStart := maxTime(cursor, dayStart)
							effectiveEnd := minTime(prevDaySlotEnd, dayEnd)
							if effectiveStart.Before(effectiveEnd) {
								breakAdjusted := bh.applyBreakTimes(effectiveStart, effectiveEnd, schedule.BreakSlots, prevDay)
								for _, adjustedSlot := range breakAdjusted {
									duration := adjustedSlot.End.Sub(adjustedSlot.Start)
									result.TotalHours += duration.Hours()
									result.Slots = append(result.Slots, adjustedSlot)
								}
							}
						}
					}
				}
			}
		}

		if !bh.IsHoliday(dayStart) {
			weekday := dayStart.Weekday()
			if schedule, exists := bh.Weekly[weekday]; exists {
				for _, slot := range schedule.OpenSlots {
					slotStart := adjustSlotTime(dayStart, slot.Start)
					slotEnd := adjustSlotTime(dayStart, slot.End)

					if isCrossDay(&slot) {
						slotEnd = slotEnd.Add(24 * time.Hour)
					}

					effectiveStart := maxTime(maxTime(cursor, slotStart), dayStart)
					effectiveEnd := minTime(slotEnd, dayEnd)

					if effectiveStart.Before(effectiveEnd) {
						breakAdjusted := bh.applyBreakTimes(effectiveStart, effectiveEnd, schedule.BreakSlots, dayStart)
						for _, adjustedSlot := range breakAdjusted {
							duration := adjustedSlot.End.Sub(adjustedSlot.Start)
							result.TotalHours += duration.Hours()
							result.Slots = append(result.Slots, adjustedSlot)
						}
					}
				}
			}
		}

		cursor = dayEnd
	}

	return result
}

func (bh *BusinessHours) applyBreakTimes(start, end time.Time, breaks []TimeSlot, dayStart time.Time) []TimeSlot {
	if len(breaks) == 0 {
		return []TimeSlot{{Start: start, End: end}}
	}

	result := []TimeSlot{}
	currentStart := start

	for _, breakSlot := range breaks {
		breakStart := adjustSlotTime(dayStart, breakSlot.Start)
		breakEnd := adjustSlotTime(dayStart, breakSlot.End)

		if isCrossDay(&breakSlot) {
			breakEnd = breakEnd.Add(24 * time.Hour)
		}

		if breakEnd.Before(currentStart) || !breakStart.Before(end) {
			continue
		}

		if breakStart.After(currentStart) {
			result = append(result, TimeSlot{Start: currentStart, End: breakStart})
		}

		currentStart = maxTime(breakEnd, currentStart)
		if !currentStart.Before(end) {
			break
		}
	}

	if currentStart.Before(end) {
		result = append(result, TimeSlot{Start: currentStart, End: end})
	}

	return result
}

func (bh *BusinessHours) findNextOpenTime(t time.Time) *CheckResult {
	for dayOffset := 0; dayOffset < 14; dayOffset++ {
		checkDate := t.AddDate(0, 0, dayOffset)
		dayStart := time.Date(checkDate.Year(), checkDate.Month(), checkDate.Day(), 0, 0, 0, 0, checkDate.Location())

		if dayOffset == 0 {
			prevDay := dayStart.AddDate(0, 0, -1)
			if !bh.IsHoliday(prevDay) {
				prevWeekday := prevDay.Weekday()
				if schedule, exists := bh.Weekly[prevWeekday]; exists {
					for i := range schedule.OpenSlots {
						slot := &schedule.OpenSlots[i]
						if isCrossDay(slot) {
							slotEnd := adjustSlotTime(dayStart, slot.End)
							if t.Before(slotEnd) {
								nextStart := maxTime(t, dayStart)
								if !isTimeInBreaks(nextStart, schedule.BreakSlots) {
									return &CheckResult{
										IsOpen:       !t.Before(dayStart) && t.Before(slotEnd) && !isTimeInBreaks(t, schedule.BreakSlots),
										NextOpenTime: nextStart,
										TimeToOpen:   nextStart.Sub(t),
										CurrentSlot:  nil,
									}
								}
							}
						}
					}
				}
			}
		}

		if bh.IsHoliday(dayStart) {
			continue
		}

		weekday := dayStart.Weekday()
		schedule, exists := bh.Weekly[weekday]
		if !exists || len(schedule.OpenSlots) == 0 {
			continue
		}

		for i := range schedule.OpenSlots {
			slot := &schedule.OpenSlots[i]
			slotStart := adjustSlotTime(dayStart, slot.Start)
			slotEnd := adjustSlotTime(dayStart, slot.End)

			if isCrossDay(slot) {
				slotEnd = slotEnd.Add(24 * time.Hour)
			}

			if t.Before(slotEnd) {
				nextStart := maxTime(t, slotStart)

				if !isTimeInBreaks(nextStart, schedule.BreakSlots) {
					isOpenNow := !t.Before(slotStart) && t.Before(slotEnd) && !isTimeInBreaks(t, schedule.BreakSlots)
					return &CheckResult{
						IsOpen:       isOpenNow,
						NextOpenTime: nextStart,
						TimeToOpen:   nextStart.Sub(t),
						CurrentSlot:  nil,
					}
				} else {
					breakEnd := findBreakEnd(nextStart, schedule.BreakSlots, dayStart)
					if breakEnd.Before(slotEnd) {
						return &CheckResult{
							IsOpen:       false,
							NextOpenTime: breakEnd,
							TimeToOpen:   breakEnd.Sub(t),
							CurrentSlot:  nil,
						}
					}
				}
			}
		}
	}

	return &CheckResult{
		IsOpen:       false,
		NextOpenTime: time.Time{},
		TimeToOpen:   -1,
		CurrentSlot:  nil,
	}
}

func isTimeInSlot(t time.Time, slot *TimeSlot) bool {
	slotDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return isTimeInSlotOnDay(t, slot, slotDate)
}

func isTimeInBreaks(t time.Time, breaks []TimeSlot) bool {
	for i := range breaks {
		breakSlot := &breaks[i]
		if isTimeInSlot(t, breakSlot) {
			return true
		}
	}
	return false
}

func isCrossDay(slot *TimeSlot) bool {
	startMinutes := slot.Start.Hour()*60 + slot.Start.Minute()
	endMinutes := slot.End.Hour()*60 + slot.End.Minute()
	return startMinutes > endMinutes
}

func adjustSlotTime(date time.Time, slotTime time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		slotTime.Hour(), slotTime.Minute(), slotTime.Second(), slotTime.Nanosecond(),
		date.Location(),
	)
}

func findBreakEnd(t time.Time, breaks []TimeSlot, dayStart time.Time) time.Time {
	for i := range breaks {
		breakSlot := &breaks[i]
		breakStart := adjustSlotTime(dayStart, breakSlot.Start)
		breakEnd := adjustSlotTime(dayStart, breakSlot.End)

		if isCrossDay(breakSlot) {
			if t.Before(breakStart) {
				breakStart = breakStart.AddDate(0, 0, -1)
			} else {
				breakEnd = breakEnd.AddDate(0, 0, 1)
			}
		}

		if !t.Before(breakStart) && t.Before(breakEnd) {
			return breakEnd
		}
	}
	return t
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func CreateTime(hour, minute int) time.Time {
	return time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC)
}
