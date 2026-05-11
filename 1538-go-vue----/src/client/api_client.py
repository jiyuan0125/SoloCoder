import os
import json
import httpx
from typing import Optional, Dict, Any


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get(
            "API_BASE_URL",
            "http://localhost:8000/api"
        )

    def _request(self, method: str, path: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{path}"
        try:
            with httpx.Client(timeout=30.0) as client:
                response = client.request(method, url, **kwargs)
                response.raise_for_status()
                if response.status_code == 204:
                    return {}
                return response.json()
        except httpx.HTTPStatusError as e:
            raise RuntimeError(f"API Error {e.response.status_code}: {e.response.text}")
        except httpx.RequestError as e:
            raise RuntimeError(f"Connection Error: {e}")

    def get(self, path: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("GET", path, params=params)

    def post(self, path: str, json_data: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("POST", path, json=json_data)

    def put(self, path: str, json_data: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("PUT", path, json=json_data)

    def delete(self, path: str) -> Dict[str, Any]:
        return self._request("DELETE", path)

    def download(self, path: str) -> str:
        url = f"{self.base_url}{path}"
        try:
            with httpx.Client(timeout=60.0) as client:
                response = client.get(url)
                response.raise_for_status()
                return response.text
        except httpx.HTTPStatusError as e:
            raise RuntimeError(f"API Error {e.response.status_code}: {e.response.text}")
        except httpx.RequestError as e:
            raise RuntimeError(f"Connection Error: {e}")
