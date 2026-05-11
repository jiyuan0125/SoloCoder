import os
import httpx
from typing import Optional, Dict, Any, List


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get(
            "GAP_SERVER_URL", "http://localhost:8000"
        )

    def _get(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client() as client:
            response = client.get(f"{self.base_url}{endpoint}", params=params)
            self._check_response(response)
            return response.json()

    def _post(self, endpoint: str, data: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client() as client:
            response = client.post(f"{self.base_url}{endpoint}", json=data)
            self._check_response(response)
            return response.json()

    def _put(self, endpoint: str, data: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client() as client:
            response = client.put(f"{self.base_url}{endpoint}", json=data)
            self._check_response(response)
            return response.json()

    def _delete(self, endpoint: str) -> Any:
        with httpx.Client() as client:
            response = client.delete(f"{self.base_url}{endpoint}")
            self._check_response(response)
            return response.json()

    def _get_text(self, endpoint: str) -> str:
        with httpx.Client() as client:
            response = client.get(f"{self.base_url}{endpoint}")
            self._check_response(response)
            return response.text

    @staticmethod
    def _check_response(response: httpx.Response):
        if response.status_code >= 400:
            try:
                error_detail = response.json().get("detail", response.text)
            except Exception:
                error_detail = response.text
            raise RuntimeError(f"API Error ({response.status_code}): {error_detail}")

    def health_check(self) -> Dict[str, Any]:
        return self._get("/api/health")

    def create_plot(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self._post("/api/plots", data)

    def list_plots(self) -> List[Dict[str, Any]]:
        return self._get("/api/plots")

    def get_plot(self, plot_id: str) -> Dict[str, Any]:
        return self._get(f"/api/plots/{plot_id}")

    def update_plot(self, plot_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return self._put(f"/api/plots/{plot_id}", data)

    def delete_plot(self, plot_id: str) -> Dict[str, Any]:
        return self._delete(f"/api/plots/{plot_id}")

    def create_operation(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self._post("/api/operations", data)

    def list_operations(self, plot_id: str) -> List[Dict[str, Any]]:
        return self._get("/api/operations", params={"plot_id": plot_id})

    def create_harvest(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self._post("/api/harvests", data)

    def list_harvests(self, plot_id: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"plot_id": plot_id} if plot_id else None
        return self._get("/api/harvests", params=params)

    def get_harvest(self, harvest_id: str) -> Dict[str, Any]:
        return self._get(f"/api/harvests/{harvest_id}")

    def update_harvest(self, harvest_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return self._put(f"/api/harvests/{harvest_id}", data)

    def create_processing(self, data: Dict[str, Any]) -> Dict[str, Any]:
        return self._post("/api/processings", data)

    def list_processings(self, harvest_id: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"harvest_id": harvest_id} if harvest_id else None
        return self._get("/api/processings", params=params)

    def update_processing(self, processing_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return self._put(f"/api/processings/{processing_id}", data)

    def list_todos(self) -> List[Dict[str, Any]]:
        return self._get("/api/todos")

    def update_todo(self, todo_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        return self._put(f"/api/todos/{todo_id}", data)

    def refresh_todos(self) -> Dict[str, Any]:
        return self._post("/api/todos/refresh")

    def export_batch(self, harvest_id: str) -> str:
        return self._get_text(f"/api/export/batch/{harvest_id}")

    def export_all(self) -> str:
        return self._get_text("/api/export/all")
