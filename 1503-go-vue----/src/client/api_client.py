import os
import json
import httpx
from typing import Dict, Any, List, Optional


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get(
            "API_BASE_URL", "http://localhost:8000/api"
        )
        self.client = httpx.Client(timeout=30.0)

    def close(self):
        self.client.close()

    def _request(self, method: str, path: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{path}"
        response = self.client.request(method, url, **kwargs)
        if response.status_code >= 400:
            try:
                detail = response.json().get("detail", response.text)
            except:
                detail = response.text
            raise Exception(f"请求失败: {response.status_code} - {detail}")
        return response.json()

    def get(self, path: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("GET", path, params=params)

    def post(self, path: str, data: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("POST", path, json=data)

    def patch(self, path: str, data: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("PATCH", path, json=data)

    def create_user(self, name: str, phone: str, role: str) -> Dict[str, Any]:
        return self.post("/users/", data={"name": name, "phone": phone, "role": role})

    def list_users(self) -> List[Dict[str, Any]]:
        return self.get("/users/")

    def get_user(self, user_id: int) -> Dict[str, Any]:
        return self.get(f"/users/{user_id}")

    def create_area(
        self,
        name: str,
        area_km2: float,
        main_tree_species: str,
        ranger_id: Optional[int] = None,
        manager_id: Optional[int] = None
    ) -> Dict[str, Any]:
        data = {
            "name": name,
            "area_km2": area_km2,
            "main_tree_species": main_tree_species,
            "ranger_id": ranger_id,
            "manager_id": manager_id
        }
        return self.post("/areas/", data=data)

    def list_areas(self) -> List[Dict[str, Any]]:
        return self.get("/areas/")

    def get_area(self, area_id: int) -> Dict[str, Any]:
        return self.get(f"/areas/{area_id}")

    def create_task(
        self,
        area_id: int,
        ranger_id: int,
        patrol_date: str,
        route_waypoints: List[Dict[str, float]]
    ) -> Dict[str, Any]:
        data = {
            "area_id": area_id,
            "ranger_id": ranger_id,
            "patrol_date": patrol_date,
            "route_waypoints": route_waypoints
        }
        return self.post("/tasks/", data=data)

    def list_tasks(self) -> List[Dict[str, Any]]:
        return self.get("/tasks/")

    def get_task(self, task_id: int) -> Dict[str, Any]:
        return self.get(f"/tasks/{task_id}")

    def submit_report(
        self,
        task_id: int,
        actual_route: List[Dict[str, float]],
        anomalies: List[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        data = {
            "report": {
                "task_id": task_id,
                "actual_route": actual_route
            },
            "anomalies": anomalies or []
        }
        return self.post("/reports/submit", data=data)

    def list_reports(self) -> List[Dict[str, Any]]:
        return self.get("/reports/")

    def get_report(self, report_id: int) -> Dict[str, Any]:
        return self.get(f"/reports/{report_id}")

    def list_todos(self, overdue_only: bool = False) -> List[Dict[str, Any]]:
        return self.get("/todos/", params={"overdue_only": overdue_only})

    def get_todo(self, todo_id: int) -> Dict[str, Any]:
        return self.get(f"/todos/{todo_id}")

    def update_todo(
        self,
        todo_id: int,
        status: Optional[str] = None,
        notes: Optional[str] = None
    ) -> Dict[str, Any]:
        data = {}
        if status:
            data["status"] = status
        if notes is not None:
            data["notes"] = notes
        return self.patch(f"/todos/{todo_id}", data=data)

    def get_monthly_pest_stats(self) -> Dict[str, Any]:
        return self.get("/statistics/pest/monthly")

    def get_area_health_scores(self) -> Dict[str, Any]:
        return self.get("/statistics/health/areas")
