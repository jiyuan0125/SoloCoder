from typing import Any, Optional

import httpx
from pydantic import BaseModel

from shared.models import APIResponse


DEFAULT_BASE_URL = "http://127.0.0.1:8000"


class APIClient:
    def __init__(self, base_url: str = DEFAULT_BASE_URL) -> None:
        self._base_url = base_url
        self._client: httpx.Client | None = None

    def _get_client(self) -> httpx.Client:
        if self._client is None:
            self._client = httpx.Client(base_url=self._base_url, timeout=30.0)
        return self._client

    def _request(
        self,
        method: str,
        endpoint: str,
        params: Optional[dict[str, Any]] = None,
        json: Optional[dict[str, Any]] = None,
    ) -> httpx.Response:
        client = self._get_client()
        return client.request(
            method=method,
            url=endpoint,
            params=params,
            json=json,
        )

    def get(
        self,
        endpoint: str,
        params: Optional[dict[str, Any]] = None,
    ) -> httpx.Response:
        return self._request("GET", endpoint, params=params)

    def post(
        self,
        endpoint: str,
        json: Optional[dict[str, Any]] = None,
        params: Optional[dict[str, Any]] = None,
    ) -> httpx.Response:
        return self._request("POST", endpoint, params=params, json=json)

    def put(
        self,
        endpoint: str,
        json: Optional[dict[str, Any]] = None,
    ) -> httpx.Response:
        return self._request("PUT", endpoint, json=json)

    def delete(self, endpoint: str) -> httpx.Response:
        return self._request("DELETE", endpoint)

    def close(self) -> None:
        if self._client is not None:
            self._client.close()
            self._client = None

    def __enter__(self) -> "APIClient":
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()
