import asyncio
import sys
sys.path.insert(0, '.')

from circuit_breaker import CircuitBreakerManager
from models import DependencyConfig, CircuitBreakerState, DegradationState


async def test_basic_flow():
    cb = CircuitBreakerManager()
    cb.failure_threshold = 2
    cb.open_duration = 0.1

    print("=== Test 1: Basic circuit breaker ===")
    assert cb.can_call("service-a") == True
    print("✓ can_call returns True for new service")

    await cb.record_failure("service-a")
    status = cb.get_service_status("service-a")
    assert status.failure_count == 1
    assert status.circuit_breaker_state == CircuitBreakerState.CLOSED
    print("✓ 1 failure: circuit still closed")

    await cb.record_failure("service-a")
    status = cb.get_service_status("service-a")
    assert status.circuit_breaker_state == CircuitBreakerState.OPEN
    assert cb.can_call("service-a") == False
    print("✓ 2 failures (threshold): circuit opened")

    await asyncio.sleep(0.15)
    status = cb.get_service_status("service-a")
    assert status.circuit_breaker_state == CircuitBreakerState.HALF_OPEN
    print("✓ After open duration: circuit half-open")

    await cb.record_success("service-a")
    status = cb.get_service_status("service-a")
    assert status.circuit_breaker_state == CircuitBreakerState.CLOSED
    assert cb.can_call("service-a") == True
    print("✓ Half-open success: circuit closed")

    print("\n=== Test 2: Dependency degradation ===")
    config = DependencyConfig(
        upstream="api-gateway",
        downstream=["payment-service"],
        degraded_timeout=1.0
    )
    dep = cb.add_dependency(config)
    assert dep.upstream == "api-gateway"
    print("✓ Dependency created")

    cb.failure_threshold = 1
    await cb.record_failure("payment-service")
    status_downstream = cb.get_service_status("payment-service")
    status_upstream = cb.get_service_status("api-gateway")

    assert status_downstream.circuit_breaker_state == CircuitBreakerState.OPEN
    assert status_upstream.degradation_state == DegradationState.DEGRADED
    assert status_upstream.current_timeout == 1.0
    print("✓ Downstream opened, upstream degraded")
    print(f"✓ Upstream timeout changed from 5.0s to {status_upstream.current_timeout}s")

    assert cb.can_call("api-gateway") == True
    print("✓ Upstream still can call (not fully blocked)")

    await asyncio.sleep(0.15)
    status_downstream = cb.get_service_status("payment-service")
    assert status_downstream.circuit_breaker_state == CircuitBreakerState.HALF_OPEN
    print("✓ Downstream circuit half-open after timeout")

    await cb.record_success("payment-service")
    await asyncio.sleep(0.01)
    status_upstream = cb.get_service_status("api-gateway")
    assert status_upstream.degradation_state == DegradationState.NORMAL
    assert status_upstream.current_timeout == 5.0
    print("✓ Downstream recovered, upstream recovered")

    print("\n=== Test 3: Idempotent degradation ===")
    config2 = DependencyConfig(
        upstream="order-service",
        downstream=["inventory-service", "payment-service"],
        degraded_timeout=0.5
    )
    cb.add_dependency(config2)

    cb.failure_threshold = 1
    await cb.record_failure("inventory-service")
    status_order = cb.get_service_status("order-service")
    assert status_order.degradation_state == DegradationState.DEGRADED
    print("✓ First downstream fails: upstream degraded")

    await cb.record_failure("payment-service")
    status_order = cb.get_service_status("order-service")
    assert status_order.degradation_state == DegradationState.DEGRADED
    print("✓ Second downstream fails: no duplicate degradation")

    print("\n=== Test 4: Remove dependency ===")
    removed = cb.remove_dependency(dep.id)
    assert removed == True
    print("✓ Dependency removed")

    removed = cb.remove_dependency("non-existent-id")
    assert removed == False
    print("✓ Non-existent dependency returns False")

    print("\n=== All tests passed! ===")


if __name__ == "__main__":
    asyncio.run(test_basic_flow())
