from pydantic import BaseModel
from typing import Optional, List
from datetime import datetime


class ParkBase(BaseModel):
    code: str
    name: str
    max_capacity: int


class ParkCreate(ParkBase):
    pass


class Park(ParkBase):
    id: int
    current_visitors: int = 0
    created_at: datetime

    class Config:
        from_attributes = True


class TicketTypeBase(BaseModel):
    name: str
    base_price: int
    description: Optional[str] = None


class TicketTypeCreate(TicketTypeBase):
    pass


class TicketType(TicketTypeBase):
    id: int
    park_id: int
    is_active: bool = True
    created_at: datetime

    class Config:
        from_attributes = True


class TicketPurchaseRequest(BaseModel):
    visitor_name: Optional[str] = None
    visitor_age: Optional[int] = None
    visitor_height: Optional[float] = None
    is_student: bool = False
    is_group: bool = False
    group_size: int = 1
    free_children_count: int = 0
    route_id: Optional[int] = None


class TicketBase(BaseModel):
    park_code: str
    ticket_type_name: str
    final_price: int
    visitor_name: Optional[str] = None
    visitor_age: Optional[int] = None
    visitor_height: Optional[float] = None
    is_student: bool = False
    is_group: bool = False
    group_size: int = 1
    insurance_fee: int = 0
    status: str = "valid"


class Ticket(TicketBase):
    id: int
    ticket_type_id: int
    purchase_time: datetime
    refund_time: Optional[datetime] = None

    class Config:
        from_attributes = True


class GuideBase(BaseModel):
    guide_id: str
    name: str
    level: str


class GuideCreate(GuideBase):
    pass


class Guide(GuideBase):
    id: int
    park_id: int
    is_available: bool = True
    total_assignments: int = 0
    total_rating: float = 0.0
    rating_count: int = 0
    created_at: datetime

    class Config:
        from_attributes = True


class GuideAssignRequest(BaseModel):
    route_id: Optional[int] = None
    ticket_ids: Optional[List[int]] = None
    duration_hours: float


class GuideRateRequest(BaseModel):
    rating: float
    comment: Optional[str] = None


class RouteBase(BaseModel):
    route_id: str
    name: str
    difficulty: str = "normal"
    capacity: int = 100
    duration_hours: float = 2.0
    base_insurance_fee: int = 0
    description: Optional[str] = None


class RouteCreate(RouteBase):
    pass


class Route(RouteBase):
    id: int
    park_id: int
    current_visitors: int = 0
    is_active: bool = True
    created_at: datetime

    class Config:
        from_attributes = True


class RouteDifficultyUpdate(BaseModel):
    difficulty: str


class RouteCapacityUpdate(BaseModel):
    capacity: int


class PriceInfo(BaseModel):
    base_price: int
    discounts: List[str]
    final_price: int
    insurance_fee: int
    total_price: int
    message: str
