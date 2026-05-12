from datetime import datetime
from sqlalchemy import Column, Integer, String, DateTime, Boolean, Enum, Float, ForeignKey
from sqlalchemy.orm import relationship
import enum

from .database import Base


class CardType(str, enum.Enum):
    SINGLE = "单程票"
    STORED = "储值票"
    STUDENT = "学生票"
    ELDERLY = "老年票"


class CardStatus(str, enum.Enum):
    ACTIVE = "正常"
    LOCKED = "锁定"
    EXPIRED = "过期"
    CONSUMED = "已使用"


class Priority(str, enum.Enum):
    LOW = "低"
    MEDIUM = "中"
    HIGH = "高"


class FaultStatus(str, enum.Enum):
    PENDING = "待处理"
    PROCESSING = "处理中"
    RESOLVED = "已解决"


class Card(Base):
    __tablename__ = "cards"

    id = Column(Integer, primary_key=True, index=True)
    card_number = Column(String(20), unique=True, index=True, nullable=False)
    card_type = Column(Enum(CardType), nullable=False)
    balance = Column(Integer, default=0)
    status = Column(Enum(CardStatus), default=CardStatus.ACTIVE)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    transactions = relationship("Transaction", back_populates="card")


class Transaction(Base):
    __tablename__ = "transactions"

    id = Column(Integer, primary_key=True, index=True)
    card_id = Column(Integer, ForeignKey("cards.id"), nullable=False)
    card_number = Column(String(20), nullable=False)
    entry_station = Column(String(50), nullable=True)
    exit_station = Column(String(50), nullable=True)
    entry_time = Column(DateTime, nullable=True)
    exit_time = Column(DateTime, nullable=True)
    distance = Column(Float, default=0)
    base_fare = Column(Integer, default=0)
    discount = Column(String(20), default="无")
    final_fare = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)

    card = relationship("Card", back_populates="transactions")


class Fault(Base):
    __tablename__ = "faults"

    id = Column(Integer, primary_key=True, index=True)
    gate_id = Column(String(20), nullable=False)
    station = Column(String(50), nullable=False)
    description = Column(String(200), nullable=False)
    priority = Column(Enum(Priority), default=Priority.MEDIUM)
    status = Column(Enum(FaultStatus), default=FaultStatus.PENDING)
    reported_at = Column(DateTime, default=datetime.utcnow)
    resolved_at = Column(DateTime, nullable=True)
    escalated = Column(Boolean, default=False)
