from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from sqlalchemy import and_
from datetime import datetime
from typing import List, Optional
from server.database import get_db, SessionLocal
from server.models import (
    Watershed, Reservoir, MonitoringStation, HydrologicalData,
    Alert, DispatchRecommendation, ReservoirOperation, FloodPhase, FloodEvent,
    DispatchStatus
)
from server.schemas import (
    WatershedCreate, WatershedUpdate, WatershedResponse,
    ReservoirCreate, ReservoirUpdate, ReservoirResponse,
    MonitoringStationCreate, MonitoringStationUpdate, MonitoringStationResponse,
    HydrologicalDataCreate, HydrologicalDataResponse,
    AlertResponse, DispatchRecommendationCreate, DispatchRecommendationResponse,
    DispatchApprove, DispatchReject, ReservoirOperationResponse,
    FloodPhaseResponse, FloodEventResponse, FloodEventDetail
)
from server.services import (
    calculate_6h_rainfall, check_and_create_alerts,
    generate_dispatch_recommendations, execute_dispatch,
    evaluate_flood_situation, end_flood_event
)

router = APIRouter()


@router.post("/watersheds/", response_model=WatershedResponse, status_code=status.HTTP_201_CREATED)
def create_watershed(data: WatershedCreate, db: Session = Depends(get_db)):
    existing = db.query(Watershed).filter(Watershed.name == data.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="Watershed name already exists")
    watershed = Watershed(**data.model_dump())
    db.add(watershed)
    db.commit()
    db.refresh(watershed)
    return watershed


