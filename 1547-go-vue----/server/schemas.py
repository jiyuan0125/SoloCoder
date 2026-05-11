from datetime import date, datetime
from typing import Optional, List
from pydantic import BaseModel, Field
from server.models import RiverLevel, IssueType, IssueSeverity, IssueStatus

class RiverBase(BaseModel):
    code: str
    name: str
    start_point: str
    end_point: str
    basin: str
    is_key_section: bool = False

class RiverCreate(RiverBase):
    pass

class RiverUpdate(BaseModel):
    code: Optional[str] = None
    name: Optional[str] = None
    start_point: Optional[str] = None
    end_point: Optional[str] = None
    basin: Optional[str] = None
    is_key_section: Optional[bool] = None

class River(RiverBase):
    id: int
    
    class Config:
        from_attributes = True

class RiverDetail(River):
    river_keepers: List["RiverKeeperSimple"] = []
    
    class Config:
        from_attributes = True

class RiverKeeperBase(BaseModel):
    name: str
    level: RiverLevel
    river_id: Optional[int] = None
    contact: Optional[str] = None

class RiverKeeperCreate(RiverKeeperBase):
    pass

class RiverKeeperUpdate(BaseModel):
    name: Optional[str] = None
    level: Optional[RiverLevel] = None
    river_id: Optional[int] = None
    contact: Optional[str] = None

class RiverKeeper(RiverKeeperBase):
    id: int
    
    class Config:
        from_attributes = True

class RiverKeeperSimple(BaseModel):
    id: int
    name: str
    level: RiverLevel
    
    class Config:
        from_attributes = True

class PatrolBase(BaseModel):
    river_id: int
    keeper_id: int
    patrol_date: date
    description: Optional[str] = None

class PatrolCreate(PatrolBase):
    pass

class PatrolUpdate(BaseModel):
    patrol_date: Optional[date] = None
    description: Optional[str] = None

class Patrol(PatrolBase):
    id: int
    created_at: datetime
    
    class Config:
        from_attributes = True

class PatrolDetail(Patrol):
    river_keeper: RiverKeeperSimple
    issues: List["IssueSimple"] = []
    
    class Config:
        from_attributes = True

class IssueBase(BaseModel):
    river_id: int
    patrol_id: Optional[int] = None
    issue_type: IssueType
    severity: IssueSeverity
    description: str

class IssueCreate(IssueBase):
    pass

class IssueUpdate(BaseModel):
    issue_type: Optional[IssueType] = None
    severity: Optional[IssueSeverity] = None
    description: Optional[str] = None
    rectification_note: Optional[str] = None
    review_note: Optional[str] = None

class IssueSimple(BaseModel):
    id: int
    issue_type: IssueType
    severity: IssueSeverity
    status: IssueStatus
    description: str
    
    class Config:
        from_attributes = True

class Issue(IssueBase):
    id: int
    status: IssueStatus
    discovered_date: date
    deadline: date
    escalated: bool
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True

class IssueDetail(Issue):
    notifications: List["Notification"] = []
    
    class Config:
        from_attributes = True

class NotificationBase(BaseModel):
    issue_id: int
    message: str
    recipient_level: RiverLevel

class Notification(NotificationBase):
    id: int
    sent_date: date
    created_at: datetime
    
    class Config:
        from_attributes = True

class MonthlyReportBase(BaseModel):
    river_id: int
    report_month: str
    content: Optional[str] = None

class MonthlyReportCreate(MonthlyReportBase):
    pass

class MonthlyReportUpdate(BaseModel):
    content: Optional[str] = None

class MonthlyReport(MonthlyReportBase):
    id: int
    patrol_count: int
    issue_count: int
    resolved_issue_count: int
    deadline: date
    is_submitted: bool
    submitted_date: Optional[date] = None
    created_at: datetime
    
    class Config:
        from_attributes = True

class PatrolFrequency(BaseModel):
    level: RiverLevel
    frequency: str
    days: int

RiverDetail.model_rebuild()
PatrolDetail.model_rebuild()
IssueDetail.model_rebuild()
