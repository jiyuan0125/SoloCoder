import os
from typing import Optional, Dict, Any, List
import httpx


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get("API_URL", "http://localhost:8000")
        self.client = httpx.Client(base_url=self.base_url, timeout=30.0)

    def close(self):
        self.client.close()

    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        try:
            response = self.client.request(method, endpoint, **kwargs)
            if response.status_code >= 400:
                try:
                    error_detail = response.json().get("detail", response.text)
                except Exception:
                    error_detail = response.text
                raise Exception(f"HTTP {response.status_code}: {error_detail}")
            return response.json() if response.content else {}
        except httpx.RequestError as e:
            raise Exception(f"连接服务器失败: {e}")

    def get(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        return self._request("GET", endpoint, params=params)

    def post(self, endpoint: str, json: Optional[Dict[str, Any]] = None, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        return self._request("POST", endpoint, json=json, params=params)

    def put(self, endpoint: str, json: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        return self._request("PUT", endpoint, json=json)

    def delete(self, endpoint: str) -> Dict[str, Any]:
        return self._request("DELETE", endpoint)

    def health_check(self) -> bool:
        try:
            self.get("/health")
            return True
        except Exception:
            return False

    def create_animal(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/animals", json=data)

    def list_animals(self, **filters) -> List[Dict[str, Any]]:
        return self.get("/animals", params={k: v for k, v in filters.items() if v is not None})

    def get_animal(self, animal_id: str) -> Dict[str, Any]:
        return self.get(f"/animals/{animal_id}")

    def update_animal(self, animal_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.put(f"/animals/{animal_id}", json=data)

    def delete_animal(self, animal_id: str) -> Dict[str, Any]:
        return self.delete(f"/animals/{animal_id}")

    def create_adopter(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/adopters", json=data)

    def list_adopters(self, **filters) -> List[Dict[str, Any]]:
        return self.get("/adopters", params={k: v for k, v in filters.items() if v is not None})

    def get_adopter(self, adopter_id: str) -> Dict[str, Any]:
        return self.get(f"/adopters/{adopter_id}")

    def update_adopter(self, adopter_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.put(f"/adopters/{adopter_id}", json=data)

    def delete_adopter(self, adopter_id: str) -> Dict[str, Any]:
        return self.delete(f"/adopters/{adopter_id}")

    def create_adoption(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/adoptions", json=data)

    def list_adoptions(self, **filters) -> List[Dict[str, Any]]:
        return self.get("/adoptions", params={k: v for k, v in filters.items() if v is not None})

    def get_adoption(self, app_id: str) -> Dict[str, Any]:
        return self.get(f"/adoptions/{app_id}")

    def approve_adoption(self, app_id: str) -> Dict[str, Any]:
        return self.post(f"/adoptions/{app_id}/approve")

    def approve_adoption_extra(self, app_id: str) -> Dict[str, Any]:
        return self.post(f"/adoptions/{app_id}/approve-extra")

    def reject_adoption(self, app_id: str) -> Dict[str, Any]:
        return self.post(f"/adoptions/{app_id}/reject")

    def cancel_adoption(self, app_id: str) -> Dict[str, Any]:
        return self.post(f"/adoptions/{app_id}/cancel")

    def finalize_adoption(self, app_id: str) -> Dict[str, Any]:
        return self.post(f"/adoptions/{app_id}/finalize")

    def list_follow_ups(self, **filters) -> List[Dict[str, Any]]:
        return self.get("/follow-ups", params={k: v for k, v in filters.items() if v is not None})

    def get_follow_up(self, fu_id: str) -> Dict[str, Any]:
        return self.get(f"/follow-ups/{fu_id}")

    def update_follow_up(self, fu_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.put(f"/follow-ups/{fu_id}", json=data)

    def complete_follow_up(self, fu_id: str, notes: Optional[str] = None) -> Dict[str, Any]:
        return self.post(f"/follow-ups/{fu_id}/complete", params={"notes": notes} if notes else None)

    def check_overdue_follow_ups(self) -> Dict[str, Any]:
        return self.post("/follow-ups/check-overdue")

    def create_donation(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/donations", json=data)

    def list_donations(self, **filters) -> List[Dict[str, Any]]:
        return self.get("/donations", params={k: v for k, v in filters.items() if v is not None})

    def get_donation_stats(self) -> Dict[str, Any]:
        return self.get("/donations/stats")

    def get_donation(self, don_id: str) -> Dict[str, Any]:
        return self.get(f"/donations/{don_id}")

    def create_appointment(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/appointments", json=data)

    def list_appointments(self, **filters) -> List[Dict[str, Any]]:
        return self.get("/appointments", params={k: v for k, v in filters.items() if v is not None})

    def get_appointment(self, appt_id: str) -> Dict[str, Any]:
        return self.get(f"/appointments/{appt_id}")

    def update_appointment(self, appt_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return self.put(f"/appointments/{appt_id}", json=data)

    def complete_appointment(self, appt_id: str, passed: bool, notes: Optional[str] = None) -> Dict[str, Any]:
        params = {"passed": passed}
        if notes:
            params["notes"] = notes
        return self.post(f"/appointments/{appt_id}/complete", params=params)

    def cancel_appointment(self, appt_id: str) -> Dict[str, Any]:
        return self.post(f"/appointments/{appt_id}/cancel")
