from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Boolean, Text
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from server.database import Base
from server.enums import (
    AgentServiceStage,
    StageStatus,
    BerthStatus,
    BerthApplicationStatus,
    SupplyType,
    SupplyStatus,
    WasteStatus,
    TodoStatus,
    SettlementStatus,
)


class AgentService(Base):
    __tablename__ = "agent_services"

    id = Column(Integer, primary_key=True, index=True)
    ship_name = Column(String(255), nullable=False)
    imo_number = Column(String(50), nullable=False)
    captain_name = Column(String(255), nullable=False)
    arrival_time = Column(DateTime, nullable=False)
    departure_time = Column(DateTime, nullable=True)
    current_stage = Column(String(50), default=AgentServiceStage.ORDER_RECEIVED.value, nullable=False)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())

    stage_history = relationship("StageHistory", back_populates="agent_service", order_by="StageHistory.id")
    berth_application = relationship("BerthApplication", back_populates="agent_service", uselist=False)
    supplies = relationship("Supply", back_populates="agent_service")
    waste_recovery = relationship("WasteRecovery", back_populates="agent_service", uselist=False)
    todos = relationship("Todo", back_populates="agent_service", order_by="Todo.due_time")
    fee_settlement = relationship("FeeSettlement", back_populates="agent_service", uselist=False)


class StageHistory(Base):
    __tablename__ = "stage_history"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    stage = Column(String(50), nullable=False)
    status = Column(String(50), default=StageStatus.PENDING.value, nullable=False)
    completed_at = Column(DateTime, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, server_default=func.now())

    agent_service = relationship("AgentService", back_populates="stage_history")


class Berth(Base):
    __tablename__ = "berths"

    id = Column(Integer, primary_key=True, index=True)
    berth_number = Column(String(50), unique=True, nullable=False)
    capacity = Column(Integer, nullable=False)
    status = Column(String(50), default=BerthStatus.AVAILABLE.value, nullable=False)


class BerthApplication(Base):
    __tablename__ = "berth_applications"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    requested_berthing_time = Column(DateTime, nullable=False)
    expected_duration_hours = Column(Integer, nullable=False)
    berth_preference = Column(String(255), nullable=True)
    status = Column(String(50), default=BerthApplicationStatus.PENDING.value, nullable=False)
    assigned_berth_id = Column(Integer, ForeignKey("berths.id"), nullable=True)
    assigned_berth_time = Column(DateTime, nullable=True)
    is_waiting_timeout = Column(Boolean, default=False, nullable=False)
    approved_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, server_default=func.now())

    agent_service = relationship("AgentService", back_populates="berth_application")
    assigned_berth = relationship("Berth")


class Supply(Base):
    __tablename__ = "supplies"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    supply_type = Column(String(50), nullable=False)
    quantity = Column(Float, nullable=False)
    unit_price = Column(Float, nullable=False)
    total_price = Column(Float, nullable=False)
    scheduled_time = Column(DateTime, nullable=False)
    status = Column(String(50), default=SupplyStatus.SCHEDULED.value, nullable=False)
    delivered_at = Column(DateTime, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, server_default=func.now())

    agent_service = relationship("AgentService", back_populates="supplies")


class WasteRecovery(Base):
    __tablename__ = "waste_recovery"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    total_weight_kg = Column(Float, nullable=False)
    unit_price_per_kg = Column(Float, nullable=False)
    discounted_weight_kg = Column(Float, nullable=True)
    discounted_unit_price = Column(Float, nullable=True)
    total_fee = Column(Float, nullable=False)
    scheduled_time = Column(DateTime, nullable=False)
    status = Column(String(50), default=WasteStatus.SCHEDULED.value, nullable=False)
    completed_at = Column(DateTime, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, server_default=func.now())

    agent_service = relationship("AgentService", back_populates="waste_recovery")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    title = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    due_time = Column(DateTime, nullable=False)
    status = Column(String(50), default=TodoStatus.PENDING.value, nullable=False)
    completed_at = Column(DateTime, nullable=True)
    sequence = Column(Integer, nullable=False)
    created_at = Column(DateTime, server_default=func.now())

    agent_service = relationship("AgentService", back_populates="todos")


class FeeSettlement(Base):
    __tablename__ = "fee_settlements"

    id = Column(Integer, primary_key=True, index=True)
    agent_service_id = Column(Integer, ForeignKey("agent_services.id"), nullable=False)
    agent_fee = Column(Float, nullable=False, default=0)
    supplies_fee = Column(Float, nullable=False, default=0)
    waste_recovery_fee = Column(Float, nullable=False, default=0)
    other_fees = Column(Float, nullable=False, default=0)
    total_amount = Column(Float, nullable=False, default=0)
    final_amount = Column(Float, nullable=False, default=0)
    status = Column(String(50), default=SettlementStatus.PENDING.value, nullable=False)
    settled_at = Column(DateTime, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, server_default=func.now())

    agent_service = relationship("AgentService", back_populates="fee_settlement")
