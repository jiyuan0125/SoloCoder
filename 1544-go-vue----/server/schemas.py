from datetime import datetime
from typing import Optional, List
from enum import Enum

from pydantic import BaseModel, ConfigDict

from .models import StationType, DataQualityStatus, AggregationGranularity, AggregationMissingStatus


class StationTypeEnum(str, Enum):
    WATER_LEVEL = "water_level"
    RAINFALL = "rainfall"
    FLOW = "flow"


class DataQualityStatusEnum(str, Enum):
    VALID = "valid"
    SUSPICIOUS = "suspicious"
    INVALID = "invalid"


class AggregationGranularityEnum(str, Enum):
    HOURLY = "hourly"
    DAILY = "daily"


class AggregationMissingStatusEnum(str, Enum):
    COMPLETE = "complete"
    MISSING = "missing"


class StationBase(BaseModel):
    name: str
    station_type: StationTypeEnum
    device_model: Optional[str] = None


class StationCreate(StationBase):
    pass


class StationUpdate(BaseModel):
    name: Optional[str] = None
    station_type: Optional[StationTypeEnum] = None
    device_model: Optional[str] = None


class StationResponse(StationBase):
    id: int
    created_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)


class CalibrationBase(BaseModel):
    last_calibration_date: datetime
    next_calibration_date: datetime
    is_done: bool = False


class CalibrationCreate(CalibrationBase):
    pass


class CalibrationUpdate(BaseModel):
    last_calibration_date: Optional[datetime] = None
    next_calibration_date: Optional[datetime] = None
    is_done: Optional[bool] = None


class CalibrationResponse(CalibrationBase):
    id: int
    station_id: int
    created_at: datetime

    model_config = ConfigDict(from_attributes=True)


class CalibrationTodoResponse(BaseModel):
    station_id: int
    station_name: str
    next_calibration_date: datetime
    days_remaining: int


class RawDataBase(BaseModel):
    data_type: StationTypeEnum
    timestamp: datetime
    value: float


class RawDataCreate(RawDataBase):
    pass


class RawDataResponse(RawDataBase):
    id: int
    station_id: int
    quality_status: DataQualityStatusEnum
    created_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)


class QualityRecordBase(BaseModel):
    raw_data_id: Optional[int] = None
    issue_type: str
    description: str


class QualityRecordCreate(QualityRecordBase):
    pass


class QualityRecordReview(BaseModel):
    reviewed: bool
    reviewer: Optional[str] = None
    review_notes: Optional[str] = None


class QualityRecordResponse(QualityRecordBase):
    id: int
    station_id: int
    detected_at: datetime
    reviewed: bool
    reviewer: Optional[str] = None
    review_notes: Optional[str] = None
    reviewed_at: Optional[datetime] = None

    model_config = ConfigDict(from_attributes=True)


class AggregatedDataResponse(BaseModel):
    id: int
    station_id: int
    data_type: StationTypeEnum
    granularity: AggregationGranularityEnum
    period_start: datetime
    period_end: datetime
    average_value: Optional[float] = None
    max_value: Optional[float] = None
    min_value: Optional[float] = None
    missing_status: AggregationMissingStatusEnum
    sample_count: int
    device_model: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)


class AggregationRequest(BaseModel):
    station_id: int
    start_time: datetime
    end_time: datetime


class PaginationInfo(BaseModel):
    page: int
    page_size: int
    total: int
    total_pages: int


class PaginatedResponse(BaseModel):
    data: List
    pagination: PaginationInfo


class BulkRawDataCreate(BaseModel):
    data: List[RawDataCreate]
