from datetime import date, datetime
from typing import List, Optional
from pydantic import BaseModel, Field, field_validator


class FarmerBase(BaseModel):
    name: str
    address: str
    livestock_type: str
    scale: int = Field(ge=1)


class FarmerCreate(FarmerBase):
    pass


class FarmerUpdate(BaseModel):
    name: Optional[str] = None
    address: Optional[str] = None
    livestock_type: Optional[str] = None
    scale: Optional[int] = Field(None, ge=1)


class Farmer(FarmerBase):
    id: int
    last_visit_date: Optional[date] = None
    created_at: datetime

    class Config:
        from_attributes = True


class MedicineBase(BaseModel):
    name: str
    specification: str
    unit: str


class MedicineCreate(MedicineBase):
    stock: float = Field(ge=0)


class MedicineUpdate(BaseModel):
    name: Optional[str] = None
    specification: Optional[str] = None
    unit: Optional[str] = None
    stock: Optional[float] = Field(None, ge=0)


class Medicine(MedicineBase):
    id: int
    stock: float
    created_at: datetime

    class Config:
        from_attributes = True


class PrescriptionBase(BaseModel):
    medicine_id: int
    dosage_per_day: float = Field(gt=0)
    treatment_days: int = Field(ge=1)
    notes: Optional[str] = None
    start_date: date


class PrescriptionCreate(PrescriptionBase):
    pass


class Prescription(PrescriptionBase):
    id: int
    case_id: int
    total_dosage: float
    medicine: Optional[Medicine] = None
    created_at: datetime

    class Config:
        from_attributes = True


class CaseBase(BaseModel):
    animal_type: str
    symptoms: str
    diagnosis: str


class CaseCreate(CaseBase):
    prescriptions: List[PrescriptionCreate] = []


class Case(CaseBase):
    id: int
    visit_id: int
    prescriptions: List[Prescription] = []
    created_at: datetime

    class Config:
        from_attributes = True


class VisitRecordBase(BaseModel):
    farmer_id: int
    visit_date: date
    has_abnormality: bool
    notes: Optional[str] = None


class VisitRecordCreate(VisitRecordBase):
    cases: List[CaseCreate] = []


class VisitRecordUpdate(BaseModel):
    has_abnormality: Optional[bool] = None
    notes: Optional[str] = None
    cases: Optional[List[CaseCreate]] = None


class VisitRecord(VisitRecordBase):
    id: int
    cases: List[Case] = []
    farmer: Optional[Farmer] = None
    created_at: datetime

    class Config:
        from_attributes = True


class RouteItemBase(BaseModel):
    farmer_id: int
    order_index: int


class RouteItemCreate(RouteItemBase):
    pass


class RouteItem(RouteItemBase):
    id: int
    route_id: int
    is_completed: bool
    farmer: Optional[Farmer] = None

    class Config:
        from_attributes = True


class RouteBase(BaseModel):
    route_date: date


class RouteCreate(RouteBase):
    pass


class RouteReorder(BaseModel):
    items: List[RouteItemCreate]

    @field_validator("items")
    def check_unique_farmer(cls, v):
        farmer_ids = [item.farmer_id for item in v]
        if len(farmer_ids) != len(set(farmer_ids)):
            raise ValueError("同一养殖户在巡诊路线中不能出现两次")
        return v


class Route(RouteBase):
    id: int
    is_generated: bool
    items: List[RouteItem] = []
    created_at: datetime

    class Config:
        from_attributes = True


class TodoBase(BaseModel):
    todo_date: date
    is_completed: bool = False
    notes: Optional[str] = None


class TodoUpdate(BaseModel):
    is_completed: Optional[bool] = None
    notes: Optional[str] = None


class Todo(TodoBase):
    id: int
    prescription_id: int
    prescription: Optional[Prescription] = None
    created_at: datetime

    class Config:
        from_attributes = True


class MedicineConsumptionRank(BaseModel):
    medicine_name: str
    total_dosage: float
    unit: str


class DashboardStats(BaseModel):
    month: str
    visit_count: int
    covered_farmers: int
    new_cases: int
    consumption_ranking: List[MedicineConsumptionRank]
