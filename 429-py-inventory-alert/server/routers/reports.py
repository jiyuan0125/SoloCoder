from datetime import datetime
from typing import Optional

from fastapi import APIRouter, Query

from shared.errors import ErrorCode
from shared.models import (
    APIResponse,
    InventoryHealthReport,
    InventoryTurnoverAnalysis,
)
from server.dependencies import get_inventory_service

router = APIRouter(prefix="/reports", tags=["reports"])


@router.post("/health", response_model=APIResponse[InventoryHealthReport])
def generate_health_report() -> APIResponse[InventoryHealthReport]:
    service = get_inventory_service()
    report = service.generate_health_report()
    return APIResponse(success=True, data=report)


@router.get("/health/latest", response_model=APIResponse[InventoryHealthReport])
def get_latest_health_report() -> APIResponse[InventoryHealthReport]:
    service = get_inventory_service()
    report = service.get_latest_health_report()
    if report is None:
        return APIResponse(
            success=False,
            error="No health report found",
            error_code=ErrorCode.REPORT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=report)


@router.get("/health/{report_id}", response_model=APIResponse[InventoryHealthReport])
def get_health_report(report_id: str) -> APIResponse[InventoryHealthReport]:
    service = get_inventory_service()
    report = service.get_health_report(report_id)
    if report is None:
        return APIResponse(
            success=False,
            error="Health report not found",
            error_code=ErrorCode.REPORT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=report)


@router.post("/turnover", response_model=APIResponse[InventoryTurnoverAnalysis])
def generate_turnover_analysis(
    start_date: datetime = Query(..., description="Start date in ISO format (UTC)"),
    end_date: datetime = Query(..., description="End date in ISO format (UTC)"),
) -> APIResponse[InventoryTurnoverAnalysis]:
    service = get_inventory_service()
    analysis = service.generate_turnover_analysis(start_date, end_date)
    return APIResponse(success=True, data=analysis)


@router.get("/turnover/{analysis_id}", response_model=APIResponse[InventoryTurnoverAnalysis])
def get_turnover_analysis(analysis_id: str) -> APIResponse[InventoryTurnoverAnalysis]:
    service = get_inventory_service()
    analysis = service.get_turnover_analysis(analysis_id)
    if analysis is None:
        return APIResponse(
            success=False,
            error="Turnover analysis not found",
            error_code=ErrorCode.INTERNAL_ERROR.value,
        )
    return APIResponse(success=True, data=analysis)
