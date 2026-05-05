from typing import Any, TypeVar, cast

import httpx
from pydantic import TypeAdapter

from shared.protocols.common import ApiResponse

DEFAULT_BASE_URL = "http://localhost:8000/api/v1"
T = TypeVar("T")


class HttpClient:
    def __init__(self, base_url: str = DEFAULT_BASE_URL, timeout: float = 10.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    def _get_client(self) -> httpx.Client:
        return httpx.Client(base_url=self.base_url, timeout=self.timeout)

    def get(self, path: str, params: dict[str, Any] | None = None) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.get(path, params=params or {})
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def post(self, path: str, data: dict[str, Any] | None = None, params: dict[str, Any] | None = None) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.post(path, json=data or {}, params=params or {})
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def put(self, path: str, data: dict[str, Any] | None = None, params: dict[str, Any] | None = None) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.put(path, json=data or {}, params=params or {})
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def parse_response(self, response_data: dict[str, Any], data_type: type[T]) -> T:
        ta = TypeAdapter(ApiResponse[T])
        api_response = ta.validate_python(response_data)
        if not api_response.is_success:
            raise ValueError(f"API Error [{api_response.code}]: {api_response.message}")
        if api_response.data is None:
            raise ValueError("API response data is None")
        return api_response.data


_default_client: HttpClient | None = None


def get_default_client() -> HttpClient:
    global _default_client
    if _default_client is None:
        _default_client = HttpClient()
    return _default_client
