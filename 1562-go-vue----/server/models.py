from datetime import datetime
from sqlalchemy import Column, Integer, String, DateTime, Float, ForeignKey, Enum, Boolean, Text
from sqlalchemy.orm import relationship
from server.database import Base
import enum


class TaskStatus(enum.Enum):
    CREATED = "created"
    ASSIGNED = "assigned"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class SubTaskType(enum.Enum):
    PASSENGER_DISEMBARK = "passenger_disembark"
    CLEANING = "cleaning"
    BAGGAGE_HANDLING = "baggage_handling"
    FUELING = "fueling"
    DEICING = "deicing"
    PASSENGER_BOARD = "passenger_board"


class FuelTruckStatus(enum.Enum):
    AVAILABLE = "available"
    IN_USE = "in_use"
    CHARGING = "charging"
    MAINTENANCE = "maintenance"


class BusStatus(enum.Enum):
    AVAILABLE = "available"
    IN_USE = "in_use"
    CHARGING = "charging"
    MAINTENANCE = "maintenance"


class Flight(Base):
    __tablename__ = "flights"

    id = Column(Integer, primary_key=True, index=True)
    flight_number = Column(String, index=True)
    aircraft_type = Column(String)
    passenger_count = Column(Integer)
    gate = Column(String)
    status = Column(String, default="scheduled")
    temperature = Column(Float, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    guarantee_tasks = relationship("GuaranteeTask", back_populates="flight")


class GuaranteeTask(Base):
    __tablename__ = "guarantee_tasks"

    id = Column(Integer, primary_key=True, index=True)
    flight_id = Column(Integer, ForeignKey("flights.id"))
    status = Column(Enum(TaskStatus), default=TaskStatus.CREATED)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    flight = relationship("Flight", back_populates="guarantee_tasks")
    subtasks = relationship("SubTask", back_populates="guarantee_task", cascade="all, delete-orphan")


class SubTask(Base):
    __tablename__ = "subtasks"

    id = Column(Integer, primary_key=True, index=True)
    guarantee_task_id = Column(Integer, ForeignKey("guarantee_tasks.id"))
    task_type = Column(Enum(SubTaskType))
    status = Column(Enum(TaskStatus), default=TaskStatus.CREATED)
    assigned_team_id = Column(Integer, nullable=True)
    start_time = Column(DateTime, nullable=True)
    end_time = Column(DateTime, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    guarantee_task = relationship("GuaranteeTask", back_populates="subtasks")
    dependencies = relationship(
        "TaskDependency",
        foreign_keys="[TaskDependency.subtask_id]",
        back_populates="subtask"
    )
    dependents = relationship(
        "TaskDependency",
        foreign_keys="[TaskDependency.dependency_id]",
        back_populates="dependency"
    )


class TaskDependency(Base):
    __tablename__ = "task_dependencies"

    id = Column(Integer, primary_key=True, index=True)
    subtask_id = Column(Integer, ForeignKey("subtasks.id"))
    dependency_id = Column(Integer, ForeignKey("subtasks.id"))

    subtask = relationship("SubTask", foreign_keys=[subtask_id], back_populates="dependencies")
    dependency = relationship("SubTask", foreign_keys=[dependency_id], back_populates="dependents")


class CleaningTeam(Base):
    __tablename__ = "cleaning_teams"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String)
    member_count = Column(Integer)
    is_available = Column(Boolean, default=True)
    current_flight_id = Column(Integer, nullable=True)


class FuelTruck(Base):
    __tablename__ = "fuel_trucks"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String)
    capacity = Column(Float)
    current_fuel = Column(Float)
    status = Column(Enum(FuelTruckStatus), default=FuelTruckStatus.AVAILABLE)


class ShuttleBus(Base):
    __tablename__ = "shuttle_buses"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String)
    capacity = Column(Integer)
    status = Column(Enum(BusStatus), default=BusStatus.AVAILABLE)


class FuelingTask(Base):
    __tablename__ = "fueling_tasks"

    id = Column(Integer, primary_key=True, index=True)
    subtask_id = Column(Integer, ForeignKey("subtasks.id"))
    planned_amount = Column(Float)
    actual_amount = Column(Float, nullable=True)
    tank_capacity = Column(Float)
    trucks_needed = Column(Integer, default=1)
    fuel_truck_ids = Column(String, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)


class DeicingTask(Base):
    __tablename__ = "deicing_tasks"

    id = Column(Integer, primary_key=True, index=True)
    subtask_id = Column(Integer, ForeignKey("subtasks.id"))
    deicing_type = Column(String)
    temperature = Column(Float, nullable=True)
    volume_used = Column(Float, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)


class BusAllocation(Base):
    __tablename__ = "bus_allocations"

    id = Column(Integer, primary_key=True, index=True)
    subtask_id = Column(Integer, ForeignKey("subtasks.id"))
    bus_ids = Column(String)
    passenger_count = Column(Integer)
    created_at = Column(DateTime, default=datetime.utcnow)
