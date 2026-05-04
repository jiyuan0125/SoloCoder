package businesshours

import (
	"time"
)

type TimeSlot struct {
	Start time.Time
	End   time.Time
}

type DailySchedule struct {
	OpenSlots  []TimeSlot
	BreakSlots []TimeSlot
}

type WeeklySchedule map[time.Weekday]*DailySchedule

type HolidaySet map[string]struct{}

type BusinessHours struct {
	Weekly   WeeklySchedule
	Holidays HolidaySet
}

type CheckResult struct {
	IsOpen        bool
	NextOpenTime  time.Time
	TimeToOpen    time.Duration
	CurrentSlot   *TimeSlot
}

type HoursResult struct {
	TotalHours float64
	Slots      []TimeSlot
}
