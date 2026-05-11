import os
import httpx
from typing import Any, Optional


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("EPMS_SERVER_URL", "http://localhost:8000")
        self.api_prefix = "/api/v1"

    def _request(self, method: str, path: str, **kwargs) -> Any:
        url = f"{self.base_url}{self.api_prefix}{path}"
        with httpx.Client(timeout=30.0) as client:
            response = client.request(method, url, **kwargs)
            response.raise_for_status()
            if response.content:
                return response.json()
            return None

    def get(self, path: str, params: Optional[dict] = None) -> Any:
        return self._request("GET", path, params=params)

    def post(self, path: str, json: Optional[dict] = None) -> Any:
        return self._request("POST", path, json=json)

    def patch(self, path: str, json: Optional[dict] = None) -> Any:
        return self._request("PATCH", path, json=json)

    def create_enterprise(self, name: str) -> dict:
        return self.post("/enterprises", json={"name": name})

    def list_enterprises(self) -> list:
        return self.get("/enterprises")

    def get_enterprise(self, enterprise_id: int) -> dict:
        return self.get(f"/enterprises/{enterprise_id}")

    def create_facility(self, enterprise_id: int, name: str, facility_type: str, installed_at: str, is_running: bool = True) -> dict:
        return self.post("/facilities", json={
            "enterprise_id": enterprise_id,
            "name": name,
            "facility_type": facility_type,
            "installed_at": installed_at,
            "is_running": is_running
        })

    def list_facilities(self, enterprise_id: Optional[int] = None) -> list:
        params = {}
        if enterprise_id:
            params["enterprise_id"] = enterprise_id
        return self.get("/facilities", params=params if params else None)

    def get_facility(self, facility_id: int) -> dict:
        return self.get(f"/facilities/{facility_id}")

    def create_maintenance(self, facility_id: int, maintenance_date: str, description: Optional[str] = None, parts: Optional[list] = None) -> dict:
        return self.post("/maintenances", json={
            "facility_id": facility_id,
            "maintenance_date": maintenance_date,
            "description": description,
            "parts": parts or []
        })

    def list_maintenances(self, facility_id: Optional[int] = None, start_date: Optional[str] = None, end_date: Optional[str] = None) -> list:
        params = {}
        if facility_id:
            params["facility_id"] = facility_id
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        return self.get("/maintenances", params=params if params else None)

    def get_maintenance(self, maintenance_id: int) -> dict:
        return self.get(f"/maintenances/{maintenance_id}")

    def monthly_maintenance_cost(self, year: int, month: int, enterprise_id: Optional[int] = None) -> dict:
        params = {}
        if enterprise_id:
            params["enterprise_id"] = enterprise_id
        return self.get(f"/maintenances/monthly-cost/{year}/{month}", params=params if params else None)

    def create_emission(self, facility_id: int, recorded_at: str, pollutant: str, value: float, limit_value: float) -> dict:
        return self.post("/emissions", json={
            "facility_id": facility_id,
            "recorded_at": recorded_at,
            "pollutant": pollutant,
            "value": value,
            "limit_value": limit_value
        })

    def list_emissions(self, facility_id: Optional[int] = None) -> list:
        params = {}
        if facility_id:
            params["facility_id"] = facility_id
        return self.get("/emissions", params=params if params else None)

    def generate_maintenance_todos(self) -> dict:
        return self.post("/todos/generate-maintenance")

    def generate_compliance_todos(self) -> dict:
        return self.post("/todos/generate-compliance")

    def list_todos(self, facility_id: Optional[int] = None, status: Optional[str] = None) -> list:
        params = {}
        if facility_id:
            params["facility_id"] = facility_id
        if status:
            params["status"] = status
        return self.get("/todos", params=params if params else None)

    def update_todo_status(self, todo_id: int, status: str) -> dict:
        return self.patch(f"/todos/{todo_id}/status", json={"status": status})

    def list_alerts(self, facility_id: Optional[int] = None, resolved: Optional[bool] = None) -> list:
        params = {}
        if facility_id:
            params["facility_id"] = facility_id
        if resolved is not None:
            params["resolved"] = str(resolved).lower()
        return self.get("/alerts", params=params if params else None)

    def resolve_alert(self, alert_id: int) -> dict:
        return self.post(f"/alerts/{alert_id}/resolve")

    def monthly_stats(self, year: int, month: int) -> list:
        return self.get(f"/statistics/monthly/{year}/{month}")

    def facility_monthly_stats(self, year: int, month: int, facility_id: int) -> dict:
        return self.get(f"/statistics/monthly/{year}/{month}/{facility_id}")
