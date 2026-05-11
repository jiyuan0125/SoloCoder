import os
import json
from typing import Dict, Any, Optional, List
from urllib.parse import urlencode
import httpx


class APIClient:
    def __init__(self, base_url: str = None, timeout: float = 30.0):
        self.base_url = base_url or os.getenv(
            "SERVER_URL", "http://localhost:8000"
        )
        self.timeout = timeout

    def _url(self, path: str, params: Dict[str, Any] = None) -> str:
        url = f"{self.base_url}{path}"
        if params:
            query = urlencode({k: v for k, v in params.items() if v is not None})
            if query:
                url = f"{url}?{query}"
        return url

    def get(self, path: str, params: Dict[str, Any] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as client:
            resp = client.get(self._url(path, params))
            resp.raise_for_status()
            return resp.json()

    def post(self, path: str, json_data: Dict[str, Any] = None,
             params: Dict[str, Any] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as client:
            resp = client.post(self._url(path, params), json=json_data)
            resp.raise_for_status()
            return resp.json()

    def put(self, path: str, json_data: Dict[str, Any] = None,
            params: Dict[str, Any] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as client:
            resp = client.put(self._url(path, params), json=json_data)
            resp.raise_for_status()
            return resp.json()

    def delete(self, path: str, params: Dict[str, Any] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as client:
            resp = client.delete(self._url(path, params))
            resp.raise_for_status()
            return resp.json()

    def get_csv(self, path: str, params: Dict[str, Any] = None) -> str:
        with httpx.Client(timeout=self.timeout) as client:
            resp = client.get(self._url(path, params))
            resp.raise_for_status()
            return resp.text
