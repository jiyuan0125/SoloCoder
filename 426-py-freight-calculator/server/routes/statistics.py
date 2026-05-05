from typing import List, Optional, Any
from decimal import Decimal

from fastapi import APIRouter, HTTPException

from shared.responses import StatisticsResponse
from shared.models import MonthlyStatistics, TransportMode

from server.store import data_store
from server.cache import freight_cache


router = APIRouter(prefix="/statistics", tags=["statistics"])


@router.get("/monthly/{year_month}", response_model=StatisticsResponse)
async def get_monthly_statistics(year_month: str) -> StatisticsResponse:
    stats = data_store.get_monthly_statistics(year_month)
    
    if stats is None:
        raise HTTPException(
            status_code=404,
            detail=f"未找到 {year_month} 的统计数据",
        )
    
    return StatisticsResponse(
        success=True,
        data=stats,
        message="获取统计数据成功",
    )


@router.get("/months", response_model=StatisticsResponse)
async def list_available_months() -> StatisticsResponse:
    months = data_store.list_available_months()
    
    empty_breakdown: dict[TransportMode, int] = {}
    
    return StatisticsResponse(
        success=True,
        data=MonthlyStatistics(
            year_month="",
            total_packages=data_store.get_total_packages_count(),
            total_base_freight=Decimal("0"),
            total_surcharge=Decimal("0"),
            total_insurance_fee=Decimal("0"),
            total_discount=Decimal("0"),
            total_final_amount=Decimal("0"),
            transport_mode_breakdown=empty_breakdown,
        ),
        message=f"可用月份: {', '.join(months) if months else '无数据'}",
    )


@router.get("/cache/stats")
async def get_cache_stats() -> dict[str, Any]:
    stats = freight_cache.get_stats()
    return {
        "success": True,
        "data": stats,
        "message": "缓存统计信息",
    }


@router.post("/cache/cleanup")
async def cleanup_expired_cache() -> dict[str, Any]:
    removed = freight_cache.cleanup_expired()
    return {
        "success": True,
        "data": {"removed_count": removed},
        "message": f"清理了 {removed} 个过期缓存项",
    }


@router.delete("/cache/all")
async def clear_all_cache() -> dict[str, Any]:
    removed = freight_cache.clear_all()
    return {
        "success": True,
        "data": {"removed_count": removed},
        "message": f"清空了 {removed} 个缓存项",
    }
