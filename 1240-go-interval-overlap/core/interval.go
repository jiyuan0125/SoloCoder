package core

import (
	"errors"
	"time"
)

type Interval struct {
	Start time.Time
	End   time.Time
}

func NewInterval(start, end time.Time) (*Interval, error) {
	if !start.Before(end) {
		return nil, errors.New("start time must be before end time")
	}
	return &Interval{
		Start: start.UTC(),
		End:   end.UTC(),
	}, nil
}

func (i *Interval) Overlaps(other *Interval) bool {
	if i == nil || other == nil {
		return false
	}
	return i.Start.Before(other.End) && other.Start.Before(i.End)
}

func (i *Interval) IsEmpty() bool {
	if i == nil {
		return true
	}
	return !i.Start.Before(i.End)
}

func (i *Interval) Contains(t time.Time) bool {
	if i == nil {
		return false
	}
	return (t.Equal(i.Start) || t.After(i.Start)) && t.Before(i.End)
}

func ParseInterval(startStr, endStr string) (*Interval, error) {
	start, err := parseTime(startStr)
	if err != nil {
		return nil, err
	}
	end, err := parseTime(endStr)
	if err != nil {
		return nil, err
	}
	return NewInterval(start, end)
}

func parseTime(timeStr string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", timeStr)
		if err == nil {
			t = t.UTC()
		}
	}
	return t, err
}
