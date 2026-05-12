from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Text, Boolean
from sqlalchemy.orm import relationship

from server.config import Base


HAZARDOUS_CATEGORIES = {
    1: "一类：爆炸品",
    2: "二类：气体",
    3: "三类：易燃液体",
    4: "四类：易燃固体、易于自燃的物质、遇水放出易燃气体的物质",
    5: "五类：氧化性物质和有机过氧化物",
    6: "六类：毒性物质和感染性物质",
    7: "七类：放射性物质",
    8: "八类：腐蚀性物质",
    9: "九类：杂项危险物质和物品",
}

NEEDS_DOUBLE_REVIEW = {1, 7}

DECLARATION_STATUS = {
    "submitted": "已提交",
    "initial_review_pending": "初审待审",
    "initial_review_approved": "初审通过",
    "initial_review_rejected": "初审驳回",
    "final_review_pending": "复审待审",
    "final_review_approved": "复审通过",
    "final_review_rejected": "复审驳回",
    "approved": "审核通过",
    "loading": "装卸中",
    "completed": "已完成",
    "closed": "已关闭",
}

EMERGENCY_LEVELS = {
    1: "一级（重大）",
    2: "二级（较大）",
    3: "三级（一般）",
}


class Ship(Base):
    __tablename__ = "ships"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, index=True, nullable=False)
    imo_number = Column(String(20), unique=True, index=True, nullable=False)
    flag = Column(String(50), nullable=False)
    has_hazardous_qualification = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    declarations = relationship("Declaration", back_populates="ship")


class Declaration(Base):
    __tablename__ = "declarations"

    id = Column(Integer, primary_key=True, index=True)
    declaration_number = Column(String(50), unique=True, index=True, nullable=False)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=False)
    voyage_number = Column(String(50), nullable=False)
    hazardous_category = Column(Integer, nullable=False)
    cargo_name = Column(String(200), nullable=False)
    cargo_quantity = Column(Float, nullable=False)
    packaging_compliant = Column(Boolean, default=False)
    submitted_by = Column(String(100), nullable=False)
    status = Column(String(50), default="submitted", nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    ship = relationship("Ship", back_populates="declarations")
    reviews = relationship("Review", back_populates="declaration", order_by="Review.id")
    loading = relationship("Loading", back_populates="declaration", uselist=False)
    emergencies = relationship("Emergency", back_populates="declaration", order_by="Emergency.id")
    audit_logs = relationship("AuditLog", back_populates="declaration", order_by="AuditLog.id")


class Review(Base):
    __tablename__ = "reviews"

    id = Column(Integer, primary_key=True, index=True)
    declaration_id = Column(Integer, ForeignKey("declarations.id"), nullable=False)
    review_type = Column(String(20), nullable=False)
    reviewer = Column(String(100), nullable=False)
    approved = Column(Boolean, nullable=False)
    comments = Column(Text, nullable=True)
    reviewed_at = Column(DateTime, default=datetime.utcnow)

    declaration = relationship("Declaration", back_populates="reviews")


class Loading(Base):
    __tablename__ = "loadings"

    id = Column(Integer, primary_key=True, index=True)
    declaration_id = Column(Integer, ForeignKey("declarations.id"), nullable=False)
    operator = Column(String(100), nullable=False)
    temperature = Column(Float, nullable=True)
    radiation_dose_rate = Column(Float, nullable=True)
    start_time = Column(DateTime, default=datetime.utcnow)
    end_time = Column(DateTime, nullable=True)
    completed = Column(Boolean, default=False)
    notes = Column(Text, nullable=True)

    declaration = relationship("Declaration", back_populates="loading")


class Emergency(Base):
    __tablename__ = "emergencies"

    id = Column(Integer, primary_key=True, index=True)
    declaration_id = Column(Integer, ForeignKey("declarations.id"), nullable=False)
    category = Column(Integer, nullable=False)
    impact_range = Column(String(50), nullable=False)
    level = Column(Integer, nullable=False)
    plan = Column(Text, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    triggered_by = Column(String(100), nullable=True)

    declaration = relationship("Declaration", back_populates="emergencies")


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id = Column(Integer, primary_key=True, index=True)
    declaration_id = Column(Integer, ForeignKey("declarations.id"), nullable=True)
    action = Column(String(100), nullable=False)
    actor = Column(String(100), nullable=False)
    details = Column(Text, nullable=True)
    timestamp = Column(DateTime, default=datetime.utcnow)

    declaration = relationship("Declaration", back_populates="audit_logs")
