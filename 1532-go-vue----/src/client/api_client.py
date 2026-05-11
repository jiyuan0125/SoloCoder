import os
import json
from typing import Optional, Dict, Any
import httpx


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("REACTOR_API_URL", "http://localhost:8000")
        self._client = httpx.Client(base_url=self.base_url, timeout=30.0, follow_redirects=True)

    def _get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        response = self._client.get(f"/api/v1{path}", params=params)
        response.raise_for_status()
        return response.json()

    def _post(self, path: str, json_data: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        response = self._client.post(f"/api/v1{path}", json=json_data)
        response.raise_for_status()
        return response.json()

    def _patch(self, path: str, json_data: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        response = self._client.patch(f"/api/v1{path}", json=json_data)
        response.raise_for_status()
        return response.json()

    def _delete(self, path: str) -> None:
        response = self._client.delete(f"/api/v1{path}")
        response.raise_for_status()

    def close(self):
        self._client.close()

    def list_reactors(self):
        return self._get("/reactors")

    def create_reactor(self, name: str, temp_max: float, press_max: float, description: str = ""):
        data = {
            "name": name,
            "description": description,
            "design_temperature_max": temp_max,
            "design_pressure_max": press_max
        }
        return self._post("/reactors", data)

    def get_reactor(self, reactor_id: str):
        return self._get(f"/reactors/{reactor_id}")

    def update_reactor(self, reactor_id: str, **kwargs):
        data = {k: v for k, v in kwargs.items() if v is not None}
        return self._patch(f"/reactors/{reactor_id}", data)

    def delete_reactor(self, reactor_id: str):
        self._delete(f"/reactors/{reactor_id}")

    def list_recipes(self):
        return self._get("/recipes")

    def create_recipe(self, name: str, materials: list, time_window: int, description: str = ""):
        data = {
            "name": name,
            "description": description,
            "materials": materials,
            "time_window_seconds": time_window
        }
        return self._post("/recipes", data)

    def get_recipe(self, recipe_id: str):
        return self._get(f"/recipes/{recipe_id}")

    def report_sensor_reading(self, reactor_id: str, temperature: float, pressure: float):
        data = {
            "reactor_id": reactor_id,
            "temperature": temperature,
            "pressure": pressure
        }
        return self._post("/sensors/reading", data)

    def list_readings(self, reactor_id: Optional[str] = None):
        params = {}
        if reactor_id:
            params["reactor_id"] = reactor_id
        return self._get("/sensors/readings", params)

    def list_alarms(self, reactor_id: Optional[str] = None, is_high_risk: Optional[bool] = None):
        params = {}
        if reactor_id:
            params["reactor_id"] = reactor_id
        if is_high_risk is not None:
            params["is_high_risk"] = is_high_risk
        return self._get("/sensors/alarms", params)

    def list_batches(self, reactor_id: Optional[str] = None, status_filter: Optional[str] = None):
        params = {}
        if reactor_id:
            params["reactor_id"] = reactor_id
        if status_filter:
            params["status_filter"] = status_filter
        return self._get("/batches", params)

    def create_batch(self, reactor_id: str, recipe_id: str):
        data = {
            "reactor_id": reactor_id,
            "recipe_id": recipe_id
        }
        return self._post("/batches", data)

    def get_batch(self, batch_id: str):
        return self._get(f"/batches/{batch_id}")

    def record_feeding(self, batch_id: str, material_name: str, req_amount: float,
                       act_amount: float, unit: str, order: int):
        data = {
            "batch_id": batch_id,
            "material_name": material_name,
            "required_amount": req_amount,
            "actual_amount": act_amount,
            "unit": unit,
            "order": order
        }
        return self._post(f"/batches/{batch_id}/feed", data)

    def complete_batch(self, batch_id: str):
        return self._post(f"/batches/{batch_id}/complete")

    def list_feedings(self, batch_id: str):
        return self._get(f"/batches/{batch_id}/feedings")

    def get_daily_stats(self, start_date: Optional[str] = None, end_date: Optional[str] = None):
        params = {}
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        return self._get("/statistics/daily", params)

    def calculate_daily_stats(self, date_str: str):
        return self._post(f"/statistics/daily/{date_str}")

    def get_scheduler_status(self):
        return self._get("/scheduler/status")

    def start_scheduler(self):
        return self._post("/scheduler/start")

    def stop_scheduler(self):
        return self._post("/scheduler/stop")
