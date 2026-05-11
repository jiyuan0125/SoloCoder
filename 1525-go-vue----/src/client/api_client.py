import os
import httpx
from typing import Dict, List, Any, Optional
from datetime import date


class ApiClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_BASE_URL", "http://localhost:8000")
        self.client = httpx.Client(base_url=self.base_url)

    def close(self) -> None:
        self.client.close()

    def _request(self, method: str, path: str, **kwargs) -> Any:
        try:
            response = self.client.request(method, path, **kwargs)
            response.raise_for_status()
            if response.status_code == 204:
                return None
            return response.json()
        except httpx.HTTPStatusError as e:
            raise RuntimeError(f"API 请求失败: {e.response.status_code} - {e.response.text}")
        except httpx.HTTPError as e:
            raise RuntimeError(f"网络错误: {e}")

    def list_certificates(self, status: Optional[str] = None, cert_type: Optional[str] = None) -> List[Dict]:
        params = {}
        if status:
            params["status"] = status
        if cert_type:
            params["cert_type"] = cert_type
        return self._request("GET", "/certificates/", params=params)

    def get_certificate(self, cert_id: int) -> Dict:
        return self._request("GET", f"/certificates/{cert_id}/")

    def create_certificate(
        self,
        name: str,
        cert_type: str,
        number: str,
        issuing_date: date,
        expiry_date: date,
        remarks: Optional[str] = None,
    ) -> Dict:
        data = {
            "name": name,
            "type": cert_type,
            "number": number,
            "issuing_date": issuing_date.isoformat(),
            "expiry_date": expiry_date.isoformat(),
            "remarks": remarks,
        }
        return self._request("POST", "/certificates/", json=data)

    def update_certificate(self, cert_id: int, **kwargs) -> Dict:
        data = {}
        if "name" in kwargs:
            data["name"] = kwargs["name"]
        if "number" in kwargs:
            data["number"] = kwargs["number"]
        if "issuing_date" in kwargs:
            data["issuing_date"] = kwargs["issuing_date"].isoformat()
        if "expiry_date" in kwargs:
            data["expiry_date"] = kwargs["expiry_date"].isoformat()
        if "remarks" in kwargs:
            data["remarks"] = kwargs["remarks"]
        return self._request("PUT", f"/certificates/{cert_id}/", json=data)

    def cancel_certificate(self, cert_id: int) -> Dict:
        return self._request("POST", f"/certificates/{cert_id}/cancel/")

    def list_inspections(self, certificate_id: Optional[int] = None) -> List[Dict]:
        params = {}
        if certificate_id:
            params["certificate_id"] = certificate_id
        return self._request("GET", "/inspections/", params=params)

    def create_inspection(
        self,
        certificate_id: int,
        year: int,
        inspection_date: date,
        result: bool,
        remarks: Optional[str] = None,
    ) -> Dict:
        data = {
            "certificate_id": certificate_id,
            "year": year,
            "inspection_date": inspection_date.isoformat(),
            "result": result,
            "remarks": remarks,
        }
        return self._request("POST", "/inspections/", json=data)

    def list_compliance_checks(self, certificate_id: Optional[int] = None) -> List[Dict]:
        params = {}
        if certificate_id:
            params["certificate_id"] = certificate_id
        return self._request("GET", "/compliance-checks/", params=params)

    def create_compliance_check(
        self,
        certificate_id: int,
        check_date: date,
        check_items: str,
        is_compliant: bool,
        has_safety_issues: bool = False,
        remarks: Optional[str] = None,
    ) -> Dict:
        data = {
            "certificate_id": certificate_id,
            "check_date": check_date.isoformat(),
            "check_items": check_items,
            "is_compliant": is_compliant,
            "has_safety_issues": has_safety_issues,
            "remarks": remarks,
        }
        return self._request("POST", "/compliance-checks/", json=data)

    def list_todos(self, status: Optional[str] = None, todo_type: Optional[str] = None) -> List[Dict]:
        params = {}
        if status:
            params["status"] = status
        if todo_type:
            params["todo_type"] = todo_type
        return self._request("GET", "/todos/", params=params)

    def update_todo_status(self, todo_id: int, status: str) -> Dict:
        data = {"status": status}
        return self._request("PUT", f"/todos/{todo_id}/", json=data)

    def process_reminders(self) -> Dict:
        return self._request("POST", "/todos/process-reminders/")