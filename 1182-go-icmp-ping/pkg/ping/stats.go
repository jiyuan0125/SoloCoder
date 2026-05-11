package ping

import (
	"math"
	"time"
)

type Statistics struct {
	rtts       []time.Duration
	timeExceeded []*TimeExceededInfo
	sent       int
	received   int
}

func NewStatistics() *Statistics {
	return &Statistics{
		rtts:       make([]time.Duration, 0),
		timeExceeded: make([]*TimeExceededInfo, 0),
	}
}

func (s *Statistics) AddSent() {
	s.sent++
}

func (s *Statistics) AddReceived(rtt time.Duration) {
	s.received++
	s.rtts = append(s.rtts, rtt)
}

func (s *Statistics) AddTimeExceeded(te *TimeExceededInfo) {
	s.timeExceeded = append(s.timeExceeded, te)
}

func (s *Statistics) GetStats() *Stats {
	stats := &Stats{
		PacketsSent:     s.sent,
		PacketsReceived: s.received,
		TimeExceeded:    s.timeExceeded,
	}

	if s.sent > 0 {
		stats.PacketLoss = float64(s.sent-s.received) / float64(s.sent) * 100
	}

	if len(s.rtts) == 0 {
		return stats
	}

	var sum time.Duration
	min := s.rtts[0]
	max := s.rtts[0]

	for _, rtt := range s.rtts {
		sum += rtt
		if rtt < min {
			min = rtt
		}
		if rtt > max {
			max = rtt
		}
	}

	stats.MinRTT = min
	stats.MaxRTT = max
	stats.AvgRTT = sum / time.Duration(len(s.rtts))

	if len(s.rtts) > 1 {
		avgNs := stats.AvgRTT.Nanoseconds()
		var variance float64
		for _, rtt := range s.rtts {
			diff := float64(rtt.Nanoseconds() - avgNs)
			variance += diff * diff
		}
		variance /= float64(len(s.rtts) - 1)
		stats.StdDevRTT = time.Duration(math.Sqrt(variance))
	}

	return stats
}
