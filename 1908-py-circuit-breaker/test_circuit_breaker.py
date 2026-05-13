import pytest
from fastapi.testclient import TestClient

from main import app
from breaker_manager import manager
from circuit_breaker import CircuitBreaker, CircuitConfig, CircuitState, TriggerReason

client = TestClient(app)


@pytest.fixture(autouse=True)
def reset_manager():
    manager._breakers.clear()
    yield


def test_health_check():
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_get_unconfigured_service():
    response = client.get("/breakers/unknown-service")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "unconfigured"
    assert data["service_name"] == "unknown-service"
    assert data["events"] == []


def test_list_empty_breakers():
    response = client.get("/breakers")
    assert response.status_code == 200
    data = response.json()
    assert "services" in data
    assert data["services"] == []


def test_configure_service():
    response = client.put(
        "/breakers/test-service/config",
        json={
            "failure_rate_threshold": 0.5,
            "slow_call_rate_threshold": 0.6,
            "slow_call_threshold_ms": 500,
            "window_size": 10,
            "half_open_probes": 2,
        },
    )
    assert response.status_code == 200
    data = response.json()
    assert data["service_name"] == "test-service"
    assert data["config"]["failure_rate_threshold"] == 0.5
    assert data["config"]["slow_call_threshold_ms"] == 500


def test_failure_rate_triggers_open():
    client.put(
        "/breakers/test-failure/config",
        json={
            "failure_rate_threshold": 0.5,
            "window_size": 4,
        },
    )
    client.post("/breakers/test-failure/call?is_success=true&duration_ms=100")
    client.post("/breakers/test-failure/call?is_success=false&duration_ms=100")
    client.post("/breakers/test-failure/call?is_success=false&duration_ms=100")
    client.post("/breakers/test-failure/call?is_success=false&duration_ms=100")
    response = client.get("/breakers/test-failure")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "open"
    assert len(data["events"]) > 0
    assert data["events"][-1]["trigger_reason"] == "失败率"


def test_slow_call_rate_triggers_open():
    client.put(
        "/breakers/test-slow/config",
        json={
            "slow_call_rate_threshold": 0.5,
            "slow_call_threshold_ms": 200,
            "window_size": 4,
        },
    )
    for _ in range(4):
        client.post("/breakers/test-slow/call?is_success=true&duration_ms=500")
    response = client.get("/breakers/test-slow")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "open"
    assert len(data["events"]) > 0
    assert data["events"][-1]["trigger_reason"] == "慢调用比例"


def test_open_circuit_rejects_call():
    client.put(
        "/breakers/test-reject/config",
        json={
            "failure_rate_threshold": 0.5,
            "window_size": 2,
        },
    )
    client.post("/breakers/test-reject/call?is_success=false&duration_ms=100")
    client.post("/breakers/test-reject/call?is_success=false&duration_ms=100")
    response = client.post("/breakers/test-reject/call?is_success=true&duration_ms=100")
    assert response.status_code == 503


def test_list_breakers():
    client.put("/breakers/svc1/config", json={"half_open_probes": 3})
    client.put("/breakers/svc2/config", json={"half_open_probes": 5})
    response = client.get("/breakers")
    assert response.status_code == 200
    data = response.json()
    assert len(data["services"]) == 2
    service_names = [s["service_name"] for s in data["services"]]
    assert "svc1" in service_names
    assert "svc2" in service_names


def test_circuit_breaker_states():
    breaker = CircuitBreaker(service_name="test-cb")
    breaker.is_configured = True
    breaker.config = CircuitConfig(
        failure_rate_threshold=0.5,
        window_size=2,
        half_open_probes=1,
    )
    breaker.call_window.maxlen = breaker.config.window_size
    assert breaker.state == CircuitState.CLOSED
    breaker.record_call(is_success=False, duration_ms=100)
    breaker.record_call(is_success=False, duration_ms=100)
    assert breaker.state == CircuitState.OPEN
    breaker.open_until = 0
    assert breaker.can_accept_call()
    breaker.record_call(is_success=True, duration_ms=100)
    assert breaker.state == CircuitState.CLOSED


def test_half_open_probes_all_succeed():
    breaker = CircuitBreaker(service_name="test-half")
    breaker.is_configured = True
    breaker.config = CircuitConfig(
        failure_rate_threshold=0.5,
        window_size=2,
        half_open_probes=3,
    )
    breaker.call_window.maxlen = breaker.config.window_size
    breaker.record_call(is_success=False, duration_ms=100)
    breaker.record_call(is_success=False, duration_ms=100)
    assert breaker.state == CircuitState.OPEN
    breaker.open_until = 0
    for i in range(3):
        assert breaker.can_accept_call()
        breaker.record_call(is_success=True, duration_ms=100)
    assert breaker.state == CircuitState.CLOSED
    assert len(breaker.events) >= 3
    assert breaker.events[-1].trigger_reason == "探测成功"


def test_half_open_probe_fails():
    breaker = CircuitBreaker(service_name="test-fail")
    breaker.is_configured = True
    breaker.config = CircuitConfig(
        failure_rate_threshold=0.5,
        window_size=2,
        half_open_probes=3,
    )
    breaker.call_window.maxlen = breaker.config.window_size
    breaker.record_call(is_success=False, duration_ms=100)
    breaker.record_call(is_success=False, duration_ms=100)
    assert breaker.state == CircuitState.OPEN
    breaker.open_until = 0
    assert breaker.can_accept_call()
    breaker.record_call(is_success=False, duration_ms=100)
    assert breaker.state == CircuitState.OPEN
    assert breaker.events[-1].trigger_reason == "探测失败"


def test_update_config_runtime():
    client.put("/breakers/test-update/config", json={"half_open_probes": 3})
    response = client.get("/breakers/test-update")
    assert response.json()["config"]["half_open_probes"] == 3
    client.put("/breakers/test-update/config", json={"half_open_probes": 5})
    response = client.get("/breakers/test-update")
    assert response.json()["config"]["half_open_probes"] == 5


def test_call_unconfigured_service():
    response = client.post("/breakers/unconfigured/call?is_success=true&duration_ms=100")
    assert response.status_code == 404
