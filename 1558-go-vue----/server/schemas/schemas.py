from datetime import datetime
from typing import Optional, List

from pydantic import BaseModel, Field


class ShipBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    imo_number: str = Field(..., min_length=1, max_length=20)
    flag: str = Field(..., min_length=1, max_length=50)
    has_hazardous_qualification: bool = False


class ShipCreate(ShipBase):
    pass


class ShipResponse(ShipBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DeclarationBase(BaseModel):
    declaration_number: str = Field(..., min_length=1, max_length=50)
    ship_id: int
    voyage_number: str = Field(..., min_length=1, max_length=50)
    hazardous_category: int = Field(..., ge=1, le=9)
    cargo_name: str = Field(..., min_length=1, max_length=200)
    cargo_quantity: float = Field(..., gt=0)
    packaging_compliant: bool = False
    submitted_by: str = Field(..., min_length=1, max_length=100)


class DeclarationCreate(DeclarationBase):
    pass


class DeclarationStatusUpdate(BaseModel):
    status: str


class ReviewBase(BaseModel):
    reviewer: str = Field(..., min_length=1, max_length=100)
    approved: bool
    comments: Optional[str] = None


class InitialReviewCreate(ReviewBase):
    pass


class FinalReviewCreate(ReviewBase):
    pass


class ReviewResponse(ReviewBase):
    id: int
    declaration_id: int
    review_type: str
    reviewed_at: datetime

    class Config:
        from_attributes = True


class LoadingBase(BaseModel):
    operator: str = Field(..., min_length=1, max_length=100)
    temperature: Optional[float] = None
    radiation_dose_rate: Optional[float] = None
    notes: Optional[str] = None


class LoadingCreate(LoadingBase):
    pass


class LoadingComplete(BaseModel):
    notes: Optional[str] = None


class LoadingResponse(LoadingBase):
    id: int
    declaration_id: int
    start_time: datetime
    end_time: Optional[datetime] = None
    completed: bool

    class Config:
        from_attributes = True


class EmergencyCreate(BaseModel):
    impact_range: str = Field(..., min_length=1, max_length=50)
    triggered_by: Optional[str] = None


class EmergencyResponse(BaseModel):
    id: int
    declaration_id: int
    category: int
    impact_range: str
    level: int
    level_description: str
    plan: str
    created_at: datetime
    triggered_by: Optional[str] = None

    class Config:
        from_attributes = True


class AuditLogResponse(BaseModel):
    id: int
    declaration_id: Optional[int] = None
    action: str
    actor: str
    details: Optional[str] = None
    timestamp: datetime

    class Config:
        from_attributes = True


class DeclarationResponse(BaseModel):
    id: int
    declaration_number: str
    ship_id: int
    ship_name: str
    voyage_number: str
    hazardous_category: int
    hazardous_category_name: str
    cargo_name: str
    cargo_quantity: float
    packaging_compliant: bool
    submitted_by: str
    status: str
    status_description: str
    created_at: datetime
    updated_at: datetime
    reviews: List[ReviewResponse] = []
    loading: Optional[LoadingResponse] = None
    emergencies: List[EmergencyResponse] = []

    class Config:
        from_attributes = True


class DeclarationListResponse(BaseModel):
    id: int
    declaration_number: str
    ship_name: str
    voyage_number: str
    hazardous_category_name: str
    cargo_name: str
    status: str
    status_description: str
    created_at: datetime
    updated_at: datetime
