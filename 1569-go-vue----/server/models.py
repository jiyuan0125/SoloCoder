import datetime
from enum import Enum as PyEnum
from sqlalchemy import (
    Column,
    Integer,
    String,
    DateTime,
    ForeignKey,
    Date,
    Enum,
    Boolean,
    Text,
)
from sqlalchemy.orm import relationship
from server.database import Base


class ProductCategory(PyEnum):
    COSMETICS = "cosmetics"
    TOBACCO = "tobacco"
    ALCOHOL = "alcohol"
    OTHER = "other"


class SaleStatus(PyEnum):
    PENDING = "pending"
    CONFIRMED = "confirmed"
    SETTLED = "settled"
    CANCELLED = "cancelled"


class RefundStatus(PyEnum):
    APPLIED = "applied"
    APPROVED = "approved"
    REJECTED = "rejected"
    REFUNDED = "refunded"
    COMPLETED = "completed"


class RefundType(PyEnum):
    FREE_RETURN = "free_return"
    NEEDS_APPROVAL = "needs_approval"


class SettlementStatus(PyEnum):
    DRAFT = "draft"
    CONFIRMED = "confirmed"


class Shop(Base):
    __tablename__ = "shops"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False, index=True)
    location = Column(String(200), nullable=False)
    manager_name = Column(String(100), nullable=True)
    phone = Column(String(20), nullable=True)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.datetime.utcnow, onupdate=datetime.datetime.utcnow)

    products = relationship("Product", back_populates="shop")
    sales = relationship("Sale", back_populates="shop")
    settlements = relationship("DailySettlement", back_populates="shop")


class Product(Base):
    __tablename__ = "products"

    id = Column(Integer, primary_key=True, index=True)
    shop_id = Column(Integer, ForeignKey("shops.id"), nullable=False, index=True)
    name = Column(String(150), nullable=False, index=True)
    sku = Column(String(50), nullable=False, unique=True, index=True)
    brand = Column(String(100), nullable=True)
    category = Column(Enum(ProductCategory), nullable=False, default=ProductCategory.OTHER)
    retail_price = Column(Integer, nullable=False)
    duty_free_price = Column(Integer, nullable=False)
    description = Column(Text, nullable=True)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.datetime.utcnow, onupdate=datetime.datetime.utcnow)

    shop = relationship("Shop", back_populates="products")
    inventory_batches = relationship("InventoryBatch", back_populates="product")
    sale_items = relationship("SaleItem", back_populates="product")


class InventoryBatch(Base):
    __tablename__ = "inventory_batches"

    id = Column(Integer, primary_key=True, index=True)
    product_id = Column(Integer, ForeignKey("products.id"), nullable=False, index=True)
    batch_number = Column(String(50), nullable=False, index=True)
    quantity = Column(Integer, nullable=False)
    available_quantity = Column(Integer, nullable=False)
    expiry_date = Column(Date, nullable=False, index=True)
    received_date = Column(DateTime, default=datetime.datetime.utcnow)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)

    product = relationship("Product", back_populates="inventory_batches")
    sale_item_batches = relationship("SaleItemBatch", back_populates="batch")


class Passenger(Base):
    __tablename__ = "passengers"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    passport_number = Column(String(50), nullable=False, unique=True, index=True)
    flight_number = Column(String(20), nullable=False)
    flight_departure_time = Column(DateTime, nullable=False)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)

    sales = relationship("Sale", back_populates="passenger")


class Sale(Base):
    __tablename__ = "sales"

    id = Column(Integer, primary_key=True, index=True)
    shop_id = Column(Integer, ForeignKey("shops.id"), nullable=False, index=True)
    passenger_id = Column(Integer, ForeignKey("passengers.id"), nullable=False, index=True)
    sale_number = Column(String(50), nullable=False, unique=True, index=True)
    total_amount = Column(Integer, nullable=False)
    payment_method = Column(String(50), nullable=False)
    status = Column(Enum(SaleStatus), nullable=False, default=SaleStatus.CONFIRMED)
    sale_time = Column(DateTime, default=datetime.datetime.utcnow)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)

    shop = relationship("Shop", back_populates="sales")
    passenger = relationship("Passenger", back_populates="sales")
    items = relationship("SaleItem", back_populates="sale", cascade="all, delete-orphan")
    refund = relationship("RefundRequest", back_populates="sale", uselist=False)


