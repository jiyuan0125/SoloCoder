from typing import List
from fastapi import APIRouter, HTTPException, Query
from app.models import (
    GrayRule,
    GrayStrategy,
    RuleUpdateRequest,
    TrafficType
)
from app.rules_engine import rule_engine
from app.stats import stats_manager

router = APIRouter(prefix="/admin", tags=["admin"])


@router.post("/rules", response_model=GrayRule)
def add_or_update_rule(rule_request: RuleUpdateRequest):
    rule = GrayRule(**rule_request.model_dump())
    rule_engine.add_or_update_rule(rule)
    return rule


@router.get("/rules", response_model=List[GrayRule])
def list_rules():
    return rule_engine.get_all_rules()


@router.get("/rules/{rule_id}", response_model=GrayRule)
def get_rule(rule_id: str):
    rule = rule_engine.get_rule(rule_id)
    if not rule:
        raise HTTPException(status_code=404, detail="Rule not found")
    return rule


@router.delete("/rules/{rule_id}")
def delete_rule(rule_id: str):
    success = rule_engine.remove_rule(rule_id)
    if not success:
        raise HTTPException(status_code=404, detail="Rule not found")
    return {"message": "Rule deleted successfully"}


@router.patch("/rules/{rule_id}/enable")
def enable_rule(rule_id: str, enabled: bool = Query(True)):
    rule = rule_engine.get_rule(rule_id)
    if not rule:
        raise HTTPException(status_code=404, detail="Rule not found")
    rule.enabled = enabled
    rule_engine.add_or_update_rule(rule)
    return {"rule_id": rule_id, "enabled": enabled}


@router.get("/stats")
def get_all_stats():
    return stats_manager.get_stats()


@router.get("/stats/normal")
def get_normal_stats():
    return stats_manager.get_stats(TrafficType.NORMAL)


@router.get("/stats/gray")
def get_gray_stats():
    return stats_manager.get_stats(TrafficType.GRAY)


@router.post("/stats/reset")
def reset_stats():
    stats_manager.reset()
    return {"message": "Stats reset successfully"}
