import os
import json
from datetime import datetime, timedelta
from typing import Optional, Dict, Any

import httpx


class HydrologyClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get("API_BASE_URL", "http://localhost:8000/api")

    def _request(self, method: str, endpoint: str, **kwargs) -> httpx.Response:
        url = f"{self.base_url}{endpoint}"
        with httpx.Client() as client:
            response = client.request(method, url, **kwargs)
            response.raise_for_status()
            return response

    def health_check(self) -> Dict[str, Any]:
        with httpx.Client() as client:
            response = client.get(self.base_url.replace("/api", ""))
            return response.json()

    def create_station(self, name: str, station_type: str, device_model: Optional[str] = None) -> Dict[str, Any]:
        data = {"name": name, "station_type": station_type}
        if device_model:
            data["device_model"] = device_model
        return self._request("POST", "/stations", json=data).json()

    def list_stations(self) -> list:
        return self._request("GET", "/stations").json()

    def get_station(self, station_id: int) -> Dict[str, Any]:
        return self._request("GET", f"/stations/{station_id}").json()

    def update_station(self, station_id: int, **kwargs) -> Dict[str, Any]:
        return self._request("PATCH", f"/stations/{station_id}", json=kwargs).json()

    def delete_station(self, station_id: int) -> None:
        self._request("DELETE", f"/stations/{station_id}")

    def create_raw_data(
        self,
        station_id: int,
        data_type: str,
        value: float,
        timestamp: Optional[datetime] = None
    ) -> Dict[str, Any]:
        ts = timestamp or datetime.utcnow()
        data = {
            "data_type": data_type,
            "value": value,
            "timestamp": ts.isoformat()
        }
        return self._request("POST", f"/stations/{station_id}/raw-data", json=data).json()

    def get_raw_data(
        self,
        station_id: int,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        limit: int = 100
    ) -> list:
        params = {"limit": limit}
        if start_time:
            params["start_time"] = start_time.isoformat()
        if end_time:
            params["end_time"] = end_time.isoformat()
        return self._request("GET", f"/stations/{station_id}/raw-data", params=params).json()

    def get_aggregated_data(
        self,
        station_id: int,
        granularity: Optional[str] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        limit: int = 100
    ) -> list:
        params = {"limit": limit}
        if granularity:
            params["granularity"] = granularity
        if start_time:
            params["start_time"] = start_time.isoformat()
        if end_time:
            params["end_time"] = end_time.isoformat()
        return self._request("GET", f"/stations/{station_id}/aggregated", params=params).json()

    def recalculate_aggregations(
        self,
        station_id: int,
        start_time: datetime,
        end_time: datetime
    ) -> list:
        params = {
            "start_time": start_time.isoformat(),
            "end_time": end_time.isoformat()
        }
        return self._request("POST", f"/stations/{station_id}/aggregated/recalculate", params=params).json()

    def get_quality_records(self, station_id: int, limit: int = 100) -> list:
        return self._request(
            "GET", f"/stations/{station_id}/quality", params={"limit": limit}
        ).json()

    def review_quality_record(
        self,
        station_id: int,
        record_id: int,
        reviewed: bool,
        reviewer: Optional[str] = None,
        review_notes: Optional[str] = None
    ) -> Dict[str, Any]:
        data = {"reviewed": reviewed}
        if reviewer:
            data["reviewer"] = reviewer
        if review_notes:
            data["review_notes"] = review_notes
        return self._request(
            "PATCH", f"/stations/{station_id}/quality/{record_id}/review", json=data
        ).json()

    def create_calibration(
        self,
        station_id: int,
        last_calibration_date: datetime,
        next_calibration_date: Optional[datetime] = None,
        is_done: bool = False
    ) -> Dict[str, Any]:
        next_date = next_calibration_date or (last_calibration_date + timedelta(days=730))
        data = {
            "last_calibration_date": last_calibration_date.isoformat(),
            "next_calibration_date": next_date.isoformat(),
            "is_done": is_done
        }
        return self._request("POST", f"/stations/{station_id}/calibrations", json=data).json()

    def get_calibrations(self, station_id: int, limit: int = 100) -> list:
        return self._request(
            "GET", f"/stations/{station_id}/calibrations", params={"limit": limit}
        ).json()

    def update_calibration(self, station_id: int, calibration_id: int, **kwargs) -> Dict[str, Any]:
        return self._request(
            "PATCH", f"/stations/{station_id}/calibrations/{calibration_id}", json=kwargs
        ).json()

    def get_calibration_todos(self) -> list:
        return self._request("GET", "/calibration-todos").json()
