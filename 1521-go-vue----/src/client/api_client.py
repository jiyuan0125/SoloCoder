import json
import os
from typing import List, Optional, Dict, Any

import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_BASE_URL", "http://localhost:8000")
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})

    def _request(self, method: str, path: str, **kwargs) -> Any:
        url = f"{self.base_url}{path}"
        try:
            response = self.session.request(method, url, **kwargs)
            response.raise_for_status()
            if response.content:
                return response.json()
            return None
        except requests.exceptions.ConnectionError:
            raise RuntimeError(f"无法连接到服务器: {self.base_url}")
        except requests.exceptions.HTTPError as e:
            if e.response is not None:
                try:
                    detail = e.response.json().get("detail", str(e))
                except json.JSONDecodeError:
                    detail = e.response.text
                raise RuntimeError(f"请求失败: {detail}")
            raise RuntimeError(f"请求失败: {str(e)}")

    def list_mines(self) -> List[Dict]:
        return self._request("GET", "/api/mines")

    def create_mine(self, data: Dict) -> Dict:
        return self._request("POST", "/api/mines", json=data)

    def get_mine(self, mine_id: int) -> Dict:
        return self._request("GET", f"/api/mines/{mine_id}")

    def update_mine(self, mine_id: int, data: Dict) -> Dict:
        return self._request("PUT", f"/api/mines/{mine_id}", json=data)

    def delete_mine(self, mine_id: int) -> None:
        self._request("DELETE", f"/api/mines/{mine_id}")

    def list_mining_operations(self, mine_id: Optional[int] = None) -> List[Dict]:
        params = {"mine_id": mine_id} if mine_id else {}
        return self._request("GET", "/api/mining-operations", params=params)

    def create_mining_operation(self, data: Dict) -> Dict:
        return self._request("POST", "/api/mining-operations", json=data)

    def get_mining_operation(self, op_id: int) -> Dict:
        return self._request("GET", f"/api/mining-operations/{op_id}")

    def update_mining_operation(self, op_id: int, data: Dict) -> Dict:
        return self._request("PUT", f"/api/mining-operations/{op_id}", json=data)

    def delete_mining_operation(self, op_id: int) -> None:
        self._request("DELETE", f"/api/mining-operations/{op_id}")

    def list_transports(self, mine_id: Optional[int] = None) -> List[Dict]:
        params = {"mine_id": mine_id} if mine_id else {}
        return self._request("GET", "/api/transports", params=params)

    def create_transport(self, data: Dict) -> Dict:
        return self._request("POST", "/api/transports", json=data)

    def get_transport(self, transport_id: int) -> Dict:
        return self._request("GET", f"/api/transports/{transport_id}")

    def update_transport(self, transport_id: int, data: Dict) -> Dict:
        return self._request("PUT", f"/api/transports/{transport_id}", json=data)

    def delete_transport(self, transport_id: int) -> None:
        self._request("DELETE", f"/api/transports/{transport_id}")

    def list_safety_checks(self, mine_id: Optional[int] = None) -> List[Dict]:
        params = {"mine_id": mine_id} if mine_id else {}
        return self._request("GET", "/api/safety-checks", params=params)

    def create_safety_check(self, data: Dict) -> Dict:
        return self._request("POST", "/api/safety-checks", json=data)

    def get_safety_check(self, check_id: int) -> Dict:
        return self._request("GET", f"/api/safety-checks/{check_id}")

    def update_safety_check(self, check_id: int, data: Dict) -> Dict:
        return self._request("PUT", f"/api/safety-checks/{check_id}", json=data)

    def delete_safety_check(self, check_id: int) -> None:
        self._request("DELETE", f"/api/safety-checks/{check_id}")

    def list_todos(self, mine_id: Optional[int] = None, status: Optional[str] = None) -> List[Dict]:
        params = {}
        if mine_id:
            params["mine_id"] = mine_id
        if status:
            params["status"] = status
        return self._request("GET", "/api/todos", params=params)

    def create_todo(self, data: Dict) -> Dict:
        return self._request("POST", "/api/todos", json=data)

    def get_todo(self, todo_id: int) -> Dict:
        return self._request("GET", f"/api/todos/{todo_id}")

    def update_todo(self, todo_id: int, data: Dict) -> Dict:
        return self._request("PUT", f"/api/todos/{todo_id}", json=data)

    def delete_todo(self, todo_id: int) -> None:
        self._request("DELETE", f"/api/todos/{todo_id}")

    def upgrade_overdue_todos(self) -> int:
        return self._request("POST", "/api/todos/upgrade-overdue")

    def get_monthly_statistics(self, year: Optional[int] = None, month: Optional[int] = None) -> Dict:
        params = {}
        if year:
            params["year"] = year
        if month:
            params["month"] = month
        return self._request("GET", "/api/statistics/monthly", params=params)
