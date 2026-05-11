import os
from typing import Any, Dict, Optional
from urllib.parse import urljoin

import urllib.request
import json


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        if base_url is None:
            host = os.environ.get("SERVER_HOST", "127.0.0.1")
            port = os.environ.get("SERVER_PORT", "8000")
            base_url = f"http://{host}:{port}"
        self.base_url = base_url.rstrip("/") + "/"

    def _request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict[str, Any]] = None,
    ) -> Any:
        url = urljoin(self.base_url, endpoint.lstrip("/"))
        
        headers = {"Content-Type": "application/json"}
        body = json.dumps(data).encode("utf-8") if data else None
        
        req = urllib.request.Request(url, data=body, headers=headers, method=method)
        
        try:
            with urllib.request.urlopen(req) as response:
                content_type = response.headers.get("Content-Type", "")
                raw = response.read().decode("utf-8")
                if "application/json" in content_type:
                    return json.loads(raw)
                return raw
        except urllib.error.HTTPError as e:
            error_body = e.read().decode("utf-8")
            try:
                error_data = json.loads(error_body)
                detail = error_data.get("detail", error_body)
            except (ValueError, KeyError):
                detail = error_body
            raise Exception(f"请求失败 [{e.code}]: {detail}")
        except urllib.error.URLError as e:
            raise Exception(f"无法连接到服务器: {e.reason}")

    def get(self, endpoint: str) -> Any:
        return self._request("GET", endpoint)

    def post(self, endpoint: str, data: Dict[str, Any]) -> Any:
        return self._request("POST", endpoint, data)

    def patch(self, endpoint: str, data: Dict[str, Any]) -> Any:
        return self._request("PATCH", endpoint, data)
