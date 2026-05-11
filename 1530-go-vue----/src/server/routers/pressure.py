from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from src.core.database import get_db
from src.core.services import PressureService
from src.server.schemas import (
    PressurePointCreate,
    PressurePointResponse,
    PressureReadingCreate,
    PressureReadingResponse,
    PressureStatsResponse,
    LeakCheckResponse
)


router = APIRouter(prefix="/pressure", tags=["pressure"])


@router.post("/points/", response_model=PressurePointResponse)
def create_pressure_point(point: PressurePointCreate, db: Session = Depends(get_db)):
    service = PressureService(db)
    return service.create_pressure_point(point.model_dump())


@router.get("/points/", response_model=List[PressurePointResponse])
def list_pressure_points(pipeline_id: int, db: Session = Depends(get_db)):
    service = PressureService(db)
    return service.list_pressure_points(pipeline_id)


@router.post("/readings/", response_model=PressureReadingResponse)
def add_pressure_reading(reading: PressureReadingCreate, db: Session = Depends(get_db)):
    service = PressureService(db)
    point = service.get_pressure_point(reading.pressure_point_id)
    if not point:
        raise HTTPException(status_code=404, detail="Pressure point not found")
    return service.add_pressure_reading(reading.pressure_point_id, reading.pressure)


@router.get("/stats/{pipeline_id}", response_model=PressureStatsResponse)
def get_pressure_stats(pipeline_id: int, db: Session = Depends(get_db)):
    service = PressureService(db)
    return service.calculate_stats(pipeline_id)


@router.get("/check-leak/{start_point_id}/{end_point_id}", response_model=LeakCheckResponse)
def check_leak(
    start_point_id: int,
    end_point_id: int,
    db: Session = Depends(get_db)
):
    service = PressureService(db)

    start_point = service.get_pressure_point(start_point_id)
    end_point = service.get_pressure_point(end_point_id)

    if not start_point or not end_point:
        raise HTTPException(status_code=404, detail="Pressure point not found")

    if start_point.pipeline_id != end_point.pipeline_id:
        raise HTTPException(status_code=400, detail="Points must be in the same pipeline")

    is_suspected, details = service.check_leak_suspected(
        start_point_id, end_point_id, start_point.pipeline_id
    )

    diff = details.get("pressure_diff")
    if diff is not None:
        service.record_pressure_diff(
            start_point.pipeline_id, start_point_id, end_point_id, diff
        )

    return LeakCheckResponse(
        is_suspected=is_suspected,
        pressure_diff=details.get("pressure_diff"),
        mean=details.get("mean"),
        stddev=details.get("stddev"),
        threshold=details.get("threshold"),
        data_count=details.get("data_count")
    )


@router.post("/aggregate")
def trigger_aggregation(db: Session = Depends(get_db)):
    service = PressureService(db)
    service.aggregate_hourly_data()
    return {"message": "Aggregation completed"}
