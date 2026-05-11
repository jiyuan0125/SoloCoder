import os
from typing import Optional, List, Dict, Any
from datetime import date
import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_URL", "http://localhost:8000/api")

    def _request(
        self,
        method: str,
        endpoint: str,
        params: Optional[Dict[str, Any]] = None,
        json: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        response = requests.request(method, url, params=params, json=json)
        
        if response.status_code >= 400:
            try:
                error_detail = response.json().get("detail", response.text)
            except ValueError:
                error_detail = response.text
            raise Exception(f"API 错误 ({response.status_code}): {error_detail}")
        
        return response.json()

    def create_store(self, name: str, address: Optional[str] = None) -> Dict[str, Any]:
        data = {"name": name}
        if address:
            data["address"] = address
        return self._request("POST", "/stores", json=data)

    def list_stores(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/stores")

    def get_store(self, store_id: int) -> Dict[str, Any]:
        return self._request("GET", f"/stores/{store_id}")

    def create_ingredient(
        self,
        store_id: int,
        name: str,
        stock_quantity: float,
        unit_price: float,
        safety_stock: float,
        expiry_date: Optional[str] = None,
    ) -> Dict[str, Any]:
        data = {
            "store_id": store_id,
            "name": name,
            "stock_quantity": stock_quantity,
            "unit_price": unit_price,
            "safety_stock": safety_stock,
        }
        if expiry_date:
            data["expiry_date"] = expiry_date
        return self._request("POST", "/ingredients", json=data)

    def list_ingredients(self, store_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {}
        if store_id:
            params["store_id"] = store_id
        return self._request("GET", "/ingredients", params=params)

    def get_ingredient(self, ingredient_id: int) -> Dict[str, Any]:
        return self._request("GET", f"/ingredients/{ingredient_id}")

    def update_ingredient(
        self,
        ingredient_id: int,
        **kwargs: Any,
    ) -> Dict[str, Any]:
        return self._request("PATCH", f"/ingredients/{ingredient_id}", json=kwargs)

    def create_purchase_order(
        self,
        store_id: int,
        ingredient_id: int,
        requested_quantity: float,
        expected_arrival_date: str,
        remarks: Optional[str] = None,
    ) -> Dict[str, Any]:
        data = {
            "store_id": store_id,
            "ingredient_id": ingredient_id,
            "requested_quantity": requested_quantity,
            "expected_arrival_date": expected_arrival_date,
        }
        if remarks:
            data["remarks"] = remarks
        return self._request("POST", "/purchase-orders", json=data)

    def list_purchase_orders(
        self,
        store_id: Optional[int] = None,
        status: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        params = {}
        if store_id:
            params["store_id"] = store_id
        if status:
            params["status"] = status
        return self._request("GET", "/purchase-orders", params=params)

    def get_purchase_order(self, po_id: int) -> Dict[str, Any]:
        return self._request("GET", f"/purchase-orders/{po_id}")

    def approve_purchase_order(self, po_id: int, approved: bool) -> Dict[str, Any]:
        return self._request(
            "POST",
            f"/purchase-orders/{po_id}/approve",
            json={"approved": approved},
        )

    def receive_purchase_order(
        self,
        po_id: int,
        received_quantity: float,
        actual_arrival_date: Optional[str] = None,
    ) -> Dict[str, Any]:
        data = {"received_quantity": received_quantity}
        if actual_arrival_date:
            data["actual_arrival_date"] = actual_arrival_date
        return self._request(
            "POST",
            f"/purchase-orders/{po_id}/receive",
            json=data,
        )

    def create_wastage(
        self,
        store_id: int,
        ingredient_id: int,
        quantity: float,
        wastage_date: Optional[str] = None,
        reason: Optional[str] = None,
    ) -> Dict[str, Any]:
        data = {
            "store_id": store_id,
            "ingredient_id": ingredient_id,
            "quantity": quantity,
        }
        if wastage_date:
            data["wastage_date"] = wastage_date
        if reason:
            data["reason"] = reason
        return self._request("POST", "/wastages", json=data)

    def list_wastages(
        self,
        store_id: Optional[int] = None,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        params = {}
        if store_id:
            params["store_id"] = store_id
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        return self._request("GET", "/wastages", params=params)

    def list_alerts(
        self,
        store_id: Optional[int] = None,
        active_only: bool = True,
    ) -> List[Dict[str, Any]]:
        params = {"active_only": active_only}
        if store_id:
            params["store_id"] = store_id
        return self._request("GET", "/alerts", params=params)

    def get_metrics(
        self,
        start_date: str,
        end_date: str,
        store_ids: Optional[List[int]] = None,
    ) -> Dict[str, Any]:
        params = {"start_date": start_date, "end_date": end_date}
        if store_ids:
            params["store_ids"] = store_ids
        return self._request("GET", "/metrics/summary", params=params)

    def get_wastage_ranking(
        self,
        start_date: str,
        end_date: str,
        store_ids: Optional[List[int]] = None,
    ) -> Dict[str, Any]:
        params = {"start_date": start_date, "end_date": end_date}
        if store_ids:
            params["store_ids"] = store_ids
        return self._request("GET", "/metrics/wastage-ranking", params=params)
