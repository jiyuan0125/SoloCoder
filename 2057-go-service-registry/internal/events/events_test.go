package events

import (
	"testing"
	"time"

	"serviceregistry/internal/model"
)

func TestEventBus(t *testing.T) {
	eb := NewEventBus()

	sub1 := NewChannelSubscriber(10)
	sub2 := NewChannelSubscriber(10)

	eb.Subscribe("test-service", sub1)
	eb.Subscribe("test-service", sub2)

	event := &model.ServiceEvent{
		Type:      model.EventRegistered,
		Instance: &model.ServiceInstance{
			ID:          "test-1",
			ServiceName: "test-service",
		},
		Timestamp: time.Now(),
	}

	eb.Publish(event)

	select {
	case e := <-sub1.Events():
		if e.Type != model.EventRegistered {
			t.Fatalf("Expected registered event, got %s", e.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	select {
	case e := <-sub2.Events():
		if e.Type != model.EventRegistered {
			t.Fatalf("Expected registered event, got %s", e.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	eb.Unsubscribe("test-service", sub1)

	event2 := &model.ServiceEvent{
		Type:      model.EventUnregistered,
		Instance: &model.ServiceInstance{
			ID:          "test-1",
			ServiceName: "test-service",
		},
		Timestamp: time.Now(),
	}

	eb.Publish(event2)

	select {
	case <-sub1.Events():
		t.Fatal("Should not receive event after unsubscribe")
	case <-time.After(100 * time.Millisecond):
	}

	select {
	case e := <-sub2.Events():
		if e.Type != model.EventUnregistered {
			t.Fatalf("Expected unregistered event, got %s", e.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestHasSubscribers(t *testing.T) {
	eb := NewEventBus()

	if eb.HasSubscribers("test-service") {
		t.Fatal("Should not have subscribers")
	}

	sub := NewChannelSubscriber(10)
	eb.Subscribe("test-service", sub)

	if !eb.HasSubscribers("test-service") {
		t.Fatal("Should have subscribers")
	}

	eb.Unsubscribe("test-service", sub)

	if eb.HasSubscribers("test-service") {
		t.Fatal("Should not have subscribers after unsubscribe")
	}
}
