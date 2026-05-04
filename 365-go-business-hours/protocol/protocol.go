package protocol

import (
	"time"
)

type CheckRequest struct {
	Time time.Time `json:"time"`
}

type CheckResponse struct {
	IsOpen        bool          `json:"is_open"`
	NextOpenTime  time.Time     `json:"next_open_time"`
	TimeToOpen    time.Duration `json:"time_to_open"`
	Error         string        `json:"error,omitempty"`
}

type CalculateHoursRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type CalculateHoursResponse struct {
	TotalHours float64    `json:"total_hours"`
	Slots      []TimeSlot `json:"slots"`
	Error      string     `json:"error,omitempty"`
}

type ConfigRequest struct {
	Weekly   WeeklyConfig   `json:"weekly"`
	Holidays []time.Time    `json:"holidays"`
}

type WeeklyConfig map[string]DaySchedule

type DaySchedule struct {
	OpenSlots  []SlotConfig `json:"open_slots"`
	BreakSlots []SlotConfig `json:"break_slots"`
}

type SlotConfig struct {
	StartHour   int `json:"start_hour"`
	StartMinute int `json:"start_minute"`
	EndHour     int `json:"end_hour"`
	EndMinute   int `json:"end_minute"`
}

type ConfigResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
