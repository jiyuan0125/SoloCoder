from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from src.core.database import get_db
from src.core.models import WasteLedger, Unit
from src.core.schemas import WasteLedgerCreate, WasteLedgerUpdate, WasteLedgerResponse
from src.core.enums import WasteType, UnitType

router = APIRouter()


@router.post("", response_model=WasteLedgerResponse, status_code=status.HTTP_201_CREATED)
def create_ledger(ledger_data: WasteLedgerCreate, db: Session = Depends(get_db)):
    producer = db.query(Unit).filter(
        Unit.id == ledger_data.producer_id,
        Unit.unit_type == UnitType.PRODUCER
    ).first()
    if not producer:
        raise HTTPException(status_code=400, detail="产废单位不存在或不是产废单位")

    if ledger_data.waste_type == WasteType.HAZARDOUS and not ledger_data.hw_code:
        raise HTTPException(status_code=400, detail="危废必须提供HW编号")

    ledger = WasteLedger(**ledger_data.model_dump())
    db.add(ledger)
    db.commit()
    db.refresh(ledger)
    return ledger


@router.get("", response_model=List[WasteLedgerResponse])
def list_ledgers(
    producer_id: int = None,
    waste_type: WasteType = None,
    hw_code: str = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(WasteLedger)
    if producer_id:
        query = query.filter(WasteLedger.producer_id == producer_id)
    if waste_type:
        query = query.filter(WasteLedger.waste_type == waste_type.value)
    if hw_code:
        query = query.filter(WasteLedger.hw_code == hw_code)
    return query.offset(skip).limit(limit).all()


@router.get("/{ledger_id}", response_model=WasteLedgerResponse)
def get_ledger(ledger_id: int, db: Session = Depends(get_db)):
    ledger = db.query(WasteLedger).filter(WasteLedger.id == ledger_id).first()
    if not ledger:
        raise HTTPException(status_code=404, detail="台账不存在")
    return ledger


@router.put("/{ledger_id}", response_model=WasteLedgerResponse)
def update_ledger(
    ledger_id: int,
    ledger_data: WasteLedgerUpdate,
    db: Session = Depends(get_db)
):
    ledger = db.query(WasteLedger).filter(WasteLedger.id == ledger_id).first()
    if not ledger:
        raise HTTPException(status_code=404, detail="台账不存在")

    update_data = ledger_data.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(ledger, key, value)

    db.commit()
    db.refresh(ledger)
    return ledger


@router.delete("/{ledger_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_ledger(ledger_id: int, db: Session = Depends(get_db)):
    ledger = db.query(WasteLedger).filter(WasteLedger.id == ledger_id).first()
    if not ledger:
        raise HTTPException(status_code=404, detail="台账不存在")
    db.delete(ledger)
    db.commit()