class SaleItem(Base):
    __tablename__ = "sale_items"

    id = Column(Integer, primary_key=True, index=True)
    sale_id = Column(Integer, ForeignKey("sales.id"), nullable=False, index=True)
    product_id = Column(Integer, ForeignKey("products.id"), nullable=False, index=True)
    quantity = Column(Integer, nullable=False)
    unit_price = Column(Integer, nullable=False)
    subtotal = Column(Integer, nullable=False)

    sale = relationship("Sale", back_populates="items")
    product = relationship("Product", back_populates="sale_items")
    batches = relationship("SaleItemBatch", back_populates="sale_item", cascade="all, delete-orphan")


class SaleItemBatch(Base):
    __tablename__ = "sale_item_batches"

    id = Column(Integer, primary_key=True, index=True)
    sale_item_id = Column(Integer, ForeignKey("sale_items.id"), nullable=False, index=True)
    batch_id = Column(Integer, ForeignKey("inventory_batches.id"), nullable=False, index=True)
    quantity = Column(Integer, nullable=False)
    is_returned = Column(Boolean, default=False)
    returned_quantity = Column(Integer, default=0)

    sale_item = relationship("SaleItem", back_populates="batches")
    batch = relationship("InventoryBatch", back_populates="sale_item_batches")


class DailySettlement(Base):
    __tablename__ = "daily_settlements"

    id = Column(Integer, primary_key=True, index=True)
    shop_id = Column(Integer, ForeignKey("shops.id"), nullable=False, index=True)
    settlement_date = Column(Date, nullable=False, index=True)
    status = Column(Enum(SettlementStatus), nullable=False, default=SettlementStatus.DRAFT)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    confirmed_at = Column(DateTime, nullable=True)

    shop = relationship("Shop", back_populates="settlements")
    summaries = relationship("SettlementSummary", back_populates="settlement", cascade="all, delete-orphan")


class SettlementSummary(Base):
    __tablename__ = "settlement_summaries"

    id = Column(Integer, primary_key=True, index=True)
    settlement_id = Column(Integer, ForeignKey("daily_settlements.id"), nullable=False, index=True)
    payment_method = Column(String(50), nullable=False)
    total_amount = Column(Integer, nullable=False)
    transaction_count = Column(Integer, nullable=False)

    settlement = relationship("DailySettlement", back_populates="summaries")


class RefundRequest(Base):
    __tablename__ = "refund_requests"

    id = Column(Integer, primary_key=True, index=True)
    sale_id = Column(Integer, ForeignKey("sales.id"), nullable=False, unique=True, index=True)
    refund_number = Column(String(50), nullable=False, unique=True, index=True)
    refund_type = Column(Enum(RefundType), nullable=False)
    status = Column(Enum(RefundStatus), nullable=False, default=RefundStatus.APPLIED)
    reason = Column(Text, nullable=True)
    is_sealed = Column(Boolean, default=True)
    total_refund_amount = Column(Integer, nullable=False)
    applicant_name = Column(String(100), nullable=True)
    supervisor_name = Column(String(100), nullable=True)
    applied_at = Column(DateTime, default=datetime.datetime.utcnow)
    approved_at = Column(DateTime, nullable=True)
    rejected_at = Column(DateTime, nullable=True)
    refunded_at = Column(DateTime, nullable=True)
    completed_at = Column(DateTime, nullable=True)
    rejection_reason = Column(Text, nullable=True)

    sale = relationship("Sale", back_populates="refund")
    items = relationship("RefundItem", back_populates="refund", cascade="all, delete-orphan")


class RefundItem(Base):
    __tablename__ = "refund_items"

    id = Column(Integer, primary_key=True, index=True)
    refund_id = Column(Integer, ForeignKey("refund_requests.id"), nullable=False, index=True)
    sale_item_id = Column(Integer, ForeignKey("sale_items.id"), nullable=False, index=True)
    quantity = Column(Integer, nullable=False)
    unit_price = Column(Integer, nullable=False)
    subtotal = Column(Integer, nullable=False)

    refund = relationship("RefundRequest", back_populates="items")
