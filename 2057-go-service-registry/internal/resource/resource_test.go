package resource

import (
	"testing"

	"serviceregistry/internal/model"
	"serviceregistry/internal/storage"
)

func TestResource(t *testing.T) {
	store, _ := storage.NewSQLiteStorage(":memory:")
	defer store.Close()

	m := NewManager(store)

	t.Run("CreateAndGet", func(t *testing.T) {
		err := m.Create("res-1", model.ResourceService, "test-service")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		r, err := m.Get("res-1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if r == nil {
			t.Fatal("Resource should exist")
		}
		if r.Name != "test-service" {
			t.Fatalf("Expected name 'test-service', got %s", r.Name)
		}
	})

	t.Run("Associations", func(t *testing.T) {
		err := m.AddAssociation("res-1", "inst-1", model.ResourceInstance, "register")
		if err != nil {
			t.Fatalf("AddAssociation failed: %v", err)
		}

		summary, err := m.GetSummary("res-1")
		if err != nil {
			t.Fatalf("GetSummary failed: %v", err)
		}
		if len(summary) != 1 {
			t.Fatalf("Expected 1 association, got %d", len(summary))
		}
		if summary[0].Operation != "register" {
			t.Fatalf("Expected operation 'register', got %s", summary[0].Operation)
		}
	})
}
