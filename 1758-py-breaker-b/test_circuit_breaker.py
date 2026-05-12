import asyncio
import pytest

from circuit_breaker import (
    CircuitBreaker,
    CircuitBreakerManager,
    CircuitBreakerOpenError,
    CircuitState,
)


class TestCircuitBreaker:
    @pytest.mark.asyncio
    async def test_closed_state_success(self):
        breaker = CircuitBreaker("test_service", failure_threshold=3)
        async def success():
            await asyncio.sleep(0)
            return "success"
        result = await breaker.call(success)
        assert result == "success"
        assert breaker.get_status().state == CircuitState.CLOSED
        assert breaker.get_status().failure_count == 0

    @pytest.mark.asyncio
    async def test_closed_state_failure_then_success(self):
        breaker = CircuitBreaker("test_service", failure_threshold=3)
        call_count = 0

        async def failing_then_success():
            nonlocal call_count
            call_count += 1
            if call_count <= 2:
                raise Exception("error")
            return "success"

        with pytest.raises(Exception):
            await breaker.call(failing_then_success)
        with pytest.raises(Exception):
            await breaker.call(failing_then_success)

        result = await breaker.call(failing_then_success)
        assert result == "success"
        assert breaker.get_status().state == CircuitState.CLOSED

    @pytest.mark.asyncio
    async def test_triggers_open_after_failures(self):
        breaker = CircuitBreaker("test_service", failure_threshold=3)

        async def always_fail():
            raise Exception("error")

        for _ in range(2):
            with pytest.raises(Exception):
                await breaker.call(always_fail)
        assert breaker.get_status().state == CircuitState.CLOSED

        with pytest.raises(Exception):
            await breaker.call(always_fail)
        assert breaker.get_status().state == CircuitState.OPEN

    @pytest.mark.asyncio
    async def test_open_raises_circuit_breaker_error(self):
        breaker = CircuitBreaker("test_service", failure_threshold=2)

        async def always_fail():
            raise Exception("error")

        for _ in range(2):
            with pytest.raises(Exception):
                await breaker.call(always_fail)

        with pytest.raises(CircuitBreakerOpenError):
            await breaker.call(always_fail)

    @pytest.mark.asyncio
    async def test_fallback_used_when_open(self):
        breaker = CircuitBreaker("test_service", failure_threshold=2)

        async def always_fail():
            raise Exception("error")

        for _ in range(2):
            with pytest.raises(Exception):
                await breaker.call(always_fail)

        result = await breaker.call(always_fail, fallback="fallback_value")
        assert result == "fallback_value"

    @pytest.mark.asyncio
    async def test_half_open_then_back_to_open(self):
        breaker = CircuitBreaker(
            "test_service",
            failure_threshold=2,
            recovery_timeout=0.1,
            success_threshold=2,
        )

        async def always_fail():
            raise Exception("error")

        for _ in range(2):
            with pytest.raises(Exception):
                await breaker.call(always_fail)
        assert breaker.get_status().state == CircuitState.OPEN

        await asyncio.sleep(0.15)

        with pytest.raises(Exception):
            await breaker.call(always_fail)
        assert breaker.get_status().state == CircuitState.OPEN

    @pytest.mark.asyncio
    async def test_half_open_then_back_to_closed(self):
        breaker = CircuitBreaker(
            "test_service",
            failure_threshold=2,
            recovery_timeout=0.1,
            success_threshold=2,
        )

        async def always_fail():
            raise Exception("error")

        for _ in range(2):
            with pytest.raises(Exception):
                await breaker.call(always_fail)
        assert breaker.get_status().state == CircuitState.OPEN

        await asyncio.sleep(0.15)

        async def always_success():
            return "ok"

        result1 = await breaker.call(always_success)
        assert result1 == "ok"
        assert breaker.get_status().state == CircuitState.HALF_OPEN

        result2 = await breaker.call(always_success)
        assert result2 == "ok"
        assert breaker.get_status().state == CircuitState.CLOSED


class TestCircuitBreakerManager:
    def test_service_isolation(self):
        manager = CircuitBreakerManager(failure_threshold=3)
        b1 = manager.get_breaker("service_a")
        b2 = manager.get_breaker("service_b")
        assert b1.service == "service_a"
        assert b2.service == "service_b"
        assert b1 is not b2

    @pytest.mark.asyncio
    async def test_history_records_state_changes(self):
        manager = CircuitBreakerManager(failure_threshold=2, max_history=10)

        async def always_fail():
            raise Exception("error")

        for _ in range(2):
            with pytest.raises(Exception):
                await manager.call("service_x", always_fail)

        await asyncio.sleep(0.01)
        history = manager.get_history("service_x")
        assert len(history) >= 1
        assert history[-1].to_state == CircuitState.OPEN
