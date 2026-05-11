from datetime import datetime
from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field, validator


class FurnaceStatus(str, Enum):
    IDLE = "idle"
    HEATING = "heating"
    MELTING = "melting"
    COOLING = "cooling"
    MAINTENANCE = "maintenance"


class Material(BaseModel):
    name: str
    weight: float
    actual_weight: Optional[float] = None


class BatchingOrderStatus(str, Enum):
    PENDING = "pending"
    ASSIGNED = "assigned"
    SMELTING = "smelting"
    COMPLETED = "completed"
    ABNORMAL = "abnormal"


class Furnace(BaseModel):
    id: str
    name: str
    design_capacity: float
    min_temperature: float
    max_temperature: float
    status: FurnaceStatus = FurnaceStatus.IDLE
    current_smelting_id: Optional[str] = None


class BatchingOrder(BaseModel):
    id: str
    furnace_id: Optional[str] = None
    materials: List[Material]
    status: BatchingOrderStatus = BatchingOrderStatus.PENDING
    created_at: datetime = Field(default_factory=datetime.now)
    assigned_at: Optional[datetime] = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    output_weight: Optional[float] = None
    energy_consumed: Optional[float] = None
    is_abnormal: bool = False
    abnormal_reason: Optional[str] = None
    
    @validator('materials')
    def check_materials(cls, v):
        if not v:
            raise ValueError('配料单必须至少包含一种原料')
        return v
    
    def get_total_weight(self) -> float:
        return sum(m.weight for m in self.materials)
    
    def get_total_actual_weight(self) -> Optional[float]:
        if all(m.actual_weight is not None for m in self.materials):
            return sum(m.actual_weight for m in self.materials)
        return None
    
    def check_deviation(self) -> bool:
        actual_total = self.get_total_actual_weight()
        if actual_total is None:
            return False
        expected_total = self.get_total_weight()
        if expected_total == 0:
            return False
        deviation = abs(actual_total - expected_total) / expected_total
        return deviation > 0.1


class TemperatureReading(BaseModel):
    id: str
    furnace_id: str
    temperature: float
    timestamp: datetime = Field(default_factory=datetime.now)
    is_alert: bool = False


class AlertType(str, Enum):
    TEMPERATURE_HIGH = "temperature_high"
    TEMPERATURE_LOW = "temperature_low"
    EMERGENCY = "emergency"


class Alert(BaseModel):
    id: str
    furnace_id: str
    type: AlertType
    message: str
    timestamp: datetime = Field(default_factory=datetime.now)
    is_resolved: bool = False


class Metrics(BaseModel):
    date: str
    total_output: float
    average_output_per_furnace: float
    total_energy_consumed: float
    temperature_exceed_count: int
    completed_orders: int


class CreateFurnaceRequest(BaseModel):
    name: str
    design_capacity: float
    min_temperature: float
    max_temperature: float


class CreateBatchingOrderRequest(BaseModel):
    materials: List[Material]


class ChargeMaterialRequest(BaseModel):
    furnace_id: str
    materials: List[Material]


class TemperatureReportRequest(BaseModel):
    furnace_id: str
    temperature: float


class UpdateFurnaceStatusRequest(BaseModel):
    status: FurnaceStatus
    output_weight: Optional[float] = None
    energy_consumed: Optional[float] = None
