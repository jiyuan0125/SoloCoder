import requests
from typing import Optional, List, Dict, Any
import os


class ApiClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get("API_URL", "http://localhost:8000")
        self.session = requests.Session()

    def _request(self, method: str, endpoint: str, **kwargs) -> Any:
        url = f"{self.base_url}{endpoint}"
        response = self.session.request(method, url, **kwargs)
        if response.status_code >= 400:
            raise Exception(f"API Error ({response.status_code}): {response.text}")
        if response.status_code == 204:
            return None
        if "json" in response.headers.get("content-type", ""):
            return response.json()
        return response.text

    def health(self):
        return self._request("GET", "/api/health")

    def get_hw_codes(self):
        return self._request("GET", "/api/hw-codes")

    def create_producer(self, producer: Dict[str, Any]):
        return self._request("POST", "/api/producers", json=producer)

    def list_producers(self):
        return self._request("GET", "/api/producers")

    def create_company(self, company: Dict[str, Any]):
        return self._request("POST", "/api/companies", json=company)

    def list_companies(self):
        return self._request("GET", "/api/companies")

    def create_waste(self, waste: Dict[str, Any]):
        return self._request("POST", "/api/wastes", json=waste)

    def list_wastes(self, producer_id: Optional[str] = None, status: Optional[str] = None):
        params = {}
        if producer_id:
            params["producer_id"] = producer_id
        if status:
            params["status"] = status
        return self._request("GET", "/api/wastes", params=params)

    def store_waste(self, waste_id: str, location: str):
        return self._request("POST", f"/api/wastes/{waste_id}/store", params={"location": location})

    def create_ledger(self, producer_id: str, waste_id: str, year: int, month: int,
                       generated: float, transferred: float, beginning: float, ending: float):
        params = {
            "producer_id": producer_id,
            "waste_id": waste_id,
            "year": year,
            "month": month,
            "generated": generated,
            "transferred": transferred,
            "beginning": beginning,
            "ending": ending
        }
        return self._request("POST", "/api/ledgers", params=params)

    def submit_ledger(self, ledger_id: str):
        return self._request("POST", f"/api/ledgers/{ledger_id}/submit")

    def list_ledgers(self, producer_id: str, year: int, month: int):
        params = {"producer_id": producer_id, "year": year, "month": month}
        return self._request("GET", "/api/ledgers", params=params)

    def create_summary(self, producer_id: str, year: int, month: int):
        params = {"producer_id": producer_id, "year": year, "month": month}
        return self._request("POST", "/api/ledgers/summary", params=params)

    def list_summaries(self, producer_id: Optional[str] = None):
        params = {"producer_id": producer_id} if producer_id else {}
        return self._request("GET", "/api/ledgers/summary", params=params)

    def create_transfer(self, producer_id: str, receiver_id: str, transporter_id: str,
                        waste_ids: List[str], total_quantity: float, transfer_date: str):
        params = {
            "producer_id": producer_id,
            "receiver_id": receiver_id,
            "transporter_id": transporter_id,
            "waste_ids": waste_ids,
            "total_quantity": total_quantity,
            "transfer_date": transfer_date
        }
        return self._request("POST", "/api/transfers", params=params)

    def confirm_transfer(self, doc_id: str, party: str):
        return self._request("POST", f"/api/transfers/{doc_id}/confirm", params={"party": party})

    def reject_transfer(self, doc_id: str, party: str, reason: str):
        return self._request("POST", f"/api/transfers/{doc_id}/reject",
                              params={"party": party, "reason": reason})

    def resolve_exception(self, exc_id: str, resolution: str):
        return self._request("POST", f"/api/exceptions/{exc_id}/resolve",
                              params={"resolution": resolution})

    def list_transfers(self, producer_id: Optional[str] = None, status: Optional[str] = None):
        params = {}
        if producer_id:
            params["producer_id"] = producer_id
        if status:
            params["status"] = status
        return self._request("GET", "/api/transfers", params=params)

    def record_disposal(self, waste_id: str, disposal_company_id: str,
                         disposal_method: str, disposal_date: str, quantity: float):
        params = {
            "waste_id": waste_id,
            "disposal_company_id": disposal_company_id,
            "disposal_method": disposal_method,
            "disposal_date": disposal_date,
            "quantity": quantity
        }
        return self._request("POST", "/api/disposals", params=params)

    def list_disposals(self, waste_id: Optional[str] = None):
        params = {"waste_id": waste_id} if waste_id else {}
        return self._request("GET", "/api/disposals", params=params)

    def check_storage_alerts(self):
        return self._request("POST", "/api/alerts/check-storage")

    def check_supervision_alerts(self):
        return self._request("POST", "/api/alerts/check-supervision")

    def list_alerts(self, unread_only: bool = False):
        return self._request("GET", "/api/alerts", params={"unread_only": unread_only})

    def mark_alert_read(self, alert_id: str):
        return self._request("POST", f"/api/alerts/{alert_id}/read")

    def validate_transfer(self, receiver_id: str, waste_ids: List[str]):
        return self._request("GET", "/api/validate/transfer",
                              params={"receiver_id": receiver_id, "waste_ids": waste_ids})

    def export_wastes(self, producer_id: Optional[str] = None):
        params = {"producer_id": producer_id} if producer_id else {}
        return self._request("GET", "/api/export/wastes.csv", params=params)

    def export_ledgers(self, producer_id: str, year: int, month: int):
        params = {"producer_id": producer_id, "year": year, "month": month}
        return self._request("GET", "/api/export/ledgers.csv", params=params)

    def export_transfers(self, producer_id: Optional[str] = None):
        params = {"producer_id": producer_id} if producer_id else {}
        return self._request("GET", "/api/export/transfers.csv", params=params)
