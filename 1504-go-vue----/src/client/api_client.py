import os
import requests
from typing import Optional, Dict, Any, List

class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("SERVER_URL", "http://localhost:8000")
        if not self.base_url.endswith("/"):
            self.base_url += "/"

    def _request(self, method: str, endpoint: str, **kwargs) -> Any:
        url = f"{self.base_url}api{endpoint}"
        try:
            response = requests.request(method, url, **kwargs)
            response.raise_for_status()
            if response.content:
                return response.json()
            return None
        except requests.exceptions.HTTPError as e:
            if e.response is not None:
                error_detail = e.response.json().get("detail", str(e))
                raise Exception(f"API 错误: {error_detail}")
            raise
        except requests.exceptions.RequestException as e:
            raise Exception(f"请求失败: {str(e)}")

    def create_batch(self, breed: str, initial_count: int, gestation_period_days: int = 114) -> Dict:
        data = {
            "breed": breed,
            "initial_count": initial_count,
            "gestation_period_days": gestation_period_days
        }
        return self._request("POST", "/batches", json=data)

    def list_batches(self) -> List[Dict]:
        return self._request("GET", "/batches")

    def get_batch(self, batch_id: int) -> Dict:
        return self._request("GET", f"/batches/{batch_id}")

    def create_breeding_record(self, batch_id: int, female_id: str, breeding_date: str) -> Dict:
        data = {
            "batch_id": batch_id,
            "female_id": female_id,
            "breeding_date": breeding_date
        }
        return self._request("POST", "/breeding", json=data)

    def list_breeding_records(self, batch_id: Optional[int] = None) -> List[Dict]:
        params = {"batch_id": batch_id} if batch_id else {}
        return self._request("GET", "/breeding", params=params)

    def record_breeding_failure(self, record_id: int, failure_reason: str) -> Dict:
        data = {"failure_reason": failure_reason}
        return self._request("POST", f"/breeding/{record_id}/failure", json=data)

    def record_breeding_success(self, record_id: int) -> Dict:
        return self._request("POST", f"/breeding/{record_id}/success")

    def create_vaccination_record(
        self, batch_id: int, vaccine_name: str, 
        vaccination_date: str, interval_days: Optional[int] = None
    ) -> Dict:
        data = {
            "batch_id": batch_id,
            "vaccine_name": vaccine_name,
            "vaccination_date": vaccination_date,
            "interval_days": interval_days
        }
        return self._request("POST", "/vaccination", json=data)

    def list_vaccination_records(self, batch_id: Optional[int] = None) -> List[Dict]:
        params = {"batch_id": batch_id} if batch_id else {}
        return self._request("GET", "/vaccination", params=params)

    def create_slaughter_record(
        self, batch_id: int, count: int, avg_weight: float,
        unit_price: float, slaughter_date: str
    ) -> Dict:
        data = {
            "batch_id": batch_id,
            "count": count,
            "avg_weight": avg_weight,
            "unit_price": unit_price,
            "slaughter_date": slaughter_date
        }
        return self._request("POST", "/slaughter", json=data)

    def list_slaughter_records(self, batch_id: Optional[int] = None) -> List[Dict]:
        params = {"batch_id": batch_id} if batch_id else {}
        return self._request("GET", "/slaughter", params=params)

    def generate_reminders(self) -> List[Dict]:
        return self._request("POST", "/reminders/generate")

    def list_reminders(self, is_read: Optional[bool] = None) -> List[Dict]:
        params = {"is_read": is_read} if is_read is not None else {}
        return self._request("GET", "/reminders", params=params)

    def mark_reminder_read(self, reminder_id: int) -> Dict:
        return self._request("POST", f"/reminders/{reminder_id}/read")

    def get_dashboard(self) -> Dict:
        return self._request("GET", "/dashboard")

api_client = APIClient()
