from sqlalchemy import Column, DateTime, Float, ForeignKey, Integer, String
from sqlalchemy.orm import relationship

from server.core.database import Base

from .base import BaseModel


class SparePart(Base, BaseModel):
    __tablename__ = "spare_parts"

    name = Column(String(100), nullable=False, unique=True)
    code = Column(String(50), nullable=False, unique=True)
    category = Column(String(50), nullable=True)
    unit = Column(String(20), nullable=False)
    stock_quantity = Column(Float, default=0)
    safety_stock = Column(Float, default=0)
    in_transit_quantity = Column(Float, default=0)
    price = Column(Float, default=0)
    supplier = Column(String(100), nullable=True)


class SpareUsage(Base, BaseModel):
    __tablename__ = "spare_usages"

    lighthouse_id = Column(Integer, ForeignKey("lighthouses.id"), nullable=False)
    spare_part_id = Column(Integer, ForeignKey("spare_parts.id"), nullable=False)
    work_order_id = Column(Integer, ForeignKey("work_orders.id"), nullable=True)
    quantity = Column(Float, nullable=False)
    usage_type = Column(String(20), default="maintenance")
    notes = Column(String(255), nullable=True)

    lighthouse = relationship("Lighthouse")
    spare_part = relationship("SparePart")
    work_order = relationship("WorkOrder")


class PurchaseRequest(Base, BaseModel):
    __tablename__ = "purchase_requests"

    request_no = Column(String(50), nullable=False, unique=True)
    status = Column(String(20), default="pending")
    requested_by = Column(String(100), nullable=False)
    requested_at = Column(DateTime, nullable=False)
    reason = Column(String(255), nullable=True)
    total_amount = Column(Float, default=0)
    approved_by = Column(String(100), nullable=True)
    approved_at = Column(DateTime, nullable=True)
    ordered_by = Column(String(100), nullable=True)
    ordered_at = Column(DateTime, nullable=True)
    received_by = Column(String(100), nullable=True)
    received_at = Column(DateTime, nullable=True)


class PurchaseRequestItem(Base, BaseModel):
    __tablename__ = "purchase_request_items"

    request_id = Column(Integer, ForeignKey("purchase_requests.id"), nullable=False)
    spare_part_id = Column(Integer, ForeignKey("spare_parts.id"), nullable=False)
    quantity = Column(Float, nullable=False)
    unit_price = Column(Float, default=0)
    subtotal = Column(Float, default=0)

    request = relationship("PurchaseRequest")
    spare_part = relationship("SparePart")
