package interval

import "time"

type IntervalType string

const (
	TypeInteger IntervalType = "integer"
	TypeTime    IntervalType = "time"
)

type IntegerInterval struct {
	Min int
	Max int
}

func NewIntegerInterval(min, max int) *IntegerInterval {
	if min > max {
		min, max = max, min
	}
	return &IntegerInterval{Min: min, Max: max}
}

func (i *IntegerInterval) Length() int {
	return i.Max - i.Min + 1
}

func (i *IntegerInterval) OverlapsWith(other *IntegerInterval) bool {
	return !(i.Max < other.Min || other.Max < i.Min)
}

func (i *IntegerInterval) IsAdjacentTo(other *IntegerInterval) bool {
	return i.Max+1 == other.Min || other.Max+1 == i.Min
}

func (i *IntegerInterval) Contains(n int) bool {
	return n >= i.Min && n <= i.Max
}

func (i *IntegerInterval) ContainsInterval(other *IntegerInterval) bool {
	return other.Min >= i.Min && other.Max <= i.Max
}

type TimeInterval struct {
	Start time.Time
	End   time.Time
	Zone  *time.Location
}

func NewTimeInterval(start, end time.Time, zone *time.Location) *TimeInterval {
	if start.After(end) {
		start, end = end, start
	}
	if zone == nil {
		zone = time.UTC
	}
	return &TimeInterval{
		Start: start.UTC(),
		End:   end.UTC(),
		Zone:  zone,
	}
}

func (t *TimeInterval) Duration() time.Duration {
	return t.End.Sub(t.Start)
}

func (t *TimeInterval) OverlapsWith(other *TimeInterval) bool {
	return !t.End.Before(other.Start) && !other.End.Before(t.Start)
}

func (t *TimeInterval) IsAdjacentTo(other *TimeInterval) bool {
	return t.End.Equal(other.Start) || other.End.Equal(t.Start)
}

func (t *TimeInterval) Contains(ts time.Time) bool {
	utc := ts.UTC()
	return (utc.After(t.Start) || utc.Equal(t.Start)) && (utc.Before(t.End) || utc.Equal(t.End))
}

func (t *TimeInterval) ContainsInterval(other *TimeInterval) bool {
	return (other.Start.After(t.Start) || other.Start.Equal(t.Start)) &&
		(other.End.Before(t.End) || other.End.Equal(t.End))
}