@router.get("/watersheds/", response_model=List[WatershedResponse])
def list_watersheds(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return db.query(Watershed).offset(skip).limit(limit).all()


@router.get("/watersheds/{id}", response_model=WatershedResponse)
def get_watershed(id: int, db: Session = Depends(get_db)):
    watershed = db.query(Watershed).filter(Watershed.id == id).first()
    if not watershed:
        raise HTTPException(status_code=404, detail="Watershed not found")
    return watershed


@router.put("/watersheds/{id}", response_model=WatershedResponse)
def update_watershed(id: int, data: WatershedUpdate, db: Session = Depends(get_db)):
    watershed = db.query(Watershed).filter(Watershed.id == id).first()
    if not watershed:
        raise HTTPException(status_code=404, detail="Watershed not found")
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(watershed, key, value)
    db.commit()
    db.refresh(watershed)
    return watershed


@router.post("/reservoirs/", response_model=ReservoirResponse, status_code=status.HTTP_201_CREATED)
def create_reservoir(data: ReservoirCreate, db: Session = Depends(get_db)):
    existing = db.query(Reservoir).filter(Reservoir.code == data.code).first()
    if existing:
        raise HTTPException(status_code=400, detail="Reservoir code already exists")
    reservoir = Reservoir(**data.model_dump())
    db.add(reservoir)
    db.commit()
    db.refresh(reservoir)
    return reservoir


@router.get("/reservoirs/", response_model=List[ReservoirResponse])
def list_reservoirs(skip: int = 0, limit: int = 100, watershed_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(Reservoir)
    if watershed_id is not None:
        query = query.filter(Reservoir.watershed_id == watershed_id)
    return query.offset(skip).limit(limit).all()


@router.get("/reservoirs/{id}", response_model=ReservoirResponse)
def get_reservoir(id: int, db: Session = Depends(get_db)):
    reservoir = db.query(Reservoir).filter(Reservoir.id == id).first()
    if not reservoir:
        raise HTTPException(status_code=404, detail="Reservoir not found")
    return reservoir


@router.put("/reservoirs/{id}", response_model=ReservoirResponse)
def update_reservoir(id: int, data: ReservoirUpdate, db: Session = Depends(get_db)):
    reservoir = db.query(Reservoir).filter(Reservoir.id == id).first()
    if not reservoir:
        raise HTTPException(status_code=404, detail="Reservoir not found")
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(reservoir, key, value)
    db.commit()
    db.refresh(reservoir)
    return reservoir


@router.post("/stations/", response_model=MonitoringStationResponse, status_code=status.HTTP_201_CREATED)
def create_station(data: MonitoringStationCreate, db: Session = Depends(get_db)):
    existing = db.query(MonitoringStation).filter(MonitoringStation.code == data.code).first()
    if existing:
        raise HTTPException(status_code=400, detail="Station code already exists")
    station = MonitoringStation(**data.model_dump())
    db.add(station)
    db.commit()
    db.refresh(station)
    return station


@router.get("/stations/", response_model=List[MonitoringStationResponse])
def list_stations(skip: int = 0, limit: int = 100, watershed_id: Optional[int] = None,
                  station_type: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(MonitoringStation)
    if watershed_id is not None:
        query = query.filter(MonitoringStation.watershed_id == watershed_id)
    if station_type:
        query = query.filter(MonitoringStation.station_type == station_type)
    return query.offset(skip).limit(limit).all()


@router.get("/stations/{id}", response_model=MonitoringStationResponse)
def get_station(id: int, db: Session = Depends(get_db)):
    station = db.query(MonitoringStation).filter(MonitoringStation.id == id).first()
    if not station:
        raise HTTPException(status_code=404, detail="Station not found")
    return station


@router.put("/stations/{id}", response_model=MonitoringStationResponse)
def update_station(id: int, data: MonitoringStationUpdate, db: Session = Depends(get_db)):
    station = db.query(MonitoringStation).filter(MonitoringStation.id == id).first()
    if not station:
        raise HTTPException(status_code=404, detail="Station not found")
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(station, key, value)
    db.commit()
    db.refresh(station)
    return station


@router.post("/hydrological-data/", response_model=HydrologicalDataResponse, status_code=status.HTTP_201_CREATED)
def report_hydrological_data(data: HydrologicalDataCreate, db: Session = Depends(get_db)):
    station = db.query(MonitoringStation).filter(
        MonitoringStation.id == data.station_id,
        MonitoringStation.is_active == True
    ).first()
    if not station:
        raise HTTPException(status_code=404, detail="Station not found or inactive")
    record_data = data.model_dump()
    if station.station_type == "rain" and record_data.get("rainfall_6h") is None:
        recorded_at = record_data.get("recorded_at") or datetime.utcnow()
        record_data["rainfall_6h"] = calculate_6h_rainfall(db, data.station_id, recorded_at) + data.value
    record = HydrologicalData(**record_data)
    db.add(record)
    db.flush()
    active_event = None
    if station.watershed_id:
        active_event = db.query(FloodEvent).filter(
            and_(
                FloodEvent.watershed_id == station.watershed_id,
                FloodEvent.is_active == True
            )
        ).first()
    if active_event:
        record.flood_event_id = active_event.id
    alerts = check_and_create_alerts(db, record)
    db.commit()
    db.refresh(record)
    return record


@router.get("/hydrological-data/", response_model=List[HydrologicalDataResponse])
def list_hydrological_data(skip: int = 0, limit: int = 100, station_id: Optional[int] = None,
                           flood_event_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(HydrologicalData)
    if station_id is not None:
        query = query.filter(HydrologicalData.station_id == station_id)
    if flood_event_id is not None:
        query = query.filter(HydrologicalData.flood_event_id == flood_event_id)
    return query.order_by(HydrologicalData.recorded_at.desc()).offset(skip).limit(limit).all()


@router.get("/alerts/", response_model=List[AlertResponse])
def list_alerts(skip: int = 0, limit: int = 100, is_active: Optional[bool] = None,
                flood_event_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(Alert)
    if is_active is not None:
        query = query.filter(Alert.is_active == is_active)
    if flood_event_id is not None:
        query = query.filter(Alert.flood_event_id == flood_event_id)
    return query.order_by(Alert.triggered_at.desc()).offset(skip).limit(limit).all()


@router.get("/alerts/{id}", response_model=AlertResponse)
def get_alert(id: int, db: Session = Depends(get_db)):
    alert = db.query(Alert).filter(Alert.id == id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    return alert


@router.post("/dispatches/", response_model=DispatchRecommendationResponse, status_code=status.HTTP_201_CREATED)
def create_manual_dispatch(data: DispatchRecommendationCreate, db: Session = Depends(get_db)):
    dispatch = DispatchRecommendation(
        reservoir_id=data.reservoir_id,
        recommended_discharge=data.recommended_discharge,
        rationale=data.rationale,
        status=DispatchStatus.PENDING.value,
        created_by="manual"
    )
    db.add(dispatch)
    db.commit()
    db.refresh(dispatch)
    return dispatch


@router.get("/dispatches/", response_model=List[DispatchRecommendationResponse])
def list_dispatches(skip: int = 0, limit: int = 100, status: Optional[str] = None,
                    flood_event_id: Optional[int] = None, reservoir_id: Optional[int] = None,
                    db: Session = Depends(get_db)):
    query = db.query(DispatchRecommendation)
    if status:
        query = query.filter(DispatchRecommendation.status == status)
    if flood_event_id is not None:
        query = query.filter(DispatchRecommendation.flood_event_id == flood_event_id)
    if reservoir_id is not None:
        query = query.filter(DispatchRecommendation.reservoir_id == reservoir_id)
    return query.order_by(DispatchRecommendation.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/dispatches/{id}", response_model=DispatchRecommendationResponse)
def get_dispatch(id: int, db: Session = Depends(get_db)):
    dispatch = db.query(DispatchRecommendation).filter(DispatchRecommendation.id == id).first()
    if not dispatch:
        raise HTTPException(status_code=404, detail="Dispatch not found")
    return dispatch


@router.post("/dispatches/{id}/approve", response_model=DispatchRecommendationResponse)
def approve_dispatch(id: int, data: DispatchApprove, db: Session = Depends(get_db)):
    dispatch = db.query(DispatchRecommendation).filter(DispatchRecommendation.id == id).first()
    if not dispatch:
        raise HTTPException(status_code=404, detail="Dispatch not found")
    if dispatch.status != DispatchStatus.PENDING.value:
        raise HTTPException(status_code=400, detail="Dispatch is not pending")
    dispatch.status = DispatchStatus.APPROVED.value
    dispatch.approved_by = data.approved_by
    dispatch.approved_at = datetime.utcnow()
    execute_dispatch(db, dispatch, data.approved_by)
    db.commit()
    db.refresh(dispatch)
    return dispatch


@router.post("/dispatches/{id}/reject", response_model=DispatchRecommendationResponse)
def reject_dispatch(id: int, data: DispatchReject, db: Session = Depends(get_db)):
    dispatch = db.query(DispatchRecommendation).filter(DispatchRecommendation.id == id).first()
    if not dispatch:
        raise HTTPException(status_code=404, detail="Dispatch not found")
    if dispatch.status != DispatchStatus.PENDING.value:
        raise HTTPException(status_code=400, detail="Dispatch is not pending")
    dispatch.status = DispatchStatus.REJECTED.value
    if data.reason:
        dispatch.rationale = (dispatch.rationale or "") + f"\n拒绝原因：{data.reason}"
    db.commit()
    db.refresh(dispatch)
    return dispatch


@router.get("/reservoir-operations/", response_model=List[ReservoirOperationResponse])
def list_reservoir_operations(skip: int = 0, limit: int = 100, reservoir_id: Optional[int] = None,
                              flood_event_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(ReservoirOperation)
    if reservoir_id is not None:
        query = query.filter(ReservoirOperation.reservoir_id == reservoir_id)
    if flood_event_id is not None:
        query = query.filter(ReservoirOperation.flood_event_id == flood_event_id)
    return query.order_by(ReservoirOperation.operated_at.desc()).offset(skip).limit(limit).all()


@router.get("/floods/", response_model=List[FloodEventResponse])
def list_flood_events(skip: int = 0, limit: int = 100, is_active: Optional[bool] = None,
                      watershed_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(FloodEvent)
    if is_active is not None:
        query = query.filter(FloodEvent.is_active == is_active)
    if watershed_id is not None:
        query = query.filter(FloodEvent.watershed_id == watershed_id)
    return query.order_by(FloodEvent.started_at.desc()).offset(skip).limit(limit).all()


@router.get("/floods/{id}", response_model=FloodEventDetail)
def get_flood_event(id: int, db: Session = Depends(get_db)):
    event = db.query(FloodEvent).filter(FloodEvent.id == id).first()
    if not event:
        raise HTTPException(status_code=404, detail="Flood event not found")
    return event


@router.get("/floods/{id}/phases", response_model=List[FloodPhaseResponse])
def get_flood_phases(id: int, db: Session = Depends(get_db)):
    event = db.query(FloodEvent).filter(FloodEvent.id == id).first()
    if not event:
        raise HTTPException(status_code=404, detail="Flood event not found")
    return db.query(FloodPhase).filter(FloodPhase.flood_event_id == id).order_by(FloodPhase.started_at).all()


@router.get("/floods/{id}/dispatches", response_model=List[DispatchRecommendationResponse])
def get_flood_dispatches(id: int, db: Session = Depends(get_db)):
    event = db.query(FloodEvent).filter(FloodEvent.id == id).first()
    if not event:
        raise HTTPException(status_code=404, detail="Flood event not found")
    return db.query(DispatchRecommendation).filter(
        DispatchRecommendation.flood_event_id == id
    ).order_by(DispatchRecommendation.created_at).all()


@router.get("/floods/{id}/hydrology", response_model=List[HydrologicalDataResponse])
def get_flood_hydrology(id: int, db: Session = Depends(get_db)):
    event = db.query(FloodEvent).filter(FloodEvent.id == id).first()
    if not event:
        raise HTTPException(status_code=404, detail="Flood event not found")
    return db.query(HydrologicalData).filter(
        HydrologicalData.flood_event_id == id
    ).order_by(HydrologicalData.recorded_at).all()


@router.post("/floods/{id}/end", response_model=FloodEventResponse)
def end_event(id: int, db: Session = Depends(get_db)):
    try:
        return end_flood_event(db, id)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.post("/scheduler/evaluate")
def trigger_evaluation():
    db = SessionLocal()
    try:
        results = evaluate_flood_situation(db)
        return {"evaluated": len(results), "events": results}
    finally:
        db.close()
