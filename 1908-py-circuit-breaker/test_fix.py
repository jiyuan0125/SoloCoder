#!/usr/bin/env python3
import sys

sys.path.insert(0, '/home/baru/Work/Private/work01/SoloCoder/1908-py-circuit-breaker')

from breaker_manager import manager
from circuit_breaker import CircuitConfig, CircuitState


def test_configure_service():
    print("Testing configure service...")
    
    service_name = "test-config-service"
    config = CircuitConfig(
        failure_rate_threshold=0.5,
        slow_call_rate_threshold=0.6,
        slow_call_threshold_ms=500,
        window_size=10,
        half_open_probes=3,
    )
    
    breaker = manager.configure(service_name, config)
    
    assert breaker.service_name == service_name
    assert breaker.is_configured == True
    assert breaker.config.window_size == 10
    assert len(breaker.call_window) == 0
    assert breaker.call_window.maxlen == 10
    assert breaker.state == CircuitState.CLOSED
    
    print("✓ Service configured successfully")


def test_full_circuit_breaker_flow():
    print("\nTesting full circuit breaker flow...")
    
    service_name = "test-full-flow"
    
    # Configure the service
    config = CircuitConfig(
        failure_rate_threshold=0.5,
        slow_call_rate_threshold=0.6,
        slow_call_threshold_ms=500,
        window_size=4,
        half_open_probes=2,
        open_duration_seconds=1,
    )
    manager.configure(service_name, config)
    breaker = manager.get_or_create(service_name)
    
    # Verify initial state
    assert breaker.state == CircuitState.CLOSED
    assert breaker.can_accept_call()
    print("✓ Initial state: CLOSED")
    
    # Record some calls - 50% failure rate should open the circuit
    for i in range(2):
        assert breaker.can_accept_call()
        breaker.record_call(is_success=False, duration_ms=100)
    
    for i in range(2):
        assert breaker.can_accept_call()
        breaker.record_call(is_success=False, duration_ms=100)
    
    # Circuit should be open now
    assert breaker.state == CircuitState.OPEN
    assert not breaker.can_accept_call()
    print("✓ Circuit opened after 4 failures")
    
    # Test that unconfigured service returns unconfigured status
    unconfigured = manager.get_or_create("unconfigured-service")
    assert unconfigured.is_configured == False
    print("✓ Unconfigured service returns correct status")
    
    print("\nAll tests passed!")


if __name__ == "__main__":
    test_configure_service()
    test_full_circuit_breaker_flow()
