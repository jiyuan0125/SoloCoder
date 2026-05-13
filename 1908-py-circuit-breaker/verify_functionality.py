#!/usr/bin/env python3
import sys

sys.path.insert(0, '/home/baru/Work/Private/work01/SoloCoder/1908-py-circuit-breaker')

from circuit_breaker import CircuitBreaker, CircuitConfig, CircuitState


def test_basic_functionality():
    print("Testing basic circuit breaker functionality...")
    
    breaker = CircuitBreaker(service_name="test-service")
    breaker.is_configured = True
    breaker.config = CircuitConfig(
        failure_rate_threshold=0.5,
        slow_call_rate_threshold=0.6,
        slow_call_threshold_ms=500,
        window_size=4,
        half_open_probes=2,
        open_duration_seconds=1,
    )
    from collections import deque
    breaker.call_window = deque(maxlen=breaker.config.window_size)
    
    assert breaker.state == CircuitState.CLOSED, "Initial state should be CLOSED"
    print("✓ Initial state is CLOSED")
    
    for i in range(4):
        breaker.record_call(is_success=False, duration_ms=100)
    
    assert breaker.state == CircuitState.OPEN, "Should open after 4 failures"
    print("✓ Opens after 4 failures (50% failure rate)")
    
    breaker.open_until = 0
    assert breaker.can_accept_call(), "Should accept call after timeout"
    print("✓ Accepts call after timeout")
    
    breaker.record_call(is_success=True, duration_ms=100)
    breaker.record_call(is_success=True, duration_ms=100)
    
    assert breaker.state == CircuitState.CLOSED, "Should close after 2 successful probes"
    print("✓ Closes after 2 successful probes")
    
    print("\nAll basic tests passed!")


def test_slow_call_trigger():
    print("\nTesting slow call trigger...")
    
    breaker = CircuitBreaker(service_name="test-slow")
    breaker.is_configured = True
    breaker.config = CircuitConfig(
        slow_call_rate_threshold=0.5,
        slow_call_threshold_ms=200,
        window_size=4,
        half_open_probes=1,
        open_duration_seconds=1,
    )
    from collections import deque
    breaker.call_window = deque(maxlen=breaker.config.window_size)
    
    for i in range(4):
        breaker.record_call(is_success=True, duration_ms=500)
    
    assert breaker.state == CircuitState.OPEN, "Should open due to slow calls"
    print("✓ Opens after 4 slow calls (50% threshold)")
    
    print("\nSlow call tests passed!")


if __name__ == "__main__":
    test_basic_functionality()
    test_slow_call_trigger()
    print("\n✓ All tests passed successfully!")
