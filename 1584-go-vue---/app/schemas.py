from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime, date, time
from app.models import TaskStatus


class LineBase(BaseModel):
    name: str = Field(..., max_length=100)
    description: Optional[str] = None


class LineCreate(LineBase):
    pass


class LineResponse(LineBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class MaintenanceWindowBase(BaseModel):
    line_id: int
    date: Optional[date] = None
    start_time: time
    end_time: time
    is_general: bool = False


class MaintenanceWindowCreate(MaintenanceWindowBase):
    pass


class MaintenanceWindowUpdate(BaseModel):
    start_time: Optional[time] = None
    end_time: Optional[time] = None
    is_general: Optional[bool] = None


class MaintenanceWindowResponse(MaintenanceWindowBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class StaffBase(BaseModel):
    name: str = Field(..., max_length=100)
    employee_id: str = Field(..., max_length=50)
    join_date: Optional[date] = None
    department: Optional[str] = None


class StaffCreate(StaffBase):
    pass


class StaffResponse(StaffBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class QualificationBase(BaseModel):
    staff_id: int
    type: str = Field(..., max_length=100)
    valid_from: date
    valid_to: Optional[date] = None


class QualificationCreate(QualificationBase):
    pass


class QualificationResponse(QualificationBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class TaskBase(BaseModel):
    title: str = Field(..., max_length=200)
    description: Optional[str] = None
    line_id: int
    start_kp: int
    end_kp: int
    scheduled_start: datetime
    scheduled_end: datetime
    responsible_person_id: int


class TaskCreate(TaskBase):
    assigned_staff_ids: List[int] = []


class TaskUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    start_kp: Optional[int] = None
    end_kp: Optional[int] = None
    scheduled_start: Optional[datetime] = None
    scheduled_end: Optional[datetime] = None
    assigned_staff_ids: Optional[List[int]] = None


class TaskStaffResponse(BaseModel):
    staff_id: int
    staff_name: str
    employee_id: str

    class Config:
        from_attributes = True


class TaskResponse(TaskBase):
    id: int
    status: TaskStatus
    submitted_at: Optional[datetime] = None
    approved_at: Optional[datetime] = None
    rejected_at: Optional[datetime] = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    timeout_at: Optional[datetime] = None
    qualification_checked: bool
    created_at: datetime
    updated_at: datetime
    assigned_staff: List[TaskStaffResponse] = []

    class Config:
        from_attributes = True


class ConflictDetail(BaseModel):
    conflict_type: str
    task_id: int
    task_title: str
    reason: str


class ConflictCheckResponse(BaseModel):
    has_conflicts: bool
    conflicts: List[ConflictDetail] = []


class TodoItem(BaseModel):
    task_id: int
    task_title: str
    scheduled_start: datetime
    scheduled_end: datetime
    status: TaskStatus
    action_required: str


class TodoResponse(BaseModel):
    staff_id: int
    staff_name: str
    todos: List[TodoItem] = []


class NotificationResponse(BaseModel):
    id: int
    task_id: int
    type: str
    message: str
    sent_at: datetime
    read: bool

    class Config:
        from_attributes = True
