from typing import List

from fastapi import APIRouter, HTTPException, status
from fastapi.responses import JSONResponse

from app.models import ColoringRule
from app.services import coloring_service

router = APIRouter(prefix="/api/coloring", tags=["coloring"])


@router.get("/rules", response_model=List[ColoringRule])
async def list_coloring_rules() -> List[ColoringRule]:
    return coloring_service.list_rules()


@router.post("/rules", response_model=ColoringRule)
async def add_coloring_rule(rule: ColoringRule) -> ColoringRule:
    coloring_service.add_rule(rule)
    return rule


@router.put("/rules/{rule_id}", response_model=ColoringRule)
async def update_coloring_rule(rule_id: str, rule: ColoringRule) -> ColoringRule:
    if rule.id != rule_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Rule ID in path must match ID in body",
        )
    coloring_service.add_rule(rule)
    return rule


@router.delete("/rules/{rule_id}")
async def delete_coloring_rule(rule_id: str) -> JSONResponse:
    deleted = coloring_service.remove_rule(rule_id)
    if not deleted:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Coloring rule '{rule_id}' not found",
        )
    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"message": f"Coloring rule '{rule_id}' deleted successfully"},
    )


@router.get("/rules/{rule_id}", response_model=ColoringRule)
async def get_coloring_rule(rule_id: str) -> ColoringRule:
    rule = coloring_service.get_rule(rule_id)
    if not rule:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Coloring rule '{rule_id}' not found",
        )
    return rule
