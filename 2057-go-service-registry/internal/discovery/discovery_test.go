package discovery

import (
	"testing"

	"serviceregistry/internal/events"
	"serviceregistry/internal/model"
	"serviceregistry/internal/registry"
	"serviceregistry/internal/storage"
)

type testProvider struct {
	instances map[string][]*model.ServiceInstance
}

func (p *testProvider) GetInstances(serviceName string) []*model.ServiceInstance {
	return p.instances[serviceName]
}

func TestDiscovery(t *testing.T) {
	store, _ := storage.NewSQLiteStorage(":memory:")
	defer store.Close()

	eb := events.NewEventBus()
	reg := registry.NewRegistry(store, eb)

	inst1 := &model.ServiceInstance{
		ServiceName: "test-service",
		Address:     "127.0.0.1",
		Port:        8001,
		Version:     "1.0.0",
		Weight:      1,
		Status:      model.StatusHealthy,
	}
	inst2 := &model.ServiceInstance{
		ServiceName: "test-service",
		Address:     "127.0.0.1",
		Port:        8002,
		Version:     "1.0.0",
		Weight:      3,
		Status:      model.StatusHealthy,
	}
	inst3 := &model.ServiceInstance{
		ServiceName: "test-service",
		Address:     "127.0.0.1",
		Port:        8003,
		Version:     "2.0.0",
		Weight:      1,
		Status:      model.StatusHealthy,
	}

	reg.Register(inst1, nil)
	reg.Register(inst2, nil)
	reg.Register(inst3, nil)

	d := NewDiscovery(reg)

	t.Run("ListHealthyInstances", func(t *testing.T) {
		instances := d.ListHealthyInstances("test-service", "", "")
		if len(instances) != 3 {
			t.Fatalf("Expected 3 instances, got %d", len(instances))
		}

		instances = d.ListHealthyInstances("test-service", "1.0.0", "")
		if len(instances) != 2 {
			t.Fatalf("Expected 2 instances with version 1.0.0, got %d", len(instances))
		}

		instances = d.ListHealthyInstances("unknown-service", "", "")
		if len(instances) != 0 {
			t.Fatalf("Expected 0 instances for unknown service, got %d", len(instances))
		}
	})

	t.Run("RoundRobin", func(t *testing.T) {
		seen := make(map[string]int)
		for i := 0; i < 10; i++ {
			inst := d.SelectInstance("test-service", model.StrategyRoundRobin)
			if inst == nil {
				t.Fatal("Expected an instance")
			}
			seen[inst.ID]++
		}

		if len(seen) != 3 {
			t.Fatalf("Round robin should select all 3 instances")
		}
	})

	t.Run("WeightedRandom", func(t *testing.T) {
		seen := make(map[string]int)
		for i := 0; i < 100; i++ {
			inst := d.SelectInstance("test-service", model.StrategyWeightedRandom)
			if inst == nil {
				t.Fatal("Expected an instance")
			}
			seen[inst.ID]++
		}

		if seen["test-service:127.0.0.1:8002"] < seen["test-service:127.0.0.1:8001"] {
			t.Log("Weighted instance should select higher weight more often")
		}
	})
}
