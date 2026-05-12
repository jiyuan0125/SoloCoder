from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List
from models import FlightStatus, BerthStatus, ShipStatus, TicketStatus


class TerminalBase(BaseModel):
    name: str
    location: Optional[str] = None
    max_berths: int = 1


class TerminalCreate(TerminalBase):
    pass


class TerminalUpdate(BaseModel):
    name: Optional[str] = None
    location: Optional[str] = None
    max_berths: Optional[int] = None


class TerminalResponse(TerminalBase):
    id: int
    
    class Config:
        from_attributes = True


class BerthBase(BaseModel):
    terminal_id: int
    berth_number: str
    capacity: int = 1


class BerthCreate(BerthBase):
    pass


class BerthUpdate(BaseModel):
    berth_number: Optional[str] = None
    capacity: Optional[int] = None
    status: Optional[BerthStatus] = None


class BerthResponse(BerthBase):
    id: int
    status: BerthStatus
    
    class Config:
        from_attributes = True


class ShipBase(BaseModel):
    name: str
    passenger_capacity: int
    vehicle_capacity: int = 0


class ShipCreate(ShipBase):
    pass


class ShipUpdate(BaseModel):
    name: Optional[str] = None
    passenger_capacity: Optional[int] = None
    vehicle_capacity: Optional[int] = None
    status: Optional[ShipStatus] = None
    inspection_start_date: Optional[datetime] = None
    inspection_end_date: Optional[datetime] = None


class ShipResponse(ShipBase):
    id: int
    status: ShipStatus
    
    class Config:
        from_attributes = True


class FlightBase(BaseModel):
    flight_number: str
    departure_terminal_id: int
    arrival_terminal_id: int
    scheduled_departure: datetime
    scheduled_arrival: datetime
    estimated_passengers: int = 0


class FlightCreate(FlightBase):
    ship_id: Optional[int] = None
    berth_id: Optional[int] = None


class FlightUpdate(BaseModel):
    status: Optional[FlightStatus] = None
    actual_passengers: Optional[int] = None
    actual_vehicles: Optional[int] = None
    delay_minutes: Optional[int] = None


class FlightResponse(FlightBase):
    id: int
    status: FlightStatus
    ship_id: Optional[int] = None
    berth_id: Optional[int] = None
    actual_departure: Optional[datetime] = None
    actual_arrival: Optional[datetime] = None
    
    class Config:
        from_attributes = True


class TicketBase(BaseModel):
    ticket_number: str
    flight_id: int
    passenger_name: Optional[str] = None
    passenger_count: int = 1
    vehicle_count: int = 0
    is_on_site: bool = False
    amount: float = 0.0


class TicketCreate(TicketBase):
    pass


class TicketResponse(TicketBase):
    id: int
    status: TicketStatus
    
    class Config:
        from_attributes = True


class WeatherBase(BaseModel):
    terminal_id: int
    wind_speed: float
    visibility: int
    weather_condition: Optional[str] = None


class WeatherCreate(WeatherBase):
    pass


class WeatherResponse(WeatherBase):
    id: int
    timestamp: datetime
    
    class Config:
        from_attributes = True


class RefundResponse(BaseModel):
    id: int
    ticket_id: int
    refund_amount: float
    reason: Optional[str] = None
    created_at: datetime
    fee_waived: bool
    
    class Config:
        from_attributes = True
