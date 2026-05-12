from datetime import datetime
from sqlalchemy import Column, Integer, String, DateTime, Boolean, Numeric, ForeignKey, Index
from sqlalchemy.orm import relationship

from .database import Base


class Station(Base):
    __tablename__ = "stations"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(50), unique=True, nullable=False, index=True)
    price_zone = Column(Integer, nullable=False)
    
    created_at = Column(DateTime, default=datetime.now)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)


class Rate(Base):
    __tablename__ = "rates"
    
    id = Column(Integer, primary_key=True, index=True)
    from_zone = Column(Integer, nullable=False)
    to_zone = Column(Integer, nullable=False)
    rate_per_ton_fen = Column(Integer, nullable=False)
    
    effective_from = Column(DateTime, nullable=False)
    effective_to = Column(DateTime, nullable=True)
    
    created_at = Column(DateTime, default=datetime.now)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)
    
    __table_args__ = (
        Index('idx_zone_pair_date', 'from_zone', 'to_zone', 'effective_from'),
    )


class Waybill(Base):
    __tablename__ = "waybills"
    
    id = Column(Integer, primary_key=True, index=True)
    waybill_no = Column(String(30), unique=True, nullable=False, index=True)
    
    from_station_id = Column(Integer, ForeignKey('stations.id'), nullable=False)
    to_station_id = Column(Integer, ForeignKey('stations.id'), nullable=False)
    from_station_name = Column(String(50), nullable=False)
    to_station_name = Column(String(50), nullable=False)
    
    weight_ton = Column(Numeric(3, 1), nullable=False)
    is_hazardous = Column(Boolean, default=False, nullable=False)
    customer_code = Column(String(20), nullable=True, index=True)
    
    basic_rate_fen = Column(Integer, nullable=False)
    basic_charge_fen = Column(Integer, nullable=False)
    hazardous_surcharge_fen = Column(Integer, default=0, nullable=False)
    discount_fen = Column(Integer, default=0, nullable=False)
    total_charge_fen = Column(Integer, nullable=False)
    
    rate_id = Column(Integer, ForeignKey('rates.id'), nullable=False)
    original_charge_fen = Column(Integer, nullable=False)
    
    status = Column(String(20), default='issued', nullable=False, index=True)
    settled_at = Column(DateTime, nullable=True)
    
    created_at = Column(DateTime, default=datetime.now, index=True)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)
    
    from_station = relationship("Station", foreign_keys=[from_station_id])
    to_station = relationship("Station", foreign_keys=[to_station_id])
    rate = relationship("Rate")
    
    __table_args__ = (
        Index('idx_customer_created', 'customer_code', 'created_at'),
        Index('idx_status_created', 'status', 'created_at'),
    )


class RateAdjustmentRecord(Base):
    __tablename__ = "rate_adjustment_records"
    
    id = Column(Integer, primary_key=True, index=True)
    waybill_id = Column(Integer, ForeignKey('waybills.id'), nullable=False, index=True)
    waybill_no = Column(String(30), nullable=False, index=True)
    
    old_rate_fen = Column(Integer, nullable=False)
    new_rate_fen = Column(Integer, nullable=False)
    adjustment_fen = Column(Integer, nullable=False)
    
    old_basic_charge_fen = Column(Integer, nullable=False)
    new_basic_charge_fen = Column(Integer, nullable=False)
    old_total_charge_fen = Column(Integer, nullable=False)
    new_total_charge_fen = Column(Integer, nullable=False)
    
    adjustment_date = Column(DateTime, default=datetime.now, nullable=False)
    
    waybill = relationship("Waybill")


class MonthlyCustomerStat(Base):
    __tablename__ = "monthly_customer_stats"
    
    id = Column(Integer, primary_key=True, index=True)
    customer_code = Column(String(20), nullable=False, index=True)
    year_month = Column(String(7), nullable=False, index=True)
    
    total_count = Column(Integer, default=0, nullable=False)
    total_amount_fen = Column(Integer, default=0, nullable=False)
    settled_amount_fen = Column(Integer, default=0, nullable=False)
    
    created_at = Column(DateTime, default=datetime.now)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)
    
    __table_args__ = (
        Index('idx_customer_month', 'customer_code', 'year_month', unique=True),
    )
