from uuid import UUID

from fastapi import APIRouter
from pydantic import UUID4

from shared.models import (
    Package,
    CreatePackageRequest,
    AddItemToPackageRequest,
    UpdateTrackingNumberRequest,
    ApiResponse,
)
from shared import protocols
from server.app.exceptions import PackingSystemException
from server.app.services.packing_service import (
    create_package,
    get_package,
    add_item_to_package,
    mark_package_shipped,
    get_packages_by_order,
)

router = APIRouter()


@router.post(protocols.PACKAGES_ENDPOINT, response_model=ApiResponse)
async def create_new_package(order_id: str, request: CreatePackageRequest) -> ApiResponse:
    try:
        package = create_package(
            order_id=order_id,
            box_type_id=request.box_type_id,
        )
        return ApiResponse(
            success=True,
            data={
                "package": package.model_dump(mode="json"),
            },
            message="包裹创建成功",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.PACKAGES_ENDPOINT, response_model=ApiResponse)
async def list_order_packages(order_id: str) -> ApiResponse:
    try:
        packages = get_packages_by_order(order_id)
        return ApiResponse(
            success=True,
            data={
                "order_id": order_id,
                "packages": [p.model_dump(mode="json") for p in packages],
            },
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.PACKAGE_ENDPOINT, response_model=ApiResponse)
async def get_single_package(package_id: UUID4) -> ApiResponse:
    try:
        package = get_package(package_id)
        return ApiResponse(
            success=True,
            data={
                "package": package.model_dump(mode="json"),
            },
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.post(protocols.PACKAGE_ITEMS_ENDPOINT, response_model=ApiResponse)
async def add_item_to_package_endpoint(
    package_id: UUID4,
    request: AddItemToPackageRequest,
) -> ApiResponse:
    try:
        package = add_item_to_package(
            package_id=package_id,
            product_id=request.product_id,
            quantity=request.quantity,
        )
        return ApiResponse(
            success=True,
            data={
                "package": package.model_dump(mode="json"),
            },
            message="商品已添加到包裹",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.post(protocols.PACKAGE_TRACKING_ENDPOINT, response_model=ApiResponse)
async def update_tracking_number(
    package_id: UUID4,
    request: UpdateTrackingNumberRequest,
) -> ApiResponse:
    try:
        package = mark_package_shipped(
            package_id=package_id,
            tracking_number=request.tracking_number,
        )
        return ApiResponse(
            success=True,
            data={
                "package": package.model_dump(mode="json"),
            },
            message="物流单号已更新，包裹标记为已发出",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )
