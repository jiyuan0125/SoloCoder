package ical

import "time"

type Calendar struct {
	Version   string
	ProdID    string
	Events    []*Event
	Timezones map[string]*Timezone
}

type Event struct {
	UID         string
	Summary     string
	Description string
	Location    string
	Organizer   *Organizer
	DTStart     time.Time
	DTEnd       time.Time
	DTStamp     time.Time
	RRULE       *RRULE
	RDATEs      []time.Time
	EXDATEs     []time.Time
}

type Organizer struct {
	CN   string
	Mail string
}

type Timezone struct {
	TZID string
}

type RRULE struct {
	FREQ       Frequency
	INTERVAL   int
	COUNT      int
	UNTIL      time.Time
	BYDAY      []WeekdayPos
	BYMONTH    []int
	BYMONTHDAY []int
	BYSETPOS   []int
}

type Frequency string

const (
	FreqDAILY   Frequency = "DAILY"
	FreqWEEKLY  Frequency = "WEEKLY"
	FreqMONTHLY Frequency = "MONTHLY"
	FreqYEARLY  Frequency = "YEARLY"
)

type Weekday struct {
	Day time.Weekday
}

type WeekdayPos struct {
	Day   time.Weekday
	Pos   int
}

func NewWeekday(d time.Weekday) Weekday {
	return Weekday{Day: d}
}

func (w WeekdayPos) HasPos() bool {
	return w.Pos != 0
}
