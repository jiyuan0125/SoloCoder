package main

import (
	"testing"
	"time"
)

func TestCreateTask(t *testing.T) {
	s := NewScheduler()

	_, err := s.CreateTask(TaskRequest{Name: "A"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = s.CreateTask(TaskRequest{Name: "A"})
	if err == nil {
		t.Error("Expected error for duplicate task, got nil")
	}

	_, err = s.CreateTask(TaskRequest{Name: "B", Dependencies: []string{"NonExistent"}})
	if err == nil {
		t.Error("Expected error for non-existent dependency, got nil")
	}
}

func TestSelfDependency(t *testing.T) {
	s := NewScheduler()

	_, err := s.CreateTask(TaskRequest{Name: "A", Dependencies: []string{"A"}})
	if err == nil {
		t.Error("Expected error for self dependency, got nil")
	}
}

func TestCycleDetection(t *testing.T) {
	s := NewScheduler()

	_, err := s.CreateTask(TaskRequest{Name: "A"})
	if err != nil {
		t.Fatalf("Failed to create task A: %v", err)
	}

	_, err = s.CreateTask(TaskRequest{Name: "B", Dependencies: []string{"A"}})
	if err != nil {
		t.Fatalf("Failed to create task B: %v", err)
	}

	_, err = s.CreateTask(TaskRequest{Name: "C", Dependencies: []string{"B"}})
	if err != nil {
		t.Fatalf("Failed to create task C: %v", err)
	}
}

func TestTaskExecution(t *testing.T) {
	s := NewScheduler()

	_, err := s.CreateTask(TaskRequest{Name: "A", Timeout: 5})
	if err != nil {
		t.Fatalf("Failed to create task A: %v", err)
	}

	err = s.StartTask("A")
	if err != nil {
		t.Fatalf("Failed to start task A: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	task, ok := s.GetTask("A")
	if !ok {
		t.Fatal("Task A not found")
	}

	if task.Status != StatusSucceeded {
		t.Errorf("Expected task A to be SUCCEEDED, got %v", task.Status)
	}
}

func TestDependencyChain(t *testing.T) {
	s := NewScheduler()

	_, err := s.CreateTask(TaskRequest{Name: "A", Timeout: 5})
	if err != nil {
		t.Fatalf("Failed to create task A: %v", err)
	}

	_, err = s.CreateTask(TaskRequest{Name: "B", Dependencies: []string{"A"}, Timeout: 5})
	if err != nil {
		t.Fatalf("Failed to create task B: %v", err)
	}

	_, err = s.CreateTask(TaskRequest{Name: "C", Dependencies: []string{"B"}, Timeout: 5})
	if err != nil {
		t.Fatalf("Failed to create task C: %v", err)
	}

	err = s.StartTask("A")
	if err != nil {
		t.Fatalf("Failed to start task A: %v", err)
	}

	time.Sleep(1 * time.Second)

	taskA, _ := s.GetTask("A")
	if taskA.Status != StatusSucceeded {
		t.Errorf("Expected task A to be SUCCEEDED, got %v", taskA.Status)
	}

	taskB, _ := s.GetTask("B")
	if taskB.Status != StatusSucceeded {
		t.Errorf("Expected task B to be SUCCEEDED, got %v", taskB.Status)
	}

	taskC, _ := s.GetTask("C")
	if taskC.Status != StatusSucceeded {
		t.Errorf("Expected task C to be SUCCEEDED, got %v", taskC.Status)
	}
}

func TestGraphResponse(t *testing.T) {
	s := NewScheduler()

	_, err := s.CreateTask(TaskRequest{Name: "A"})
	if err != nil {
		t.Fatalf("Failed to create task A: %v", err)
	}

	_, err = s.CreateTask(TaskRequest{Name: "B", Dependencies: []string{"A"}})
	if err != nil {
		t.Fatalf("Failed to create task B: %v", err)
	}

	graph := s.GetGraph()

	if len(graph.Tasks) != 2 {
		t.Errorf("Expected 2 tasks in graph, got %d", len(graph.Tasks))
	}

	if len(graph.Edges) != 1 {
		t.Errorf("Expected 1 edge in graph, got %d", len(graph.Edges))
	}

	if graph.Edges[0].From != "A" || graph.Edges[0].To != "B" {
		t.Errorf("Expected edge A→B, got %s→%s", graph.Edges[0].From, graph.Edges[0].To)
	}
}
