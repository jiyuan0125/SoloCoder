package metrics

import (
	"testing"
)

func TestCounterIncrement(t *testing.T) {
	s := NewStore()
	err := s.IncrementCounter("test", "key1", 1)
	if err != nil {
		t.Errorf("increment failed: %v", err)
	}
	err = s.IncrementCounter("test", "key1", 2)
	if err != nil {
		t.Errorf("increment failed: %v", err)
	}

	data, ok := s.GetCounter("test", "key1")
	if !ok {
		t.Error("counter not found")
	}
	if data.Value != 3 {
		t.Errorf("expected 3, got %v", data.Value)
	}
}

func TestCounterNegativeValue(t *testing.T) {
	s := NewStore()
	err := s.IncrementCounter("test", "key1", -1)
	if err == nil {
		t.Error("expected error for negative value")
	}
}

func TestGaugeSet(t *testing.T) {
	s := NewStore()
	s.SetGauge("test", "key1", 100)
	s.SetGauge("test", "key1", 200)

	data, ok := s.GetGauge("test", "key1")
	if !ok {
		t.Error("gauge not found")
	}
	if data.Value != 200 {
		t.Errorf("expected 200, got %v", data.Value)
	}
}

func TestHistogramObserve(t *testing.T) {
	s := NewStore()
	s.ObserveHistogram("test", "key1", 0.1)
	s.ObserveHistogram("test", "key1", 0.5)
	s.ObserveHistogram("test", "key1", 2.0)

	data, ok := s.GetHistogram("test", "key1")
	if !ok {
		t.Error("histogram not found")
	}
	if data.Count != 3 {
		t.Errorf("expected 3, got %d", data.Count)
	}
	if data.Sum != 2.6 {
		t.Errorf("expected 2.6, got %v", data.Sum)
	}
}

func TestHistogramPercentile(t *testing.T) {
	hist := NewHistogramData(DefaultBuckets)
	for i := 1; i <= 100; i++ {
		hist.Observe(float64(i))
	}

	p50 := hist.Percentile(0.50)
	p90 := hist.Percentile(0.90)
	p99 := hist.Percentile(0.99)

	if p50 < 49 || p50 > 51 {
		t.Errorf("expected P50 around 50, got %v", p50)
	}
	if p90 < 89 || p90 > 91 {
		t.Errorf("expected P90 around 90, got %v", p90)
	}
	if p99 < 98 || p99 > 100 {
		t.Errorf("expected P99 around 99, got %v", p99)
	}
}
