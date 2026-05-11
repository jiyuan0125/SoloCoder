import os
import json
from typing import Optional, Dict, Any, List
import requests

class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_URL", "http://localhost:8000")
    
    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        try:
            response = requests.request(method, url, **kwargs)
            if response.status_code >= 400:
                try:
                    error_data = response.json()
                    error_msg = error_data.get("error") or error_data.get("detail") or str(response.status_code)
                    raise RuntimeError(f"API 错误: {error_msg}")
                except ValueError:
                    raise RuntimeError(f"API 错误: {response.status_code} - {response.text}")
            if response.status_code == 204:
                return {}
            return response.json()
        except requests.ConnectionError:
            raise RuntimeError(f"无法连接到服务端: {self.base_url}")
    
    def get(self, endpoint: str, params: Optional[Dict] = None) -> Any:
        return self._request("GET", endpoint, params=params)
    
    def post(self, endpoint: str, data: Optional[Dict] = None) -> Any:
        return self._request("POST", endpoint, json=data)
    
    def put(self, endpoint: str, data: Optional[Dict] = None) -> Any:
        return self._request("PUT", endpoint, json=data)
    
    def delete(self, endpoint: str) -> Any:
        return self._request("DELETE", endpoint)

client = APIClient()
