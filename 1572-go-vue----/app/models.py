from sqlalchemy import Column, Integer, String, DateTime, Float, Boolean, ForeignKey, Text, Enum
from sqlalchemy.orm import relationship
from datetime import datetime
from app.database import Base
import enum


class TrainStatus(str, enum.Enum):
    IDLE = "idle"
    IN_SERVICE = "in_service"
    MAINTENANCE = "maintenance"


class RouteStatus(str, enum.Enum):
    SCHEDULED = "scheduled"
    RUNNING = "running"
    DELAYED = "delayed"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class CrewType(str, enum.Enum):
    DRIVER = "driver"
    CONDUCTOR = "conductor"
    ATTENDANT = "attendant"


class AlertType(str, enum.Enum):
    OVERTIME = "overtime"
    DELAY = "delay"
    MAINTENANCE = "maintenance"


class AlertStatus(str, enum.Enum):
    PENDING = "pending"
    ACKNOWLEDGED = "acknowledged"
    RESOLVED = "resolved"


class Train(Base):
    __tablename__ = "trains"

    id = Column(Integer, primary_key=True, index=True)
    train_number = Column(String(20), unique=True, index=True, nullable=False)
    train_type = Column(String(20), nullable=False)
    seat_capacity = Column(Integer, nullable=False)
    crew_quota = Column(Integer, nullable=False)
    status = Column(Enum(TrainStatus), default=TrainStatus.IDLE)
    current_route_id = Column(Integer, ForeignKey("routes.id"), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    routes = relationship("Route", back_populates="train", foreign_keys="Route.train_id")


class Route(Base):
    __tablename__ = "routes"

    id = Column(Integer, primary_key=True, index=True)
    route_code = Column(String(20), unique=True, index=True, nullable=False)
    train_id = Column(Integer, ForeignKey("trains.id"), nullable=False)
    departure_station = Column(String(50), nullable=False)
    arrival_station = Column(String(50), nullable=False)
    scheduled_departure = Column(DateTime, nullable=False)
    scheduled_arrival = Column(DateTime, nullable=False)
    actual_departure = Column(DateTime, nullable=True)
    actual_arrival = Column(DateTime, nullable=True)
    status = Column(Enum(RouteStatus), default=RouteStatus.SCHEDULED)
    delay_minutes = Column(Integer, default=0)
    crew_group_id = Column(Integer, ForeignKey("crew_groups.id"), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    train = relationship("Train", back_populates="routes", foreign_keys=[train_id])
    crew_group = relationship("CrewGroup", back_populates="routes")


class CrewMember(Base):
    __tablename__ = "crew_members"

    id = Column(Integer, primary_key=True, index=True)
    employee_id = Column(String(20), unique=True, index=True, nullable=False)
    name = Column(String(50), nullable=False)
    crew_type = Column(Enum(CrewType), nullable=False)
    phone = Column(String(20), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew_assignments = relationship("CrewAssignment", back_populates="crew_member")


class CrewGroup(Base):
    __tablename__ = "crew_groups"

    id = Column(Integer, primary_key=True, index=True)
    group_code = Column(String(20), unique=True, index=True, nullable=False)
    name = Column(String(100), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    routes = relationship("Route", back_populates="crew_group")
    assignments = relationship("CrewAssignment", back_populates="crew_group")


class CrewAssignment(Base):
    __tablename__ = "crew_assignments"

    id = Column(Integer, primary_key=True, index=True)
    crew_group_id = Column(Integer, ForeignKey("crew_groups.id"), nullable=False)
    crew_member_id = Column(Integer, ForeignKey("crew_members.id"), nullable=False)
    is_lead = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    crew_group = relationship("CrewGroup", back_populates="assignments")
    crew_member = relationship("CrewMember", back_populates="crew_assignments")


class DutyRecord(Base):
    __tablename__ = "duty_records"

    id = Column(Integer, primary_key=True, index=True)
    crew_member_id = Column(Integer, ForeignKey("crew_members.id"), nullable=False)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=True)
    work_duration_minutes = Column(Integer, default=0)
    rest_period_minutes = Column(Integer, default=0)
    is_completed = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)


class MonthlyWorkHours(Base):
    __tablename__ = "monthly_work_hours"

    id = Column(Integer, primary_key=True, index=True)
    crew_member_id = Column(Integer, ForeignKey("crew_members.id"), nullable=False)
    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)
    total_minutes = Column(Integer, default=0)
    max_limit_minutes = Column(Integer, default=9600)
    is_over_limit = Column(Boolean, default=False)
    last_recalculated = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class DelayAdjustment(Base):
    __tablename__ = "delay_adjustments"

    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    original_schedule_id = Column(Integer, nullable=True)
    delay_minutes = Column(Integer, nullable=False)
    adjustment_type = Column(String(50), nullable=False)
    adjustment_plan = Column(Text, nullable=False)
    notified_to = Column(String(50), default="值班调度")
    created_at = Column(DateTime, default=datetime.utcnow)


class RouteOptimization(Base):
    __tablename__ = "route_optimizations"

    id = Column(Integer, primary_key=True, index=True)
    train_id = Column(Integer, ForeignKey("trains.id"), nullable=False)
    optimization_type = Column(String(50), nullable=False)
    description = Column(Text, nullable=False)
    suggestion = Column(Text, nullable=False)
    avg_utilization_hours = Column(Float, nullable=True)
    avg_turnaround_minutes = Column(Float, nullable=True)
    is_implemented = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)


class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    alert_type = Column(Enum(AlertType), nullable=False)
    title = Column(String(100), nullable=False)
    message = Column(Text, nullable=False)
    target_role = Column(String(50), default="调度主任")
    status = Column(Enum(AlertStatus), default=AlertStatus.PENDING)
    related_crew_member_id = Column(Integer, ForeignKey("crew_members.id"), nullable=True)
    related_route_id = Column(Integer, ForeignKey("routes.id"), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    acknowledged_at = Column(DateTime, nullable=True)
