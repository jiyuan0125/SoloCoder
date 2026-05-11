import os
import requests
from typing import Optional, Dict, Any, List
from datetime import date


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get(
            "API_BASE_URL", "http://localhost:8000/api/v1"
        )
        self.session = requests.Session()

    def _request(self, method: str, endpoint: str, **kwargs) -> Any:
        url = f"{self.base_url}{endpoint}"
        try:
            response = self.session.request(method, url, **kwargs)
            response.raise_for_status()
            if response.status_code == 204:
                return None
            return response.json()
        except requests.RequestException as e:
            if hasattr(e, 'response') and e.response is not None:
                try:
                    detail = e.response.json().get("detail", str(e))
                except Exception:
                    detail = str(e)
                raise Exception(f"API Error: {detail}")
            raise Exception(f"Connection Error: {e}")

    def list_projects(self) -> List[Dict]:
        return self._request("GET", "/projects/")

    def create_project(self, name: str, duration: int, price: float) -> Dict:
        return self._request(
            "POST",
            "/projects/",
            json={"name": name, "duration_minutes": duration, "price": price}
        )

    def delete_project(self, project_id: int) -> Dict:
        return self._request("DELETE", f"/projects/{project_id}")

    def list_stylists(self) -> List[Dict]:
        return self._request("GET", "/stylists/")

    def create_stylist(self, name: str, status: str = "available", skilled_projects: List[int] = None) -> Dict:
        return self._request(
            "POST",
            "/stylists/",
            json={
                "name": name,
                "status": status,
                "skilled_projects": skilled_projects or []
            }
        )

    def list_clients(self) -> List[Dict]:
        return self._request("GET", "/clients/")

    def create_client(self, name: str, phone: str, pets: List[str] = None) -> Dict:
        return self._request(
            "POST",
            "/clients/",
            json={"name": name, "phone": phone, "pets": pets or []}
        )

    def get_available_slots(self, stylist_id: int, project_id: int, appointment_date: date) -> List[str]:
        result = self._request(
            "GET",
            "/appointments/slots",
            params={
                "stylist_id": stylist_id,
                "project_id": project_id,
                "appointment_date": appointment_date.isoformat()
            }
        )
        return result.get("available_slots", [])

    def create_appointment(self, client_id: int, pet_name: str, project_id: int,
                          stylist_id: int, appointment_date: date, start_time: str) -> Dict:
        return self._request(
            "POST",
            "/appointments/",
            json={
                "client_id": client_id,
                "pet_name": pet_name,
                "project_id": project_id,
                "stylist_id": stylist_id,
                "appointment_date": appointment_date.isoformat(),
                "start_time": start_time
            }
        )

    def list_appointments(self) -> List[Dict]:
        return self._request("GET", "/appointments/")

    def cancel_appointment(self, appointment_id: int) -> Dict:
        return self._request("POST", f"/appointments/{appointment_id}/cancel")

    def complete_appointment(self, appointment_id: int, actual_duration: int) -> Dict:
        return self._request(
            "POST",
            f"/appointments/{appointment_id}/complete",
            json={"actual_duration_minutes": actual_duration}
        )

    def create_review(self, appointment_id: int, client_id: int, rating: int, comment: str = None) -> Dict:
        data = {
            "appointment_id": appointment_id,
            "client_id": client_id,
            "rating": rating
        }
        if comment:
            data["comment"] = comment
        return self._request("POST", "/reviews/", json=data)

    def list_reviews(self) -> List[Dict]:
        return self._request("GET", "/reviews/")

    def list_supplies(self) -> List[Dict]:
        return self._request("GET", "/supplies/")

    def create_supply(self, name: str, current_stock: int, min_stock: int, unit: str) -> Dict:
        return self._request(
            "POST",
            "/supplies/",
            json={
                "name": name,
                "current_stock": current_stock,
                "min_stock": min_stock,
                "unit": unit
            }
        )

    def list_purchase_todos(self) -> List[Dict]:
        return self._request("GET", "/purchase-todos/")

    def complete_purchase_todo(self, todo_id: int) -> Dict:
        return self._request("POST", f"/purchase-todos/{todo_id}/complete")
