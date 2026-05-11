import os
import json
from datetime import datetime
from typing import List, Optional
import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_URL", "http://localhost:8000")

    def _post(self, endpoint: str, data: dict) -> dict:
        url = f"{self.base_url}{endpoint}"
        response = requests.post(url, json=data)
        response.raise_for_status()
        return response.json()

    def _get(self, endpoint: str, params: Optional[dict] = None) -> dict:
        url = f"{self.base_url}{endpoint}"
        response = requests.get(url, params=params)
        response.raise_for_status()
        return response.json()

    def create_greenhouse(self, code: str, area: float, crop: str, greenhouse_type: str) -> dict:
        return self._post(
            "/api/greenhouses",
            {"code": code, "area": area, "crop": crop, "type": greenhouse_type},
        )

    def list_greenhouses(self) -> List[dict]:
        return self._get("/api/greenhouses")

    def get_greenhouse(self, greenhouse_id: str) -> dict:
        return self._get(f"/api/greenhouses/{greenhouse_id}")

    def report_environment(
        self,
        greenhouse_id: str,
        temperature: float,
        humidity: float,
        soil_moisture: float,
        light: float,
    ) -> dict:
        return self._post(
            f"/api/greenhouses/{greenhouse_id}/environment",
            {
                "temperature": temperature,
                "humidity": humidity,
                "soil_moisture": soil_moisture,
                "light": light,
            },
        )

    def aggregate_environment(
        self,
        greenhouse_id: str,
        metric: str,
        start_time: datetime,
        end_time: datetime,
    ) -> dict:
        return self._get(
            f"/api/greenhouses/{greenhouse_id}/environment/aggregate",
            {
                "metric": metric,
                "start_time": start_time.isoformat(),
                "end_time": end_time.isoformat(),
            },
        )

    def create_irrigation_plan(
        self,
        greenhouse_id: str,
        execution_time: datetime,
        water_amount: float,
    ) -> dict:
        return self._post(
            f"/api/greenhouses/{greenhouse_id}/irrigation-plans",
            {
                "execution_time": execution_time.isoformat(),
                "water_amount": water_amount,
            },
        )

    def list_irrigation_plans(
        self,
        greenhouse_id: Optional[str] = None,
        status: Optional[str] = None,
    ) -> List[dict]:
        params = {}
        if greenhouse_id:
            params["greenhouse_id"] = greenhouse_id
        if status:
            params["status"] = status
        return self._get("/api/irrigation-plans", params=params if params else None)

    def create_fertilizer_plan(
        self,
        greenhouse_id: str,
        execution_time: datetime,
        fertilizer_amount: float,
    ) -> dict:
        return self._post(
            f"/api/greenhouses/{greenhouse_id}/fertilizer-plans",
            {
                "execution_time": execution_time.isoformat(),
                "fertilizer_amount": fertilizer_amount,
            },
        )

    def list_fertilizer_plans(
        self,
        greenhouse_id: Optional[str] = None,
        status: Optional[str] = None,
    ) -> List[dict]:
        params = {}
        if greenhouse_id:
            params["greenhouse_id"] = greenhouse_id
        if status:
            params["status"] = status
        return self._get("/api/fertilizer-plans", params=params if params else None)

    def run_scheduler(self) -> dict:
        return self._post("/api/scheduler/run", {})

    def health_check(self) -> dict:
        return self._get("/health")
