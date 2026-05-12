from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from .. import schemas, crud, services
from ..database import get_db

router = APIRouter(prefix="/signals", tags=["signals"])


@router.get("/", response_model=List[schemas.SignalResponse])
def read_signals(station_id: Optional[int] = None, db: Session = Depends(get_db)):
    return crud.get_signals(db, station_id=station_id)


@router.get("/{signal_id}", response_model=schemas.SignalResponse)
def read_signal(signal_id: int, db: Session = Depends(get_db)):
    db_signal = crud.get_signal(db, signal_id=signal_id)
    if db_signal is None:
        raise HTTPException(status_code=404, detail="Signal not found")
    return db_signal


@router.post("/", response_model=schemas.SignalResponse, status_code=status.HTTP_201_CREATED)
def create_signal(signal: schemas.SignalCreate, db: Session = Depends(get_db)):
    db_signal = crud.get_signal_by_code(db, signal_code=signal.signal_code)
    if db_signal:
        raise HTTPException(status_code=400, detail="Signal with this code already exists")
    db_station = crud.get_station(db, station_id=signal.station_id)
    if db_station is None:
        raise HTTPException(status_code=404, detail="Station not found")
    return crud.create_signal(db=db, signal=signal)


@router.put("/{signal_id}", response_model=schemas.SignalResponse)
def update_signal(signal_id: int, signal: schemas.SignalUpdate, db: Session = Depends(get_db)):
    db_signal = crud.get_signal(db, signal_id=signal_id)
    if db_signal is None:
        raise HTTPException(status_code=404, detail="Signal not found")
    return crud.update_signal(db=db, signal_id=signal_id, signal=signal)


@router.post("/{signal_id}/fault", response_model=schemas.SignalFaultReport)
def report_signal_fault(
    signal_id: int,
    description: Optional[str] = None,
    db: Session = Depends(get_db),
):
    signal_service = services.get_signal_service(db)
    try:
        signal, report = signal_service.report_signal_fault(signal_id, description)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    section = None
    if signal.section_id:
        section = crud.get_power_section(db, section_id=signal.section_id)

    return schemas.SignalFaultReport(
        signal_id=signal.id,
        signal_code=signal.signal_code,
        fault_time=signal.fault_time,
        fault_description=signal.fault_description,
        affected_section_id=signal.section_id,
        affected_section_name=section.name if section else None,
    )


@router.post("/{signal_id}/confirm-recovery", response_model=schemas.SignalResponse)
def confirm_signal_recovery(
    signal_id: int,
    operator: str,
    db: Session = Depends(get_db),
):
    signal_service = services.get_signal_service(db)
    try:
        signal = signal_service.confirm_signal_recovery(signal_id, operator)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    return signal


@router.post("/{signal_id}/restore-normal", response_model=schemas.SignalResponse)
def restore_signal_to_normal(
    signal_id: int,
    db: Session = Depends(get_db),
):
    signal_service = services.get_signal_service(db)
    try:
        signal = signal_service.restore_signal_to_normal(signal_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    return signal


@router.get("/{signal_id}/fault-reports", response_model=List[schemas.FaultReportResponse])
def read_signal_fault_reports(signal_id: int, db: Session = Depends(get_db)):
    db_signal = crud.get_signal(db, signal_id=signal_id)
    if db_signal is None:
        raise HTTPException(status_code=404, detail="Signal not found")
    return crud.get_fault_reports_by_signal(db, signal_id=signal_id)
