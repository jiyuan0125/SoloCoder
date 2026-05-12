from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field
from app.models.performance import PerformanceStatus, TicketStatus


class PerformanceBase(BaseModel):
    name: str = Field(..., max_length=200)
    description: Optional[str] = Field(None, max_length=500)
    start_time: datetime
    end_time: datetime
    venue_id: int
    total_tickets: int = Field(..., ge=1)
    price: int = Field(..., ge=0)
    status: PerformanceStatus = PerformanceStatus.PLANNED


class PerformanceCreate(PerformanceBase):
    pass


class PerformanceUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    venue_id: Optional[int] = None
    total_tickets: Optional[int] = None
    price: Optional[int] = None
    status: Optional[PerformanceStatus] = None


class PerformanceOut(PerformanceBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class TicketBase(BaseModel):
    customer_name: str = Field(..., max_length=100)
    customer_phone: Optional[str] = Field(None, max_length=20)
    seat_number: Optional[str] = Field(None, max_length=20)


class TicketCreate(TicketBase):
    performance_id: int


class TicketOut(TicketBase):
    id: int
    performance_id: int
    price: int
    status: TicketStatus
    sold_at: Optional[datetime]
    issued_at: Optional[datetime]
    used_at: Optional[datetime]
    refunded_at: Optional[datetime]
    refund_amount: Optional[int]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RefundResponse(BaseModel):
    ticket_id: int
    refund_amount: int
    message: str
