from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, HTTPException

from src.core.models.schemas import DailyStatistics
from src.core.storage.memory_store import get_store
from src.core.services.aggregation_service import AggregationService


router = APIRouter(prefix="/statistics", tags=["statistics"])
store = get_store()
agg_service = AggregationService()


@router.get("/daily", response_model=List[DailyStatistics])
def get_daily_statistics(
    start_date: Optional[str] = None,
    end_date: Optional[str] = None
):
    return store.list_daily_statistics(start_date, end_date)


@router.post("/daily/{date_str}", response_model=DailyStatistics)
def calculate_daily_statistics(date_str: str):
    try:
        target_date = datetime.strptime(date_str, "%Y-%m-%d").date()
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid date format. Use YYYY-MM-DD")
    
    start_of_day = datetime.combine(target_date, datetime.min.time())
    end_of_day = datetime.combine(target_date, datetime.max.time())
    
    batches = store.list_batches(
        start_date=start_of_day,
        end_date=end_of_day
    )
    
    stats = agg_service.aggregate_daily_statistics(date_str, batches)
    stats = store.save_daily_statistics(stats)
    
    return stats


@router.get("/daily/{date_str}", response_model=DailyStatistics)
def get_specific_daily_statistics(date_str: str):
    stats = store.get_daily_statistics(date_str)
    if not stats:
        raise HTTPException(status_code=404, detail="Statistics not found for this date")
    return stats
