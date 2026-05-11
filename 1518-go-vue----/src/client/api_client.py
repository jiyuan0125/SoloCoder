from datetime import date
from typing import Any, Dict, List, Optional

import httpx

from client.config import BASE_URL


class VetAPIClient:
    def __init__(self, base_url: str = BASE_URL):
        self.base_url = base_url
        self.client = httpx.Client(base_url=base_url)

    def close(self):
        self.client.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def health_check(self) -> Dict[str, Any]:
        resp = self.client.get("/health")
        resp.raise_for_status()
        return resp.json()

    def create_farmer(self, name: str, address: str, livestock_type: str, scale: int) -> Dict[str, Any]:
        data = {"name": name, "address": address, "livestock_type": livestock_type, "scale": scale}
        resp = self.client.post("/farmers", json=data)
        resp.raise_for_status()
        return resp.json()

    def list_farmers(self) -> List[Dict[str, Any]]:
        resp = self.client.get("/farmers")
        resp.raise_for_status()
        return resp.json()

    def get_farmer(self, farmer_id: int) -> Dict[str, Any]:
        resp = self.client.get(f"/farmers/{farmer_id}")
        resp.raise_for_status()
        return resp.json()

    def update_farmer(self, farmer_id: int, **kwargs) -> Dict[str, Any]:
        data = {k: v for k, v in kwargs.items() if v is not None}
        resp = self.client.put(f"/farmers/{farmer_id}", json=data)
        resp.raise_for_status()
        return resp.json()

    def delete_farmer(self, farmer_id: int) -> None:
        resp = self.client.delete(f"/farmers/{farmer_id}")
        resp.raise_for_status()

    def create_medicine(self, name: str, specification: str, unit: str, stock: float) -> Dict[str, Any]:
        data = {"name": name, "specification": specification, "unit": unit, "stock": stock}
        resp = self.client.post("/medicines", json=data)
        resp.raise_for_status()
        return resp.json()

    def list_medicines(self) -> List[Dict[str, Any]]:
        resp = self.client.get("/medicines")
        resp.raise_for_status()
        return resp.json()

    def get_medicine(self, medicine_id: int) -> Dict[str, Any]:
        resp = self.client.get(f"/medicines/{medicine_id}")
        resp.raise_for_status()
        return resp.json()

    def add_medicine_stock(self, medicine_id: int, amount: float) -> Dict[str, Any]:
        resp = self.client.post(f"/medicines/{medicine_id}/add-stock", params={"amount": amount})
        resp.raise_for_status()
        return resp.json()

    def generate_route(self, route_date: Optional[date] = None) -> Dict[str, Any]:
        params = {}
        if route_date:
            params["route_date"] = route_date.isoformat()
        resp = self.client.post("/routes/generate", params=params)
        resp.raise_for_status()
        return resp.json()

    def get_route(self, route_date: date) -> Dict[str, Any]:
        resp = self.client.get(f"/routes/{route_date.isoformat()}")
        resp.raise_for_status()
        return resp.json()

    def reorder_route(self, route_id: int, items: List[Dict[str, Any]]) -> Dict[str, Any]:
        data = {"items": items}
        resp = self.client.put(f"/routes/{route_id}/reorder", json=data)
        resp.raise_for_status()
        return resp.json()

    def complete_route_item(self, item_id: int, completed: bool = True) -> Dict[str, Any]:
        resp = self.client.post(f"/routes/items/{item_id}/complete", params={"completed": completed})
        resp.raise_for_status()
        return resp.json()

    def create_visit(
        self,
        farmer_id: int,
        visit_date: date,
        has_abnormality: bool,
        notes: Optional[str] = None,
        cases: Optional[List[Dict[str, Any]]] = None
    ) -> Dict[str, Any]:
        data = {
            "farmer_id": farmer_id,
            "visit_date": visit_date.isoformat(),
            "has_abnormality": has_abnormality,
            "notes": notes,
            "cases": cases or []
        }
        resp = self.client.post("/visits", json=data)
        resp.raise_for_status()
        return resp.json()

    def list_visits(self, farmer_id: Optional[int] = None, visit_date: Optional[date] = None) -> List[Dict[str, Any]]:
        params = {}
        if farmer_id:
            params["farmer_id"] = farmer_id
        if visit_date:
            params["visit_date"] = visit_date.isoformat()
        resp = self.client.get("/visits", params=params)
        resp.raise_for_status()
        return resp.json()

    def get_visit(self, visit_id: int) -> Dict[str, Any]:
        resp = self.client.get(f"/visits/{visit_id}")
        resp.raise_for_status()
        return resp.json()

    def list_todos(self, todo_date: Optional[date] = None, pending: bool = False) -> List[Dict[str, Any]]:
        params = {"pending": pending}
        if todo_date:
            params["todo_date"] = todo_date.isoformat()
        resp = self.client.get("/todos", params=params)
        resp.raise_for_status()
        return resp.json()

    def update_todo(self, todo_id: int, is_completed: Optional[bool] = None, notes: Optional[str] = None) -> Dict[str, Any]:
        data = {}
        if is_completed is not None:
            data["is_completed"] = is_completed
        if notes is not None:
            data["notes"] = notes
        resp = self.client.put(f"/todos/{todo_id}", json=data)
        resp.raise_for_status()
        return resp.json()

    def get_dashboard_stats(self, target_date: Optional[date] = None) -> Dict[str, Any]:
        params = {}
        if target_date:
            params["target_date"] = target_date.isoformat()
        resp = self.client.get("/dashboard/stats", params=params)
        resp.raise_for_status()
        return resp.json()
