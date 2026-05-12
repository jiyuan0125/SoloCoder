from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class VenueBase(BaseModel):
    name: str = Field(..., max_length=100)
    capacity: int = Field(..., ge=1)
    hourly_rate: int = Field(..., ge=0)
    half_day_rate: int = Field(..., ge=0)
    full_day_rate: int = Field(..., ge=0)
    description: Optional[str] = Field(None, max_length=300)


class VenueCreate(VenueBase):
    pass


class VenueUpdate(BaseModel):
    name: Optional[str] = None
    capacity: Optional[int] = None
    hourly_rate: Optional[int] = None
    half_day_rate: Optional[int] = None
    full_day_rate: Optional[int] = None
    description: Optional[str] = None


class VenueOut(VenueBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RentalBase(BaseModel):
    venue_id: int
    customer_name: str = Field(..., max_length=100)
    customer_phone: Optional[str] = Field(None, max_length=20)
    purpose: Optional[str] = Field(None, max_length=200)
    start_time: datetime
    end_time: datetime


class RentalCreate(RentalBase):
    pass


class RentalOut(RentalBase):
    id: int
    total_amount: int
    created_at: datetime

    class Config:
        from_attributes = True
