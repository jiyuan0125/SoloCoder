from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from server.models import TaskStatus, SubTaskType, FuelTruckStatus, BusStatus


class FlightBase(BaseModel):
    flight_number: str
    aircraft_type: str
    passenger_count: int
    gate: str
    status: str = "scheduled"
    temperature: Optional[float] = None


class FlightCreate(FlightBase):
    pass


class FlightUpdate(BaseModel):
    flight_number: Optional[str] = None
    aircraft_type: Optional[str] = None
    passenger_count: Optional[int] = None
    gate: Optional[str] = None
    status: Optional[str] = None
    temperature: Optional[float] = None


class FlightResponse(FlightBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class SubTaskBase(BaseModel):
    task_type: SubTaskType
    status: TaskStatus = TaskStatus.CREATED


class SubTaskResponse(BaseModel):
    id: int
    guarantee_task_id: int
    task_type: SubTaskType
    status: TaskStatus
    assigned_team_id: Optional[int] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    notes: Optional[str] = None
    dependencies: List[int] = []

    class Config:
        from_attributes = True


class GuaranteeTaskBase(BaseModel):
    flight_id: int


class GuaranteeTaskResponse(BaseModel):
    id: int
    flight_id: int
    status: TaskStatus
    subtasks: List[SubTaskResponse] = []
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CleaningTeamBase(BaseModel):
    name: str
    member_count: int


class CleaningTeamCreate(CleaningTeamBase):
    pass


class CleaningTeamResponse(CleaningTeamBase):
    id: int
    is_available: bool
    current_flight_id: Optional[int] = None

    class Config:
        from_attributes = True


class FuelTruckBase(BaseModel):
    name: str
    capacity: float
    current_fuel: float


class FuelTruckCreate(FuelTruckBase):
    pass


class FuelTruckResponse(FuelTruckBase):
    id: int
    status: FuelTruckStatus

    class Config:
        from_attributes = True


class ShuttleBusBase(BaseModel):
    name: str
    capacity: int


class ShuttleBusCreate(ShuttleBusBase):
    pass


class ShuttleBusResponse(ShuttleBusBase):
    id: int
    status: BusStatus

    class Config:
        from_attributes = True


class FuelingTaskBase(BaseModel):
    planned_amount: float
    tank_capacity: float


class FuelingTaskCreate(FuelingTaskBase):
    pass


class FuelingTaskUpdate(BaseModel):
    actual_amount: Optional[float] = None


class FuelingTaskResponse(BaseModel):
    id: int
    subtask_id: int
    planned_amount: float
    actual_amount: Optional[float] = None
    tank_capacity: float
    trucks_needed: int
    fuel_truck_ids: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class DeicingTaskResponse(BaseModel):
    id: int
    subtask_id: int
    deicing_type: str
    temperature: Optional[float] = None
    volume_used: Optional[float] = None
    notes: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class PassengerTaskResponse(BaseModel):
    subtask: SubTaskResponse
    bus_allocation: Optional[dict] = None
    passenger_count: int


class CleaningTaskResponse(BaseModel):
    subtask: SubTaskResponse
    cleaning_team: Optional[CleaningTeamResponse] = None


class BaggageTaskResponse(BaseModel):
    subtask: SubTaskResponse


class TaskStatusUpdate(BaseModel):
    status: TaskStatus
    notes: Optional[str] = None
