from fastapi import APIRouter

from shared.models import (
    BoxType,
    BoxTypeCreateRequest,
    ApiResponse,
)
from shared import protocols
from server.app.exceptions import PackingSystemException
from server.app.services.box_type_service import (
    create_box_type,
    get_box_type,
    get_all_box_types,
    get_all_inventories,
    add_stock,
)

router = APIRouter()


@router.post(protocols.BOX_TYPES_ENDPOINT, response_model=ApiResponse)
async def create_new_box_type(request: BoxTypeCreateRequest) -> ApiResponse:
    try:
        box_type = create_box_type(
            box_type_id=request.box_type_id,
            name=request.name,
            length_cm=request.length_cm,
            width_cm=request.width_cm,
            height_cm=request.height_cm,
            max_weight_kg=request.max_weight_kg,
            initial_stock=request.initial_stock,
        )
        return ApiResponse(
            success=True,
            data={
                "box_type": box_type.model_dump(mode="json"),
            },
            message="箱型创建成功",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.BOX_TYPES_ENDPOINT, response_model=ApiResponse)
async def list_box_types() -> ApiResponse:
    box_types = get_all_box_types()
    return ApiResponse(
        success=True,
        data={
            "box_types": [b.model_dump(mode="json") for b in box_types],
        },
    )


@router.get(protocols.BOX_TYPE_ENDPOINT, response_model=ApiResponse)
async def get_single_box_type(box_type_id: str) -> ApiResponse:
    try:
        box_type = get_box_type(box_type_id)
        return ApiResponse(
            success=True,
            data={
                "box_type": box_type.model_dump(mode="json"),
            },
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.MATERIAL_STATS_ENDPOINT, response_model=ApiResponse)
async def get_material_stats() -> ApiResponse:
    inventories = get_all_inventories()
    return ApiResponse(
        success=True,
        data={
            "stats": [
                {
                    "box_type": inv.box_type.model_dump(mode="json"),
                    "stock_quantity": inv.stock_quantity,
                    "used_quantity": inv.used_quantity,
                }
                for inv in inventories
            ],
        },
    )
