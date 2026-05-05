package cron

import (
	"time"
)

func (c *CronExpr) Next(after time.Time) (time.Time, bool) {
	t := after.Add(time.Second).Truncate(time.Second)

	for i := 0; i < maxIterations; i++ {
		if c.matches(t) {
			return t, true
		}
		t = t.Add(time.Second)
	}

	return time.Time{}, false
}

func (c *CronExpr) NextN(after time.Time, n int) ([]time.Time, bool) {
	var results []time.Time
	t := after

	for i := 0; i < n; i++ {
		next, found := c.Next(t)
		if !found {
			return results, false
		}
		results = append(results, next)
		t = next
	}

	return results, true
}

func (c *CronExpr) matches(t time.Time) bool {
	second := t.Second()
	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	month := int(t.Month())
	weekday := int(t.Weekday())

	if !contains(c.Second.Range, second) {
		return false
	}
	if !contains(c.Minute.Range, minute) {
		return false
	}
	if !contains(c.Hour.Range, hour) {
		return false
	}
	if !contains(c.Month.Range, month) {
		return false
	}

	domSpecified := len(c.DayOfMonth.Range) < 31
	dowSpecified := len(c.DayOfWeek.Range) < 8

	if domSpecified && dowSpecified {
		domMatch := contains(c.DayOfMonth.Range, day)
		dowMatch := contains(c.DayOfWeek.Range, weekday)
		return domMatch || dowMatch
	} else if domSpecified {
		if !contains(c.DayOfMonth.Range, day) {
			return false
		}
	} else if dowSpecified {
		if !contains(c.DayOfWeek.Range, weekday) {
			return false
		}
	}

	return true
}

func (c *CronExpr) WillNeverFire() bool {
	domSpecified := len(c.DayOfMonth.Range) < 31

	if domSpecified {
		has29 := contains(c.DayOfMonth.Range, 29)
		if has29 {
			hasOtherDays := false
			for _, d := range c.DayOfMonth.Range {
				if d != 29 {
					hasOtherDays = true
					break
				}
			}
			if !hasOtherDays {
				monSpecified := len(c.Month.Range) < 12
				if monSpecified {
					hasFebruary := contains(c.Month.Range, 2)
					if hasFebruary {
						hasOtherMonths := false
						for _, m := range c.Month.Range {
							if m != 2 {
								hasOtherMonths = true
								break
							}
						}
						if !hasOtherMonths {
							return false
						}
					} else {
						for _, d := range c.DayOfMonth.Range {
							if d == 31 {
								hasValidMonthFor31 := false
								for _, m := range c.Month.Range {
									if m == 1 || m == 3 || m == 5 || m == 7 || m == 8 || m == 10 || m == 12 {
										hasValidMonthFor31 = true
										break
									}
								}
								if !hasValidMonthFor31 {
									return true
								}
							}
						}
					}
				}
			}
		}
	}

	domSpecified = len(c.DayOfMonth.Range) < 31
	dowSpecified := len(c.DayOfWeek.Range) < 8

	if domSpecified && dowSpecified {
		if len(c.DayOfMonth.Range) == 0 || len(c.DayOfWeek.Range) == 0 {
			return true
		}
	} else if domSpecified {
		if len(c.DayOfMonth.Range) == 0 {
			return true
		}
	} else if dowSpecified {
		if len(c.DayOfWeek.Range) == 0 {
			return true
		}
	}

	return false
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
