from fastapi import APIRouter, HTTPException
from typing import List

from src.core import storage
from src.core.models import MonitoringPoint
from src.server.schemas import (
    MonitoringPointCreate,
    MonitoringPointUpdate,
    MonitoringPointResponse,
)

router = APIRouter(prefix="/monitoring-points", tags=["monitoring-points"])


@router.post("/", response_model=MonitoringPointResponse, status_code=201)
def create_point(point: MonitoringPointCreate):
    existing = storage.query(MonitoringPoint, code=point.code)
    if existing:
        raise HTTPException(status_code=400, detail=f"Point with code {point.code} already exists")
    
    db_point = MonitoringPoint(
        id=0,
        name=point.name,
        code=point.code,
        zone_type=point.zone_type,
        address=point.address,
        latitude=point.latitude,
        longitude=point.longitude,
        is_active=point.is_active,
        last_calibration_date=point.last_calibration_date,
        next_calibration_date=point.next_calibration_date,
        calibration_due=False,
    )
    return storage.create(db_point)


@router.get("/", response_model=List[MonitoringPointResponse])
def get_points():
    return storage.get_all(MonitoringPoint)


@router.get("/{point_id}", response_model=MonitoringPointResponse)
def get_point(point_id: int):
    point = storage.get_by_id(MonitoringPoint, point_id)
    if not point:
        raise HTTPException(status_code=404, detail="Monitoring point not found")
    return point


@router.put("/{point_id}", response_model=MonitoringPointResponse)
def update_point(point_id: int, point: MonitoringPointUpdate):
    db_point = storage.get_by_id(MonitoringPoint, point_id)
    if not db_point:
        raise HTTPException(status_code=404, detail="Monitoring point not found")
    
    update_data = point.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_point, key, value)
    
    return storage.update(db_point)


@router.delete("/{point_id}", status_code=204)
def delete_point(point_id: int):
    if not storage.delete(MonitoringPoint, point_id):
        raise HTTPException(status_code=404, detail="Monitoring point not found")
