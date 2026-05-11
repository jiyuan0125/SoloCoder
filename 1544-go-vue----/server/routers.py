from typing import List, Optional
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from .database import get_db
from .models import AggregationGranularity
from .schemas import (
    StationCreate, StationUpdate, StationResponse,
    RawDataCreate, BulkRawDataCreate, RawDataResponse,
    CalibrationCreate, CalibrationUpdate, CalibrationResponse, CalibrationTodoResponse,
    QualityRecordResponse, QualityRecordReview,
    AggregatedDataResponse, AggregationGranularityEnum
)
from .services import (
    StationService, RawDataService, QualityService,
    AggregationService, CalibrationService
)

router = APIRouter()


@router.post("/stations", response_model=StationResponse, status_code=201)
def create_station(station: StationCreate, db: Session = Depends(get_db)):
    return StationService.create(db, station)


@router.get("/stations", response_model=List[StationResponse])
def list_stations(
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db)
):
    stations, _ = StationService.get_all(db, skip=skip, limit=limit)
    return stations


@router.get("/stations/{station_id}", response_model=StationResponse)
def get_station(station_id: int, db: Session = Depends(get_db)):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    return station


@router.patch("/stations/{station_id}", response_model=StationResponse)
def update_station(station_id: int, data: StationUpdate, db: Session = Depends(get_db)):
    station = StationService.update(db, station_id, data)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    return station


@router.delete("/stations/{station_id}", status_code=204)
def delete_station(station_id: int, db: Session = Depends(get_db)):
    if not StationService.delete(db, station_id):
        raise HTTPException(status_code=404, detail="监测站不存在")


@router.get("/stations/{station_id}/raw-data", response_model=List[RawDataResponse])
def get_station_raw_data(
    station_id: int,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    data, _ = RawDataService.get_by_station(
        db, station_id, start_time=start_time, end_time=end_time, skip=skip, limit=limit
    )
    return data


@router.post("/stations/{station_id}/raw-data", response_model=RawDataResponse, status_code=201)
def create_raw_data(
    station_id: int,
    data: RawDataCreate,
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    return RawDataService.create(db, station_id, data)


@router.post("/stations/{station_id}/raw-data/bulk", response_model=List[RawDataResponse], status_code=201)
def create_bulk_raw_data(
    station_id: int,
    bulk_data: BulkRawDataCreate,
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    return RawDataService.create_bulk(db, station_id, bulk_data.data)


@router.get("/stations/{station_id}/aggregated", response_model=List[AggregatedDataResponse])
def get_station_aggregated_data(
    station_id: int,
    granularity: Optional[AggregationGranularityEnum] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    gran = AggregationGranularity(granularity.value) if granularity else None

    data, _ = AggregationService.get_aggregated_data(
        db, station_id, granularity=gran, start_time=start_time, end_time=end_time,
        skip=skip, limit=limit
    )
    return data


@router.post("/stations/{station_id}/aggregated/recalculate", response_model=List[AggregatedDataResponse])
def recalculate_aggregations(
    station_id: int,
    start_time: datetime,
    end_time: datetime,
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    return AggregationService.aggregate_for_period(db, station, start_time, end_time)


@router.get("/stations/{station_id}/quality", response_model=List[QualityRecordResponse])
def get_station_quality_records(
    station_id: int,
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    records, _ = QualityService.get_quality_records(db, station_id, skip=skip, limit=limit)
    return records


@router.patch("/stations/{station_id}/quality/{record_id}/review", response_model=QualityRecordResponse)
def review_quality_record(
    station_id: int,
    record_id: int,
    review_data: QualityRecordReview,
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    record = QualityService.review_quality_record(db, station_id, record_id, review_data)
    if not record:
        raise HTTPException(status_code=404, detail="质量记录不存在")

    return record


@router.get("/stations/{station_id}/calibrations", response_model=List[CalibrationResponse])
def get_station_calibrations(
    station_id: int,
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    calibrations, _ = CalibrationService.get_by_station(db, station_id, skip=skip, limit=limit)
    return calibrations


@router.post("/stations/{station_id}/calibrations", response_model=CalibrationResponse, status_code=201)
def create_calibration(
    station_id: int,
    data: CalibrationCreate,
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    return CalibrationService.create(db, station_id, data)


@router.patch("/stations/{station_id}/calibrations/{calibration_id}", response_model=CalibrationResponse)
def update_calibration(
    station_id: int,
    calibration_id: int,
    data: CalibrationUpdate,
    db: Session = Depends(get_db)
):
    station = StationService.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")

    calibration = CalibrationService.update(db, station_id, calibration_id, data)
    if not calibration:
        raise HTTPException(status_code=404, detail="检定记录不存在")

    return calibration


@router.get("/calibration-todos", response_model=List[CalibrationTodoResponse])
def get_calibration_todos(db: Session = Depends(get_db)):
    todos = CalibrationService.get_pending_todos(db)
    return todos
