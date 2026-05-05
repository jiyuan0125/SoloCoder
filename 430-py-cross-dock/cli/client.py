from typing import Optional, List, Dict, Any
from datetime import datetime

import httpx
from pydantic import TypeAdapter

from shared.models import (
    CrossDockOrderCreate,
    CrossDockOrderResponse,
    InboundScanRequest,
    OutboundScanRequest,
    BatchCrossDockCreate,
    BatchCrossDockResponse,
    EfficiencyStatistics,
    DailyReport,
    ZoneMonitorResponse,
    ExceptionPlan,
)


class CrossDockClient:
    def __init__(self, server_url: str, timeout: float = 30.0) -> None:
        self._server_url = server_url.rstrip("/")
        self._timeout = timeout
        self._client: Optional[httpx.Client] = None

    def _get_client(self) -> httpx.Client:
        if self._client is None:
            self._client = httpx.Client(timeout=self._timeout)
        return self._client

    def _make_request(
        self,
        method: str,
        endpoint: str,
        json_data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
    ) -> httpx.Response:
        url = f"{self._server_url}{endpoint}"
        client = self._get_client()

        try:
            if method.upper() == "GET":
                response = client.get(url, params=params)
            elif method.upper() == "POST":
                response = client.post(url, json=json_data, params=params)
            else:
                raise ValueError(f"不支持的HTTP方法: {method}")

            if response.status_code >= 400:
                try:
                    error_data = response.json()
                    error_message = error_data.get("message", str(response.status_code))
                    raise httpx.HTTPStatusError(
                        f"请求失败: {error_message}",
                        request=response.request,
                        response=response,
                    )
                except (ValueError, KeyError):
                    raise httpx.HTTPStatusError(
                        f"请求失败: {response.status_code}",
                        request=response.request,
                        response=response,
                    )

            return response
        except httpx.HTTPStatusError:
            raise
        except httpx.RequestError as e:
            raise RuntimeError(f"连接服务端失败: {e}")

    def create_order(self, order_create: CrossDockOrderCreate) -> CrossDockOrderResponse:
        response = self._make_request(
            "POST",
            "/api/cross-dock/orders",
            json_data=order_create.model_dump(mode="json"),
        )
        return CrossDockOrderResponse(**response.json())

    def get_order(self, order_id: str) -> CrossDockOrderResponse:
        response = self._make_request("GET", f"/api/cross-dock/orders/{order_id}")
        return CrossDockOrderResponse(**response.json())

    def list_orders(
        self,
        status: Optional[str] = None,
        warehouse_id: Optional[str] = None,
    ) -> List[CrossDockOrderResponse]:
        params: Dict[str, Any] = {}
        if status:
            params["status"] = status
        if warehouse_id:
            params["warehouse_id"] = warehouse_id

        response = self._make_request("GET", "/api/cross-dock/orders", params=params)
        adapter = TypeAdapter(List[CrossDockOrderResponse])
        return adapter.validate_python(response.json())

    def inbound_scan(self, request: InboundScanRequest) -> CrossDockOrderResponse:
        response = self._make_request(
            "POST",
            "/api/cross-dock/inbound-scan",
            json_data=request.model_dump(mode="json"),
        )
        return CrossDockOrderResponse(**response.json())

    def outbound_scan(self, request: OutboundScanRequest) -> CrossDockOrderResponse:
        response = self._make_request(
            "POST",
            "/api/cross-dock/outbound-scan",
            json_data=request.model_dump(mode="json"),
        )
        return CrossDockOrderResponse(**response.json())

    def create_batch_cross_dock(self, batch_create: BatchCrossDockCreate) -> BatchCrossDockResponse:
        response = self._make_request(
            "POST",
            "/api/cross-dock/batch",
            json_data=batch_create.model_dump(mode="json"),
        )
        return BatchCrossDockResponse(**response.json())

    def get_batch_status(self, outbound_order_number: str) -> BatchCrossDockResponse:
        response = self._make_request(
            "GET", f"/api/cross-dock/batch/{outbound_order_number}"
        )
        return BatchCrossDockResponse(**response.json())

    def get_efficiency_statistics(
        self, start_time: datetime, end_time: datetime
    ) -> EfficiencyStatistics:
        params = {
            "start_time": start_time.isoformat(),
            "end_time": end_time.isoformat(),
        }
        response = self._make_request(
            "GET", "/api/cross-dock/statistics/efficiency", params=params
        )
        return EfficiencyStatistics(**response.json())

    def get_daily_report(self, report_date: datetime) -> DailyReport:
        params = {"report_date": report_date.isoformat()}
        response = self._make_request(
            "GET", "/api/cross-dock/reports/daily", params=params
        )
        return DailyReport(**response.json())

    def get_zone_monitor(self, warehouse_id: str) -> ZoneMonitorResponse:
        response = self._make_request(
            "GET", f"/api/cross-dock/zone-monitor/{warehouse_id}"
        )
        return ZoneMonitorResponse(**response.json())

    def get_exception_plan(self, exception_type: str) -> ExceptionPlan:
        response = self._make_request(
            "GET", f"/api/cross-dock/exception-plans/{exception_type}"
        )
        return ExceptionPlan(**response.json())

    def health_check(self) -> Dict[str, Any]:
        response = self._make_request("GET", "/api/health")
        result: Dict[str, Any] = response.json()
        return result

    def close(self) -> None:
        if self._client is not None:
            self._client.close()
            self._client = None

    def __enter__(self) -> "CrossDockClient":
        return self

    def __exit__(
        self,
        exc_type: Optional[type[BaseException]],
        exc_val: Optional[BaseException],
        exc_tb: Optional[object],
    ) -> None:
        self.close()
