from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from .. import models, schemas, services
from ..database import get_db

router = APIRouter()


@router.post("/", response_model=schemas.OperationRecord)
def create_operation_record(
    record: schemas.OperationRecordCreate, db: Session = Depends(get_db)
):
    vessel = (
        db.query(models.Vessel)
        .filter(models.Vessel.id == record.vessel_id)
        .first()
    )
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    db_record = models.OperationRecord(**record.model_dump())
    db.add(db_record)
    db.commit()
    db.refresh(db_record)

    services.check_illegal_operation(db, record.vessel_id)

    return db_record


@router.get("/", response_model=List[schemas.OperationRecord])
def list_operation_records(
    vessel_id: Optional[int] = Query(None),
    status: Optional[str] = Query(None),
    is_illegal: Optional[bool] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(models.OperationRecord)

    if vessel_id:
        query = query.filter(models.OperationRecord.vessel_id == vessel_id)
    if status:
        query = query.filter(models.OperationRecord.status == status)
    if is_illegal is not None:
        query = query.filter(models.OperationRecord.is_illegal == is_illegal)

    records = query.order_by(models.OperationRecord.start_date.desc()).offset(skip).limit(limit).all()
    return records


@router.get("/{record_id}", response_model=schemas.OperationRecord)
def get_operation_record(record_id: int, db: Session = Depends(get_db)):
    record = (
        db.query(models.OperationRecord)
        .filter(models.OperationRecord.id == record_id)
        .first()
    )
    if not record:
        raise HTTPException(status_code=404, detail="营运记录不存在")
    return record


@router.put("/{record_id}", response_model=schemas.OperationRecord)
def update_operation_record(
    record_id: int,
    record_update: schemas.OperationRecordUpdate,
    db: Session = Depends(get_db),
):
    record = (
        db.query(models.OperationRecord)
        .filter(models.OperationRecord.id == record_id)
        .first()
    )
    if not record:
        raise HTTPException(status_code=404, detail="营运记录不存在")

    update_data = record_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(record, key, value)

    db.commit()
    db.refresh(record)

    services.check_illegal_operation(db, record.vessel_id)

    return record


@router.get("/check-illegal/{vessel_id}")
def check_vessel_illegal_operation(vessel_id: int, db: Session = Depends(get_db)):
    vessel = (
        db.query(models.Vessel)
        .filter(models.Vessel.id == vessel_id)
        .first()
    )
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    is_illegal = services.check_illegal_operation(db, vessel_id)
    return {"vessel_id": vessel_id, "is_illegal": is_illegal}
