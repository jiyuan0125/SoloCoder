from __future__ import annotations

import os
from typing import Any, Dict, List, Optional

import httpx


class ColdChainClient:
    def __init__(
        self,
        base_url: Optional[str] = None,
        timeout: float = 30.0,
    ) -> None:
        self.base_url = base_url or os.environ.get(
            "COLD_CHAIN_BASE_URL", "http://localhost:8000"
        )
        self.timeout = timeout
        self._client: Optional[httpx.Client] = None

    def __enter__(self) -> "ColdChainClient":
        self._client = httpx.Client(base_url=self.base_url, timeout=self.timeout)
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        if self._client:
            self._client.close()
            self._client = None

    def _get_client(self) -> httpx.Client:
        if self._client is None:
            raise RuntimeError("Client is not initialized. Use 'with' statement.")
        return self._client

    def health(self) -> Dict[str, Any]:
        response = self._get_client().get("/health")
        response.raise_for_status()
        return response.json()

    def list_locations(self) -> List[Dict[str, Any]]:
        response = self._get_client().get("/locations")
        response.raise_for_status()
        return response.json()

    def create_location(self, location: Dict[str, Any]) -> Dict[str, Any]:
        response = self._get_client().post("/locations", json=location)
        response.raise_for_status()
        return response.json()

    def list_vehicles(self) -> List[Dict[str, Any]]:
        response = self._get_client().get("/vehicles")
        response.raise_for_status()
        return response.json()

    def get_vehicle(self, vehicle_id: str) -> Dict[str, Any]:
        response = self._get_client().get(f"/vehicles/{vehicle_id}")
        response.raise_for_status()
        return response.json()

    def create_vehicle(self, vehicle: Dict[str, Any]) -> Dict[str, Any]:
        response = self._get_client().post("/vehicles", json=vehicle)
        response.raise_for_status()
        return response.json()

    def list_tasks(self, status: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"status": status} if status else None
        response = self._get_client().get("/tasks", params=params)
        response.raise_for_status()
        return response.json()

    def get_task(self, task_id: str) -> Dict[str, Any]:
        response = self._get_client().get(f"/tasks/{task_id}")
        response.raise_for_status()
        return response.json()

    def create_task(self, task: Dict[str, Any]) -> Dict[str, Any]:
        response = self._get_client().post("/tasks", json=task)
        response.raise_for_status()
        return response.json()

    def assign_task(self, task_id: str) -> Dict[str, Any]:
        response = self._get_client().post(f"/tasks/{task_id}/assign")
        response.raise_for_status()
        return response.json()

    def start_task(self, task_id: str) -> Dict[str, Any]:
        response = self._get_client().post(f"/tasks/{task_id}/start")
        response.raise_for_status()
        return response.json()

    def complete_task(self, task_id: str) -> Dict[str, Any]:
        response = self._get_client().post(f"/tasks/{task_id}/complete")
        response.raise_for_status()
        return response.json()

    def report_temperature(
        self, task_id: str, vehicle_id: str, temperature: float
    ) -> Dict[str, Any]:
        params = {
            "task_id": task_id,
            "vehicle_id": vehicle_id,
            "temperature": temperature,
        }
        response = self._get_client().post(
            f"/tasks/{task_id}/report/temperature", params=params
        )
        response.raise_for_status()
        return response.json()

    def report_location(
        self, task_id: str, vehicle_id: str, latitude: float, longitude: float
    ) -> Dict[str, Any]:
        params = {
            "task_id": task_id,
            "vehicle_id": vehicle_id,
            "latitude": latitude,
            "longitude": longitude,
        }
        response = self._get_client().post(
            f"/tasks/{task_id}/report/location", params=params
        )
        response.raise_for_status()
        return response.json()

    def check_communication(self) -> List[Dict[str, Any]]:
        response = self._get_client().post("/system/check-communication")
        response.raise_for_status()
        return response.json()

    def export_delivery_records(self, start_date: str, end_date: str) -> str:
        params = {"start_date": start_date, "end_date": end_date}
        response = self._get_client().get(
            "/export/delivery-records", params=params
        )
        response.raise_for_status()
        return response.text
