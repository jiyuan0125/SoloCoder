from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime

class StationBase(BaseModel):
    code: str = Field(..., description="车站代码")
    name: str = Field(..., description="车站名称")
    max_capacity: Optional[int] = Field(5000, description="最大容纳人数")

class StationCreate(StationBase):
    pass

class StationUpdate(BaseModel):
    name: Optional[str] = None
    max_capacity: Optional[int] = None

class StationResponse(StationBase):
    id: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True

class ZoneBase(BaseModel):
    name: str = Field(..., description="区域名称")
    zone_type: str = Field(..., description="区域类型：候车区、站台区、出站区等")
    max_capacity: Optional[int] = Field(1000, description="区域最大容纳人数")

class ZoneCreate(ZoneBase):
    pass

class ZoneUpdate(BaseModel):
    name: Optional[str] = None
    zone_type: Optional[str] = None
    max_capacity: Optional[int] = None
    current_count: Optional[int] = None

class ZoneResponse(ZoneBase):
    id: int
    station_id: int
    current_count: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True

class SecurityGateBase(BaseModel):
    gate_number: str = Field(..., description="安检通道号")

class SecurityGateCreate(SecurityGateBase):
    pass

class SecurityGateUpdate(BaseModel):
    is_open: Optional[bool] = None
    is_faulty: Optional[bool] = None
    queue_length: Optional[int] = None

class SecurityGateResponse(SecurityGateBase):
    id: int
    zone_id: int
    is_open: bool
    is_faulty: bool
    queue_length: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True

class TrainBase(BaseModel):
    train_number: str = Field(..., description="车次号")
    platform: str = Field(..., description="站台号")
    departure_time: datetime = Field(..., description="发车时间")

class TrainCreate(TrainBase):
    pass

class TrainResponse(TrainBase):
    id: int
    station_id: int
    checkin_start_time: datetime
    checkin_end_time: datetime
    is_checkin_active: bool
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True

class PassengerCountResponse(BaseModel):
    id: int
    zone_id: int
    count: int
    timestamp: datetime
    count_type: str
    
    class Config:
        orm_mode = True

class HourlyStats(BaseModel):
    hour: int
    total_count: int

class DailyStats(BaseModel):
    date: str
    total_count: int

class AlertResponse(BaseModel):
    id: int
    station_id: int
    zone_id: Optional[int]
    alert_type: str
    severity: str
    message: str
    is_active: bool
    created_at: datetime
    resolved_at: Optional[datetime]
    
    class Config:
        orm_mode = True
