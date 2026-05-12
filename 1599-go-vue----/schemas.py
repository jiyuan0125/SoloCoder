from pydantic import BaseModel, Field
from datetime import datetime, date
from typing import List, Optional
from models import PlantStatus, ProtectionLevel


class ObservationRecordBase(BaseModel):
    growth_condition: str
    notes: Optional[str] = None


class ObservationRecordCreate(ObservationRecordBase):
    pass


class ObservationRecord(ObservationRecordBase):
    id: int
    plant_id: int
    record_date: datetime
    created_at: datetime

    class Config:
        from_attributes = True


class PlantBase(BaseModel):
    scientific_name: str
    family: str
    genus: str
    origin: str
    location: Optional[str] = None
    description: Optional[str] = None
    protection_level: ProtectionLevel = ProtectionLevel.NONE
    is_published: bool = False


class PlantCreate(PlantBase):
    status: PlantStatus = PlantStatus.NORMAL


class PlantUpdate(BaseModel):
    scientific_name: Optional[str] = None
    family: Optional[str] = None
    genus: Optional[str] = None
    origin: Optional[str] = None
    location: Optional[str] = None
    description: Optional[str] = None
    status: Optional[PlantStatus] = None
    protection_level: Optional[ProtectionLevel] = None
    is_published: Optional[bool] = None


class Plant(PlantBase):
    id: int
    status: PlantStatus
    introduction_date: Optional[datetime] = None
    observation_end_date: Optional[datetime] = None
    previous_status: Optional[PlantStatus] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class PlantWithObservations(Plant):
    observation_records: List[ObservationRecord] = []


class PublishedPlant(BaseModel):
    id: int
    scientific_name: str
    family: str
    genus: str
    origin: str
    description: Optional[str] = None
    protection_level: ProtectionLevel
    location: Optional[str] = None
    status: PlantStatus

    class Config:
        from_attributes = True


class ExhibitionBase(BaseModel):
    name: str
    description: Optional[str] = None
    start_date: date
    end_date: Optional[date] = None


class ExhibitionCreate(ExhibitionBase):
    plant_ids: List[int] = []


class Exhibition(ExhibitionBase):
    id: int
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True


class ExhibitionWithPlants(Exhibition):
    plant_ids: List[int] = []


class PlantFilter(BaseModel):
    family: Optional[str] = None
    genus: Optional[str] = None
    status: Optional[PlantStatus] = None
    protection_level: Optional[ProtectionLevel] = None
    is_published: Optional[bool] = None
    scientific_name: Optional[str] = None


class CSVExportResponse(BaseModel):
    total_count: int
    message: str
