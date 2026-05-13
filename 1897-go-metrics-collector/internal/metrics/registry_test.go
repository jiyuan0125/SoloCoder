package metrics

import (
	"testing"
)

func TestLabelsKey(t *testing.T) {
	labels := Labels{"method": "GET", "path": "/api"}
	key := LabelsKey(labels)
	if key != "method=GET,path=/api" && key != "path=/api,method=GET" {
		t.Errorf("unexpected labels key: %s", key)
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()

	err := r.Register("test_counter", TypeCounter)
	if err != nil {
		t.Errorf("register failed: %v", err)
	}

	desc, ok := r.Get("test_counter")
	if !ok {
		t.Error("metric not found after registration")
	}
	if desc.Type != TypeCounter {
		t.Errorf("unexpected type: %s", desc.Type)
	}
}

func TestRegistryTypeMismatch(t *testing.T) {
	r := NewRegistry()
	r.Register("test", TypeCounter)
	err := r.Register("test", TypeGauge)
	if err == nil {
		t.Error("expected type mismatch error")
	}
	if err != ErrMetricTypeMismatch {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegistryAddLabelSeries(t *testing.T) {
	r := NewRegistry()
	r.Register("test", TypeGauge)

	for i := 0; i < 5; i++ {
		labels := Labels{"i": string(rune('0' + i))}
		err := r.AddLabelSeries("test", LabelsKey(labels))
		if err != nil {
			t.Errorf("add label series failed: %v", err)
		}
	}

	if r.SeriesCount() != 5 {
		t.Errorf("expected 5 series, got %d", r.SeriesCount())
	}
}
