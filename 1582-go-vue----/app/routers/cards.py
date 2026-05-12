from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Card, CardStatus
from ..schemas import CardCreate, CardResponse, CardBalanceUpdate, RechargeResponse
from ..fare import calculate_recharge_bonus

router = APIRouter(prefix="/cards", tags=["票卡管理"])


@router.post("/", response_model=CardResponse, status_code=status.HTTP_201_CREATED)
def create_card(card_data: CardCreate, db: Session = Depends(get_db)):
    existing_card = db.query(Card).filter(Card.card_number == card_data.card_number).first()
    if existing_card:
        raise HTTPException(status_code=400, detail="票卡号已存在")

    card = Card(
        card_number=card_data.card_number,
        card_type=card_data.card_type,
        balance=card_data.initial_balance,
        status=CardStatus.ACTIVE
    )

    db.add(card)
    db.commit()
    db.refresh(card)
    return card


@router.get("/", response_model=List[CardResponse])
def list_cards(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    cards = db.query(Card).offset(skip).limit(limit).all()
    return cards


@router.get("/{card_number}", response_model=CardResponse)
def get_card(card_number: str, db: Session = Depends(get_db)):
    card = db.query(Card).filter(Card.card_number == card_number).first()
    if not card:
        raise HTTPException(status_code=404, detail="票卡不存在")
    return card


@router.post("/{card_number}/recharge", response_model=RechargeResponse)
def recharge_card(
    card_number: str,
    recharge_data: CardBalanceUpdate,
    db: Session = Depends(get_db)
):
    card = db.query(Card).filter(Card.card_number == card_number).first()
    if not card:
        raise HTTPException(status_code=404, detail="票卡不存在")

    if card.status != CardStatus.ACTIVE:
        raise HTTPException(status_code=400, detail=f"票卡状态为{card.status.value}，无法充值")

    amount = recharge_data.amount

    if amount < 1000 or amount > 50000:
        raise HTTPException(status_code=400, detail="充值金额必须在10-500元之间")

    if amount % 1000 != 0:
        raise HTTPException(status_code=400, detail="充值金额必须是10元的整数倍")

    original_balance = card.balance
    bonus = calculate_recharge_bonus(card.card_type, amount)
    total_added = amount + bonus

    card.balance += total_added
    db.commit()
    db.refresh(card)

    message = f"充值成功！充值{amount/100:.0f}元"
    if bonus > 0:
        message += f"，赠送{bonus/100:.0f}元"

    return RechargeResponse(
        success=True,
        card_number=card.card_number,
        original_balance=original_balance,
        recharge_amount=amount,
        bonus_amount=bonus,
        new_balance=card.balance,
        message=message
    )


@router.put("/{card_number}/lock", response_model=CardResponse)
def lock_card(card_number: str, db: Session = Depends(get_db)):
    card = db.query(Card).filter(Card.card_number == card_number).first()
    if not card:
        raise HTTPException(status_code=404, detail="票卡不存在")

    if card.status == CardStatus.LOCKED:
        raise HTTPException(status_code=400, detail="票卡已被锁定")

    card.status = CardStatus.LOCKED
    db.commit()
    db.refresh(card)
    return card


@router.put("/{card_number}/unlock", response_model=CardResponse)
def unlock_card(card_number: str, db: Session = Depends(get_db)):
    card = db.query(Card).filter(Card.card_number == card_number).first()
    if not card:
        raise HTTPException(status_code=404, detail="票卡不存在")

    if card.status != CardStatus.LOCKED:
        raise HTTPException(status_code=400, detail="票卡未被锁定")

    card.status = CardStatus.ACTIVE
    db.commit()
    db.refresh(card)
    return card
