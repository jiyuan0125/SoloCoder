from pydantic import BaseModel
from typing import Optional
from datetime import datetime
from models import ItemStatus, ComplaintStatus, ComplaintSeverity


class InquiryRecordCreate(BaseModel):
    inquirer_name: Optional[str] = None
    inquirer_phone: Optional[str] = None
    content: str
    response: Optional[str] = None


class InquiryRecordUpdate(BaseModel):
    inquirer_name: Optional[str] = None
    inquirer_phone: Optional[str] = None
    content: Optional[str] = None
    response: Optional[str] = None


class InquiryRecordResponse(BaseModel):
    id: int
    record_number: str
    record_date: str
    sequence_number: int
    inquirer_name: Optional[str]
    inquirer_phone: Optional[str]
    content: str
    response: Optional[str]
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class FoundItemCreate(BaseModel):
    item_type: str
    description: str
    location: str
    value: Optional[float] = 0
    finder_name: Optional[str] = None
    finder_phone: Optional[str] = None
    found_time: datetime


class FoundItemUpdate(BaseModel):
    item_type: Optional[str] = None
    description: Optional[str] = None
    location: Optional[str] = None
    value: Optional[float] = None
    finder_name: Optional[str] = None
    finder_phone: Optional[str] = None
    found_time: Optional[datetime] = None


class FoundItemResponse(BaseModel):
    id: int
    item_type: str
    description: str
    location: str
    value: float
    finder_name: Optional[str]
    finder_phone: Optional[str]
    found_time: datetime
    status: ItemStatus
    is_secondary_confirmed: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class LostItemCreate(BaseModel):
    item_type: str
    description: Optional[str] = None
    location: str
    value: Optional[float] = 0
    owner_name: Optional[str] = None
    owner_phone: Optional[str] = None
    lost_time: datetime


class LostItemUpdate(BaseModel):
    item_type: Optional[str] = None
    description: Optional[str] = None
    location: Optional[str] = None
    value: Optional[float] = None
    owner_name: Optional[str] = None
    owner_phone: Optional[str] = None
    lost_time: Optional[datetime] = None


class LostItemResponse(BaseModel):
    id: int
    item_type: str
    description: Optional[str]
    location: str
    value: float
    owner_name: Optional[str]
    owner_phone: Optional[str]
    lost_time: datetime
    status: ItemStatus
    is_secondary_confirmed: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class ItemMatchResponse(BaseModel):
    id: int
    lost_item_id: int
    found_item_id: int
    similarity_score: float
    location_score: float
    total_score: float
    is_confirmed: int
    is_claimed: int
    created_at: datetime
    
    class Config:
        from_attributes = True


class ComplaintCreate(BaseModel):
    complainant_name: Optional[str] = None
    complainant_phone: Optional[str] = None
    content: str
    severity: ComplaintSeverity = ComplaintSeverity.NORMAL


class ComplaintUpdate(BaseModel):
    complainant_name: Optional[str] = None
    complainant_phone: Optional[str] = None
    content: Optional[str] = None
    severity: Optional[ComplaintSeverity] = None
    current_handler: Optional[str] = None
    response_content: Optional[str] = None


class ComplaintResponse(BaseModel):
    id: int
    complainant_name: Optional[str]
    complainant_phone: Optional[str]
    content: str
    severity: ComplaintSeverity
    status: ComplaintStatus
    current_handler: Optional[str]
    response_content: Optional[str]
    is_overdue: int
    supervisor_notified: int
    registered_at: Optional[datetime]
    processing_at: Optional[datetime]
    pending_confirmation_at: Optional[datetime]
    closed_at: Optional[datetime]
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class StatisticsResponse(BaseModel):
    today_inquiry_count: int
    today_lost_items_count: int
    today_found_items_count: int
    today_matches_count: int
    
    monthly_complaint_count: int
    avg_processing_hours: float
    overdue_count: int


class ConfirmSecondaryRequest(BaseModel):
    match_id: int
    is_confirmed: bool
