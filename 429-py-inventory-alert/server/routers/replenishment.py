from typing import List, Optional

from fastapi import APIRouter, Query

from shared.errors import ErrorCode
from shared.models import (
    APIResponse,
    ReplenishmentOrderDraft,
)
from server.dependencies import get_inventory_service

router = APIRouter(prefix="/replenishment", tags=["replenishment"])


@router.get("", response_model=APIResponse[List[ReplenishmentOrderDraft]])
def list_drafts(
    sku: Optional[str] = Query(None),
    confirmed: Optional[bool] = Query(None),
) -> APIResponse[List[ReplenishmentOrderDraft]]:
    service = get_inventory_service()
    drafts = service.list_replenishment_drafts(sku, confirmed)
    return APIResponse(success=True, data=drafts)


@router.get("/{draft_id}", response_model=APIResponse[ReplenishmentOrderDraft])
def get_draft(draft_id: str) -> APIResponse[ReplenishmentOrderDraft]:
    service = get_inventory_service()
    draft = service.get_replenishment_draft(draft_id)
    if draft is None:
        return APIResponse(
            success=False,
            error="Replenishment draft not found",
            error_code=ErrorCode.REPLENISHMENT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=draft)


@router.post("/{draft_id}/confirm", response_model=APIResponse[ReplenishmentOrderDraft])
def confirm_draft(draft_id: str) -> APIResponse[ReplenishmentOrderDraft]:
    service = get_inventory_service()
    draft = service.confirm_replenishment_draft(draft_id)
    if draft is None:
        return APIResponse(
            success=False,
            error="Draft not found or already confirmed",
            error_code=ErrorCode.REPLENISHMENT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=draft)
