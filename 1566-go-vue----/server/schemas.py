from pydantic import BaseModel, Field, field_validator
from datetime import datetime
from typing import Optional, List
from enum import Enum

class RouteStatus(str, Enum):
    PLANNING = "planning"
    TRIAL = "trial"
    FORMAL = "formal"
    PAUSED = "paused"
    TERMINATED = "terminated"

class SuggestionType(str, Enum):
    INCREASE_FREQUENCY = "increase_frequency"
    DECREASE_FREQUENCY = "decrease_frequency"
    SUSPEND = "suspend"
    LOSS_WARNING = "loss_warning"

class RouteCreate(BaseModel):
    route_code: str
    origin_city: str
    destination_city: str
    aircraft_type: Optional[str] = None
    seating_capacity: Optional[int] = None
    weekly_frequency: Optional[int] = None
    base_fuel_consumption: Optional[int] = None

class RouteResponse(BaseModel):
    id: int
    route_code: str
    origin_city: str
    destination_city: str
    status: str
    aircraft_type: Optional[str] = None
    seating_capacity: Optional[int] = None
    weekly_frequency: Optional[int] = None
    base_fuel_consumption: Optional[int] = None
    planning_date: Optional[datetime] = None
    trial_start_date: Optional[datetime] = None
    formal_start_date: Optional[datetime] = None
    pause_date: Optional[datetime] = None
    termination_date: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

class FlightCreate(BaseModel):
    flight_number: str
    departure_datetime: datetime
    arrival_datetime: datetime
    total_seats: int

class FlightResponse(BaseModel):
    id: int
    route_id: int
    flight_number: str
    departure_datetime: datetime
    arrival_datetime: datetime
    status: str
    total_seats: int
    booked_seats: int
    load_factor: Optional[float] = None

    class Config:
        from_attributes = True

class FuelPriceCreate(BaseModel):
    year: int
    month: int
    price_per_liter_cents: int

class FuelPriceResponse(BaseModel):
    id: int
    year: int
    month: int
    price_per_liter_cents: int
    price_per_liter_yuan: float

    class Config:
        from_attributes = True

class RevenueDataResponse(BaseModel):
    id: int
    route_id: int
    year: int
    month: int
    total_flights: int
    total_booked_seats: int
    total_available_seats: int
    passenger_revenue_cents: int
    fuel_cost_cents: int
    other_costs_cents: int
    route_status_at_month: str
    total_revenue_cents: int
    total_cost_cents: int
    net_profit_cents: int
    load_factor: float
    average_ticket_price: float
    passenger_revenue_yuan: int
    fuel_cost_yuan: int
    other_costs_yuan: int
    total_revenue_yuan: int
    total_cost_yuan: int
    net_profit_yuan: int

    class Config:
        from_attributes = True

    @field_validator('passenger_revenue_yuan', mode='before')
    @classmethod
    def calculate_passenger_revenue_yuan(cls, v, values):
        if isinstance(v, int):
            return v
        if 'passenger_revenue_cents' in values.data:
            return round(values.data['passenger_revenue_cents'] / 100)
        return 0

    @field_validator('fuel_cost_yuan', mode='before')
    @classmethod
    def calculate_fuel_cost_yuan(cls, v, values):
        if isinstance(v, int):
            return v
        if 'fuel_cost_cents' in values.data:
            return round(values.data['fuel_cost_cents'] / 100)
        return 0

    @field_validator('other_costs_yuan', mode='before')
    @classmethod
    def calculate_other_costs_yuan(cls, v, values):
        if isinstance(v, int):
            return v
        if 'other_costs_cents' in values.data:
            return round(values.data['other_costs_cents'] / 100)
        return 0

    @field_validator('total_revenue_yuan', mode='before')
    @classmethod
    def calculate_total_revenue_yuan(cls, v, values):
        if isinstance(v, int):
            return v
        if 'total_revenue_cents' in values.data:
            return round(values.data['total_revenue_cents'] / 100)
        return 0

    @field_validator('total_cost_yuan', mode='before')
    @classmethod
    def calculate_total_cost_yuan(cls, v, values):
        if isinstance(v, int):
            return v
        if 'total_cost_cents' in values.data:
            return round(values.data['total_cost_cents'] / 100)
        return 0

    @field_validator('net_profit_yuan', mode='before')
    @classmethod
    def calculate_net_profit_yuan(cls, v, values):
        if isinstance(v, int):
            return v
        if 'net_profit_cents' in values.data:
            return round(values.data['net_profit_cents'] / 100)
        return 0

class RevenueDataCreate(BaseModel):
    total_flights: int
    total_booked_seats: int
    total_available_seats: int
    passenger_revenue_cents: int
    other_costs_cents: int

class OptimizationSuggestionResponse(BaseModel):
    id: int
    route_id: int
    suggestion_type: str
    reason: str
    generated_at: datetime
    is_implemented: bool
    implemented_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class RouteDetailResponse(BaseModel):
    route: RouteResponse
    flights: List[FlightResponse] = []
    revenue_data: List[RevenueDataResponse] = []
    suggestions: List[OptimizationSuggestionResponse] = []
