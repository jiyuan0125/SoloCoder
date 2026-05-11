from datetime import date, datetime
from enum import Enum as PyEnum
from sqlalchemy import (
    Column,
    Integer,
    String,
    Float,
    Date,
    DateTime,
    ForeignKey,
    Boolean,
    UniqueConstraint,
)
from sqlalchemy.orm import relationship, DeclarativeBase


class Base(DeclarativeBase):
    pass


class Store(Base):
    __tablename__ = "stores"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False)
    address = Column(String(255), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    ingredients = relationship("Ingredient", back_populates="store")
    purchase_orders = relationship("PurchaseOrder", back_populates="store")
    wastages = relationship("Wastage", back_populates="store")
    alerts = relationship("LowStockAlert", back_populates="store")


class Ingredient(Base):
    __tablename__ = "ingredients"
    __table_args__ = (
        UniqueConstraint("store_id", "name", name="uq_ingredient_store_name"),
    )

    id = Column(Integer, primary_key=True, index=True)
    store_id = Column(Integer, ForeignKey("stores.id"), nullable=False)
    name = Column(String(100), nullable=False)
    stock_quantity = Column(Float, nullable=False, default=0.0)
    unit_price = Column(Float, nullable=False, default=0.0)
    expiry_date = Column(Date, nullable=True)
    safety_stock = Column(Float, nullable=False, default=0.0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    store = relationship("Store", back_populates="ingredients")
    purchase_orders = relationship("PurchaseOrder", back_populates="ingredient")
    wastages = relationship("Wastage", back_populates="ingredient")
    alerts = relationship("LowStockAlert", back_populates="ingredient")


class PurchaseOrderStatus(str, PyEnum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    RECEIVED = "received"


class PurchaseOrder(Base):
    __tablename__ = "purchase_orders"

    id = Column(Integer, primary_key=True, index=True)
    store_id = Column(Integer, ForeignKey("stores.id"), nullable=False)
    ingredient_id = Column(Integer, ForeignKey("ingredients.id"), nullable=False)
    requested_quantity = Column(Float, nullable=False)
    received_quantity = Column(Float, nullable=True)
    expected_arrival_date = Column(Date, nullable=False)
    actual_arrival_date = Column(Date, nullable=True)
    status = Column(String(20), nullable=False, default=PurchaseOrderStatus.PENDING.value)
    remarks = Column(String(255), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    store = relationship("Store", back_populates="purchase_orders")
    ingredient = relationship("Ingredient", back_populates="purchase_orders")


class Wastage(Base):
    __tablename__ = "wastages"

    id = Column(Integer, primary_key=True, index=True)
    store_id = Column(Integer, ForeignKey("stores.id"), nullable=False)
    ingredient_id = Column(Integer, ForeignKey("ingredients.id"), nullable=False)
    quantity = Column(Float, nullable=False)
    actual_deducted = Column(Float, nullable=False)
    wastage_date = Column(Date, nullable=False, default=date.today)
    reason = Column(String(255), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    store = relationship("Store", back_populates="wastages")
    ingredient = relationship("Ingredient", back_populates="wastages")


class LowStockAlert(Base):
    __tablename__ = "low_stock_alerts"
    __table_args__ = (
        UniqueConstraint("ingredient_id", "alert_date", name="uq_alert_ingredient_date"),
    )

    id = Column(Integer, primary_key=True, index=True)
    store_id = Column(Integer, ForeignKey("stores.id"), nullable=False)
    ingredient_id = Column(Integer, ForeignKey("ingredients.id"), nullable=False)
    alert_date = Column(Date, nullable=False, default=date.today)
    current_stock = Column(Float, nullable=False)
    safety_stock = Column(Float, nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    store = relationship("Store", back_populates="alerts")
    ingredient = relationship("Ingredient", back_populates="alerts")
