from typing import List, Optional

from fastapi import APIRouter, HTTPException, Query

from shared.errors import ErrorCode
from shared.models import (
    APIResponse,
    BulkSafetyStockUpdate,
    Product,
    ProductCreate,
    ProductUpdate,
    SafetyStock,
)
from server.dependencies import get_inventory_service

router = APIRouter(prefix="/products", tags=["products"])


@router.post("", response_model=APIResponse[Product])
def create_product(product_data: ProductCreate) -> APIResponse[Product]:
    service = get_inventory_service()
    try:
        product = service.create_product(product_data)
        return APIResponse(success=True, data=product)
    except ValueError as e:
        if str(e) == ErrorCode.PRODUCT_ALREADY_EXISTS.value:
            return APIResponse(
                success=False,
                error="Product already exists",
                error_code=ErrorCode.PRODUCT_ALREADY_EXISTS.value,
            )
        return APIResponse(
            success=False,
            error=str(e),
            error_code=ErrorCode.INTERNAL_ERROR.value,
        )


@router.get("/{sku}", response_model=APIResponse[Product])
def get_product(sku: str) -> APIResponse[Product]:
    service = get_inventory_service()
    product = service.get_product(sku)
    if product is None:
        return APIResponse(
            success=False,
            error="Product not found",
            error_code=ErrorCode.PRODUCT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=product)


@router.get("", response_model=APIResponse[List[Product]])
def list_products(category: Optional[str] = Query(None)) -> APIResponse[List[Product]]:
    service = get_inventory_service()
    products = service.list_products(category)
    return APIResponse(success=True, data=products)


@router.put("/{sku}", response_model=APIResponse[Product])
def update_product(sku: str, update_data: ProductUpdate) -> APIResponse[Product]:
    service = get_inventory_service()
    product = service.update_product(sku, update_data)
    if product is None:
        return APIResponse(
            success=False,
            error="Product not found",
            error_code=ErrorCode.PRODUCT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=product)


@router.delete("/{sku}", response_model=APIResponse[bool])
def delete_product(sku: str) -> APIResponse[bool]:
    service = get_inventory_service()
    success = service.delete_product(sku)
    if not success:
        return APIResponse(
            success=False,
            error="Product not found",
            error_code=ErrorCode.PRODUCT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=True)


@router.put("/{sku}/safety-stock", response_model=APIResponse[Product])
def update_safety_stock(sku: str, safety_stock: SafetyStock) -> APIResponse[Product]:
    service = get_inventory_service()
    product = service.update_safety_stock(sku, safety_stock)
    if product is None:
        return APIResponse(
            success=False,
            error="Product not found",
            error_code=ErrorCode.PRODUCT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=product)


@router.post("/bulk-safety-stock", response_model=APIResponse[int])
def bulk_update_safety_stock(bulk_update: BulkSafetyStockUpdate) -> APIResponse[int]:
    service = get_inventory_service()
    count = service.bulk_update_safety_stock(bulk_update)
    return APIResponse(success=True, data=count)
