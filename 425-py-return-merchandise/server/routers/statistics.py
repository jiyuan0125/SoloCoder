from datetime import datetime
from typing import Optional

from fastapi import APIRouter, HTTPException, Query, status

from shared.models import (
    ReturnStatisticsRequest,
    ReturnStatisticsResponse,
    ScrapLedgerRequest,
    ScrapLedgerResponse,
    QualityAlertRequest,
    QualityAlertResponse,
    DisposalReason,
)
from server.services.statistics_service import StatisticsService

router = APIRouter()

_service = StatisticsService()


@router.get("/returns", response_model=ReturnStatisticsResponse)
async def get_return_statistics(
    start_date: datetime = Query(..., description="开始时间"),
    end_date: datetime = Query(..., description="结束时间"),
    category: Optional[str] = Query(None, description="商品类别（可选）"),
) -> ReturnStatisticsResponse:
    try:
        request = ReturnStatisticsRequest(
            start_date=start_date,
            end_date=end_date,
            category=category,
        )
        return _service.get_return_statistics(request)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.get("/scrap-ledger", response_model=ScrapLedgerResponse)
async def get_scrap_ledger(
    year: int = Query(..., ge=2000, le=2100, description="年份"),
    month: int = Query(..., ge=1, le=12, description="月份"),
    reason: Optional[DisposalReason] = Query(None, description="退货原因（可选）"),
) -> ScrapLedgerResponse:
    try:
        request = ScrapLedgerRequest(
            year=year,
            month=month,
            reason=reason,
        )
        return _service.get_scrap_ledger(request)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.get("/quality-alert", response_model=QualityAlertResponse)
async def check_quality_alert(
    supplier_id: str = Query(..., description="供应商ID"),
    period_days: int = Query(30, ge=1, le=365, description="统计周期天数"),
    threshold: Optional[float] = Query(None, ge=0.0, le=1.0, description="质量问题阈值（可选，默认10%）"),
) -> QualityAlertResponse:
    try:
        request = QualityAlertRequest(
            supplier_id=supplier_id,
            period_days=period_days,
            threshold=threshold,
        )
        return _service.check_quality_alert(request)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )
