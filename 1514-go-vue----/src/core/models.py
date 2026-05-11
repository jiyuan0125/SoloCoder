from datetime import date, datetime, time
from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field, field_validator


class RegistrationStatus(str, Enum):
    WAITING = "waiting"
    TREATING = "treating"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class AppointmentStatus(str, Enum):
    CONFIRMED = "confirmed"
    WAITLIST = "waitlist"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class TodoStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    SKIPPED = "skipped"


class Owner(BaseModel):
    id: Optional[str] = None
    name: str
    phone: str
    address: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class Pet(BaseModel):
    id: Optional[str] = None
    name: str
    species: str
    breed: Optional[str] = None
    gender: Optional[str] = None
    birth_date: Optional[date] = None
    owner_id: str
    created_at: datetime = Field(default_factory=datetime.now)


class Doctor(BaseModel):
    id: Optional[str] = None
    name: str
    specialty: Optional[str] = None
    phone: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class Registration(BaseModel):
    id: Optional[str] = None
    pet_id: str
    owner_id: str
    doctor_id: Optional[str] = None
    symptoms: Optional[str] = None
    status: RegistrationStatus = RegistrationStatus.WAITING
    created_at: datetime = Field(default_factory=datetime.now)
    completed_at: Optional[datetime] = None


class PrescriptionItem(BaseModel):
    id: Optional[str] = None
    diagnosis_id: Optional[str] = None
    medicine_name: str
    unit_price: float
    quantity: int
    dosage: str
    days: int = 1
    remarks: Optional[str] = None

    @field_validator("unit_price")
    @classmethod
    def unit_price_non_negative(cls, v: float) -> float:
        if v < 0:
            raise ValueError("单价不能为负数")
        return v

    @property
    def subtotal(self) -> float:
        if self.unit_price <= 0:
            return 0.0
        return self.unit_price * self.quantity


class Treatment(BaseModel):
    id: Optional[str] = None
    diagnosis_id: Optional[str] = None
    item_name: str
    unit_price: float
    quantity: int
    remarks: Optional[str] = None

    @field_validator("unit_price")
    @classmethod
    def unit_price_non_negative(cls, v: float) -> float:
        if v < 0:
            raise ValueError("单价不能为负数")
        return v

    @property
    def subtotal(self) -> float:
        if self.unit_price <= 0:
            return 0.0
        return self.unit_price * self.quantity


class Diagnosis(BaseModel):
    id: Optional[str] = None
    registration_id: str
    doctor_id: str
    diagnosis: str
    prescription_items: List[PrescriptionItem] = Field(default_factory=list)
    treatments: List[Treatment] = Field(default_factory=list)
    remarks: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class FeeItem(BaseModel):
    id: Optional[str] = None
    fee_record_id: Optional[str] = None
    item_name: str
    item_type: str
    unit_price: float
    quantity: int
    subtotal: float
    is_charged: bool = True


class FeeRecord(BaseModel):
    id: Optional[str] = None
    registration_id: str
    diagnosis_id: str
    items: List[FeeItem] = Field(default_factory=list)
    total_amount: float = 0.0
    paid: bool = False
    created_at: datetime = Field(default_factory=datetime.now)
    paid_at: Optional[datetime] = None


class Vaccine(BaseModel):
    id: Optional[str] = None
    name: str
    manufacturer: Optional[str] = None
    recommended_interval_days: Optional[int] = None
    created_at: datetime = Field(default_factory=datetime.now)


class VaccineRecord(BaseModel):
    id: Optional[str] = None
    pet_id: str
    vaccine_id: str
    vaccine_name: str
    doctor_id: Optional[str] = None
    batch_number: Optional[str] = None
    inoculation_date: date
    next_inoculation_date: Optional[date] = None
    remarks: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)

    @field_validator("next_inoculation_date")
    @classmethod
    def next_date_not_earlier(
        cls, v: Optional[date], info: dict
    ) -> Optional[date]:
        inoculation_date = info.data.get("inoculation_date")
        if v and inoculation_date and v < inoculation_date:
            raise ValueError("下次接种日期不能早于本次接种日期")
        return v


class WorkSlot(BaseModel):
    id: Optional[str] = None
    doctor_id: str
    date: date
    start_time: time
    end_time: time
    max_appointments: int = 1
    current_appointments: int = 0
    is_available: bool = True
    created_at: datetime = Field(default_factory=datetime.now)


class Appointment(BaseModel):
    id: Optional[str] = None
    owner_id: str
    pet_id: str
    doctor_id: str
    work_slot_id: str
    appointment_date: date
    start_time: time
    end_time: time
    status: AppointmentStatus = AppointmentStatus.CONFIRMED
    waitlist_position: Optional[int] = None
    remarks: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class MedicationCourse(BaseModel):
    id: Optional[str] = None
    pet_id: str
    owner_id: str
    diagnosis_id: Optional[str] = None
    medicine_name: str
    dosage: str
    total_days: int
    frequency_per_day: int = 1
    start_date: date
    notes: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class MedicationTodo(BaseModel):
    id: Optional[str] = None
    course_id: str
    pet_id: str
    owner_id: str
    due_date: date
    medicine_name: str
    dosage: str
    status: TodoStatus = TodoStatus.PENDING
    notes: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)
    completed_at: Optional[datetime] = None
