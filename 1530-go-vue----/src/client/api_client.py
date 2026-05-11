import os
import httpx
from typing import Optional, Dict, Any, List
from datetime import date, datetime


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get(
            "PIPELINE_API_URL",
            "http://127.0.0.1:8000"
        )
        self.client = httpx.Client(base_url=self.base_url, timeout=30.0)

    def close(self):
        self.client.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def _get(self, endpoint: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        response = self.client.get(endpoint, params=params)
        response.raise_for_status()
        return response.json()

    def _post(self, endpoint: str, data: Optional[Dict] = None, params: Optional[Dict] = None) -> Dict[str, Any]:
        response = self.client.post(endpoint, json=data, params=params)
        response.raise_for_status()
        return response.json()

    def _put(self, endpoint: str, data: Optional[Dict] = None) -> Dict[str, Any]:
        response = self.client.put(endpoint, json=data)
        response.raise_for_status()
        return response.json()

    def _delete(self, endpoint: str) -> Dict[str, Any]:
        response = self.client.delete(endpoint)
        response.raise_for_status()
        return response.json()

    def health(self) -> Dict[str, Any]:
        return self._get("/health")

    def create_pipeline(self, data: Dict) -> Dict[str, Any]:
        return self._post("/pipelines/", data=data)

    def list_pipelines(self, status: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"status": status} if status else None
        return self._get("/pipelines/", params=params)

    def get_pipeline(self, pipeline_id: int) -> Dict[str, Any]:
        return self._get(f"/pipelines/{pipeline_id}")

    def update_pipeline(self, pipeline_id: int, data: Dict) -> Dict[str, Any]:
        return self._put(f"/pipelines/{pipeline_id}", data=data)

    def delete_pipeline(self, pipeline_id: int) -> Dict[str, Any]:
        return self._delete(f"/pipelines/{pipeline_id}")

    def create_segment(self, data: Dict) -> Dict[str, Any]:
        return self._post("/segments/", data=data)

    def list_segments(self, pipeline_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {"pipeline_id": pipeline_id} if pipeline_id else None
        return self._get("/segments/", params=params)

    def get_segment(self, segment_id: int) -> Dict[str, Any]:
        return self._get(f"/segments/{segment_id}")

    def update_segment(self, segment_id: int, data: Dict) -> Dict[str, Any]:
        return self._put(f"/segments/{segment_id}", data=data)

    def delete_segment(self, segment_id: int) -> Dict[str, Any]:
        return self._delete(f"/segments/{segment_id}")

    def create_pressure_point(self, data: Dict) -> Dict[str, Any]:
        return self._post("/pressure/points/", data=data)

    def list_pressure_points(self, pipeline_id: int) -> List[Dict[str, Any]]:
        return self._get("/pressure/points/", params={"pipeline_id": pipeline_id})

    def add_pressure_reading(self, pressure_point_id: int, pressure: float) -> Dict[str, Any]:
        return self._post("/pressure/readings/", data={
            "pressure_point_id": pressure_point_id,
            "pressure": pressure
        })

    def get_pressure_stats(self, pipeline_id: int) -> Dict[str, Any]:
        return self._get(f"/pressure/stats/{pipeline_id}")

    def check_leak(self, start_point_id: int, end_point_id: int) -> Dict[str, Any]:
        return self._get(f"/pressure/check-leak/{start_point_id}/{end_point_id}")

    def aggregate_pressure(self) -> Dict[str, Any]:
        return self._post("/pressure/aggregate")

    def create_alarm(self, data: Dict) -> Dict[str, Any]:
        return self._post("/alarms/", data=data)

    def list_alarms(self, status: Optional[str] = None, severity: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {}
        if status:
            params["status"] = status
        if severity:
            params["severity"] = severity
        return self._get("/alarms/", params=params if params else None)

    def get_alarm(self, alarm_id: int) -> Dict[str, Any]:
        return self._get(f"/alarms/{alarm_id}")

    def update_alarm_status(self, alarm_id: int, status: str, handled_by: Optional[str] = None, false_alarm_reason: Optional[str] = None) -> Dict[str, Any]:
        data = {"status": status}
        if handled_by:
            data["handled_by"] = handled_by
        if false_alarm_reason:
            data["false_alarm_reason"] = false_alarm_reason
        return self._put(f"/alarms/{alarm_id}/status", data=data)

    def escalate_alarms(self) -> Dict[str, Any]:
        return self._post("/alarms/escalate")

    def send_reminders(self) -> Dict[str, Any]:
        return self._post("/alarms/reminders")

    def generate_patrol_plan(self) -> Dict[str, Any]:
        return self._post("/patrol/plans/generate")

    def list_patrol_plans(self, plan_date: Optional[date] = None) -> List[Dict[str, Any]]:
        params = {"plan_date": plan_date.isoformat()} if plan_date else None
        return self._get("/patrol/plans/", params=params)

    def get_patrol_plan(self, plan_id: int) -> Dict[str, Any]:
        return self._get(f"/patrol/plans/{plan_id}")

    def add_patrol_record(self, data: Dict) -> Dict[str, Any]:
        if "patrol_date" in data and isinstance(data["patrol_date"], date):
            data["patrol_date"] = data["patrol_date"].isoformat()
        return self._post("/patrol/records/", data=data)

    def list_patrol_records(self, segment_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {"segment_id": segment_id} if segment_id else None
        return self._get("/patrol/records/", params=params)

    def add_integrity_record(self, data: Dict) -> Dict[str, Any]:
        if "inspection_date" in data and isinstance(data["inspection_date"], date):
            data["inspection_date"] = data["inspection_date"].isoformat()
        return self._post("/integrity/records/", data=data)

    def list_integrity_records(self, pipeline_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {"pipeline_id": pipeline_id} if pipeline_id else None
        return self._get("/integrity/records/", params=params)

    def check_wall_thickness(self, record_id: int) -> Dict[str, Any]:
        return self._get(f"/integrity/records/{record_id}/check")

    def list_maintenance_tasks(self, status: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"status": status} if status else None
        return self._get("/integrity/maintenance/", params=params)

    def update_maintenance_task(self, task_id: int, data: Dict) -> Dict[str, Any]:
        return self._put(f"/integrity/maintenance/{task_id}", data=data)

    def generate_report(self, report_date: Optional[date] = None) -> Dict[str, Any]:
        params = {"report_date": report_date.isoformat()} if report_date else None
        return self._post("/reports/generate", params=params)

    def list_reports(self) -> List[Dict[str, Any]]:
        return self._get("/reports/")

    def get_report(self, report_date: date) -> Dict[str, Any]:
        return self._get(f"/reports/{report_date.isoformat()}")
