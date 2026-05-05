from typing import List

from fastapi import APIRouter, HTTPException, status

from shared.models import (
    DefectiveToNormalRequest,
    DefectiveItemResponse,
    DefectiveItem,
)
from server.services.defective_service import DefectiveService

router = APIRouter()

_service = DefectiveService()


def _to_response(item: DefectiveItem) -> DefectiveItemResponse:
    return DefectiveItemResponse(
        defective_id=item.defective_id,
        sku=item.sku,
        product_name=item.product_name,
        quantity=item.quantity,
        unit_price=item.unit_price,
        defective_price=item.defective_price,
        category=item.category,
        return_order_id=item.return_order_id,
        created_at=item.created_at,
        converted_to_normal=item.converted_to_normal,
        converted_at=item.converted_at,
        converted_by=item.converted_by,
    )


@router.post("/convert", response_model=DefectiveItemResponse)
async def convert_to_normal(request: DefectiveToNormalRequest) -> DefectiveItemResponse:
    try:
        item = _service.convert_to_normal(request)
        return _to_response(item)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.get("/{defective_id}", response_model=DefectiveItemResponse)
async def get_defective_item(defective_id: str) -> DefectiveItemResponse:
    try:
        item = _service.get_defective_item(defective_id)
        return _to_response(item)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e),
        )


@router.get("/", response_model=List[DefectiveItemResponse])
async def list_defective_items() -> List[DefectiveItemResponse]:
    items = _service.list_defective_items()
    return [_to_response(i) for i in items]
