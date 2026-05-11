from datetime import datetime
from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field


class AlarmLevel(str, Enum):
    NORMAL = "normal"
    WARNING = "warning"
    CRITICAL = "critical"
    HIGH_RISK = "high_risk"


class BatchStatus(str, Enum):
    RUNNING = "running"
    COMPLETED = "completed"
    ABORTED = "aborted"
    ABNORMAL = "abnormal"


class Reactor(BaseModel):
    id: str
    name: str
    description: str = ""
    design_temperature_max: float
    design_pressure_max: float
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)


class ReactorCreate(BaseModel):
    name: str
    description: str = ""
    design_temperature_max: float
    design_pressure_max: float


class ReactorUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    design_temperature_max: Optional[float] = None
    design_pressure_max: Optional[float] = None


class SensorReading(BaseModel):
    id: str = ""
    reactor_id: str
    temperature: float
    pressure: float
    timestamp: datetime = Field(default_factory=datetime.now)
    alarm_level: AlarmLevel = AlarmLevel.NORMAL
    is_high_risk: bool = False


class SensorReadingCreate(BaseModel):
    reactor_id: str
    temperature: float
    pressure: float


class Material(BaseModel):
    name: str
    required_amount: float
    unit: str
    order: int


class Recipe(BaseModel):
    id: str
    name: str
    description: str = ""
    materials: List[Material]
    time_window_seconds: int
    created_at: datetime = Field(default_factory=datetime.now)


class RecipeCreate(BaseModel):
    name: str
    description: str = ""
    materials: List[Material]
    time_window_seconds: int


class FeedingRecord(BaseModel):
    id: str = ""
    batch_id: str
    material_name: str
    required_amount: float
    actual_amount: float
    unit: str
    order: int
    deviation_percent: float = 0.0
    is_abnormal: bool = False
    timestamp: datetime = Field(default_factory=datetime.now)


class FeedingRecordCreate(BaseModel):
    batch_id: str
    material_name: str
    required_amount: float
    actual_amount: float
    unit: str
    order: int


class ProductionBatch(BaseModel):
    id: str
    reactor_id: str
    recipe_id: str
    start_time: datetime = Field(default_factory=datetime.now)
    end_time: Optional[datetime] = None
    status: BatchStatus = BatchStatus.RUNNING
    sensor_readings: List[SensorReading] = []
    feeding_records: List[FeedingRecord] = []
    is_abnormal: bool = False
    abnormal_reasons: List[str] = []
    created_at: datetime = Field(default_factory=datetime.now)


class ProductionBatchCreate(BaseModel):
    reactor_id: str
    recipe_id: str


class DailyStatistics(BaseModel):
    date: str
    total_batches: int = 0
    completed_batches: int = 0
    abnormal_batches: int = 0
    abnormal_rate: float = 0.0
    created_at: datetime = Field(default_factory=datetime.now)


class AlarmRecord(BaseModel):
    id: str = ""
    reactor_id: str
    alarm_level: AlarmLevel
    is_high_risk: bool
    message: str
    sensor_reading_id: str
    timestamp: datetime = Field(default_factory=datetime.now)
