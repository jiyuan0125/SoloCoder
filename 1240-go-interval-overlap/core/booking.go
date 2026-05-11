package core

import (
	"crypto/rand"
	"encoding/hex"
)

type Booking struct {
	ID       string
	Resource string
	Interval *Interval
	Booker   string
}

func NewBooking(resource, booker string, interval *Interval) (*Booking, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	}
	return &Booking{
		ID:       id,
		Resource: resource,
		Interval: interval,
		Booker:   booker,
	}, nil
}

func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (b *Booking) Overlaps(other *Booking) bool {
	if b == nil || other == nil {
		return false
	}
	if b.Resource != other.Resource {
		return false
	}
	return b.Interval.Overlaps(other.Interval)
}

func (b *Booking) OverlapsInterval(interval *Interval) bool {
	if b == nil || interval == nil {
		return false
	}
	return b.Interval.Overlaps(interval)
}

func (b *Booking) IsWithinInterval(interval *Interval) bool {
	if b == nil || interval == nil || interval.IsEmpty() {
		return false
	}
	return b.Interval.Overlaps(interval)
}

type ByStartTime []*Booking

func (a ByStartTime) Len() int           { return len(a) }
func (a ByStartTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStartTime) Less(i, j int) bool { return a[i].Interval.Start.Before(a[j].Interval.Start) }
