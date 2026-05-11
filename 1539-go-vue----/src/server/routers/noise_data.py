from fastapi import APIRouter, HTTPException, Query
from fastapi.responses import StreamingResponse
from typing import List, Optional
from datetime import datetime

from src.core import storage, process_noise_data, export_to_csv
from src.core.models import NoiseData, MonitoringPoint
from src.server.schemas import (
    NoiseDataCreate,
    NoiseDataResponse,
)

router = APIRouter(prefix="/noise-data", tags=["noise-data"])


@router.post("/", response_model=NoiseDataResponse, status_code=201)
def create_noise_data(data: NoiseDataCreate):
    try:
        return process_noise_data(data.point_id, data.value, data.timestamp)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[NoiseDataResponse])
def get_noise_data(
    point_id: Optional[int] = Query(None, description="Filter by monitoring point ID"),
    start_time: Optional[datetime] = Query(None, description="Start time for filtering"),
    end_time: Optional[datetime] = Query(None, description="End time for filtering"),
):
    results = storage.get_all(NoiseData)
    
    if point_id is not None:
        results = [d for d in results if d.point_id == point_id]
    
    if start_time:
        results = [d for d in results if d.timestamp >= start_time]
    
    if end_time:
        results = [d for d in results if d.timestamp <= end_time]
    
    results.sort(key=lambda x: x.timestamp, reverse=True)
    return results


@router.get("/{data_id}", response_model=NoiseDataResponse)
def get_noise_data_item(data_id: int):
    data = storage.get_by_id(NoiseData, data_id)
    if not data:
        raise HTTPException(status_code=404, detail="Noise data not found")
    return data


@router.get("/export/csv")
def export_noise_data_csv(
    point_id: Optional[int] = Query(None, description="Filter by monitoring point ID"),
    start_time: Optional[datetime] = Query(None, description="Start time for filtering"),
    end_time: Optional[datetime] = Query(None, description="End time for filtering"),
):
    results = storage.get_all(NoiseData)
    
    if point_id is not None:
        results = [d for d in results if d.point_id == point_id]
    
    if start_time:
        results = [d for d in results if d.timestamp >= start_time]
    
    if end_time:
        results = [d for d in results if d.timestamp <= end_time]
    
    results.sort(key=lambda x: x.timestamp)
    csv_content = export_to_csv(results)
    
    filename = f"noise_data_{datetime.now().strftime('%Y%m%d%H%M%S')}.csv"
    
    return StreamingResponse(
        iter([csv_content]),
        media_type="text/csv",
        headers={
            "Content-Disposition": f"attachment; filename={filename}"
        }
    )
