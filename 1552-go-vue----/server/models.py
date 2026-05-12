from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Text, Boolean, Enum as SQLEnum
from sqlalchemy.orm import relationship
from enum import Enum
from .database import Base


class AgentServiceStatus(str, Enum):
    ACCEPTED = "accepted"
    DECLARED = "declared"
    BERTHING_ARRANGED = "berthing_arranged"
    OPERATION_EXECUTED = "operation_executed"
    SETTLED = "settled"
    DEPARTURE_CONFIRMED = "departure_confirmed"


class BerthingRequestStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    ASSIGNED = "assigned"


class MaterialType(str, Enum):
    FUEL = "fuel"
    FRESH_WATER = "fresh_water"


class DeliveryStatus(str, Enum):
    PENDING = "pending"
    DELIVERED = "delivered"


class WasteCollectionStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"


class TodoStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"


class AgentService(Base):
    __tablename__ = "agent_services"

    id = Column(Integer, primary_key=True, index=True)
    ship_name = Column(String(255), nullable=False)
    imo_number = Column(String(50))
    port_of_call = Column(String(100))
    arrival_time = Column(DateTime, nullable=False)
    estimated_departure_time = Column(DateTime)
    status = Column(SQLEnum(AgentServiceStatus), default=AgentServiceStatus.ACCEPTED, nullable=False)
    agency_fee = Column(Float, default=0)
    total_fee = Column(Float, default=0)
    material_fee = Column(Float, default=0)
    waste_fee = Column(Float, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    notes = Column(Text)

    berthing_request = relationship("BerthingRequest", back_populates="agent_service", uselist=False, cascade="all, delete-orphan")
    material_deliveries = relationship("MaterialDelivery", back_populates="agent_service", cascade="all, delete-orphan")
    waste_collections = relationship("WasteCollection", back_populates="agent_service", cascade="all, delete-orphan")
    todos = relationship("Todo", back_populates="agent_service", cascade="all, delete-orphan", order_by="Todo.due_time")
    settlements = relationship("Settlement", back_populates="agent_service", cascade="all, delete-orphan")


class Berth(Base):
    __tablename__ = "berths"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False, unique=True)
    location = Column(String(255))
    max_length = Column(Float)
    max_draft = Column(Float)
    is_available = Column(Boolean, default=True)
    description = Column(Text)


class BerthingRequest(Base):
    __tablename__ = "berthing_requests"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False, unique=True)
    requested_berthing_time = Column(DateTime, nullable=False)
    estimated_duration_hours = Column(Float, nullable=False)
    berth_preference = Column(String(255))
    status = Column(SQLEnum(BerthingRequestStatus), default=BerthingRequestStatus.PENDING, nullable=False)
    assigned_berth_id = Column(Integer, ForeignKey("berths.id"))
    actual_berthing_time = Column(DateTime)
    approved_by = Column(String(100))
    approval_time = Column(DateTime)
    is_timeout = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    agent_service = relationship("AgentService", back_populates="berthing_request")
    assigned_berth = relationship("Berth")


class MaterialDelivery(Base):
    __tablename__ = "material_deliveries"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    material_type = Column(SQLEnum(MaterialType), nullable=False)
    quantity_tons = Column(Float, nullable=False)
    unit_price = Column(Float, nullable=False)
    total_price = Column(Float, nullable=False)
    delivery_time = Column(DateTime)
    status = Column(SQLEnum(DeliveryStatus), default=DeliveryStatus.PENDING, nullable=False)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    agent_service = relationship("AgentService", back_populates="material_deliveries")


class WasteCollection(Base):
    __tablename__ = "waste_collections"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    waste_type = Column(String(100))
    weight_kg = Column(Float, nullable=False)
    unit_price = Column(Float, nullable=False)
    total_price = Column(Float, nullable=False)
    collection_time = Column(DateTime)
    status = Column(SQLEnum(WasteCollectionStatus), default=WasteCollectionStatus.PENDING, nullable=False)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    agent_service = relationship("AgentService", back_populates="waste_collections")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    title = Column(String(255), nullable=False)
    description = Column(Text)
    due_time = Column(DateTime, nullable=False)
    status = Column(SQLEnum(TodoStatus), default=TodoStatus.PENDING, nullable=False)
    completed_time = Column(DateTime)
    created_at = Column(DateTime, default=datetime.utcnow)

    agent_service = relationship("AgentService", back_populates="todos")


class Settlement(Base):
    __tablename__ = "settlements"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    agency_fee = Column(Float, default=0)
    material_fee = Column(Float, default=0)
    waste_fee = Column(Float, default=0)
    total_amount = Column(Float, default=0)
    settlement_time = Column(DateTime)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    agent_service = relationship("AgentService", back_populates="settlements")
