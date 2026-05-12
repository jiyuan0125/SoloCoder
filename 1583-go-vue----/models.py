from sqlalchemy import Column, Integer, String, Text, DateTime, Float, ForeignKey, Enum
from sqlalchemy.orm import relationship
from datetime import datetime
import enum
from database import Base


class InquiryRecord(Base):
    __tablename__ = "inquiry_records"
    
    id = Column(Integer, primary_key=True, index=True)
    record_number = Column(String(20), unique=True, index=True, nullable=False)
    record_date = Column(String(10), nullable=False)
    sequence_number = Column(Integer, nullable=False)
    
    inquirer_name = Column(String(100))
    inquirer_phone = Column(String(20))
    content = Column(Text, nullable=False)
    response = Column(Text)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class ItemStatus(enum.Enum):
    PENDING = "待认领"
    MATCHED = "已匹配"
    CLAIMED = "已认领"
    STALE = "待处理"


class FoundItem(Base):
    __tablename__ = "found_items"
    
    id = Column(Integer, primary_key=True, index=True)
    item_type = Column(String(50), nullable=False)
    description = Column(Text, nullable=False)
    location = Column(String(200), nullable=False)
    value = Column(Float, default=0)
    finder_name = Column(String(100))
    finder_phone = Column(String(20))
    found_time = Column(DateTime, nullable=False)
    
    status = Column(Enum(ItemStatus), default=ItemStatus.PENDING)
    is_secondary_confirmed = Column(Integer, default=0)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class LostItem(Base):
    __tablename__ = "lost_items"
    
    id = Column(Integer, primary_key=True, index=True)
    item_type = Column(String(50), nullable=False)
    description = Column(Text)
    location = Column(String(200), nullable=False)
    value = Column(Float, default=0)
    owner_name = Column(String(100))
    owner_phone = Column(String(20))
    lost_time = Column(DateTime, nullable=False)
    
    status = Column(Enum(ItemStatus), default=ItemStatus.PENDING)
    is_secondary_confirmed = Column(Integer, default=0)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class ItemMatch(Base):
    __tablename__ = "item_matches"
    
    id = Column(Integer, primary_key=True, index=True)
    lost_item_id = Column(Integer, ForeignKey("lost_items.id"))
    found_item_id = Column(Integer, ForeignKey("found_items.id"))
    
    similarity_score = Column(Float, nullable=False)
    location_score = Column(Float, nullable=False)
    total_score = Column(Float, nullable=False)
    
    is_confirmed = Column(Integer, default=0)
    is_claimed = Column(Integer, default=0)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    
    lost_item = relationship("LostItem", backref="matches")
    found_item = relationship("FoundItem", backref="matches")


class ComplaintStatus(enum.Enum):
    REGISTERED = "登记"
    PROCESSING = "处理中"
    PENDING_CONFIRMATION = "待旅客确认"
    CLOSED = "已关闭"


class ComplaintSeverity(enum.Enum):
    NORMAL = "一般投诉"
    SERIOUS = "严重投诉"


class Complaint(Base):
    __tablename__ = "complaints"
    
    id = Column(Integer, primary_key=True, index=True)
    
    complainant_name = Column(String(100))
    complainant_phone = Column(String(20))
    
    content = Column(Text, nullable=False)
    severity = Column(Enum(ComplaintSeverity), default=ComplaintSeverity.NORMAL)
    
    status = Column(Enum(ComplaintStatus), default=ComplaintStatus.REGISTERED)
    current_handler = Column(String(100))
    
    response_content = Column(Text)
    
    is_overdue = Column(Integer, default=0)
    supervisor_notified = Column(Integer, default=0)
    
    registered_at = Column(DateTime, default=datetime.utcnow)
    processing_at = Column(DateTime)
    pending_confirmation_at = Column(DateTime)
    closed_at = Column(DateTime)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
