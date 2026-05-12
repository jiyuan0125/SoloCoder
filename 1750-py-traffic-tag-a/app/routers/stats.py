from typing import List, Optional

from fastapi import APIRouter, HTTPException, status

from app.models import ServiceStats
from app.services import stats_service

router = APIRouter(prefix="/api/stats", tags=["stats"])


@router.get("", response_model=List[ServiceStats])
async def list_stats() -> List[ServiceStats]:
    return stats_service.list_service_stats()


@router.get("/{service}", response_model=ServiceStats)
async def get_service_stats(service: str) -> ServiceStats:
    stats = stats_service.get_service_stats(service)
    if not stats:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"No stats found for service '{service}'",
        )
    return stats


@router.post("/reset")
async def reset_all_stats() -> dict:
    stats_service.reset_stats()
    return {"message": "All stats reset successfully"}


@router.post("/{service}/reset")
async def reset_service_stats(service: str) -> dict:
    stats_service.reset_stats(service)
    return {"message": f"Stats for service '{service}' reset successfully"}
