from sqlalchemy import Column, Integer, String, DateTime, Index
from sqlalchemy.sql import func
from database import Base


class QuotaLimit(Base):
    __tablename__ = "quota_limits"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(String, index=True)
    resource_type = Column(String, index=True)
    limit = Column(Integer, default=0)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    __table_args__ = (
        Index("idx_user_resource_quota", "user_id", "resource_type", unique=True),
    )


class QuotaReservation(Base):
    __tablename__ = "quota_reservations"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(String, index=True)
    resource_type = Column(String, index=True)
    reserved_amount = Column(Integer, default=0)
    reserved_used = Column(Integer, default=0)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    __table_args__ = (
        Index("idx_user_resource_reservation", "user_id", "resource_type", unique=True),
    )


class QuotaUsage(Base):
    __tablename__ = "quota_usage"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(String, index=True)
    resource_type = Column(String, index=True)
    usage = Column(Integer, default=0)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    __table_args__ = (
        Index("idx_user_resource_usage", "user_id", "resource_type", unique=True),
    )


class QuotaLog(Base):
    __tablename__ = "quota_logs"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(String, index=True)
    resource_type = Column(String, index=True)
    action = Column(String, index=True)
    amount = Column(Integer, default=1)
    result = Column(String, index=True)
    timestamp = Column(DateTime, server_default=func.now(), index=True)
