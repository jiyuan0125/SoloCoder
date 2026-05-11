import os
from datetime import datetime
from typing import Optional, List, Dict, Any

import httpx

from core.models import (
    Mine,
    MonitorZone,
    ThresholdConfig,
    SensorReading,
    AlarmEvent,
    ShiftStats,
    DailyReport,
)


class MonitorAPIClient:
    def __init__(
        self,
        base_url: Optional[str] = None,
        timeout: int = 30,
    ):
        self.base_url = base_url or os.environ.get(
            "MONITOR_API_URL",
            "http://localhost:8000",
        )
        self.api_prefix = "/api/v1"
        self.timeout = timeout
    
    def _get_full_url(self, endpoint: str) -> str:
        return f"{self.base_url}{self.api_prefix}{endpoint}"
    
    def _client(self) -> httpx.Client:
        return httpx.Client(timeout=self.timeout)
    
    def list_mines(self) -> List[Dict[str, Any]]:
        with self._client() as client:
            response = client.get(self._get_full_url("/mines"))
            response.raise_for_status()
            return response.json()
    
    def create_mine(self, name: str, description: Optional[str] = None) -> Dict[str, Any]:
        payload = {"name": name}
        if description:
            payload["description"] = description
        with self._client() as client:
            response = client.post(self._get_full_url("/mines"), json=payload)
            response.raise_for_status()
            return response.json()
    
    def get_mine(self, mine_id: int) -> Dict[str, Any]:
        with self._client() as client:
            response = client.get(self._get_full_url(f"/mines/{mine_id}"))
            response.raise_for_status()
            return response.json()
    
    def delete_mine(self, mine_id: int) -> Dict[str, Any]:
        with self._client() as client:
            response = client.delete(self._get_full_url(f"/mines/{mine_id}"))
            response.raise_for_status()
            return response.json()
    
    def list_zones(self, mine_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {}
        if mine_id is not None:
            params["mine_id"] = mine_id
        with self._client() as client:
            response = client.get(self._get_full_url("/zones"), params=params)
            response.raise_for_status()
            return response.json()
    
    def create_zone(self, mine_id: int, name: str, description: Optional[str] = None) -> Dict[str, Any]:
        payload = {"mine_id": mine_id, "name": name}
        if description:
            payload["description"] = description
        with self._client() as client:
            response = client.post(self._get_full_url("/zones"), json=payload)
            response.raise_for_status()
            return response.json()
    
    def get_zone(self, zone_id: int) -> Dict[str, Any]:
        with self._client() as client:
            response = client.get(self._get_full_url(f"/zones/{zone_id}"))
            response.raise_for_status()
            return response.json()
    
    def delete_zone(self, zone_id: int) -> Dict[str, Any]:
        with self._client() as client:
            response = client.delete(self._get_full_url(f"/zones/{zone_id}"))
            response.raise_for_status()
            return response.json()
    
    def list_thresholds(self, zone_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {}
        if zone_id is not None:
            params["zone_id"] = zone_id
        with self._client() as client:
            response = client.get(self._get_full_url("/thresholds"), params=params)
            response.raise_for_status()
            return response.json()
    
    def create_threshold(
        self,
        zone_id: int,
        metric_type: str,
        level1: float,
        level2: float,
        level3: float,
    ) -> Dict[str, Any]:
        payload = {
            "zone_id": zone_id,
            "metric_type": metric_type,
            "level1_threshold": level1,
            "level2_threshold": level2,
            "level3_threshold": level3,
        }
        with self._client() as client:
            response = client.post(self._get_full_url("/thresholds"), json=payload)
            response.raise_for_status()
            return response.json()
    
    def update_threshold(
        self,
        threshold_id: int,
        zone_id: int,
        metric_type: str,
        level1: float,
        level2: float,
        level3: float,
    ) -> Dict[str, Any]:
        payload = {
            "id": threshold_id,
            "zone_id": zone_id,
            "metric_type": metric_type,
            "level1_threshold": level1,
            "level2_threshold": level2,
            "level3_threshold": level3,
        }
        with self._client() as client:
            response = client.put(self._get_full_url(f"/thresholds/{threshold_id}"), json=payload)
            response.raise_for_status()
            return response.json()
    
    def submit_reading(
        self,
        zone_id: int,
        methane: float,
        co: float,
        wind_speed: float,
        temperature: float,
        dust: float,
        timestamp: Optional[datetime] = None,
    ) -> Dict[str, Any]:
        if timestamp is None:
            timestamp = datetime.utcnow()
        payload = {
            "zone_id": zone_id,
            "timestamp": timestamp.isoformat(),
            "methane": methane,
            "co": co,
            "wind_speed": wind_speed,
            "temperature": temperature,
            "dust": dust,
        }
        with self._client() as client:
            response = client.post(self._get_full_url("/readings"), json=payload)
            response.raise_for_status()
            return response.json()
    
    def list_readings(
        self,
        zone_id: Optional[int] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        limit: int = 100,
    ) -> List[Dict[str, Any]]:
        params = {"limit": limit}
        if zone_id is not None:
            params["zone_id"] = zone_id
        if start_time is not None:
            params["start_time"] = start_time.isoformat()
        if end_time is not None:
            params["end_time"] = end_time.isoformat()
        with self._client() as client:
            response = client.get(self._get_full_url("/readings"), params=params)
            response.raise_for_status()
            return response.json()
    
    def export_readings(
        self,
        output_path: str,
        zone_id: Optional[int] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
    ) -> None:
        params = {}
        if zone_id is not None:
            params["zone_id"] = zone_id
        if start_time is not None:
            params["start_time"] = start_time.isoformat()
        if end_time is not None:
            params["end_time"] = end_time.isoformat()
        with self._client() as client:
            response = client.get(self._get_full_url("/readings/export"), params=params)
            response.raise_for_status()
            with open(output_path, "wb") as f:
                f.write(response.content)
    
    def list_alarms(
        self,
        zone_id: Optional[int] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        acknowledged: Optional[bool] = None,
        limit: int = 100,
    ) -> List[Dict[str, Any]]:
        params = {"limit": limit}
        if zone_id is not None:
            params["zone_id"] = zone_id
        if start_time is not None:
            params["start_time"] = start_time.isoformat()
        if end_time is not None:
            params["end_time"] = end_time.isoformat()
        if acknowledged is not None:
            params["acknowledged"] = acknowledged
        with self._client() as client:
            response = client.get(self._get_full_url("/alarms"), params=params)
            response.raise_for_status()
            return response.json()
    
    def acknowledge_alarm(self, alarm_id: int, acknowledged_by: Optional[str] = None) -> Dict[str, Any]:
        params = {}
        if acknowledged_by:
            params["acknowledged_by"] = acknowledged_by
        with self._client() as client:
            response = client.post(
                self._get_full_url(f"/alarms/{alarm_id}/acknowledge"),
                params=params,
            )
            response.raise_for_status()
            return response.json()
    
    def export_alarms(
        self,
        output_path: str,
        zone_id: Optional[int] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
    ) -> None:
        params = {}
        if zone_id is not None:
            params["zone_id"] = zone_id
        if start_time is not None:
            params["start_time"] = start_time.isoformat()
        if end_time is not None:
            params["end_time"] = end_time.isoformat()
        with self._client() as client:
            response = client.get(self._get_full_url("/alarms/export"), params=params)
            response.raise_for_status()
            with open(output_path, "wb") as f:
                f.write(response.content)
    
    def generate_shift_stats(self, zone_id: Optional[int] = None) -> Dict[str, Any]:
        params = {}
        if zone_id is not None:
            params["zone_id"] = zone_id
        with self._client() as client:
            response = client.post(self._get_full_url("/stats/generate-shift"), params=params)
            response.raise_for_status()
            return response.json()
    
    def list_shift_stats(
        self,
        zone_id: Optional[int] = None,
        shift_date: Optional[str] = None,
        limit: int = 100,
    ) -> List[Dict[str, Any]]:
        params = {"limit": limit}
        if zone_id is not None:
            params["zone_id"] = zone_id
        if shift_date is not None:
            params["shift_date"] = shift_date
        with self._client() as client:
            response = client.get(self._get_full_url("/stats/shift"), params=params)
            response.raise_for_status()
            return response.json()
    
    def export_shift_stats(
        self,
        output_path: str,
        zone_id: Optional[int] = None,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
    ) -> None:
        params = {}
        if zone_id is not None:
            params["zone_id"] = zone_id
        if start_date is not None:
            params["start_date"] = start_date
        if end_date is not None:
            params["end_date"] = end_date
        with self._client() as client:
            response = client.get(self._get_full_url("/stats/shift/export"), params=params)
            response.raise_for_status()
            with open(output_path, "wb") as f:
                f.write(response.content)
    
    def generate_daily_reports(self, zone_id: Optional[int] = None) -> Dict[str, Any]:
        params = {}
        if zone_id is not None:
            params["zone_id"] = zone_id
        with self._client() as client:
            response = client.post(self._get_full_url("/stats/generate-daily"), params=params)
            response.raise_for_status()
            return response.json()
    
    def list_daily_reports(
        self,
        zone_id: Optional[int] = None,
        report_date: Optional[str] = None,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
        limit: int = 100,
    ) -> List[Dict[str, Any]]:
        params = {"limit": limit}
        if zone_id is not None:
            params["zone_id"] = zone_id
        if report_date is not None:
            params["report_date"] = report_date
        if start_date is not None:
            params["start_date"] = start_date
        if end_date is not None:
            params["end_date"] = end_date
        with self._client() as client:
            response = client.get(self._get_full_url("/stats/daily"), params=params)
            response.raise_for_status()
            return response.json()
    
    def export_daily_reports(
        self,
        output_path: str,
        zone_id: Optional[int] = None,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
    ) -> None:
        params = {}
        if zone_id is not None:
            params["zone_id"] = zone_id
        if start_date is not None:
            params["start_date"] = start_date
        if end_date is not None:
            params["end_date"] = end_date
        with self._client() as client:
            response = client.get(self._get_full_url("/stats/daily/export"), params=params)
            response.raise_for_status()
            with open(output_path, "wb") as f:
                f.write(response.content)
