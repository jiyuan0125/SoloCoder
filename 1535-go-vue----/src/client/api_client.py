import os
from typing import Optional, List, Any, Dict
import requests


class ApiClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("APP_SERVER_URL", "http://localhost:8000")
        self._session = requests.Session()

    def _request(self, method: str, endpoint: str, **kwargs) -> Any:
        url = f"{self.base_url}{endpoint}"
        try:
            response = self._session.request(method, url, **kwargs)
            response.raise_for_status()
            if response.text:
                return response.json()
            return None
        except requests.exceptions.RequestException as e:
            raise RuntimeError(f"API 请求失败: {e}")

    def get_root(self) -> Dict:
        return self._request("GET", "/")

    def create_project(self, name: str, company: str, description: str) -> Dict:
        return self._request(
            "POST",
            "/api/projects",
            json={
                "name": name,
                "company": company,
                "description": description,
            },
        )

    def list_projects(self, status: Optional[str] = None) -> List[Dict]:
        params = {"status": status} if status else {}
        return self._request("GET", "/api/projects", params=params)

    def get_project(self, project_id: str) -> Dict:
        return self._request("GET", f"/api/projects/{project_id}")

    def update_project(self, project_id: str, name: Optional[str] = None, company: Optional[str] = None, description: Optional[str] = None) -> Dict:
        data = {}
        if name:
            data["name"] = name
        if company:
            data["company"] = company
        if description:
            data["description"] = description
        return self._request("PATCH", f"/api/projects/{project_id}", json=data)

    def start_preparation(self, project_id: str) -> Dict:
        return self._request("POST", f"/api/projects/{project_id}/start-preparation")

    def complete_preparation(self, project_id: str) -> Dict:
        return self._request("POST", f"/api/projects/{project_id}/complete-preparation")

    def submit_evaluation(self, project_id: str, compliance: float, technology: float, environmental: float, feasibility: float) -> Dict:
        return self._request(
            "POST",
            f"/api/projects/{project_id}/submit-evaluation",
            json={
                "compliance_score": compliance,
                "technology_score": technology,
                "environmental_score": environmental,
                "feasibility_score": feasibility,
            },
        )

    def start_publicity(self, project_id: str) -> Dict:
        return self._request("POST", f"/api/projects/{project_id}/start-publicity")

    def add_public_opinion(self, project_id: str, content: str) -> Dict:
        return self._request(
            "POST",
            f"/api/projects/{project_id}/public-opinions",
            json={"content": content},
        )

    def respond_to_opinion(self, project_id: str, opinion_id: str, response: str) -> Dict:
        return self._request(
            "POST",
            f"/api/projects/{project_id}/public-opinions/{opinion_id}/respond",
            params={"response": response},
        )

    def complete_publicity(self, project_id: str) -> Dict:
        return self._request("POST", f"/api/projects/{project_id}/complete-publicity")

    def approve_project(self, project_id: str, decision: str) -> Dict:
        return self._request(
            "POST",
            f"/api/projects/{project_id}/approve",
            params={"decision": decision},
        )

    def resubmit_project(self, project_id: str) -> Dict:
        return self._request("POST", f"/api/projects/{project_id}/resubmit")

    def list_todos(self, completed: Optional[bool] = None) -> List[Dict]:
        params = {}
        if completed is not None:
            params["completed"] = completed
        return self._request("GET", "/api/todos", params=params)

    def get_todo(self, todo_id: str) -> Dict:
        return self._request("GET", f"/api/todos/{todo_id}")

    def complete_todo(self, todo_id: str) -> Dict:
        return self._request(
            "PATCH",
            f"/api/todos/{todo_id}",
            json={"is_completed": True},
        )

    def list_project_todos(self, project_id: str) -> List[Dict]:
        return self._request("GET", f"/api/projects/{project_id}/todos")
