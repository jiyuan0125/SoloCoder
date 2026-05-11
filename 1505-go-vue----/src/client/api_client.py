import os
from typing import Any, Dict, List, Optional

import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get("API_URL", "http://127.0.0.1:8000")

    def _get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        resp = requests.get(f"{self.base_url}{path}", params=params)
        if resp.status_code >= 400:
            raise RuntimeError(f"请求失败 [{resp.status_code}]: {resp.text}")
        if resp.headers.get("content-type", "").startswith("text"):
            return resp.text
        return resp.json()

    def _post(self, path: str, data: Optional[Dict[str, Any]] = None) -> Any:
        resp = requests.post(f"{self.base_url}{path}", json=data or {})
        if resp.status_code >= 400:
            raise RuntimeError(f"请求失败 [{resp.status_code}]: {resp.text}")
        return resp.json()

    def create_recipe(self, name: str, category: str, raw_materials: List[Dict[str, Any]]) -> Dict[str, Any]:
        return self._post("/recipes", {
            "name": name,
            "category": category,
            "raw_materials": raw_materials,
        })

    def list_recipes(self) -> List[Dict[str, Any]]:
        return self._get("/recipes")

    def get_recipe(self, recipe_id: str) -> Dict[str, Any]:
        return self._get(f"/recipes/{recipe_id}")

    def upsert_material(self, name: str, current_stock: float, safety_stock: float) -> Dict[str, Any]:
        return self._post("/raw-materials", {
            "name": name,
            "current_stock": current_stock,
            "safety_stock": safety_stock,
        })

    def list_materials(self, low_stock: bool = False) -> List[Dict[str, Any]]:
        return self._get("/raw-materials", params={"low_stock": low_stock})

    def get_material(self, material_id: str) -> Dict[str, Any]:
        return self._get(f"/raw-materials/{material_id}")

    def create_batch(self, recipe_id: str, plan_quantity: int, actual_quantity: int) -> Dict[str, Any]:
        return self._post("/batches", {
            "recipe_id": recipe_id,
            "plan_quantity": plan_quantity,
            "actual_quantity": actual_quantity,
        })

    def list_batches(self, start: Optional[str] = None, end: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {}
        if start:
            params["start"] = start
        if end:
            params["end"] = end
        return self._get("/batches", params=params if params else None)

    def get_batch(self, batch_id: str) -> Dict[str, Any]:
        return self._get(f"/batches/{batch_id}")

    def add_quality_check(
        self,
        batch_id: str,
        inspector: str,
        result: str,
        measured_value: Optional[float],
        notes: str = "",
    ) -> Dict[str, Any]:
        return self._post(f"/batches/{batch_id}/quality-checks", {
            "inspector": inspector,
            "result": result,
            "measured_value": measured_value,
            "notes": notes,
        })

    def list_todos(self) -> List[Dict[str, Any]]:
        return self._get("/todos")

    def resolve_todo(self, todo_id: str, disposition: str, resolved_by: str) -> Dict[str, Any]:
        return self._post(f"/todos/{todo_id}/resolve", {
            "disposition": disposition,
            "resolved_by": resolved_by,
        })

    def export_batches(self, start: Optional[str] = None, end: Optional[str] = None) -> str:
        params = {}
        if start:
            params["start"] = start
        if end:
            params["end"] = end
        return self._get("/export/batches", params=params if params else None)
