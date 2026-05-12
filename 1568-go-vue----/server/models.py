from enum import Enum as PyEnum
from datetime import datetime, date, time
from sqlalchemy import Column, Integer, String, DateTime, Date, Time, ForeignKey, Enum, Boolean
from sqlalchemy.orm import relationship, DeclarativeBase


class ShiftType(PyEnum):
    MORNING = "morning"
    MIDDAY = "midday"
    NIGHT = "night"


class RoleType(PyEnum):
    SCANNER = "scanner"
    HANDCHECK = "handcheck"
    VERIFIER = "verifier"


class ContrabandType(PyEnum):
    EXPLOSIVE = "explosive"
    SIMULATED_WEAPON = "simulated_weapon"
    KNIFE = "knife"
    OTHER = "other"


class ChannelType(PyEnum):
    STANDARD = "standard"
    VIP = "vip"


class ChannelStatus(PyEnum):
    OPEN = "open"
    CLOSED = "closed"


class DisposalType(PyEnum):
    SELF_ABANDON = "self_abandon"
    HANDOVER = "handover"


class Base(DeclarativeBase):
    pass


class Officer(Base):
    __tablename__ = "officers"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    badge_number = Column(String(50), unique=True, nullable=False, index=True)
    role = Column(Enum(RoleType), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    schedules = relationship("Schedule", back_populates="officer")
    check_ins = relationship("CheckIn", back_populates="officer")


class Channel(Base):
    __tablename__ = "channels"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(50), nullable=False)
    channel_type = Column(Enum(ChannelType), default=ChannelType.STANDARD)
    status = Column(Enum(ChannelStatus), default=ChannelStatus.CLOSED)
    capacity_per_hour = Column(Integer, default=120)
    created_at = Column(DateTime, default=datetime.utcnow)

    schedules = relationship("Schedule", back_populates="channel")
    contrabands = relationship("Contraband", back_populates="channel")
    traffic = relationship("TrafficLog", back_populates="channel")


class Schedule(Base):
    __tablename__ = "schedules"

    id = Column(Integer, primary_key=True, index=True)
    officer_id = Column(Integer, ForeignKey("officers.id"), nullable=False, index=True)
    channel_id = Column(Integer, ForeignKey("channels.id"), nullable=False, index=True)
    schedule_date = Column(Date, nullable=False, index=True)
    shift = Column(Enum(ShiftType), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    officer = relationship("Officer", back_populates="schedules")
    channel = relationship("Channel", back_populates="schedules")
    check_ins = relationship("CheckIn", back_populates="schedule")


class CheckIn(Base):
    __tablename__ = "check_ins"

    id = Column(Integer, primary_key=True, index=True)
    officer_id = Column(Integer, ForeignKey("officers.id"), nullable=False, index=True)
    schedule_id = Column(Integer, ForeignKey("schedules.id"), nullable=True, index=True)
    check_in_time = Column(DateTime, default=datetime.utcnow, nullable=False)
    is_temporary = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    officer = relationship("Officer", back_populates="check_ins")
    schedule = relationship("Schedule", back_populates="check_ins")


class Contraband(Base):
    __tablename__ = "contrabands"

    id = Column(Integer, primary_key=True, index=True)
    channel_id = Column(Integer, ForeignKey("channels.id"), nullable=False, index=True)
    item_type = Column(Enum(ContrabandType), nullable=False)
    description = Column(String(255), nullable=True)
    disposal_type = Column(Enum(DisposalType), nullable=False)
    police_badge = Column(String(50), nullable=True)
    passenger_name = Column(String(100), nullable=True)
    recorded_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    channel = relationship("Channel", back_populates="contrabands")


class TrafficLog(Base):
    __tablename__ = "traffic_logs"

    id = Column(Integer, primary_key=True, index=True)
    channel_id = Column(Integer, ForeignKey("channels.id"), nullable=False, index=True)
    log_date = Column(Date, nullable=False, index=True)
    hour = Column(Integer, nullable=False)
    passenger_count = Column(Integer, default=0)
    queue_length = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)

    channel = relationship("Channel", back_populates="traffic")
