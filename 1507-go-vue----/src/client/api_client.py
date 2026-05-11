import os
from typing import List, Optional, Dict, Any

import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get(
            "SERVER_URL", "http://localhost:8000"
        )

    def _request(self, method: str, path: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{path}"
        try:
            response = requests.request(method, url, timeout=10, **kwargs)
            response.raise_for_status()
            if response.content:
                return response.json()
            return {}
        except requests.exceptions.RequestException as e:
            raise RuntimeError(f"请求失败: {e}")

    def create_rider(self, rider_id: str) -> Dict[str, Any]:
        return self._request("POST", "/riders", json={"rider_id": rider_id})

    def list_riders(self, status: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"status": status} if status else {}
        return self._request("GET", "/riders", params=params)

    def get_rider(self, rider_id: str) -> Dict[str, Any]:
        return self._request("GET", f"/riders/{rider_id}")

    def heartbeat(self, rider_id: str) -> Dict[str, Any]:
        return self._request("POST", f"/riders/{rider_id}/heartbeat")

    def set_rider_status(self, rider_id: str, status: str) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/riders/{rider_id}/status",
            json={"status": status}
        )

    def accept_order(self, rider_id: str, order_id: str) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/riders/{rider_id}/accept-order",
            json={"order_id": order_id}
        )

    def create_order(self) -> Dict[str, Any]:
        return self._request("POST", "/orders")

    def list_orders(self, status: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"status": status} if status else {}
        return self._request("GET", "/orders", params=params)

    def get_order(self, order_id: str) -> Dict[str, Any]:
        return self._request("GET", f"/orders/{order_id}")

    def complete_order(self, order_id: str) -> Dict[str, Any]:
        return self._request("POST", f"/orders/{order_id}/complete")

    def reassign_order(self, order_id: str, rider_id: str) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/orders/{order_id}/reassign",
            json={"rider_id": rider_id}
        )

    def get_metrics(self) -> Dict[str, Any]:
        return self._request("GET", "/metrics")
