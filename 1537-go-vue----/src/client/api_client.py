import json
from typing import Optional, Dict, Any, List
import requests

from .config import DEFAULT_BASE_URL

class APIClient:
    def __init__(self, base_url: str = DEFAULT_BASE_URL):
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()
        self.session.headers.update({
            "Content-Type": "application/json",
            "Accept": "application/json"
        })
    
    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        try:
            response = self.session.request(method, url, **kwargs)
            response.raise_for_status()
            if response.status_code == 204:
                return {"status": "success"}
            return response.json()
        except requests.exceptions.ConnectionError:
            raise Exception(f"无法连接到服务器: {self.base_url}")
        except requests.exceptions.HTTPError as e:
            try:
                error_detail = e.response.json().get("detail", str(e))
            except:
                error_detail = str(e)
            raise Exception(f"请求失败: {error_detail}")
    
    def get(self, endpoint: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("GET", endpoint, params=params)
    
    def post(self, endpoint: str, data: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("POST", endpoint, json=data)
    
    def put(self, endpoint: str, data: Optional[Dict] = None) -> Dict[str, Any]:
        return self._request("PUT", endpoint, json=data)
    
    def delete(self, endpoint: str) -> Dict[str, Any]:
        return self._request("DELETE", endpoint)
    
    def health_check(self) -> Dict[str, Any]:
        return self.get("/health")
