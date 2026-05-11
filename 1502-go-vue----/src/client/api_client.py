import os
from typing import Any, Dict, Optional, List
from datetime import datetime

import requests


DEFAULT_HOST = os.environ.get("SERVER_HOST", "http://localhost")
DEFAULT_PORT = int(os.environ.get("SERVER_PORT", 8000))


class ApiClient:
    def __init__(self, host: Optional[str] = None, port: Optional[int] = None):
        self.base_url = f"{host or DEFAULT_HOST}:{port or DEFAULT_PORT}"

    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        try:
            response = requests.request(method, url, timeout=30, **kwargs)
            if response.status_code == 204:
                return {}
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            if hasattr(e, "response") and e.response is not None:
                detail = e.response.json().get("detail", str(e))
                raise Exception(f"API错误: {detail}")
            raise Exception(f"连接失败: {e}")

    def create_pond(self, code: str, area: float, species: str, initial_stock: int) -> Dict[str, Any]:
        return self._request(
            "POST",
            "/api/ponds",
            json={
                "code": code,
                "area": area,
                "species": species,
                "initial_stock": initial_stock,
            },
        )

    def list_ponds(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/ponds")

    def get_pond(self, pond_id: int) -> Dict[str, Any]:
        return self._request("GET", f"/api/ponds/{pond_id}")

    def update_pond(self, pond_id: int, area: Optional[float] = None, species: Optional[str] = None) -> Dict[str, Any]:
        data = {}
        if area is not None:
            data["area"] = area
        if species is not None:
            data["species"] = species
        return self._request("PATCH", f"/api/ponds/{pond_id}", json=data)

    def delete_pond(self, pond_id: int) -> None:
        self._request("DELETE", f"/api/ponds/{pond_id}")

    def set_threshold(self, pond_id: int, indicator: str, min_value: Optional[float], max_value: Optional[float]) -> Dict[str, Any]:
        data = {"indicator": indicator}
        if min_value is not None:
            data["min_value"] = min_value
        if max_value is not None:
            data["max_value"] = max_value
        return self._request("POST", f"/api/ponds/{pond_id}/thresholds", json=data)

    def list_thresholds(self, pond_id: int) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/ponds/{pond_id}/thresholds")

    def create_water_record(
        self, pond_id: int, temperature: float, dissolved_oxygen: float, ph: float, ammonia_nitrogen: float
    ) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/api/ponds/{pond_id}/water-records",
            json={
                "temperature": temperature,
                "dissolved_oxygen": dissolved_oxygen,
                "ph": ph,
                "ammonia_nitrogen": ammonia_nitrogen,
            },
        )

    def list_water_records(self, pond_id: int, limit: int = 100) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/ponds/{pond_id}/water-records?limit={limit}")

    def get_water_stats(self, pond_id: int, start_time: str, end_time: str) -> Dict[str, Any]:
        return self._request(
            "GET",
            f"/api/ponds/{pond_id}/water-stats?start_time={start_time}&end_time={end_time}",
        )

    def create_feeding_plan(self, pond_id: int, plan_date: str, plan_time: str, feed_amount: float) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/api/ponds/{pond_id}/feeding-plans",
            json={
                "plan_date": plan_date,
                "plan_time": plan_time,
                "feed_amount": feed_amount,
            },
        )

    def list_feeding_plans(self, pond_id: int, limit: int = 100) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/ponds/{pond_id}/feeding-plans?limit={limit}")

    def list_feeding_records(self, pond_id: int, limit: int = 100) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/ponds/{pond_id}/feeding-records?limit={limit}")

    def create_harvest(self, pond_id: int, species: str, quantity: int, weight: float) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/api/ponds/{pond_id}/harvests",
            json={
                "species": species,
                "quantity": quantity,
                "weight": weight,
            },
        )

    def list_harvests(self, pond_id: int, limit: int = 100) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/ponds/{pond_id}/harvests?limit={limit}")
