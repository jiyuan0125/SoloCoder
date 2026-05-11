import os
from typing import Optional, Dict, Any
import requests


class ApiClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get("API_BASE_URL", "http://localhost:8000")

    def _request(self, method: str, path: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{path}"
        response = requests.request(method, url, **kwargs)
        if response.status_code >= 400:
            try:
                detail = response.json().get("detail", response.text)
            except Exception:
                detail = response.text
            raise RuntimeError(f"API 错误 ({response.status_code}): {detail}")
        if response.status_code == 204:
            return {}
        return response.json()

    def get(self, path: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("GET", path, params=params)

    def post(self, path: str, json: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("POST", path, json=json)


def get_client() -> ApiClient:
    return ApiClient()
