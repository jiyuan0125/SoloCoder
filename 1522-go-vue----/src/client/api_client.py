import os
import json
from typing import Dict, List, Optional, Any
from datetime import date

import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_BASE_URL", "http://localhost:8000")
        self.api_prefix = "/api"

    def _get_url(self, endpoint: str) -> str:
        if endpoint.startswith("/api"):
            return f"{self.base_url}{endpoint}"
        return f"{self.base_url}{self.api_prefix}{endpoint}"

    def _handle_response(self, response: requests.Response) -> Any:
        try:
            response.raise_for_status()
            if response.headers.get("content-type", "").startswith("text/"):
                return response.text
            if response.content:
                return response.json()
            return None
        except requests.exceptions.HTTPError as e:
            if response.content:
                try:
                    error_data = response.json()
                    raise Exception(f"API错误: {error_data.get('detail', str(e))}")
                except json.JSONDecodeError:
                    raise Exception(f"API错误: {response.text}")
            raise Exception(f"HTTP错误: {e}")

    def health_check(self) -> Dict:
        response = requests.get(f"{self.base_url}/health")
        return self._handle_response(response)

    def list_projects(self) -> List[Dict]:
        response = requests.get(self._get_url("/projects/"))
        return self._handle_response(response)

    def get_project(self, project_id: str) -> Dict:
        response = requests.get(self._get_url(f"/projects/{project_id}"))
        return self._handle_response(response)

    def create_project(self, project_data: Dict) -> Dict:
        response = requests.post(self._get_url("/projects/"), json=project_data)
        return self._handle_response(response)

    def update_project(self, project_id: str, project_data: Dict) -> Dict:
        response = requests.put(self._get_url(f"/projects/{project_id}"), json=project_data)
        return self._handle_response(response)

    def delete_project(self, project_id: str) -> Dict:
        response = requests.delete(self._get_url(f"/projects/{project_id}"))
        return self._handle_response(response)

    def list_boreholes(self, project_id: Optional[str] = None) -> List[Dict]:
        params = {"project_id": project_id} if project_id else None
        response = requests.get(self._get_url("/boreholes/"), params=params)
        return self._handle_response(response)

    def get_borehole(self, borehole_id: str) -> Dict:
        response = requests.get(self._get_url(f"/boreholes/{borehole_id}"))
        return self._handle_response(response)

    def create_borehole(self, borehole_data: Dict) -> Dict:
        response = requests.post(self._get_url("/boreholes/"), json=borehole_data)
        return self._handle_response(response)

    def update_borehole(self, borehole_id: str, borehole_data: Dict) -> Dict:
        response = requests.put(self._get_url(f"/boreholes/{borehole_id}"), json=borehole_data)
        return self._handle_response(response)

    def delete_borehole(self, borehole_id: str) -> Dict:
        response = requests.delete(self._get_url(f"/boreholes/{borehole_id}"))
        return self._handle_response(response)

    def list_samples(self, borehole_id: Optional[str] = None) -> List[Dict]:
        params = {"borehole_id": borehole_id} if borehole_id else None
        response = requests.get(self._get_url("/samples/"), params=params)
        return self._handle_response(response)

    def get_sample(self, sample_id: str) -> Dict:
        response = requests.get(self._get_url(f"/samples/{sample_id}"))
        return self._handle_response(response)

    def create_sample(self, sample_data: Dict) -> Dict:
        response = requests.post(self._get_url("/samples/"), json=sample_data)
        return self._handle_response(response)

    def update_sample(self, sample_id: str, sample_data: Dict) -> Dict:
        response = requests.put(self._get_url(f"/samples/{sample_id}"), json=sample_data)
        return self._handle_response(response)

    def delete_sample(self, sample_id: str) -> Dict:
        response = requests.delete(self._get_url(f"/samples/{sample_id}"))
        return self._handle_response(response)

    def update_analysis_results(self, sample_id: str, results: Dict[str, float]) -> Dict:
        response = requests.patch(
            self._get_url(f"/samples/{sample_id}/analysis"),
            json=results
        )
        return self._handle_response(response)

    def list_todos(self) -> List[Dict]:
        response = requests.get(self._get_url("/todos/"))
        return self._handle_response(response)

    def list_incomplete_todos(self) -> List[Dict]:
        response = requests.get(self._get_url("/todos/incomplete"))
        return self._handle_response(response)

    def complete_todo(self, todo_id: str) -> Dict:
        response = requests.patch(self._get_url(f"/todos/{todo_id}/complete"))
        return self._handle_response(response)

    def check_and_create_todos(self) -> List[Dict]:
        response = requests.post(self._get_url("/projects/check-todos"))
        return self._handle_response(response)

    def export_borehole(self, borehole_id: str) -> str:
        response = requests.get(self._get_url(f"/exports/borehole/{borehole_id}"))
        return self._handle_response(response)

    def export_project(self, project_id: str) -> str:
        response = requests.get(self._get_url(f"/exports/project/{project_id}"))
        return self._handle_response(response)
