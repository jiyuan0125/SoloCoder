from datetime import datetime, date
from enum import Enum as PyEnum
from sqlalchemy import Column, Integer, String, Float, DateTime, Date, Boolean, ForeignKey, Text, Enum
from sqlalchemy.orm import relationship
from database import Base


class AnimalStatus(PyEnum):
    HEALTHY = "healthy"
    SICK = "sick"
    ISOLATION = "isolation"
    DEAD = "dead"
    PREGNANT = "pregnant"


class ConservationLevel(PyEnum):
    LEAST_CONCERN = "least_concern"
    NEAR_THREATENED = "near_threatened"
    VULNERABLE = "vulnerable"
    ENDANGERED = "endangered"
    CRITICALLY_ENDANGERED = "critically_endangered"
    EXTINCT_IN_WILD = "extinct_in_wild"


class FeedingShift(PyEnum):
    MORNING = "morning"
    AFTERNOON = "afternoon"
    EVENING = "evening"
    NIGHT = "night"


class FeedingStatus(PyEnum):
    NORMAL = "normal"
    REFUSED = "refused"
    PARTIAL = "partial"


class TodoType(PyEnum):
    PURCHASE = "purchase"
    VETERINARY = "veterinary"
    REPORT_DEATH = "report_death"
    UPGRADE_TREATMENT = "upgrade_treatment"
    VACCINATION = "vaccination"


class TodoStatus(PyEnum):
    PENDING = "pending"
    COMPLETED = "completed"


class Employee(Base):
    __tablename__ = "employees"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    position = Column(String(50))
    phone = Column(String(20))
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    shift_assignments = relationship("ShiftAssignment", back_populates="employee")
    feeding_records = relationship("FeedingRecord", back_populates="feeder")
    medical_records = relationship("MedicalRecord", back_populates="veterinarian")


class Zone(Base):
    __tablename__ = "zones"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False, unique=True)
    description = Column(String(255))
    created_at = Column(DateTime, default=datetime.utcnow)
    
    animals = relationship("Animal", back_populates="zone")
    shift_assignments = relationship("ShiftAssignment", back_populates="zone")


class AnimalSpecies(Base):
    __tablename__ = "animal_species"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False, unique=True)
    scientific_name = Column(String(200))
    conservation_level = Column(Enum(ConservationLevel), nullable=False, default=ConservationLevel.LEAST_CONCERN)
    description = Column(Text)
    
    animals = relationship("Animal", back_populates="species")
    feeding_standards = relationship("FeedingStandard", back_populates="species")


class FeedingStandard(Base):
    __tablename__ = "feeding_standards"
    
    id = Column(Integer, primary_key=True, index=True)
    species_id = Column(Integer, ForeignKey("animal_species.id"), nullable=False)
    min_weight = Column(Float, nullable=False)
    max_weight = Column(Float, nullable=False)
    feed_type = Column(String(100), nullable=False)
    daily_amount = Column(Float, nullable=False)
    frequency = Column(Integer, default=2)
    is_special = Column(Boolean, default=False)
    for_health_status = Column(Enum(AnimalStatus), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    species = relationship("AnimalSpecies", back_populates="feeding_standards")


class Animal(Base):
    __tablename__ = "animals"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    species_id = Column(Integer, ForeignKey("animal_species.id"), nullable=False)
    zone_id = Column(Integer, ForeignKey("zones.id"), nullable=False)
    chip_id = Column(String(50), unique=True)
    gender = Column(String(10))
    birth_date = Column(Date)
    weight = Column(Float, nullable=False)
    health_status = Column(Enum(AnimalStatus), nullable=False, default=AnimalStatus.HEALTHY)
    is_isolation = Column(Boolean, default=False)
    isolation_start_date = Column(Date, nullable=True)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    species = relationship("AnimalSpecies", back_populates="animals")
    zone = relationship("Zone", back_populates="animals")
    feeding_plans = relationship("FeedingPlan", back_populates="animal")
    feeding_records = relationship("FeedingRecord", back_populates="animal")
    medical_records = relationship("MedicalRecord", back_populates="animal")
    vaccination_records = relationship("VaccinationRecord", back_populates="animal")


class Feed(Base):
    __tablename__ = "feeds"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False, unique=True)
    unit = Column(String(20), nullable=False)
    current_stock = Column(Float, nullable=False, default=0)
    safety_stock = Column(Float, nullable=False, default=10)
    supplier = Column(String(100))
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    feeding_records = relationship("FeedingRecord", back_populates="feed")


