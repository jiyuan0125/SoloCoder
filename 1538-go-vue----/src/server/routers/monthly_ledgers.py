from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from src.core.database import get_db
from src.core.models import MonthlyLedger
from src.core.schemas import MonthlyLedgerCalculate, MonthlyLedgerResponse, MonthlyLedgerBase
from src.core.services import calculate_monthly_balance

router = APIRouter()


@router.post("/calculate", response_model=MonthlyLedgerResponse)
def calculate_ledger(
    calc_data: MonthlyLedgerCalculate,
    db: Session = Depends(get_db)
):
    return calculate_monthly_balance(
        db=db,
        unit_id=calc_data.unit_id,
        year=calc_data.year,
        month=calc_data.month
    )


@router.post("/update-closing", response_model=MonthlyLedgerResponse)
def update_closing_balance(
    ledger_data: MonthlyLedgerBase,
    db: Session = Depends(get_db)
):
    ledger = calculate_monthly_balance(
        db=db,
        unit_id=ledger_data.unit_id,
        year=ledger_data.year,
        month=ledger_data.month
    )

    if ledger_data.closing_balance is not None:
        from src.core.config import MONTHLY_BALANCE_TOLERANCE
        ledger.closing_balance = ledger_data.closing_balance
        ledger.difference = abs(ledger.closing_balance - ledger.calculated_closing)
        base = max(abs(ledger.calculated_closing), 1e-9)
        ledger.is_approved = (ledger.difference / base) <= MONTHLY_BALANCE_TOLERANCE
        db.commit()
        db.refresh(ledger)

        if not ledger.is_approved:
            from src.core.services import check_consecutive_failures
            check_consecutive_failures(db, ledger.unit_id, ledger.year, ledger.month)

    return ledger


@router.get("", response_model=List[MonthlyLedgerResponse])
def list_monthly_ledgers(
    unit_id: int = None,
    year: int = None,
    month: int = None,
    is_approved: bool = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(MonthlyLedger)
    if unit_id:
        query = query.filter(MonthlyLedger.unit_id == unit_id)
    if year:
        query = query.filter(MonthlyLedger.year == year)
    if month:
        query = query.filter(MonthlyLedger.month == month)
    if is_approved is not None:
        query = query.filter(MonthlyLedger.is_approved == is_approved)
    return query.offset(skip).limit(limit).all()


@router.get("/{ledger_id}", response_model=MonthlyLedgerResponse)
def get_monthly_ledger(ledger_id: int, db: Session = Depends(get_db)):
    ledger = db.query(MonthlyLedger).filter(MonthlyLedger.id == ledger_id).first()
    if not ledger:
        raise HTTPException(status_code=404, detail="月度台账不存在")
    return ledger
