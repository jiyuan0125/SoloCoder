from datetime import datetime, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Card, CardStatus, CardType, Transaction
from ..schemas import EntryRecord, ExitRecord, TransactionResponse, FareCalculation
from ..fare import calculate_fare

router = APIRouter(prefix="/transactions", tags=["闸机通行"])


@router.post("/entry", response_model=TransactionResponse)
def record_entry(entry: EntryRecord, db: Session = Depends(get_db)):
    card = db.query(Card).filter(Card.card_number == entry.card_number).first()
    if not card:
        raise HTTPException(status_code=404, detail="票卡不存在")

    if card.status != CardStatus.ACTIVE:
        raise HTTPException(status_code=400, detail=f"票卡状态为{card.status.value}，无法通行")

    existing_open = db.query(Transaction).filter(
        Transaction.card_number == entry.card_number,
        Transaction.exit_time.is_(None),
        Transaction.entry_time.isnot(None)
    ).first()

    if existing_open:
        raise HTTPException(status_code=400, detail="该票卡已有未完成的通行记录")

    transaction = Transaction(
        card_id=card.id,
        card_number=entry.card_number,
        entry_station=entry.station,
        entry_time=entry.entry_time
    )

    db.add(transaction)
    db.commit()
    db.refresh(transaction)
    return transaction


@router.post("/exit", response_model=TransactionResponse)
def record_exit(exit_record: ExitRecord, db: Session = Depends(get_db)):
    card = db.query(Card).filter(Card.card_number == exit_record.card_number).first()
    if not card:
        raise HTTPException(status_code=404, detail="票卡不存在")

    if card.status != CardStatus.ACTIVE and card.status != CardStatus.LOCKED:
        raise HTTPException(status_code=400, detail=f"票卡状态为{card.status.value}，无法出站")

    transaction = db.query(Transaction).filter(
        Transaction.card_number == exit_record.card_number,
        Transaction.exit_time.is_(None),
        Transaction.entry_time.isnot(None)
    ).first()

    if not transaction:
        raise HTTPException(status_code=404, detail="未找到对应的进站记录")

    time_diff = (exit_record.exit_time - transaction.entry_time).total_seconds()
    is_misoperation = time_diff < 120

    if is_misoperation:
        final_fare = 0
        base_fare = 0
        discount = "误入(2分钟内)"
    else:
        fare_calc = calculate_fare(exit_record.distance, card.card_type)
        base_fare = fare_calc.base_fare
        final_fare = fare_calc.final_fare
        discount = fare_calc.discount

    existing_same = db.query(Transaction).filter(
        Transaction.card_number == exit_record.card_number,
        Transaction.entry_time == transaction.entry_time,
        Transaction.exit_time == exit_record.exit_time
    ).first()

    if existing_same:
        return existing_same

    transaction.exit_station = exit_record.station
    transaction.exit_time = exit_record.exit_time
    transaction.distance = exit_record.distance
    transaction.base_fare = base_fare
    transaction.discount = discount
    transaction.final_fare = final_fare

    if not is_misoperation and card.status == CardStatus.ACTIVE:
        card.balance -= final_fare

        if card.card_type == CardType.ELDERLY and card.balance < 0:
            card.status = CardStatus.LOCKED

        if card.card_type == CardType.SINGLE:
            card.status = CardStatus.CONSUMED

    db.commit()
    db.refresh(transaction)
    db.refresh(card)

    return transaction


@router.get("/", response_model=List[TransactionResponse])
def list_transactions(
    card_number: Optional[str] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Transaction)
    if card_number:
        query = query.filter(Transaction.card_number == card_number)
    transactions = query.order_by(Transaction.id.desc()).offset(skip).limit(limit).all()
    return transactions


@router.get("/{transaction_id}", response_model=TransactionResponse)
def get_transaction(transaction_id: int, db: Session = Depends(get_db)):
    transaction = db.query(Transaction).filter(Transaction.id == transaction_id).first()
    if not transaction:
        raise HTTPException(status_code=404, detail="通行记录不存在")
    return transaction


@router.post("/calculate-fare", response_model=FareCalculation)
def test_fare_calculation(distance: float, card_type: CardType):
    return calculate_fare(distance, card_type)
