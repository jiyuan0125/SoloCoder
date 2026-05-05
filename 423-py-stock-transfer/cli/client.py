from typing import Any
from uuid import UUID

import httpx

from shared.enums import (
    ApprovalLevel,
    ApprovalStatus,
    TransferStatus,
    TransferType,
)


class ApiClient:
    def __init__(self, base_url: str = "http://localhost:8000") -> None:
        self.base_url = base_url.rstrip("/")
        self._client: httpx.Client | None = None

    @property
    def client(self) -> httpx.Client:
        if self._client is None:
            self._client = httpx.Client(
                base_url=self.base_url,
                timeout=30.0,
            )
        return self._client

    def close(self) -> None:
        if self._client is not None:
            self._client.close()
            self._client = None

    def __enter__(self) -> "ApiClient":
        return self

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        self.close()

    def _request(
        self,
        method: str,
        path: str,
        params: dict[str, Any] | None = None,
        json: dict[str, Any] | None = None,
    ) -> Any:
        url = f"{self.base_url}{path}"
        response = self.client.request(method, url, params=params, json=json)
        response.raise_for_status()
        return response.json()

    def health_check(self) -> dict[str, Any]:
        result = self._request("GET", "/health")
        assert isinstance(result, dict)
        return result

    def list_warehouses(self) -> list[dict[str, Any]]:
        result = self._request("GET", "/api/v1/warehouses")
        assert isinstance(result, list)
        return result

    def list_products(self) -> list[dict[str, Any]]:
        result = self._request("GET", "/api/v1/products")
        assert isinstance(result, list)
        return result

    def list_inventory(
        self,
        warehouse_id: UUID | None = None,
        product_id: UUID | None = None,
        page: int = 1,
        page_size: int = 100,
    ) -> dict[str, Any]:
        params: dict[str, Any] = {"page": page, "page_size": page_size}
        if warehouse_id:
            params["warehouse_id"] = str(warehouse_id)
        if product_id:
            params["product_id"] = str(product_id)
        result = self._request("GET", "/api/v1/inventory", params=params)
        assert isinstance(result, dict)
        return result

    def create_transfer(
        self,
        source_warehouse_id: UUID,
        target_warehouse_id: UUID,
        items: list[dict[str, Any]],
        transfer_type: TransferType = TransferType.INTRA_COMPANY,
        created_by: UUID | None = None,
        remark: str | None = None,
    ) -> dict[str, Any]:
        if created_by is None:
            created_by = UUID("00000000-0000-0000-0000-000000000001")

        json_data: dict[str, Any] = {
            "source_warehouse_id": str(source_warehouse_id),
            "target_warehouse_id": str(target_warehouse_id),
            "items": items,
            "transfer_type": transfer_type.value,
            "created_by": str(created_by),
        }
        if remark:
            json_data["remark"] = remark

        result = self._request("POST", "/api/v1/transfers", json=json_data)
        assert isinstance(result, dict)
        return result

    def get_transfer(self, transfer_id: UUID) -> dict[str, Any]:
        result = self._request("GET", f"/api/v1/transfers/{transfer_id}")
        assert isinstance(result, dict)
        return result

    def get_transfer_by_no(self, transfer_no: str) -> dict[str, Any]:
        result = self._request("GET", f"/api/v1/transfers/no/{transfer_no}")
        assert isinstance(result, dict)
        return result

    def list_transfers(
        self,
        status: TransferStatus | None = None,
        source_warehouse_id: UUID | None = None,
        target_warehouse_id: UUID | None = None,
        transfer_type: TransferType | None = None,
        page: int = 1,
        page_size: int = 20,
    ) -> dict[str, Any]:
        params: dict[str, Any] = {"page": page, "page_size": page_size}
        if status:
            params["status"] = status.value
        if source_warehouse_id:
            params["source_warehouse_id"] = str(source_warehouse_id)
        if target_warehouse_id:
            params["target_warehouse_id"] = str(target_warehouse_id)
        if transfer_type:
            params["transfer_type"] = transfer_type.value

        result = self._request("GET", "/api/v1/transfers", params=params)
        assert isinstance(result, dict)
        return result

    def confirm_transfer(
        self,
        transfer_id: UUID,
        operator_id: UUID | None = None,
        remark: str | None = None,
    ) -> dict[str, Any]:
        if operator_id is None:
            operator_id = UUID("00000000-0000-0000-0000-000000000001")

        json_data: dict[str, Any] = {"operator_id": str(operator_id)}
        if remark:
            json_data["remark"] = remark

        result = self._request(
            "POST",
            f"/api/v1/transfers/{transfer_id}/confirm",
            json=json_data,
        )
        assert isinstance(result, dict)
        return result

    def approve_transfer(
        self,
        transfer_id: UUID,
        approval_level: ApprovalLevel,
        approval_status: ApprovalStatus,
        approver_id: UUID,
        approver_name: str,
        comment: str | None = None,
    ) -> dict[str, Any]:
        json_data: dict[str, Any] = {
            "approval_level": approval_level.value,
            "approval_status": approval_status.value,
            "approver_id": str(approver_id),
            "approver_name": approver_name,
        }
        if comment:
            json_data["comment"] = comment

        result = self._request(
            "POST",
            f"/api/v1/transfers/{transfer_id}/approve",
            json=json_data,
        )
        assert isinstance(result, dict)
        return result

    def approve_inter_company(
        self,
        transfer_id: UUID,
        operator_id: UUID,
        operator_name: str,
        approved: bool,
        comment: str | None = None,
    ) -> dict[str, Any]:
        json_data: dict[str, Any] = {
            "operator_id": str(operator_id),
            "operator_name": operator_name,
            "approved": approved,
        }
        if comment:
            json_data["comment"] = comment

        result = self._request(
            "POST",
            f"/api/v1/transfers/{transfer_id}/approve-inter-company",
            json=json_data,
        )
        assert isinstance(result, dict)
        return result

    def ship_transfer(
        self,
        transfer_id: UUID,
        operator_id: UUID | None = None,
    ) -> dict[str, Any]:
        if operator_id is None:
            operator_id = UUID("00000000-0000-0000-0000-000000000001")

        json_data: dict[str, Any] = {"operator_id": str(operator_id)}

        result = self._request(
            "POST",
            f"/api/v1/transfers/{transfer_id}/ship",
            json=json_data,
        )
        assert isinstance(result, dict)
        return result

    def arrive_transfer(
        self,
        transfer_id: UUID,
        arrival_items: list[dict[str, Any]],
        operator_id: UUID | None = None,
    ) -> dict[str, Any]:
        if operator_id is None:
            operator_id = UUID("00000000-0000-0000-0000-000000000001")

        json_data: dict[str, Any] = {
            "operator_id": str(operator_id),
            "arrival_items": arrival_items,
        }

        result = self._request(
            "POST",
            f"/api/v1/transfers/{transfer_id}/arrive",
            json=json_data,
        )
        assert isinstance(result, dict)
        return result

    def stock_transfer(
        self,
        transfer_id: UUID,
        operator_id: UUID | None = None,
    ) -> dict[str, Any]:
        if operator_id is None:
            operator_id = UUID("00000000-0000-0000-0000-000000000001")

        json_data: dict[str, Any] = {"operator_id": str(operator_id)}

        result = self._request(
            "POST",
            f"/api/v1/transfers/{transfer_id}/stock",
            json=json_data,
        )
        assert isinstance(result, dict)
        return result

    def cancel_transfer(
        self,
        transfer_id: UUID,
        reason: str,
        operator_id: UUID | None = None,
    ) -> dict[str, Any]:
        if operator_id is None:
            operator_id = UUID("00000000-0000-0000-0000-000000000001")

        json_data: dict[str, Any] = {
            "operator_id": str(operator_id),
            "reason": reason,
        }

        result = self._request(
            "POST",
            f"/api/v1/transfers/{transfer_id}/cancel",
            json=json_data,
        )
        assert isinstance(result, dict)
        return result

    def print_outbound(self, transfer_id: UUID) -> dict[str, Any]:
        result = self._request(
            "GET",
            f"/api/v1/transfers/{transfer_id}/print/outbound",
        )
        assert isinstance(result, dict)
        return result

    def print_inbound(self, transfer_id: UUID) -> dict[str, Any]:
        result = self._request(
            "GET",
            f"/api/v1/transfers/{transfer_id}/print/inbound",
        )
        assert isinstance(result, dict)
        return result

    def list_losses(
        self,
        warehouse_id: UUID | None = None,
        transfer_id: UUID | None = None,
        page: int = 1,
        page_size: int = 20,
    ) -> dict[str, Any]:
        params: dict[str, Any] = {"page": page, "page_size": page_size}
        if warehouse_id:
            params["warehouse_id"] = str(warehouse_id)
        if transfer_id:
            params["transfer_id"] = str(transfer_id)

        result = self._request("GET", "/api/v1/losses", params=params)
        assert isinstance(result, dict)
        return result

    def get_loss_summary(
        self,
        warehouse_id: UUID | None = None,
        year: int | None = None,
        month: int | None = None,
    ) -> dict[str, Any]:
        params: dict[str, Any] = {}
        if warehouse_id:
            params["warehouse_id"] = str(warehouse_id)
        if year:
            params["year"] = year
        if month:
            params["month"] = month

        result = self._request("GET", "/api/v1/losses/summary", params=params)
        assert isinstance(result, dict)
        return result

    def list_transactions(
        self,
        warehouse_id: UUID | None = None,
        product_id: UUID | None = None,
        reference_id: UUID | None = None,
        page: int = 1,
        page_size: int = 20,
    ) -> dict[str, Any]:
        params: dict[str, Any] = {"page": page, "page_size": page_size}
        if warehouse_id:
            params["warehouse_id"] = str(warehouse_id)
        if product_id:
            params["product_id"] = str(product_id)
        if reference_id:
            params["reference_id"] = str(reference_id)

        result = self._request("GET", "/api/v1/transactions", params=params)
        assert isinstance(result, dict)
        return result
