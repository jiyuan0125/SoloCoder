from fastapi import APIRouter, HTTPException, Query
from fastapi.responses import StreamingResponse
from typing import List, Optional
from datetime import date, datetime

from src.core import storage, calculate_daily_statistics, export_daily_stats_to_csv
from src.core.models import DailyStatistics
from src.server.schemas import DailyStatisticsResponse

router = APIRouter(prefix="/statistics", tags=["statistics"])


@router.post("/daily/{point_id}/{target_date}", response_model=DailyStatisticsResponse)
def compute_daily_statistics(point_id: int, target_date: date):
    try:
        return calculate_daily_statistics(point_id, target_date)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/daily/", response_model=List[DailyStatisticsResponse])
def get_daily_statistics(
    point_id: Optional[int] = Query(None, description="Filter by monitoring point ID"),
    start_date: Optional[date] = Query(None, description="Start date for filtering"),
    end_date: Optional[date] = Query(None, description="End date for filtering"),
):
    results = storage.get_all(DailyStatistics)
    
    if point_id is not None:
        results = [s for s in results if s.point_id == point_id]
    
    if start_date:
        results = [s for s in results if s.date >= start_date]
    
    if end_date:
        results = [s for s in results if s.date <= end_date]
    
    results.sort(key=lambda x: (x.point_id, x.date))
    return results


@router.get("/daily/{stats_id}", response_model=DailyStatisticsResponse)
def get_daily_statistic(stats_id: int):
    stats = storage.get_by_id(DailyStatistics, stats_id)
    if not stats:
        raise HTTPException(status_code=404, detail="Daily statistics not found")
    return stats


@router.get("/export/csv")
def export_statistics_csv(
    point_id: Optional[int] = Query(None, description="Filter by monitoring point ID"),
    start_date: Optional[date] = Query(None, description="Start date for filtering"),
    end_date: Optional[date] = Query(None, description="End date for filtering"),
):
    results = storage.get_all(DailyStatistics)
    
    if point_id is not None:
        results = [s for s in results if s.point_id == point_id]
    
    if start_date:
        results = [s for s in results if s.date >= start_date]
    
    if end_date:
        results = [s for s in results if s.date <= end_date]
    
    results.sort(key=lambda x: (x.point_id, x.date))
    csv_content = export_daily_stats_to_csv(results)
    
    filename = f"daily_statistics_{datetime.now().strftime('%Y%m%d%H%M%S')}.csv"
    
    return StreamingResponse(
        iter([csv_content]),
        media_type="text/csv",
        headers={
            "Content-Disposition": f"attachment; filename={filename}"
        }
    )
