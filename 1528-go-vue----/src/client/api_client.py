from __future__ import annotations

import json
import os
from datetime import date, datetime
from typing import Any, Dict, List, Optional
from uuid import UUID

import httpx


class APIClient:
    def __init__(self, base_url: Optional[str] = None) -> None:
        self.base_url = base_url or os.getenv(
            "TAILING_API_URL",
            "http://localhost:8000",
        )
        self.client = httpx.Client(base_url=self.base_url, timeout=30.0)

    def close(self) -> None:
        self.client.close()

    def _json_serialize(self, obj: Any) -> Any:
        if isinstance(obj, (datetime, date)):
            return obj.isoformat()
        if isinstance(obj, UUID):
            return str(obj)
        raise TypeError(f"Type {type(obj)} not serializable")

    def _request(
        self,
        method: str,
        path: str,
        json_data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
    ) -> Any:
        if json_data:
            json_data = json.loads(
                json.dumps(json_data, default=self._json_serialize)
            )
        response = self.client.request(
            method,
            path,
            json=json_data,
            params=params,
        )
        if response.status_code >= 400:
            raise Exception(
                f"API Error {response.status_code}: {response.text}"
            )
        if response.status_code == 204:
            return None
        return response.json()

    def list_ponds(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/ponds")

    def create_pond(
        self,
        name: str,
        capacity: float,
        dam_height: float,
        safety_level: str,
    ) -> Dict[str, Any]:
        return self._request(
            "POST",
            "/ponds",
            json_data={
                "name": name,
                "capacity": capacity,
                "dam_height": dam_height,
                "safety_level": safety_level,
            },
        )

    def get_pond(self, pond_id: UUID) -> Dict[str, Any]:
        return self._request("GET", f"/ponds/{pond_id}")

    def update_pond(
        self,
        pond_id: UUID,
        name: Optional[str] = None,
        capacity: Optional[float] = None,
        dam_height: Optional[float] = None,
        safety_level: Optional[str] = None,
    ) -> Dict[str, Any]:
        data: Dict[str, Any] = {}
        if name is not None:
            data["name"] = name
        if capacity is not None:
            data["capacity"] = capacity
        if dam_height is not None:
            data["dam_height"] = dam_height
        if safety_level is not None:
            data["safety_level"] = safety_level
        return self._request("PATCH", f"/ponds/{pond_id}", json_data=data)

    def delete_pond(self, pond_id: UUID) -> None:
        self._request("DELETE", f"/ponds/{pond_id}")

    def create_section(self, pond_id: UUID, name: str) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/ponds/{pond_id}/sections",
            json_data={"pond_id": str(pond_id), "name": name},
        )

    def list_sections(self, pond_id: UUID) -> List[Dict[str, Any]]:
        return self._request("GET", f"/ponds/{pond_id}/sections")

    def add_monitoring_data(
        self,
        section_id: UUID,
        timestamp: datetime,
        dry_beach_length: Optional[float] = None,
        phreatic_line: Optional[float] = None,
        dam_displacement: Optional[float] = None,
        water_level: Optional[float] = None,
    ) -> Dict[str, Any]:
        data: Dict[str, Any] = {
            "section_id": str(section_id),
            "timestamp": timestamp,
        }
        if dry_beach_length is not None:
            data["dry_beach_length"] = dry_beach_length
        if phreatic_line is not None:
            data["phreatic_line"] = phreatic_line
        if dam_displacement is not None:
            data["dam_displacement"] = dam_displacement
        if water_level is not None:
            data["water_level"] = water_level
        return self._request("POST", "/monitoring-data", json_data=data)

    def list_monitoring_data(
        self,
        section_id: UUID,
        start: Optional[datetime] = None,
        end: Optional[datetime] = None,
    ) -> List[Dict[str, Any]]:
        params: Dict[str, Any] = {}
        if start:
            params["start"] = start.isoformat()
        if end:
            params["end"] = end.isoformat()
        return self._request(
            "GET",
            f"/sections/{section_id}/monitoring-data",
            params=params if params else None,
        )

    def run_aggregation(self, section_id: UUID) -> None:
        self._request("POST", f"/sections/{section_id}/aggregate")

    def list_aggregated_data(
        self,
        section_id: UUID,
        data_type: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        params: Dict[str, Any] = {}
        if data_type:
            params["data_type"] = data_type
        return self._request(
            "GET",
            f"/sections/{section_id}/aggregated",
            params=params if params else None,
        )

    def generate_inspections(
        self,
        pond_id: UUID,
        days_ahead: int = 30,
    ) -> List[Dict[str, Any]]:
        return self._request(
            "POST",
            f"/ponds/{pond_id}/inspections/generate",
            params={"days_ahead": days_ahead},
        )

    def list_inspections(self, pond_id: UUID) -> List[Dict[str, Any]]:
        return self._request("GET", f"/ponds/{pond_id}/inspections")

    def complete_inspection(
        self,
        task_id: UUID,
        actual_date: date,
        inspector: str,
        remarks: Optional[str] = None,
    ) -> Dict[str, Any]:
        data: Dict[str, Any] = {
            "actual_date": actual_date.isoformat(),
            "inspector": inspector,
        }
        if remarks:
            data["remarks"] = remarks
        return self._request(
            "PATCH",
            f"/inspections/{task_id}",
            json_data=data,
        )

    def create_acceptance(self, pond_id: UUID) -> Dict[str, Any]:
        return self._request(
            "POST",
            "/acceptances",
            json_data={"pond_id": str(pond_id)},
        )

    def get_acceptance_by_pond(self, pond_id: UUID) -> Dict[str, Any]:
        return self._request("GET", f"/ponds/{pond_id}/acceptance")

    def submit_material(
        self,
        material_id: UUID,
        content: str,
    ) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/materials/{material_id}/submit",
            json_data={"content": content},
        )

    def review_material(
        self,
        material_id: UUID,
        reviewer: str,
        opinion: str,
        is_approved: bool,
    ) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/materials/{material_id}/review",
            json_data={
                "reviewer": reviewer,
                "opinion": opinion,
                "is_approved": is_approved,
            },
        )

    def list_warnings(
        self,
        acknowledged: Optional[bool] = None,
    ) -> List[Dict[str, Any]]:
        params: Dict[str, Any] = {}
        if acknowledged is not None:
            params["acknowledged"] = acknowledged
        return self._request(
            "GET",
            "/warnings",
            params=params if params else None,
        )

    def acknowledge_warning(self, warning_id: UUID) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/warnings/{warning_id}/acknowledge",
        )
