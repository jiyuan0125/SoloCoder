from typing import Any

import httpx

from shared.models import (
    AlertResponse,
    BatchQueryRequest,
    BatchQueryResponse,
    CustomStatusUpdateRequest,
    TrackingNode,
    TrackingNodeCreate,
    Waybill,
    WaybillCreate,
    WaybillQueryResponse,
    WaybillSplitRequest,
    WaybillSplitResponse,
)


class ShippingClient:
    def __init__(self, base_url: str = "http://localhost:8000/api/v1") -> None:
        self._base_url = base_url.rstrip("/")

    def _get_client(self) -> httpx.Client:
        return httpx.Client(base_url=self._base_url, timeout=30.0)

    def create_waybill(self, waybill_create: WaybillCreate) -> Waybill:
        with self._get_client() as client:
            response = client.post(
                "/waybills",
                json=waybill_create.model_dump(mode="json"),
            )
            response.raise_for_status()
            return Waybill.model_validate(response.json())

    def get_waybill(self, waybill_number: str) -> WaybillQueryResponse:
        with self._get_client() as client:
            response = client.get(f"/waybills/{waybill_number}")
            response.raise_for_status()
            return WaybillQueryResponse.model_validate(response.json())

    def batch_query_waybills(self, waybill_numbers: list[str]) -> BatchQueryResponse:
        request = BatchQueryRequest(waybill_numbers=waybill_numbers)
        with self._get_client() as client:
            response = client.post(
                "/waybills/batch",
                json=request.model_dump(mode="json"),
            )
            response.raise_for_status()
            return BatchQueryResponse.model_validate(response.json())

    def add_tracking_node(
        self, waybill_number: str, node_create: TrackingNodeCreate
    ) -> TrackingNode:
        with self._get_client() as client:
            response = client.post(
                f"/waybills/{waybill_number}/nodes",
                json=node_create.model_dump(mode="json"),
            )
            response.raise_for_status()
            return TrackingNode.model_validate(response.json())

    def split_waybill(self, split_request: WaybillSplitRequest) -> WaybillSplitResponse:
        with self._get_client() as client:
            response = client.post(
                "/waybills/split",
                json=split_request.model_dump(mode="json"),
            )
            response.raise_for_status()
            return WaybillSplitResponse.model_validate(response.json())

    def get_sub_waybills(self, parent_waybill_number: str) -> list[Waybill]:
        with self._get_client() as client:
            response = client.get(f"/waybills/{parent_waybill_number}/subs")
            response.raise_for_status()
            return [Waybill.model_validate(item) for item in response.json()]

    def update_custom_status(
        self, waybill_number: str, request: CustomStatusUpdateRequest
    ) -> Waybill:
        with self._get_client() as client:
            response = client.put(
                f"/waybills/{waybill_number}/custom-status",
                json=request.model_dump(mode="json"),
            )
            response.raise_for_status()
            return Waybill.model_validate(response.json())

    def check_abnormal_waybills(self) -> list[AlertResponse]:
        with self._get_client() as client:
            response = client.post("/alerts/check")
            response.raise_for_status()
            return [AlertResponse.model_validate(item) for item in response.json()]

    def generate_alerts(self) -> list[AlertResponse]:
        with self._get_client() as client:
            response = client.post("/alerts/generate")
            response.raise_for_status()
            return [AlertResponse.model_validate(item) for item in response.json()]

    def health_check(self) -> dict[str, Any]:
        base_url = self._base_url
        if base_url.endswith("/api/v1"):
            base_url = base_url[:-7]
        elif base_url.endswith("/api/v1/"):
            base_url = base_url[:-8]
        with httpx.Client(base_url=base_url, timeout=10.0) as client:
            response = client.get("/api/v1/health")
            response.raise_for_status()
            result: dict[str, Any] = response.json()
            return result
