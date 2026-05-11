import os
from typing import Dict, List, Any, Optional
import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("SERVER_URL", "http://localhost:8000")
        self.session = requests.Session()

    def _url(self, path: str) -> str:
        return f"{self.base_url}{path}"

    def health_check(self) -> Dict[str, Any]:
        response = self.session.get(self._url("/health"))
        response.raise_for_status()
        return response.json()

    def create_audit(
        self,
        company_name: str,
        company_id: Optional[str] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {}
        if user_id:
            params["user_id"] = user_id
        if user_name:
            params["user_name"] = user_name
        data = {"company_name": company_name}
        if company_id:
            data["company_id"] = company_id
        response = self.session.post(self._url("/audits"), json=data, params=params)
        response.raise_for_status()
        return response.json()

    def list_audits(self) -> List[Dict[str, Any]]:
        response = self.session.get(self._url("/audits"))
        response.raise_for_status()
        return response.json()

    def get_audit(self, audit_id: int) -> Dict[str, Any]:
        response = self.session.get(self._url(f"/audits/{audit_id}"))
        response.raise_for_status()
        return response.json()

    def advance_stage(
        self,
        audit_id: int,
        notes: Optional[str] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {}
        if user_id:
            params["user_id"] = user_id
        if user_name:
            params["user_name"] = user_name
        json_data = {"notes": notes} if notes else None
        response = self.session.post(
            self._url(f"/audits/{audit_id}/advance"),
            params=params,
            json=json_data,
        )
        response.raise_for_status()
        return response.json()

    def create_solution(
        self,
        audit_id: int,
        name: str,
        solution_type: str,
        description: Optional[str] = None,
        expected_energy_saving: Optional[float] = None,
        expected_investment: Optional[float] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {}
        if user_id:
            params["user_id"] = user_id
        if user_name:
            params["user_name"] = user_name
        data = {
            "name": name,
            "solution_type": solution_type,
        }
        if description:
            data["description"] = description
        if expected_energy_saving is not None:
            data["expected_energy_saving"] = expected_energy_saving
        if expected_investment is not None:
            data["expected_investment"] = expected_investment
        response = self.session.post(
            self._url(f"/audits/{audit_id}/solutions"),
            json=data,
            params=params,
        )
        response.raise_for_status()
        return response.json()

    def add_dimensions(
        self,
        solution_id: int,
        dimensions: List[Dict[str, Any]],
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {}
        if user_id:
            params["user_id"] = user_id
        if user_name:
            params["user_name"] = user_name
        response = self.session.post(
            self._url(f"/solutions/{solution_id}/dimensions"),
            json=dimensions,
            params=params,
        )
        response.raise_for_status()
        return response.json()

    def filter_solution(
        self,
        solution_id: int,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {}
        if user_id:
            params["user_id"] = user_id
        if user_name:
            params["user_name"] = user_name
        response = self.session.post(
            self._url(f"/solutions/{solution_id}/filter"),
            params=params,
        )
        response.raise_for_status()
        return response.json()

    def implement_solution(
        self,
        solution_id: int,
        actual_energy_saving: float,
        actual_investment: float,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {}
        if user_id:
            params["user_id"] = user_id
        if user_name:
            params["user_name"] = user_name
        data = {
            "actual_energy_saving": actual_energy_saving,
            "actual_investment": actual_investment,
        }
        response = self.session.post(
            self._url(f"/solutions/{solution_id}/implement"),
            json=data,
            params=params,
        )
        response.raise_for_status()
        return response.json()

    def perform_acceptance(
        self,
        audit_id: int,
        score: float,
        notes: Optional[str] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {}
        if user_id:
            params["user_id"] = user_id
        if user_name:
            params["user_name"] = user_name
        data = {"score": score}
        if notes:
            data["notes"] = notes
        response = self.session.post(
            self._url(f"/audits/{audit_id}/acceptance"),
            json=data,
            params=params,
        )
        response.raise_for_status()
        return response.json()

    def get_audit_summary(self, audit_id: int) -> Dict[str, Any]:
        response = self.session.get(self._url(f"/audits/{audit_id}/summary"))
        response.raise_for_status()
        return response.json()

    def list_logs(self, audit_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {}
        if audit_id is not None:
            params["audit_id"] = audit_id
        response = self.session.get(self._url("/logs"), params=params)
        response.raise_for_status()
        return response.json()
