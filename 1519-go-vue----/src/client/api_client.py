import os
import httpx
from typing import Optional, Dict, Any, List


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get(
            "API_BASE_URL",
            "http://localhost:8000"
        )

    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        try:
            with httpx.Client(timeout=30.0) as client:
                response = client.request(method, url, **kwargs)
                if response.status_code >= 400:
                    try:
                        error_detail = response.json().get("detail", response.text)
                    except Exception:
                        error_detail = response.text
                    raise RuntimeError(f"API 请求失败 [{response.status_code}]: {error_detail}")
                if response.status_code == 204:
                    return {}
                if "text/plain" in response.headers.get("content-type", ""):
                    return {"text": response.text}
                return response.json() if response.content else {}
        except httpx.RequestError as e:
            raise RuntimeError(f"连接服务端失败: {e}")

    def get(self, endpoint: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("GET", endpoint, params=params)

    def post(self, endpoint: str, json: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("POST", endpoint, json=json)

    def put(self, endpoint: str, json: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("PUT", endpoint, json=json)

    def delete(self, endpoint: str) -> Dict[str, Any]:
        return self._request("DELETE", endpoint)


client = APIClient()
