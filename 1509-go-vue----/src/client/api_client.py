import httpx
from typing import List, Optional, Dict, Any
from datetime import datetime


class TraceabilityAPIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")
        self.client = httpx.Client()

    def _request(self, method: str, path: str, **kwargs) -> Any:
        url = f"{self.base_url}{path}"
        response = self.client.request(method, url, **kwargs)
        if response.status_code >= 400:
            error_detail = response.json().get("detail", response.text)
            raise Exception(f"API 错误 ({response.status_code}): {error_detail}")
        if response.status_code == 204:
            return None
        return response.json()

    def get_status(self) -> Dict[str, Any]:
        return self._request("GET", "/")

    def add_supply_chain_record(
        self,
        batch_number: str,
        stage: str,
        operation_time: datetime,
        operator: str,
        location: Optional[str] = None,
        remarks: Optional[str] = None
    ) -> Dict[str, Any]:
        payload = {
            "batch_number": batch_number,
            "stage": stage,
            "operation_time": operation_time.isoformat(),
            "operator": operator
        }
        if location:
            payload["location"] = location
        if remarks:
            payload["remarks"] = remarks
        return self._request("POST", "/api/supply-chain", json=payload)

    def get_supply_chain_timeline(self, batch_number: str) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/supply-chain/{batch_number}")

    def list_supply_chain(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/supply-chain")

    def add_inspection_record(
        self,
        batch_number: str,
        inspector: str,
        status: str,
        items: List[str] = None,
        report: Optional[str] = None,
        remarks: Optional[str] = None
    ) -> Dict[str, Any]:
        payload = {
            "batch_number": batch_number,
            "inspection_time": datetime.now().isoformat(),
            "inspector": inspector,
            "status": status,
            "items": items or []
        }
        if report:
            payload["report"] = report
        if remarks:
            payload["remarks"] = remarks
        return self._request("POST", "/api/inspections", json=payload)

    def get_inspections_by_batch(self, batch_number: str) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/inspections/{batch_number}")

    def list_inspections(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/inspections")

    def create_recall(self, batch_number: str, reason: str) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/api/recalls?batch_number={batch_number}&reason={reason}"
        )

    def get_recall(self, recall_id: str) -> Dict[str, Any]:
        return self._request("GET", f"/api/recalls/{recall_id}")

    def list_recalls(self, batch_number: Optional[str] = None) -> List[Dict[str, Any]]:
        if batch_number:
            return self._request("GET", f"/api/recalls?batch_number={batch_number}")
        return self._request("GET", "/api/recalls")

    def complete_recall(self, recall_id: str, completed_by: str) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/api/recalls/{recall_id}/complete?completed_by={completed_by}"
        )

    def cancel_recall(self, recall_id: str, cancelled_by: str) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/api/recalls/{recall_id}/cancel?cancelled_by={cancelled_by}"
        )

    def get_recall_todos(self, recall_id: str) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/recalls/{recall_id}/todos")

    def list_todos(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/todos")

    def confirm_todo(
        self,
        todo_id: str,
        confirmed_by: str,
        remarks: Optional[str] = None
    ) -> Dict[str, Any]:
        url = f"/api/todos/{todo_id}/confirm?confirmed_by={confirmed_by}"
        if remarks:
            url += f"&remarks={remarks}"
        return self._request("POST", url)

    def update_overdue_todos(self) -> Dict[str, Any]:
        return self._request("POST", "/api/todos/update-overdue")

    def export_batch_traceability(self, batch_number: str) -> str:
        url = f"{self.base_url}/api/export/{batch_number}"
        response = self.client.get(url)
        if response.status_code >= 400:
            raise Exception(f"导出失败: {response.text}")
        return response.text

    def close(self):
        self.client.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()
