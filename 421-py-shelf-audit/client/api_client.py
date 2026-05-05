import httpx
from typing import Any, Optional
from decimal import Decimal

from shared import (
    Zone, Shelf, Product, Location, LocationStatus, ErrorCode,
    InventorySession, Alert,
    ZoneCreateRequest, ShelfCreateRequest, ProductCreateRequest,
    StockInRequest, StockInResult, StockOutRequest,
    InventoryStartRequest, InventoryCompleteRequest, InventoryItem,
    LocationUpdateStatusRequest, LocationSetCapacityRequest,
    ApiResponse, LocationQueryResult, ProductLocationsResult
)


class ApiClient:
    def __init__(self, base_url: str = "http://localhost:8000") -> None:
        self.base_url = base_url
        self.client = httpx.Client(base_url=base_url, timeout=10.0)

    def _handle_response(self, response: httpx.Response) -> dict[str, Any]:
        try:
            data: dict[str, Any] = response.json()
        except ValueError:
            data = {"success": False, "error_message": f"Invalid JSON response: {response.text}"}

        if not response.is_success:
            if "success" in data:
                return data
            return {
                "success": False,
                "error_code": ErrorCode.INTERNAL_ERROR.value,
                "error_message": f"HTTP {response.status_code}: {response.text}"
            }

        return data

    def create_zone(self, code: str, name: str, description: Optional[str] = None) -> dict[str, Any]:
        request = ZoneCreateRequest(code=code, name=name, description=description)
        response = self.client.post("/api/v1/zones", json=request.model_dump())
        return self._handle_response(response)

    def list_zones(self) -> dict[str, Any]:
        response = self.client.get("/api/v1/zones")
        return self._handle_response(response)

    def get_zone_detail(self, code: str) -> dict[str, Any]:
        response = self.client.get(f"/api/v1/zones/{code}")
        return self._handle_response(response)

    def create_shelf(
        self,
        zone_code: str,
        shelf_number: int,
        name: Optional[str],
        layers: int,
        columns: int,
        default_capacity: int
    ) -> dict[str, Any]:
        request = ShelfCreateRequest(
            zone_code=zone_code,
            shelf_number=shelf_number,
            name=name,
            layers=layers,
            columns=columns,
            default_capacity_per_location=default_capacity
        )
        response = self.client.post("/api/v1/shelves", json=request.model_dump())
        return self._handle_response(response)

    def list_shelves(self, zone_code: Optional[str] = None) -> dict[str, Any]:
        params: dict[str, Any] = {}
        if zone_code:
            params["zone_code"] = zone_code
        response = self.client.get("/api/v1/shelves", params=params)
        return self._handle_response(response)

    def get_shelf_detail(self, zone_code: str, shelf_number: int) -> dict[str, Any]:
        response = self.client.get(f"/api/v1/shelves/{zone_code}/{shelf_number}")
        return self._handle_response(response)

    def create_product(
        self,
        sku: str,
        name: str,
        description: Optional[str],
        unit_price: Decimal,
        volume: Decimal
    ) -> dict[str, Any]:
        request = ProductCreateRequest(
            sku=sku,
            name=name,
            description=description,
            unit_price=unit_price,
            volume=volume
        )
        response = self.client.post("/api/v1/products", json=request.model_dump())
        return self._handle_response(response)

    def list_products(self) -> dict[str, Any]:
        response = self.client.get("/api/v1/products")
        return self._handle_response(response)

    def get_product(self, sku: str) -> dict[str, Any]:
        response = self.client.get(f"/api/v1/products/{sku}")
        return self._handle_response(response)

    def list_locations(
        self,
        zone_code: Optional[str] = None,
        status: Optional[LocationStatus] = None,
        product_sku: Optional[str] = None
    ) -> dict[str, Any]:
        params: dict[str, Any] = {}
        if zone_code:
            params["zone_code"] = zone_code
        if status:
            params["status"] = status.value
        if product_sku:
            params["product_sku"] = product_sku
        response = self.client.get("/api/v1/locations", params=params)
        return self._handle_response(response)

    def get_location(self, code: str) -> dict[str, Any]:
        response = self.client.get(f"/api/v1/locations/{code}")
        return self._handle_response(response)

    def update_location_status(self, code: str, status: LocationStatus) -> dict[str, Any]:
        request = LocationUpdateStatusRequest(status=status)
        response = self.client.put(f"/api/v1/locations/{code}/status", json=request.model_dump())
        return self._handle_response(response)

    def set_location_capacity(self, code: str, max_capacity: int) -> dict[str, Any]:
        request = LocationSetCapacityRequest(max_capacity=max_capacity)
        response = self.client.put(f"/api/v1/locations/{code}/capacity", json=request.model_dump())
        return self._handle_response(response)

    def stock_in(
        self,
        product_sku: str,
        quantity: int,
        preferred_zone: Optional[str] = None,
        location_code: Optional[str] = None
    ) -> dict[str, Any]:
        request = StockInRequest(
            product_sku=product_sku,
            quantity=quantity,
            preferred_zone=preferred_zone,
            location_code=location_code
        )
        response = self.client.post("/api/v1/stock/in", json=request.model_dump())
        return self._handle_response(response)

    def stock_out(self, product_sku: str, quantity: int) -> dict[str, Any]:
        request = StockOutRequest(product_sku=product_sku, quantity=quantity)
        response = self.client.post("/api/v1/stock/out", json=request.model_dump())
        return self._handle_response(response)

    def query_location(self, code: str) -> dict[str, Any]:
        response = self.client.get(f"/api/v1/query/location/{code}")
        return self._handle_response(response)

    def query_product_locations(self, sku: str) -> dict[str, Any]:
        response = self.client.get(f"/api/v1/query/product/{sku}/locations")
        return self._handle_response(response)

    def start_inventory(
        self,
        zone_code: Optional[str] = None,
        location_codes: Optional[list[str]] = None,
        include_all: bool = False
    ) -> dict[str, Any]:
        request = InventoryStartRequest(
            zone_code=zone_code,
            location_codes=location_codes,
            include_all=include_all
        )
        response = self.client.post("/api/v1/inventory/start", json=request.model_dump())
        return self._handle_response(response)

    def complete_inventory(
        self,
        session_id: str,
        items: list[tuple[str, int]]
    ) -> dict[str, Any]:
        inventory_items = [
            InventoryItem(location_code=code, actual_quantity=qty)
            for code, qty in items
        ]
        request = InventoryCompleteRequest(session_id=session_id, items=inventory_items)
        response = self.client.post("/api/v1/inventory/complete", json=request.model_dump())
        return self._handle_response(response)

    def get_inventory_session(self, session_id: str) -> dict[str, Any]:
        response = self.client.get(f"/api/v1/inventory/{session_id}")
        return self._handle_response(response)

    def list_alerts(self, resolved: Optional[bool] = None) -> dict[str, Any]:
        params: dict[str, Any] = {}
        if resolved is not None:
            params["resolved"] = str(resolved).lower()
        response = self.client.get("/api/v1/alerts", params=params)
        return self._handle_response(response)

    def resolve_alert(self, alert_id: str) -> dict[str, Any]:
        response = self.client.post(f"/api/v1/alerts/{alert_id}/resolve")
        return self._handle_response(response)

    def close(self) -> None:
        self.client.close()

    def __enter__(self) -> "ApiClient":
        return self

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        self.close()
