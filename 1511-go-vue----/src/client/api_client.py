import os
import json
from typing import Optional, Any, Dict
from datetime import date, datetime

import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("TEA_API_URL", "http://localhost:8000")

    def _make_url(self, path: str) -> str:
        return f"{self.base_url.rstrip('/')}{path}"

    def _request(self, method: str, path: str, **kwargs) -> Any:
        url = self._make_url(path)
        try:
            response = requests.request(method, url, **kwargs)
            if response.status_code >= 400:
                try:
                    error_data = response.json()
                    if "detail" in error_data:
                        detail = error_data["detail"]
                        if isinstance(detail, dict) and "message" in detail:
                            raise Exception(detail["message"])
                        raise Exception(str(detail))
                    raise Exception(response.text)
                except ValueError:
                    raise Exception(response.text)
            if response.status_code == 204 or not response.content:
                return None
            return response.json()
        except requests.ConnectionError:
            raise Exception(f"无法连接到服务端: {self.base_url}")

    def get(self, path: str, params: Optional[Dict] = None) -> Any:
        return self._request("GET", path, params=params)

    def post(self, path: str, json_data: Optional[Dict] = None) -> Any:
        return self._request("POST", path, json=json_data)

    def put(self, path: str, json_data: Optional[Dict] = None) -> Any:
        return self._request("PUT", path, json=json_data)

    def delete(self, path: str) -> Any:
        return self._request("DELETE", path)

    def list_plots(self) -> list:
        return self.get("/plots")["plots"]

    def create_plot(self, name: str, location: str, area: float) -> dict:
        return self.post("/plots", {"name": name, "location": location, "area": area})

    def get_plot(self, plot_id: str) -> dict:
        return self.get(f"/plots/{plot_id}")

    def update_plot(self, plot_id: str, data: dict) -> dict:
        return self.put(f"/plots/{plot_id}", data)

    def delete_plot(self, plot_id: str) -> dict:
        return self.delete(f"/plots/{plot_id}")

    def list_harvest_plans(self) -> list:
        return self.get("/harvest-plans")["plans"]

    def create_harvest_plan(self, plot_id: str, plan_date: date, expected_quantity: float) -> dict:
        return self.post("/harvest-plans", {
            "plot_id": plot_id,
            "plan_date": plan_date.isoformat(),
            "expected_quantity": expected_quantity
        })

    def get_harvest_plan(self, plan_id: str) -> dict:
        return self.get(f"/harvest-plans/{plan_id}")

    def update_harvest_plan(self, plan_id: str, data: dict) -> dict:
        if "plan_date" in data and isinstance(data["plan_date"], date):
            data["plan_date"] = data["plan_date"].isoformat()
        return self.put(f"/harvest-plans/{plan_id}", data)

    def delete_harvest_plan(self, plan_id: str) -> dict:
        return self.delete(f"/harvest-plans/{plan_id}")

    def list_harvest_records(self) -> list:
        return self.get("/harvest-records")["records"]

    def create_harvest_record(self, plan_id: str, actual_quantity: float,
                              fresh_leaf_grade: str, harvest_time: Optional[datetime] = None) -> dict:
        data = {
            "plan_id": plan_id,
            "actual_quantity": actual_quantity,
            "fresh_leaf_grade": fresh_leaf_grade
        }
        if harvest_time:
            data["harvest_time"] = harvest_time.isoformat()
        return self.post("/harvest-records", data)

    def get_harvest_record(self, record_id: str) -> dict:
        return self.get(f"/harvest-records/{record_id}")

    def update_harvest_record(self, record_id: str, data: dict) -> dict:
        if "harvest_time" in data and isinstance(data["harvest_time"], datetime):
            data["harvest_time"] = data["harvest_time"].isoformat()
        return self.put(f"/harvest-records/{record_id}", data)

    def delete_harvest_record(self, record_id: str) -> dict:
        return self.delete(f"/harvest-records/{record_id}")

    def list_processing_batches(self) -> list:
        return self.get("/processing-batches")["batches"]

    def create_processing_batch(self, harvest_record_id: str, input_quantity: float,
                                start_time: Optional[datetime] = None) -> dict:
        data = {
            "harvest_record_id": harvest_record_id,
            "input_quantity": input_quantity
        }
        if start_time:
            data["start_time"] = start_time.isoformat()
        return self.post("/processing-batches", data)

    def get_processing_batch(self, batch_id: str) -> dict:
        return self.get(f"/processing-batches/{batch_id}")

    def complete_processing_batch(self, batch_id: str, output_quantity: float) -> dict:
        return self.post(f"/processing-batches/{batch_id}/complete", {
            "output_quantity": output_quantity
        })

    def delete_processing_batch(self, batch_id: str) -> dict:
        return self.delete(f"/processing-batches/{batch_id}")

    def list_quality_ratings(self) -> list:
        return self.get("/quality-ratings")["ratings"]

    def create_quality_rating(self, batch_id: str, grade: str, sensory_description: str) -> dict:
        return self.post("/quality-ratings", {
            "batch_id": batch_id,
            "grade": grade,
            "sensory_description": sensory_description
        })

    def get_quality_rating(self, rating_id: str) -> dict:
        return self.get(f"/quality-ratings/{rating_id}")

    def update_quality_rating(self, rating_id: str, data: dict) -> dict:
        return self.put(f"/quality-ratings/{rating_id}", data)

    def delete_quality_rating(self, rating_id: str) -> dict:
        return self.delete(f"/quality-ratings/{rating_id}")

    def run_scheduler(self) -> dict:
        return self.post("/scheduler/run")

    def list_todos(self, status: Optional[str] = None) -> list:
        params = {"status": status} if status else None
        return self.get("/todos", params=params)["todos"]

    def complete_todo(self, todo_id: str) -> dict:
        return self.post(f"/todos/{todo_id}/complete")

    def list_alerts(self, status: Optional[str] = None) -> list:
        params = {"status": status} if status else None
        return self.get("/alerts", params=params)["alerts"]

    def resolve_alert(self, alert_id: str) -> dict:
        return self.post(f"/alerts/{alert_id}/resolve")
