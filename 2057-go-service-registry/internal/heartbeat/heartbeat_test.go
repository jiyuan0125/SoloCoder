package heartbeat

import (
	"testing"
	"time"

	"serviceregistry/internal/events"
	"serviceregistry/internal/model"
	"serviceregistry/internal/storage"
)

func TestHeartbeat(t *testing.T) {
	store, _ := storage.NewSQLiteStorage(":memory:")
	defer store.Close()

	eb := events.NewEventBus()

	eb.Subscribe("test-service", events.NewChannelSubscriber(10))

	hb := NewCheckerWithStorage(store, eb)

	inst := &model.ServiceInstance{
		ID:          "test:127.0.0.1:8000",
		ServiceName: "test-service",
		Address:     "127.0.0.1",
		Port:        8000,
		Version:     "1.0.0",
		Weight:      1,
		Status:      model.StatusHealthy,
		LastHeartbeat: time.Now().Add(-40 * time.Second),
		CreatedAt:   time.Now().Add(-1 * time.Hour),
	}

	store.SaveInstance(inst)

	if err := hb.CheckAll(); err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	deleted, err := store.GetInstance(inst.ID)
	if err != nil {
		t.Fatalf("GetInstance failed: %v", err)
	}
	if deleted != nil {
		t.Fatal("Instance should have been deleted due to timeout")
	}
}

func TestRecordHeartbeat(t *testing.T) {
	store, _ := storage.NewSQLiteStorage(":memory:")
	defer store.Close()

	eb := events.NewEventBus()
	hb := NewCheckerWithStorage(store, eb)

	inst := &model.ServiceInstance{
		ID:          "test:127.0.0.1:8000",
		ServiceName: "test-service",
		Address:     "127.0.0.1",
		Port:        8000,
		Version:     "1.0.0",
		Weight:      1,
		Status:      model.StatusHealthy,
		LastHeartbeat: time.Now().Add(-20 * time.Second),
		CreatedAt:   time.Now(),
	}
	store.SaveInstance(inst)

	before := inst.LastHeartbeat

	if err := hb.RecordHeartbeat(inst.ID); err != nil {
		t.Fatalf("RecordHeartbeat failed: %v", err)
	}

	updated, _ := store.GetInstance(inst.ID)
	if updated.LastHeartbeat.Before(before) {
		t.Fatal("Heartbeat should have been updated")
	}
}
