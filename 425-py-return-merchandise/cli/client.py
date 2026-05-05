from datetime import datetime
from typing import Any, Dict, Optional

import httpx

from shared.models import (
    OutboundOrderCreate,
    OutboundOrderResponse,
    ReturnApplyRequest,
    ReturnApplyResponse,
    WarehouseReceiveRequest,
    InspectionRequest,
    StockInRequest,
    DefectiveToNormalRequest,
    ReturnStatisticsResponse,
    ScrapLedgerResponse,
    QualityAlertResponse,
    ReturnOrderResponse,
    DefectiveItemResponse,
)


class APIClient:
    def __init__(self, base_url: str = "http://localhost:8000") -> None:
        self._base_url = base_url.rstrip("/")
        self._client = httpx.Client(timeout=30.0)

    def _make_request(
        self,
        method: str,
        endpoint: str,
        json: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
    ) -> httpx.Response:
        url = f"{self._base_url}{endpoint}"
        try:
            response = self._client.request(
                method=method,
                url=url,
                json=json,
                params=params,
            )
            response.raise_for_status()
            return response
        except httpx.HTTPStatusError as e:
            raise ValueError(f"API请求失败: {e.response.status_code} - {e.response.text}")
        except httpx.RequestError as e:
            raise ValueError(f"连接服务器失败: {e}")

    def create_outbound_order(self, request: OutboundOrderCreate) -> OutboundOrderResponse:
        response = self._make_request(
            "POST",
            "/api/v1/outbound/",
            json=request.model_dump(mode="json"),
        )
        return OutboundOrderResponse.model_validate(response.json())

    def get_outbound_order(self, order_id: str) -> OutboundOrderResponse:
        response = self._make_request("GET", f"/api/v1/outbound/{order_id}")
        return OutboundOrderResponse.model_validate(response.json())

    def list_outbound_orders(self) -> list[OutboundOrderResponse]:
        response = self._make_request("GET", "/api/v1/outbound/")
        return [OutboundOrderResponse.model_validate(item) for item in response.json()]

    def ship_outbound_order(self, order_id: str) -> OutboundOrderResponse:
        response = self._make_request("POST", f"/api/v1/outbound/{order_id}/ship")
        return OutboundOrderResponse.model_validate(response.json())

    def deliver_outbound_order(self, order_id: str) -> OutboundOrderResponse:
        response = self._make_request("POST", f"/api/v1/outbound/{order_id}/deliver")
        return OutboundOrderResponse.model_validate(response.json())

    def complete_outbound_order(self, order_id: str) -> OutboundOrderResponse:
        response = self._make_request("POST", f"/api/v1/outbound/{order_id}/complete")
        return OutboundOrderResponse.model_validate(response.json())

    def apply_return(self, request: ReturnApplyRequest) -> ReturnApplyResponse:
        response = self._make_request(
            "POST",
            "/api/v1/returns/apply",
            json=request.model_dump(mode="json"),
        )
        return ReturnApplyResponse.model_validate(response.json())

    def warehouse_receive(self, request: WarehouseReceiveRequest) -> ReturnOrderResponse:
        response = self._make_request(
            "POST",
            "/api/v1/returns/receive",
            json=request.model_dump(mode="json"),
        )
        return ReturnOrderResponse.model_validate(response.json())

    def inspect_return(self, request: InspectionRequest) -> ReturnOrderResponse:
        response = self._make_request(
            "POST",
            "/api/v1/returns/inspect",
            json=request.model_dump(mode="json"),
        )
        return ReturnOrderResponse.model_validate(response.json())

    def stock_in_return(self, request: StockInRequest) -> ReturnOrderResponse:
        response = self._make_request(
            "POST",
            "/api/v1/returns/stock-in",
            json=request.model_dump(mode="json"),
        )
        return ReturnOrderResponse.model_validate(response.json())

    def get_return_order(self, return_id: str) -> ReturnOrderResponse:
        response = self._make_request("GET", f"/api/v1/returns/{return_id}")
        return ReturnOrderResponse.model_validate(response.json())

    def list_return_orders(self) -> list[ReturnOrderResponse]:
        response = self._make_request("GET", "/api/v1/returns/")
        return [ReturnOrderResponse.model_validate(item) for item in response.json()]

    def convert_defective_to_normal(self, request: DefectiveToNormalRequest) -> DefectiveItemResponse:
        response = self._make_request(
            "POST",
            "/api/v1/defective/convert",
            json=request.model_dump(mode="json"),
        )
        return DefectiveItemResponse.model_validate(response.json())

    def get_defective_item(self, defective_id: str) -> DefectiveItemResponse:
        response = self._make_request("GET", f"/api/v1/defective/{defective_id}")
        return DefectiveItemResponse.model_validate(response.json())

    def list_defective_items(self) -> list[DefectiveItemResponse]:
        response = self._make_request("GET", "/api/v1/defective/")
        return [DefectiveItemResponse.model_validate(item) for item in response.json()]

    def get_return_statistics(
        self,
        start_date: datetime,
        end_date: datetime,
        category: Optional[str] = None,
    ) -> ReturnStatisticsResponse:
        params: Dict[str, Any] = {
            "start_date": start_date.isoformat(),
            "end_date": end_date.isoformat(),
        }
        if category:
            params["category"] = category

        response = self._make_request(
            "GET",
            "/api/v1/statistics/returns",
            params=params,
        )
        return ReturnStatisticsResponse.model_validate(response.json())

    def get_scrap_ledger(
        self,
        year: int,
        month: int,
        reason: Optional[str] = None,
    ) -> ScrapLedgerResponse:
        params: Dict[str, Any] = {
            "year": year,
            "month": month,
        }
        if reason:
            params["reason"] = reason

        response = self._make_request(
            "GET",
            "/api/v1/statistics/scrap-ledger",
            params=params,
        )
        return ScrapLedgerResponse.model_validate(response.json())

    def check_quality_alert(
        self,
        supplier_id: str,
        period_days: int = 30,
        threshold: Optional[float] = None,
    ) -> QualityAlertResponse:
        params: Dict[str, Any] = {
            "supplier_id": supplier_id,
            "period_days": period_days,
        }
        if threshold is not None:
            params["threshold"] = threshold

        response = self._make_request(
            "GET",
            "/api/v1/statistics/quality-alert",
            params=params,
        )
        return QualityAlertResponse.model_validate(response.json())

    def close(self) -> None:
        self._client.close()

    def __enter__(self) -> "APIClient":
        return self

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        self.close()
