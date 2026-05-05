from typing import Any, List, Optional

from fastapi import APIRouter, Query

from shared.errors import ErrorCode
from shared.models import (
    APIResponse,
    InventoryTransaction,
    SeasonalSafetyStock,
    StockAdjustment,
    TransactionType,
)
from server.dependencies import get_inventory_service

router = APIRouter(prefix="/inventory", tags=["inventory"])


@router.post("/{sku}/inbound", response_model=APIResponse[InventoryTransaction])
def stock_inbound(sku: str, adjustment: StockAdjustment) -> APIResponse[InventoryTransaction]:
    service = get_inventory_service()
    try:
        transaction = service.stock_inbound(sku, adjustment)
        return APIResponse(success=True, data=transaction)
    except ValueError as e:
        if str(e) == ErrorCode.PRODUCT_NOT_FOUND.value:
            return APIResponse(
                success=False,
                error="Product not found",
                error_code=ErrorCode.PRODUCT_NOT_FOUND.value,
            )
        return APIResponse(
            success=False,
            error=str(e),
            error_code=ErrorCode.INTERNAL_ERROR.value,
        )


@router.post("/{sku}/outbound", response_model=APIResponse[InventoryTransaction])
def stock_outbound(sku: str, adjustment: StockAdjustment) -> APIResponse[InventoryTransaction]:
    service = get_inventory_service()
    try:
        transaction = service.stock_outbound(sku, adjustment)
        return APIResponse(success=True, data=transaction)
    except ValueError as e:
        if str(e) == ErrorCode.PRODUCT_NOT_FOUND.value:
            return APIResponse(
                success=False,
                error="Product not found",
                error_code=ErrorCode.PRODUCT_NOT_FOUND.value,
            )
        if str(e) == ErrorCode.INSUFFICIENT_STOCK.value:
            return APIResponse(
                success=False,
                error="Insufficient stock",
                error_code=ErrorCode.INSUFFICIENT_STOCK.value,
            )
        return APIResponse(
            success=False,
            error=str(e),
            error_code=ErrorCode.INTERNAL_ERROR.value,
        )


@router.post("/{sku}/transfer", response_model=APIResponse[InventoryTransaction])
def stock_transfer(sku: str, adjustment: StockAdjustment) -> APIResponse[InventoryTransaction]:
    service = get_inventory_service()
    try:
        transaction = service.stock_transfer(sku, adjustment)
        return APIResponse(success=True, data=transaction)
    except ValueError as e:
        if str(e) == ErrorCode.PRODUCT_NOT_FOUND.value:
            return APIResponse(
                success=False,
                error="Product not found",
                error_code=ErrorCode.PRODUCT_NOT_FOUND.value,
            )
        if str(e) == ErrorCode.INSUFFICIENT_STOCK.value:
            return APIResponse(
                success=False,
                error="Insufficient stock",
                error_code=ErrorCode.INSUFFICIENT_STOCK.value,
            )
        return APIResponse(
            success=False,
            error=str(e),
            error_code=ErrorCode.INTERNAL_ERROR.value,
        )


@router.get("/transactions", response_model=APIResponse[List[InventoryTransaction]])
def list_transactions(
    sku: Optional[str] = Query(None),
    transaction_type: Optional[TransactionType] = Query(None),
) -> APIResponse[List[InventoryTransaction]]:
    service = get_inventory_service()
    transactions = service.list_transactions(sku, transaction_type)
    return APIResponse(success=True, data=transactions)


@router.get("/transactions/{transaction_id}", response_model=APIResponse[InventoryTransaction])
def get_transaction(transaction_id: str) -> APIResponse[InventoryTransaction]:
    service = get_inventory_service()
    transaction = service.get_transaction(transaction_id)
    if transaction is None:
        return APIResponse(
            success=False,
            error="Transaction not found",
            error_code=ErrorCode.INTERNAL_ERROR.value,
        )
    return APIResponse(success=True, data=transaction)


@router.get("/{sku}/status", response_model=APIResponse[dict[str, Any]])
def get_inventory_status(sku: str) -> APIResponse[dict[str, Any]]:
    service = get_inventory_service()
    result = service.get_product_inventory_status(sku)
    if result is None:
        return APIResponse(
            success=False,
            error="Product not found",
            error_code=ErrorCode.PRODUCT_NOT_FOUND.value,
        )
    product, status, ss = result
    return APIResponse(
        success=True,
        data={
            "sku": product.sku,
            "name": product.name,
            "current_stock": product.current_stock,
            "min_stock": ss.min_stock,
            "max_stock": ss.max_stock,
            "status": status.value,
            "is_seasonal": product.safety_stock.is_seasonal,
        },
    )
