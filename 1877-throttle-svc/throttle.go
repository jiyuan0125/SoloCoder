package main

import (
	"net"
	"net/http"
	"strconv"
	"time"
)

type ThrottleResult struct {
	Allowed    bool
	RetryAfter time.Duration
	RuleID     string
	Dimension  Dimension
}

func (s *StatsStore) CheckAndRecord(key StatsKey, rule Rule) ThrottleResult {
	now := time.Now()
	currentCount, retryAfter := s.GetCurrentCount(key, rule, now)

	if currentCount >= int64(rule.MaxRequests) {
		s.RecordTrigger(key, rule.Dimension)
		return ThrottleResult{
			Allowed:    false,
			RetryAfter: retryAfter,
			RuleID:     rule.ID,
			Dimension:  rule.Dimension,
		}
	}

	exceeded := s.IncrementCount(key, rule, now)
	if exceeded {
		s.RecordTrigger(key, rule.Dimension)
		return ThrottleResult{
			Allowed:    false,
			RetryAfter: retryAfter,
			RuleID:     rule.ID,
			Dimension:  rule.Dimension,
		}
	}

	return ThrottleResult{
		Allowed:   true,
		RuleID:    rule.ID,
		Dimension: rule.Dimension,
	}
}

func getClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func getAPIKey(r *http.Request) string {
	return r.Header.Get("X-API-Key")
}

func intToSeconds(d time.Duration) string {
	sec := int64(d.Seconds())
	if sec <= 0 {
		sec = 1
	}
	return strconv.FormatInt(sec, 10)
}
