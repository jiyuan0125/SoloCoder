import os
import requests
from typing import List, Dict, Any, Optional
from datetime import date, datetime
import json


class ApiClient:
    def __init__(self, base_url: str = None):
        self.base_url = base_url or os.environ.get("SERVER_URL", "http://localhost:8000")
    
    def _get(self, endpoint: str, params: Dict[str, Any] = None) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        response = requests.get(url, params=params)
        response.raise_for_status()
        return response.json()
    
    def _post(self, endpoint: str, data: Dict[str, Any] = None) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        response = requests.post(url, json=data)
        response.raise_for_status()
        return response.json()
    
    def _put(self, endpoint: str, data: Dict[str, Any] = None) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        response = requests.put(url, json=data)
        response.raise_for_status()
        return response.json()
    
    def get_users(self) -> List[Dict[str, Any]]:
        return self._get("/users/")
    
    def get_user(self, user_id: int) -> Dict[str, Any]:
        return self._get(f"/users/{user_id}")
    
    def create_user(self, name: str, role: str, phone: str = None) -> Dict[str, Any]:
        data = {"name": name, "role": role, "phone": phone}
        return self._post("/users/", data=data)
    
    def get_areas(self) -> List[Dict[str, Any]]:
        return self._get("/areas/")
    
    def get_area(self, area_id: int) -> Dict[str, Any]:
        return self._get(f"/areas/{area_id}")
    
    def create_area(self, name: str, area_km2: float, main_tree_species: str, ranger_id: int = None) -> Dict[str, Any]:
        data = {
            "name": name,
            "area_km2": area_km2,
            "main_tree_species": main_tree_species,
            "ranger_id": ranger_id
        }
        return self._post("/areas/", data=data)
    
    def get_tasks(self) -> List[Dict[str, Any]]:
        return self._get("/tasks/")
    
    def get_pending_tasks(self) -> List[Dict[str, Any]]:
        return self._get("/tasks/pending/")
    
    def get_task(self, task_id: int) -> Dict[str, Any]:
        return self._get(f"/tasks/{task_id}")
    
    def create_task(self, area_id: int, ranger_id: int, patrol_date: str, route_keypoints: List[Dict[str, float]]) -> Dict[str, Any]:
        data = {
            "area_id": area_id,
            "ranger_id": ranger_id,
            "patrol_date": patrol_date,
            "route_keypoints": route_keypoints
        }
        return self._post("/tasks/", data=data)
    
    def get_reports(self) -> List[Dict[str, Any]]:
        return self._get("/reports/")
    
    def get_report(self, report_id: int) -> Dict[str, Any]:
        return self._get(f"/reports/{report_id}")
    
    def create_report(self, task_id: int, actual_route: List[Dict[str, float]], anomalies: List[Dict[str, Any]] = None) -> Dict[str, Any]:
        data = {
            "task_id": task_id,
            "actual_route": actual_route,
            "anomalies": anomalies or []
        }
        return self._post("/reports/", data=data)
    
    def get_todos(self) -> List[Dict[str, Any]]:
        return self._get("/todos/")
    
    def get_todo(self, todo_id: int) -> Dict[str, Any]:
        return self._get(f"/todos/{todo_id}")
    
    def resolve_todo(self, todo_id: int) -> Dict[str, Any]:
        return self._put(f"/todos/{todo_id}/resolve")
    
    def get_monthly_pest_stats(self, year: int, month: int) -> Dict[str, Any]:
        return self._get("/stats/pest/monthly/", params={"year": year, "month": month})
    
    def get_area_health_scores(self) -> List[Dict[str, Any]]:
        return self._get("/stats/area-health/")
