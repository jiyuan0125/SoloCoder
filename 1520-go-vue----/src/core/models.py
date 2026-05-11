from datetime import datetime, date
from typing import Optional, List
from enum import Enum
from pydantic import BaseModel, Field, validator


class WaterType(str, Enum):
    FRESH = "fresh"
    SALT = "salt"
    BRACKISH = "brackish"


class BreedingStage(str, Enum):
    SPAWNING = "spawning"
    INCUBATION = "incubation"
    EMERGENCE = "emergence"
    COMPLETED = "completed"


class TodoPriority(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    URGENT = "urgent"


class TodoStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    RESOLVED = "resolved"


class DeliveryStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class Species(BaseModel):
    id: str
    name: str
    min_temperature: float
    max_temperature: float
    description: Optional[str] = None


class Pond(BaseModel):
    id: str
    name: str
    water_type: WaterType
    capacity: float
    current_temperature: Optional[float] = None
    species_id: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class PondTemperatureLog(BaseModel):
    id: str
    pond_id: str
    temperature: float
    recorded_at: datetime = Field(default_factory=datetime.utcnow)


class BreedingRecord(BaseModel):
    id: str
    cycle_id: str
    stage: BreedingStage
    stage_date: date
    quantity: Optional[int] = None
    notes: Optional[str] = None


class BreedingCycle(BaseModel):
    id: str
    pond_id: str
    species_id: str
    spawning_date: date
    incubation_date: Optional[date] = None
    emergence_date: Optional[date] = None
    expected_quantity: Optional[int] = None
    actual_quantity: Optional[int] = None
    notes: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    @validator('incubation_date')
    def check_incubation_after_spawning(cls, v, values):
        if v and 'spawning_date' in values and v < values['spawning_date']:
            raise ValueError('孵化日期不能早于产卵日期')
        return v

    @validator('emergence_date')
    def check_emergence_after_spawning(cls, v, values):
        if v and 'spawning_date' in values and v < values['spawning_date']:
            raise ValueError('出苗日期不能早于产卵日期')
        return v


class InventoryItem(BaseModel):
    id: str
    species_id: str
    quantity: int
    pond_id: Optional[str] = None
    stage: BreedingStage = BreedingStage.EMERGENCE
    last_updated: datetime = Field(default_factory=datetime.utcnow)


class SalesOrder(BaseModel):
    id: str
    customer_name: str
    customer_phone: Optional[str] = None
    species_id: str
    requested_quantity: int
    approved_quantity: Optional[int] = None
    unit_price: float
    total_amount: Optional[float] = None
    notes: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    delivery_task_id: Optional[str] = None


class DeliveryVehicle(BaseModel):
    id: str
    license_plate: str
    name: str
    capacity: Optional[int] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)


class DeliveryTask(BaseModel):
    id: str
    order_id: str
    vehicle_id: str
    driver_name: str
    driver_phone: Optional[str] = None
    delivery_address: str
    scheduled_start: datetime
    scheduled_end: Optional[datetime] = None
    actual_start: Optional[datetime] = None
    actual_end: Optional[datetime] = None
    status: DeliveryStatus = DeliveryStatus.PENDING
    notes: Optional[str] = None


class TodoItem(BaseModel):
    id: str
    title: str
    description: str
    priority: TodoPriority = TodoPriority.MEDIUM
    status: TodoStatus = TodoStatus.PENDING
    category: Optional[str] = None
    related_entity_id: Optional[str] = None
    related_entity_type: Optional[str] = None
    consecutive_violations: int = 0
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    resolved_at: Optional[datetime] = None