class FeedingPlan(Base):
    __tablename__ = "feeding_plans"
    
    id = Column(Integer, primary_key=True, index=True)
    animal_id = Column(Integer, ForeignKey("animals.id"), nullable=False)
    feed_type = Column(String(100), nullable=False)
    daily_amount = Column(Float, nullable=False)
    frequency = Column(Integer, default=2)
    is_active = Column(Boolean, default=True)
    start_date = Column(Date, default=date.today)
    end_date = Column(Date, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    animal = relationship("Animal", back_populates="feeding_plans")


class ShiftAssignment(Base):
    __tablename__ = "shift_assignments"
    
    id = Column(Integer, primary_key=True, index=True)
    zone_id = Column(Integer, ForeignKey("zones.id"), nullable=False)
    employee_id = Column(Integer, ForeignKey("employees.id"), nullable=False)
    shift_date = Column(Date, nullable=False)
    shift = Column(Enum(FeedingShift), nullable=False)
    notes = Column(String(255))
    created_at = Column(DateTime, default=datetime.utcnow)
    
    zone = relationship("Zone", back_populates="shift_assignments")
    employee = relationship("Employee", back_populates="shift_assignments")


class FeedingRecord(Base):
    __tablename__ = "feeding_records"
    
    id = Column(Integer, primary_key=True, index=True)
    animal_id = Column(Integer, ForeignKey("animals.id"), nullable=False)
    feed_id = Column(Integer, ForeignKey("feeds.id"), nullable=False)
    feeder_id = Column(Integer, ForeignKey("employees.id"), nullable=True)
    feeding_date = Column(Date, nullable=False, default=date.today)
    shift = Column(Enum(FeedingShift), nullable=False)
    planned_amount = Column(Float, nullable=False)
    actual_amount = Column(Float, nullable=False)
    status = Column(Enum(FeedingStatus), nullable=False, default=FeedingStatus.NORMAL)
    notes = Column(String(255))
    created_at = Column(DateTime, default=datetime.utcnow)
    
    animal = relationship("Animal", back_populates="feeding_records")
    feed = relationship("Feed", back_populates="feeding_records")
    feeder = relationship("Employee", back_populates="feeding_records")


class MedicalRecord(Base):
    __tablename__ = "medical_records"
    
    id = Column(Integer, primary_key=True, index=True)
    animal_id = Column(Integer, ForeignKey("animals.id"), nullable=False)
    veterinarian_id = Column(Integer, ForeignKey("employees.id"), nullable=True)
    diagnosis = Column(String(255), nullable=False)
    symptoms = Column(Text)
    treatment = Column(Text)
    record_date = Column(Date, nullable=False, default=date.today)
    is_upgraded = Column(Boolean, default=False)
    next_check_date = Column(Date, nullable=True)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    animal = relationship("Animal", back_populates="medical_records")
    veterinarian = relationship("Employee", back_populates="medical_records")
    medications = relationship("MedicationRecord", back_populates="medical_record")


class MedicationRecord(Base):
    __tablename__ = "medication_records"
    
    id = Column(Integer, primary_key=True, index=True)
    medical_record_id = Column(Integer, ForeignKey("medical_records.id"), nullable=False)
    medication_name = Column(String(100), nullable=False)
    dosage = Column(String(50), nullable=False)
    frequency = Column(String(50))
    start_date = Column(Date, nullable=False, default=date.today)
    end_date = Column(Date, nullable=True)
    notes = Column(String(255))
    created_at = Column(DateTime, default=datetime.utcnow)
    
    medical_record = relationship("MedicalRecord", back_populates="medications")


class VaccinationRecord(Base):
    __tablename__ = "vaccination_records"
    
    id = Column(Integer, primary_key=True, index=True)
    animal_id = Column(Integer, ForeignKey("animals.id"), nullable=False)
    vaccine_name = Column(String(100), nullable=False)
    vaccine_batch = Column(String(50))
    vaccination_date = Column(Date, nullable=False, default=date.today)
    next_vaccination_date = Column(Date, nullable=False)
    min_interval_days = Column(Integer, nullable=False, default=30)
    veterinarian_name = Column(String(100))
    notes = Column(String(255))
    created_at = Column(DateTime, default=datetime.utcnow)
    
    animal = relationship("Animal", back_populates="vaccination_records")


class Todo(Base):
    __tablename__ = "todos"
    
    id = Column(Integer, primary_key=True, index=True)
    todo_type = Column(Enum(TodoType), nullable=False)
    status = Column(Enum(TodoStatus), nullable=False, default=TodoStatus.PENDING)
    title = Column(String(255), nullable=False)
    description = Column(Text)
    related_id = Column(Integer)
    related_type = Column(String(50))
    due_date = Column(Date, nullable=True)
    completed_at = Column(DateTime, nullable=True)
    completed_by = Column(String(100), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
