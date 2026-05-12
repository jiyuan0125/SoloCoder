import os
from typing import Optional, Dict, Any

import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_BASE_URL", "http://localhost:8000")

    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        response = requests.request(method, url, **kwargs)
        if not response.ok:
            try:
                error_detail = response.json().get("detail", response.text)
            except Exception:
                error_detail = response.text
            raise Exception(f"请求失败 [{response.status_code}]: {error_detail}")
        if response.content:
            return response.json()
        return {}

    def create_ship(self, name: str, imo_number: str, flag: str, has_hazardous_qualification: bool, actor: str):
        return self._request(
            "POST",
            f"/ships?actor={actor}",
            json={
                "name": name,
                "imo_number": imo_number,
                "flag": flag,
                "has_hazardous_qualification": has_hazardous_qualification,
            },
        )

    def list_ships(self):
        return self._request("GET", "/ships")

    def get_ship(self, ship_id: int):
        return self._request("GET", f"/ships/{ship_id}")

    def create_declaration(
        self,
        declaration_number: str,
        ship_id: int,
        voyage_number: str,
        hazardous_category: int,
        cargo_name: str,
        cargo_quantity: float,
        packaging_compliant: bool,
        submitted_by: str,
        actor: str,
    ):
        return self._request(
            "POST",
            f"/declarations?actor={actor}",
            json={
                "declaration_number": declaration_number,
                "ship_id": ship_id,
                "voyage_number": voyage_number,
                "hazardous_category": hazardous_category,
                "cargo_name": cargo_name,
                "cargo_quantity": cargo_quantity,
                "packaging_compliant": packaging_compliant,
                "submitted_by": submitted_by,
            },
        )

    def list_declarations(self):
        return self._request("GET", "/declarations")

    def get_declaration(self, declaration_id: int):
        return self._request("GET", f"/declarations/{declaration_id}")

    def initial_review(self, declaration_id: int, reviewer: str, approved: bool, comments: Optional[str], actor: str):
        data = {"reviewer": reviewer, "approved": approved}
        if comments:
            data["comments"] = comments
        return self._request(
            "POST",
            f"/declarations/{declaration_id}/review/initial?actor={actor}",
            json=data,
        )

    def final_review(self, declaration_id: int, reviewer: str, approved: bool, comments: Optional[str], actor: str):
        data = {"reviewer": reviewer, "approved": approved}
        if comments:
            data["comments"] = comments
        return self._request(
            "POST",
            f"/declarations/{declaration_id}/review/final?actor={actor}",
            json=data,
        )

    def get_reviews(self, declaration_id: int):
        return self._request("GET", f"/declarations/{declaration_id}/review")

    def start_loading(
        self,
        declaration_id: int,
        operator: str,
        temperature: Optional[float],
        radiation_dose_rate: Optional[float],
        notes: Optional[str],
        actor: str,
    ):
        data = {"operator": operator}
        if temperature is not None:
            data["temperature"] = temperature
        if radiation_dose_rate is not None:
            data["radiation_dose_rate"] = radiation_dose_rate
        if notes:
            data["notes"] = notes
        return self._request(
            "POST",
            f"/declarations/{declaration_id}/loading/start?actor={actor}",
            json=data,
        )

    def complete_loading(self, declaration_id: int, notes: Optional[str], actor: str):
        data = {}
        if notes:
            data["notes"] = notes
        return self._request(
            "POST",
            f"/declarations/{declaration_id}/loading/complete?actor={actor}",
            json=data,
        )

    def get_loading(self, declaration_id: int):
        return self._request("GET", f"/declarations/{declaration_id}/loading")

    def create_emergency(self, declaration_id: int, impact_range: str, triggered_by: Optional[str], actor: str):
        data = {"impact_range": impact_range}
        if triggered_by:
            data["triggered_by"] = triggered_by
        return self._request(
            "POST",
            f"/declarations/{declaration_id}/emergency?actor={actor}",
            json=data,
        )

    def get_emergencies(self, declaration_id: int):
        return self._request("GET", f"/declarations/{declaration_id}/emergency")

    def list_audit_logs(self, declaration_id: Optional[int] = None):
        endpoint = "/audit"
        if declaration_id is not None:
            endpoint += f"?declaration_id={declaration_id}"
        return self._request("GET", endpoint)
