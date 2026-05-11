from datetime import datetime, date
from sqlalchemy import Column, Integer, String, Text, Date, DateTime, ForeignKey, Enum, Boolean
from sqlalchemy.orm import relationship
from server.database import Base
import enum

class RiverLevel(str, enum.Enum):
    PROVINCIAL = "省级"
    MUNICIPAL = "市级"
    COUNTY = "县区级"
    TOWNSHIP = "乡镇级"

class IssueType(str, enum.Enum):
    WATER_POLLUTION = "水污染"
    WATER_QUALITY = "水质问题"
    GARBAGE = "垃圾"
    ILLEGAL_BUILDING = "违建"
    POACHER = "非法捕捞"
    OTHER = "其他"

class IssueSeverity(str, enum.Enum):
    NORMAL = "一般"
    SERIOUS = "较重"
    SEVERE = "严重"

class IssueStatus(str, enum.Enum):
    DISCOVERED = "待整改"
    RECTIFYING = "整改中"
    PENDING_REVIEW = "待复核"
    REVIEWING = "复核中"
    CLOSED = "已关闭"

class River(Base):
    __tablename__ = "rivers"
    
    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(50), unique=True, index=True, nullable=False)
    name = Column(String(100), nullable=False)
    start_point = Column(String(200), nullable=False)
    end_point = Column(String(200), nullable=False)
    basin = Column(String(100), nullable=False)
    is_key_section = Column(Boolean, default=False)
    
    river_keepers = relationship("RiverKeeper", back_populates="river")
    patrols = relationship("Patrol", back_populates="river")
    issues = relationship("Issue", back_populates="river")
    reports = relationship("MonthlyReport", back_populates="river")

class RiverKeeper(Base):
    __tablename__ = "river_keepers"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    level = Column(Enum(RiverLevel), nullable=False)
    river_id = Column(Integer, ForeignKey("rivers.id"))
    contact = Column(String(100))
    
    river = relationship("River", back_populates="river_keepers")
    patrols = relationship("Patrol", back_populates="river_keeper")

class Patrol(Base):
    __tablename__ = "patrols"
    
    id = Column(Integer, primary_key=True, index=True)
    river_id = Column(Integer, ForeignKey("rivers.id"), nullable=False)
    keeper_id = Column(Integer, ForeignKey("river_keepers.id"), nullable=False)
    patrol_date = Column(Date, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    river = relationship("River", back_populates="patrols")
    river_keeper = relationship("RiverKeeper", back_populates="patrols")
    issues = relationship("Issue", back_populates="patrol")

class Issue(Base):
    __tablename__ = "issues"
    
    id = Column(Integer, primary_key=True, index=True)
    river_id = Column(Integer, ForeignKey("rivers.id"), nullable=False)
    patrol_id = Column(Integer, ForeignKey("patrols.id"))
    issue_type = Column(Enum(IssueType), nullable=False)
    severity = Column(Enum(IssueSeverity), nullable=False)
    description = Column(Text, nullable=False)
    status = Column(Enum(IssueStatus), default=IssueStatus.DISCOVERED)
    discovered_date = Column(Date, default=date.today)
    deadline = Column(Date, nullable=False)
    rectification_note = Column(Text)
    rectification_date = Column(Date)
    review_note = Column(Text)
    review_date = Column(Date)
    escalated = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    river = relationship("River", back_populates="issues")
    patrol = relationship("Patrol", back_populates="issues")
    notifications = relationship("Notification", back_populates="issue")

class Notification(Base):
    __tablename__ = "notifications"
    
    id = Column(Integer, primary_key=True, index=True)
    issue_id = Column(Integer, ForeignKey("issues.id"), nullable=False)
    message = Column(Text, nullable=False)
    recipient_level = Column(Enum(RiverLevel), nullable=False)
    sent_date = Column(Date, default=date.today)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    issue = relationship("Issue", back_populates="notifications")

class MonthlyReport(Base):
    __tablename__ = "monthly_reports"
    
    id = Column(Integer, primary_key=True, index=True)
    river_id = Column(Integer, ForeignKey("rivers.id"), nullable=False)
    report_month = Column(String(7), nullable=False)
    patrol_count = Column(Integer, default=0)
    issue_count = Column(Integer, default=0)
    resolved_issue_count = Column(Integer, default=0)
    content = Column(Text)
    submitted_date = Column(Date)
    deadline = Column(Date, nullable=False)
    is_submitted = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    river = relationship("River", back_populates="reports")
