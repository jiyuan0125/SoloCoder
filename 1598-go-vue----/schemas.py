from datetime import date, datetime
from typing import List, Optional
from pydantic import BaseModel, Field
from models import AnimalStatus, ConservationLevel, FeedingShift, FeedingStatus, TodoType, TodoStatus


class EmployeeBase(BaseModel):
    name: str
    position: Optional[str] = None
    phone: Optional[str] = None
    is_active: bool = True


class EmployeeCreate(EmployeeBase):
    pass


class EmployeeUpdate(BaseModel):
    name: Optional[str] = None
    position: Optional[str] = None
    phone: Optional[str] = None
    is_active: Optional[bool] = None


class EmployeeResponse(EmployeeBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class ZoneBase(BaseModel):
    name: str
    description: Optional[str] = None


class ZoneCreate(ZoneBase):
    pass


class ZoneUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None


class ZoneResponse(ZoneBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class AnimalSpeciesBase(BaseModel):
    name: str
    scientific_name: Optional[str] = None
    conservation_level: ConservationLevel = ConservationLevel.LEAST_CONCERN
    description: Optional[str] = None


class AnimalSpeciesCreate(AnimalSpeciesBase):
    pass


class AnimalSpeciesUpdate(BaseModel):
    name: Optional[str] = None
    scientific_name: Optional[str] = None
    conservation_level: Optional[ConservationLevel] = None
    description: Optional[str] = None


class AnimalSpeciesResponse(AnimalSpeciesBase):
    id: int

    class Config:
        from_attributes = True


class FeedingStandardBase(BaseModel):
    species_id: int
    min_weight: float
    max_weight: float
    feed_type: str
    daily_amount: float
    frequency: int = 2
    is_special: bool = False
    for_health_status: Optional[AnimalStatus] = None


class FeedingStandardCreate(FeedingStandardBase):
    pass


class FeedingStandardUpdate(BaseModel):
    species_id: Optional[int] = None
    min_weight: Optional[float] = None
    max_weight: Optional[float] = None
    feed_type: Optional[str] = None
    daily_amount: Optional[float] = None
    frequency: Optional[int] = None
    is_special: Optional[bool] = None
    for_health_status: Optional[AnimalStatus] = None


class FeedingStandardResponse(FeedingStandardBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class AnimalBase(BaseModel):
    name: str
    species_id: int
    zone_id: int
    chip_id: Optional[str] = None
    gender: Optional[str] = None
    birth_date: Optional[date] = None
    weight: float
    health_status: AnimalStatus = AnimalStatus.HEALTHY
    notes: Optional[str] = None


class AnimalCreate(AnimalBase):
    pass


class AnimalUpdate(BaseModel):
    name: Optional[str] = None
    species_id: Optional[int] = None
    zone_id: Optional[int] = None
    chip_id: Optional[str] = None
    gender: Optional[str] = None
    birth_date: Optional[date] = None
    weight: Optional[float] = None
    health_status: Optional[AnimalStatus] = None
    notes: Optional[str] = None


class AnimalResponse(AnimalBase):
    id: int
    is_isolation: bool
    isolation_start_date: Optional[date] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class FeedBase(BaseModel):
    name: str
    unit: str
    current_stock: float = 0
    safety_stock: float = 10
    supplier: Optional[str] = None
    notes: Optional[str] = None


class FeedCreate(FeedBase):
    pass


class FeedUpdate(BaseModel):
    name: Optional[str] = None
    unit: Optional[str] = None
    current_stock: Optional[float] = None
    safety_stock: Optional[float] = None
    supplier: Optional[str] = None
    notes: Optional[str] = None


class FeedResponse(FeedBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class FeedingPlanBase(BaseModel):
    animal_id: int
    feed_type: str
    daily_amount: float
    frequency: int = 2
    is_active: bool = True
    start_date: Optional[date] = None
    end_date: Optional[date] = None


class FeedingPlanCreate(FeedingPlanBase):
    pass


class FeedingPlanUpdate(BaseModel):
    animal_id: Optional[int] = None
    feed_type: Optional[str] = None
    daily_amount: Optional[float] = None
    frequency: Optional[int] = None
    is_active: Optional[bool] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None


class FeedingPlanResponse(FeedingPlanBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ShiftAssignmentBase(BaseModel):
    zone_id: int
    employee_id: int
    shift_date: date
    shift: FeedingShift
    notes: Optional[str] = None


class ShiftAssignmentCreate(ShiftAssignmentBase):
    pass


class ShiftAssignmentResponse(ShiftAssignmentBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class FeedingRecordBase(BaseModel):
    animal_id: int
    feed_id: int
    feeder_id: Optional[int] = None
    feeding_date: date
    shift: FeedingShift
    planned_amount: float
    actual_amount: float
    status: FeedingStatus = FeedingStatus.NORMAL
    notes: Optional[str] = None


class FeedingRecordCreate(FeedingRecordBase):
    pass


class FeedingRecordUpdate(BaseModel):
    animal_id: Optional[int] = None
    feed_id: Optional[int] = None
    feeder_id: Optional[int] = None
    feeding_date: Optional[date] = None
    shift: Optional[FeedingShift] = None
    planned_amount: Optional[float] = None
    actual_amount: Optional[float] = None
    status: Optional[FeedingStatus] = None
    notes: Optional[str] = None


class FeedingRecordResponse(FeedingRecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class MedicationRecordBase(BaseModel):
    medical_record_id: int
    medication_name: str
    dosage: str
    frequency: Optional[str] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    notes: Optional[str] = None


class MedicationRecordCreate(MedicationRecordBase):
    pass


class MedicationRecordResponse(MedicationRecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class MedicalRecordBase(BaseModel):
    animal_id: int
    veterinarian_id: Optional[int] = None
    diagnosis: str
    symptoms: Optional[str] = None
    treatment: Optional[str] = None
    record_date: date
    next_check_date: Optional[date] = None
    notes: Optional[str] = None


class MedicalRecordCreate(MedicalRecordBase):
    medications: List[MedicationRecordCreate] = []


class MedicalRecordUpdate(BaseModel):
    animal_id: Optional[int] = None
    veterinarian_id: Optional[int] = None
    diagnosis: Optional[str] = None
    symptoms: Optional[str] = None
    treatment: Optional[str] = None
    record_date: Optional[date] = None
    is_upgraded: Optional[bool] = None
    next_check_date: Optional[date] = None
    notes: Optional[str] = None


class MedicalRecordResponse(MedicalRecordBase):
    id: int
    is_upgraded: bool
    created_at: datetime
    medications: List[MedicationRecordResponse] = []

    class Config:
        from_attributes = True


class VaccinationRecordBase(BaseModel):
    animal_id: int
    vaccine_name: str
    vaccine_batch: Optional[str] = None
    vaccination_date: date
    next_vaccination_date: date
    min_interval_days: int = 30
    veterinarian_name: Optional[str] = None
    notes: Optional[str] = None


class VaccinationRecordCreate(VaccinationRecordBase):
    pass


class VaccinationRecordResponse(VaccinationRecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class TodoBase(BaseModel):
    todo_type: TodoType
    title: str
    description: Optional[str] = None
    related_id: Optional[int] = None
    related_type: Optional[str] = None
    due_date: Optional[date] = None


class TodoCreate(TodoBase):
    pass


class TodoUpdate(BaseModel):
    todo_type: Optional[TodoType] = None
    status: Optional[TodoStatus] = None
    title: Optional[str] = None
    description: Optional[str] = None
    due_date: Optional[date] = None


class TodoResponse(TodoBase):
    id: int
    status: TodoStatus
    completed_at: Optional[datetime] = None
    completed_by: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True
