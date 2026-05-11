from datetime import datetime, date
from typing import List, Optional
from pydantic import BaseModel, Field
from server.models import (
    FacilityType, FacilityStatus, AlertLevel, CommandStatus, FloodModeStatus
)


class ReservoirBase(BaseModel):
    name: str
    normal_storage_level: float = Field(description="正常蓄水位(米)")
    design_flood_level: float = Field(description="设计洪水位(米)")
    main_flood_season_start: int = Field(default=6, ge=1, le=12, description="主汛期开始月份")
    main_flood_season_end: int = Field(default=9, ge=1, le=12, description="主汛期结束月份")


class ReservoirCreate(ReservoirBase):
    pass


class ReservoirUpdate(BaseModel):
    name: Optional[str] = None
    normal_storage_level: Optional[float] = None
    design_flood_level: Optional[float] = None
    main_flood_season_start: Optional[int] = None
    main_flood_season_end: Optional[int] = None


class ReservoirResponse(ReservoirBase):
    id: int
    current_flood_mode: FloodModeStatus
    flood_mode_start_time: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class FloodFacilityBase(BaseModel):
    name: str
    facility_type: FacilityType
    priority: int = Field(default=0, description="优先级，数字越大越优先")


class FloodFacilityCreate(FloodFacilityBase):
    reservoir_id: int


class FloodFacilityUpdate(BaseModel):
    name: Optional[str] = None
    priority: Optional[int] = None
    status: Optional[FacilityStatus] = None


class FloodFacilityResponse(FloodFacilityBase):
    id: int
    reservoir_id: int
    status: FacilityStatus
    total_run_hours: float
    run_start_time: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class GateBase(BaseModel):
    name: str
    gate_number: int


class GateCreate(GateBase):
    facility_id: int


class GateUpdate(BaseModel):
    name: Optional[str] = None


class GateResponse(GateBase):
    id: int
    facility_id: int
    current_open_percent: float
    last_operation_time: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class WaterLevelRecordCreate(BaseModel):
    reservoir_id: int
    level: float = Field(description="库水位(米)")
    record_time: datetime = Field(default_factory=datetime.utcnow)


class WaterLevelRecordResponse(BaseModel):
    id: int
    reservoir_id: int
    level: float
    record_time: datetime
    is_aggregated: bool
    aggregate_hour: Optional[int] = None
    aggregate_date: Optional[date] = None
    created_at: datetime

    class Config:
        from_attributes = True


class InflowRecordCreate(BaseModel):
    reservoir_id: int
    inflow_rate: float = Field(description="入库流量(立方米/秒)")
    record_time: datetime = Field(default_factory=datetime.utcnow)


class InflowRecordResponse(BaseModel):
    id: int
    reservoir_id: int
    inflow_rate: float
    record_time: datetime
    created_at: datetime

    class Config:
        from_attributes = True


class DispatchCommandBase(BaseModel):
    command_type: str
    target_open_percent: float = Field(ge=0, le=100, description="目标开度(%)")
    operator_id: str
    operator_name: str
    remark: Optional[str] = None


class DispatchCommandCreate(DispatchCommandBase):
    gate_id: int
    reservoir_id: int


class DispatchCommandConfirm(BaseModel):
    confirmer_id: str
    confirmer_name: str


class DispatchCommandAuthorize(BaseModel):
    authorize_id: str
    authorize_name: str


class DispatchCommandResponse(BaseModel):
    id: int
    reservoir_id: int
    gate_id: int
    command_type: str
    target_open_percent: float
    operator_id: str
    operator_name: str
    confirmer_id: Optional[str] = None
    confirmer_name: Optional[str] = None
    authorize_id: Optional[str] = None
    authorize_name: Optional[str] = None
    status: CommandStatus
    need_dual_confirm: bool
    need_headquarters_auth: bool
    executed_time: Optional[datetime] = None
    remark: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class FloodOperationRecordResponse(BaseModel):
    id: int
    reservoir_id: int
    start_time: datetime
    end_time: Optional[datetime] = None
    max_water_level: Optional[float] = None
    max_water_level_time: Optional[datetime] = None
    summary: Optional[str] = None
    commands_count: int
    created_at: datetime

    class Config:
        from_attributes = True


class AlertResponse(BaseModel):
    id: int
    reservoir_id: int
    alert_level: AlertLevel
    alert_type: str
    message: str
    trigger_value: Optional[float] = None
    threshold_value: Optional[float] = None
    is_acknowledged: bool
    acknowledged_time: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class FloodLimitLevelInfo(BaseModel):
    reservoir_id: int
    reservoir_name: str
    current_month: int
    is_main_flood_season: bool
    normal_storage_level: float
    design_flood_level: float
    flood_limit_level: float
    red_alert_threshold: float


class CurrentStatusResponse(BaseModel):
    reservoir: ReservoirResponse
    current_level: Optional[float] = None
    current_inflow: Optional[float] = None
    flood_limit_level: float
    is_flood_mode: bool
    is_red_alert: bool
    facilities: List[FloodFacilityResponse]
    gates: List[GateResponse]


class PaginatedResponse(BaseModel):
    total: int
    items: List
