from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field

from .models import CardType, CardStatus, Priority, FaultStatus


class CardBase(BaseModel):
    card_number: str = Field(..., min_length=1, max_length=20)
    card_type: CardType


class CardCreate(CardBase):
    initial_balance: int = Field(default=0, ge=0)


class CardResponse(BaseModel):
    id: int
    card_number: str
    card_type: CardType
    balance: int
    status: CardStatus
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CardBalanceUpdate(BaseModel):
    amount: int = Field(..., ge=1000, le=50000)


class RechargeResponse(BaseModel):
    success: bool
    card_number: str
    original_balance: int
    recharge_amount: int
    bonus_amount: int
    new_balance: int
    message: str


class EntryRecord(BaseModel):
    card_number: str
    station: str
    entry_time: datetime


class ExitRecord(BaseModel):
    card_number: str
    station: str
    exit_time: datetime
    distance: float = Field(..., gt=0)


class TransactionResponse(BaseModel):
    id: int
    card_number: str
    entry_station: Optional[str]
    exit_station: Optional[str]
    entry_time: Optional[datetime]
    exit_time: Optional[datetime]
    distance: float
    base_fare: int
    discount: str
    final_fare: int
    created_at: datetime

    class Config:
        from_attributes = True


class FareCalculation(BaseModel):
    distance: float
    base_fare: int
    card_type: CardType
    discount: str
    final_fare: int


class FaultCreate(BaseModel):
    gate_id: str
    station: str
    description: str
    priority: Priority = Priority.MEDIUM


class FaultResponse(BaseModel):
    id: int
    gate_id: str
    station: str
    description: str
    priority: Priority
    status: FaultStatus
    reported_at: datetime
    resolved_at: Optional[datetime]
    escalated: bool

    class Config:
        from_attributes = True


class FaultUpdate(BaseModel):
    status: FaultStatus
