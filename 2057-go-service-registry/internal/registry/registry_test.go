package registry

import (
	"testing"

	"serviceregistry/internal/events"
	"serviceregistry/internal/model"
	"serviceregistry/internal/storage"
)

func TestRegister(t *testing.T) {
	store, err := storage.NewSQLiteStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	eb := events.NewEventBus()
	reg := NewRegistry(store, eb)

	inst := &model.ServiceInstance{
		ServiceName: "test-service",
		Address:     "127.0.0.1",
		Port:        8080,
		Version:     "1.0.0",
		Weight:      3,
	}

	if err := reg.Register(inst, nil); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if err := reg.Register(inst, nil); err == nil {
		t.Fatal("Duplicate register should fail")
	}

	instances := reg.GetInstances("test-service")
	if len(instances) != 1 {
		t.Fatalf("Expected 1 instance, got %d", len(instances))
	}
}

func TestValidation(t *testing.T) {
	store, _ := storage.NewSQLiteStorage(":memory:")
	defer store.Close()

	reg := NewRegistry(store, events.NewEventBus())

	inst1 := &model.ServiceInstance{
		ServiceName: "test-service",
		Address:     "127.0.0.1",
	}

	if err := reg.Register(inst1, nil); err == nil {
		t.Fatal("Should fail without port")
	}

	inst2 := &model.ServiceInstance{
		ServiceName: "test-service",
		Address:     "127.0.0.1",
		Port:        8080,
	}

	if err := reg.Register(inst2, nil); err == nil {
		t.Fatal("Should fail without version")
	}
}

func TestUnregister(t *testing.T) {
	store, _ := storage.NewSQLiteStorage(":memory:")
	defer store.Close()

	reg := NewRegistry(store, events.NewEventBus())

	inst := &model.ServiceInstance{
		ServiceName: "test-service",
		Address:     "127.0.0.1",
		Port:        8080,
		Version:     "1.0.0",
	}

	if err := reg.Register(inst, nil); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if err := reg.Unregister(inst.ID, nil); err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	instances := reg.GetInstances("test-service")
	if len(instances) != 0 {
		t.Fatalf("Expected 0 instances after unregister, got %d", len(instances))
	}
}
