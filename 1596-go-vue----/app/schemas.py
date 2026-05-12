from datetime import datetime, date, time
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field

from app.models import (
    BookingType, PaymentStatus, BookingStatus, ChargeType,
    MatchStatus, TournamentStatus, ClassStatus
)


class VenueTypeBase(BaseModel):
    name: str
    charge_type: ChargeType
    price: float
    description: Optional[str] = None


class VenueTypeCreate(VenueTypeBase):
    pass


class VenueTypeResponse(VenueTypeBase):
    id: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class VenueBase(BaseModel):
    name: str
    venue_type_id: int
    capacity: int = 10
    is_active: bool = True


class VenueCreate(VenueBase):
    pass


class VenueResponse(VenueBase):
    id: int
    venue_type: Optional[VenueTypeResponse] = None
    created_at: datetime
    
    class Config:
        from_attributes = True


class UserBase(BaseModel):
    username: str
    phone: str


class UserCreate(UserBase):
    is_admin: bool = False


class UserResponse(UserBase):
    id: int
    is_admin: bool
    created_at: datetime
    
    class Config:
        from_attributes = True


class BookingSlot(BaseModel):
    start_time: time
    end_time: time


class BookingCreate(BaseModel):
    user_id: int
    venue_id: int
    booking_date: date
    slots: List[BookingSlot]
    notes: Optional[str] = None


class PaymentCreate(BaseModel):
    booking_id: int
    payment_method: str = "cash"
    transaction_id: Optional[str] = None


class PaymentResponse(BaseModel):
    id: int
    payment_no: str
    booking_id: int
    amount: float
    status: PaymentStatus
    payment_method: Optional[str] = None
    created_at: datetime
    paid_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True


class RefundResponse(BaseModel):
    id: int
    refund_no: str
    booking_id: int
    refund_amount: float
    refund_rate: float
    reason: Optional[str] = None
    status: PaymentStatus
    created_at: datetime
    
    class Config:
        from_attributes = True


class BookingResponse(BaseModel):
    id: int
    booking_no: str
    user_id: int
    venue_id: int
    booking_type: BookingType
    booking_date: date
    start_time: time
    end_time: time
    hours: float
    original_amount: float
    discount_amount: float
    final_amount: float
    status: BookingStatus
    is_continuous: bool
    notes: Optional[str] = None
    created_at: datetime
    paid_at: Optional[datetime] = None
    payment: Optional[PaymentResponse] = None
    refund: Optional[RefundResponse] = None
    
    class Config:
        from_attributes = True


class TeamBase(BaseModel):
    name: str
    captain_name: Optional[str] = None
    phone: Optional[str] = None


class TeamCreate(TeamBase):
    tournament_id: int


class TeamResponse(TeamBase):
    id: int
    tournament_id: int
    is_forfeited: bool
    is_withdrawn: bool
    
    class Config:
        from_attributes = True


class TournamentBase(BaseModel):
    name: str
    venue_type_id: int
    start_date: date
    end_date: date
    description: Optional[str] = None


class TournamentCreate(TournamentBase):
    pass


class TournamentResponse(TournamentBase):
    id: int
    status: TournamentStatus
    teams: List[TeamResponse] = []
    created_at: datetime
    
    class Config:
        from_attributes = True


class MatchBase(BaseModel):
    tournament_id: int
    match_no: str
    round_no: int = 1
    home_team_id: Optional[int] = None
    away_team_id: Optional[int] = None
    match_date: date
    start_time: time
    end_time: time
    home_score: Optional[int] = None
    away_score: Optional[int] = None


class MatchCreate(MatchBase):
    pass


class MatchUpdateScore(BaseModel):
    home_score: int
    away_score: int


class MatchResponse(MatchBase):
    id: int
    status: MatchStatus
    winner_id: Optional[int] = None
    is_forfeit: bool = False
    booking_id: Optional[int] = None
    
    class Config:
        from_attributes = True


class TeamStatisticsResponse(BaseModel):
    id: int
    tournament_id: int
    team_id: int
    team_name: str
    matches_played: Optional[int] = None
    wins: Optional[int] = None
    losses: Optional[int] = None
    draws: Optional[int] = None
    goals_for: Optional[int] = None
    goals_against: Optional[int] = None
    goal_difference: Optional[int] = None
    points: Optional[int] = None
    rank: Optional[int] = None
    
    class Config:
        from_attributes = True


class TrainingClassBase(BaseModel):
    name: str
    instructor: Optional[str] = None
    venue_type_id: int
    start_date: date
    end_date: date
    class_time: time
    duration_hours: float = 2.0
    min_students: int = 5
    max_students: int = 20
    price_per_student: float
    description: Optional[str] = None


class TrainingClassCreate(TrainingClassBase):
    pass


class ClassRegistrationBase(BaseModel):
    class_id: int
    user_id: int
    student_name: str
    phone: Optional[str] = None


class ClassRegistrationCreate(ClassRegistrationBase):
    pass


class ClassRegistrationResponse(BaseModel):
    id: int
    registration_no: str
    class_id: int
    user_id: int
    student_name: str
    phone: Optional[str] = None
    amount_paid: float
    is_refunded: bool
    refund_amount: float
    registered_at: datetime
    
    class Config:
        from_attributes = True


class TrainingClassResponse(TrainingClassBase):
    id: int
    class_no: str
    status: ClassStatus
    registrations: List[ClassRegistrationResponse] = []
    created_at: datetime
    
    class Config:
        from_attributes = True


class MaintenanceCreate(BaseModel):
    venue_id: int
    start_time: datetime
    end_time: datetime
    reason: Optional[str] = None


class MaintenanceResponse(BaseModel):
    id: int
    venue_id: int
    start_time: datetime
    end_time: datetime
    reason: Optional[str] = None
    created_at: datetime
    
    class Config:
        from_attributes = True


class QRVerifyResponse(BaseModel):
    valid: bool
    booking_no: Optional[str] = None
    venue_name: Optional[str] = None
    booking_date: Optional[date] = None
    start_time: Optional[time] = None
    end_time: Optional[time] = None
    status: Optional[str] = None
    message: str


class VenueAvailability(BaseModel):
    venue_id: int
    venue_name: str
    available_slots: List[Dict[str, Any]]
