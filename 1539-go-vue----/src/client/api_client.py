import os
import httpx
from typing import Optional, Dict, Any, List
from datetime import datetime, date


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_BASE_URL", "http://localhost:8000")
        
    def _get(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(base_url=self.base_url) as client:
            response = client.get(endpoint, params=params)
            response.raise_for_status()
            return response.json()
    
    def _post(self, endpoint: str, json: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(base_url=self.base_url) as client:
            response = client.post(endpoint, json=json)
            response.raise_for_status()
            return response.json()
    
    def _put(self, endpoint: str, json: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(base_url=self.base_url) as client:
            response = client.put(endpoint, json=json)
            response.raise_for_status()
            return response.json()
    
    def _delete(self, endpoint: str) -> None:
        with httpx.Client(base_url=self.base_url) as client:
            response = client.delete(endpoint)
            response.raise_for_status()
    
    def _get_stream(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> bytes:
        with httpx.Client(base_url=self.base_url) as client:
            response = client.get(endpoint, params=params)
            response.raise_for_status()
            return response.content
    
    # Zone limits
    def create_zone_limit(self, zone_type: str, daytime_limit: float, nighttime_limit: float):
        return self._post("/zone-limits/", json={
            "zone_type": zone_type,
            "daytime_limit": daytime_limit,
            "nighttime_limit": nighttime_limit,
        })
    
    def get_zone_limits(self):
        return self._get("/zone-limits/")
    
    def get_zone_limit(self, zone_limit_id: int):
        return self._get(f"/zone-limits/{zone_limit_id}")
    
    def update_zone_limit(self, zone_limit_id: int, **kwargs):
        return self._put(f"/zone-limits/{zone_limit_id}", json=kwargs)
    
    def delete_zone_limit(self, zone_limit_id: int):
        self._delete(f"/zone-limits/{zone_limit_id}")
    
    # Monitoring points
    def create_point(self, name: str, code: str, zone_type: str, last_calibration_date: str, 
                    next_calibration_date: str, address: Optional[str] = None, 
                    latitude: Optional[float] = None, longitude: Optional[float] = None, is_active: bool = True):
        return self._post("/monitoring-points/", json={
            "name": name,
            "code": code,
            "zone_type": zone_type,
            "address": address,
            "latitude": latitude,
            "longitude": longitude,
            "is_active": is_active,
            "last_calibration_date": last_calibration_date,
            "next_calibration_date": next_calibration_date,
        })
    
    def get_points(self):
        return self._get("/monitoring-points/")
    
    def get_point(self, point_id: int):
        return self._get(f"/monitoring-points/{point_id}")
    
    def update_point(self, point_id: int, **kwargs):
        return self._put(f"/monitoring-points/{point_id}", json=kwargs)
    
    def delete_point(self, point_id: int):
        self._delete(f"/monitoring-points/{point_id}")
    
    # Noise data
    def create_noise_data(self, point_id: int, value: float, timestamp: Optional[datetime] = None):
        data = {"point_id": point_id, "value": value}
        if timestamp:
            data["timestamp"] = timestamp.isoformat()
        return self._post("/noise-data/", json=data)
    
    def get_noise_data(self, point_id: Optional[int] = None, start_time: Optional[datetime] = None, 
                       end_time: Optional[datetime] = None):
        params = {}
        if point_id is not None:
            params["point_id"] = point_id
        if start_time:
            params["start_time"] = start_time.isoformat()
        if end_time:
            params["end_time"] = end_time.isoformat()
        return self._get("/noise-data/", params=params if params else None)
    
    def get_noise_data_item(self, data_id: int):
        return self._get(f"/noise-data/{data_id}")
    
    def export_noise_data_csv(self, point_id: Optional[int] = None, start_time: Optional[datetime] = None,
                              end_time: Optional[datetime] = None):
        params = {}
        if point_id is not None:
            params["point_id"] = point_id
        if start_time:
            params["start_time"] = start_time.isoformat()
        if end_time:
            params["end_time"] = end_time.isoformat()
        return self._get_stream("/noise-data/export/csv", params=params if params else None)
    
    # Alerts
    def check_calibration(self):
        return self._post("/alerts/check-calibration")
    
    def get_alerts(self, point_id: Optional[int] = None, alert_type: Optional[str] = None, is_resolved: Optional[bool] = None):
        params = {}
        if point_id is not None:
            params["point_id"] = point_id
        if alert_type:
            params["alert_type"] = alert_type
        if is_resolved is not None:
            params["is_resolved"] = is_resolved
        return self._get("/alerts/", params=params if params else None)
    
    def get_alert(self, alert_id: int):
        return self._get(f"/alerts/{alert_id}")
    
    def resolve_alert(self, alert_id: int):
        return self._post(f"/alerts/{alert_id}/resolve")
    
    def delete_alert(self, alert_id: int):
        self._delete(f"/alerts/{alert_id}")
    
    # Construction permits
    def create_permit(self, point_ids: List[int], permit_number: str, start_date: str, end_date: str,
                     description: Optional[str] = None, is_active: bool = True):
        return self._post("/construction-permits/", json={
            "point_ids": point_ids,
            "permit_number": permit_number,
            "start_date": start_date,
            "end_date": end_date,
            "description": description,
            "is_active": is_active,
        })
    
    def get_permits(self):
        return self._get("/construction-permits/")
    
    def get_permit(self, permit_id: int):
        return self._get(f"/construction-permits/{permit_id}")
    
    def update_permit(self, permit_id: int, **kwargs):
        return self._put(f"/construction-permits/{permit_id}", json=kwargs)
    
    def delete_permit(self, permit_id: int):
        self._delete(f"/construction-permits/{permit_id}")
    
    # Statistics
    def compute_daily_statistics(self, point_id: int, target_date: date):
        return self._post(f"/statistics/daily/{point_id}/{target_date.isoformat()}")
    
    def get_daily_statistics(self, point_id: Optional[int] = None, start_date: Optional[date] = None,
                            end_date: Optional[date] = None):
        params = {}
        if point_id is not None:
            params["point_id"] = point_id
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()
        return self._get("/statistics/daily/", params=params if params else None)
    
    def get_daily_statistic(self, stats_id: int):
        return self._get(f"/statistics/daily/{stats_id}")
    
    def export_statistics_csv(self, point_id: Optional[int] = None, start_date: Optional[date] = None,
                              end_date: Optional[date] = None):
        params = {}
        if point_id is not None:
            params["point_id"] = point_id
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()
        return self._get_stream("/statistics/export/csv", params=params if params else None)
