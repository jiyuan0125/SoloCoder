from datetime import datetime, timedelta
from sqlalchemy import Column, Integer, String, Float, DateTime, Boolean, ForeignKey, Text
from sqlalchemy.orm import relationship

from .database import Base


class Part(Base):
    __tablename__ = "parts"

    id = Column(Integer, primary_key=True, index=True)
    part_number = Column(String(100), index=True, nullable=False)
    serial_number = Column(String(100), index=True, nullable=False)
    name = Column(String(200), nullable=False)
    unit_price = Column(Float, nullable=False, default=0.0)
    is_controllable = Column(Boolean, default=False)
    minimum_stock = Column(Integer, nullable=False, default=0)
    available_quantity = Column(Integer, nullable=False, default=0)
    status = Column(String(20), default="normal")
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    requisitions = relationship("Requisition", back_populates="part")
    repairs = relationship("Repair", back_populates="part")
    purchases = relationship("Purchase", back_populates="part")
    purchase_histories = relationship("PurchaseHistory", back_populates="part")
    inventory_counts = relationship("InventoryCount", back_populates="part")

    @property
    def is_out_of_stock(self):
        return self.available_quantity <= 0

    @property
    def stock_status(self):
        if self.available_quantity <= 0:
            return "out_of_stock"
        elif self.available_quantity < self.minimum_stock * 0.5:
            return "urgent"
        elif self.available_quantity < self.minimum_stock:
            return "low"
        else:
            return "normal"


class Requisition(Base):
    __tablename__ = "requisitions"

    id = Column(Integer, primary_key=True, index=True)
    part_id = Column(Integer, ForeignKey("parts.id"), nullable=False)
    requester = Column(String(100), nullable=False)
    quantity = Column(Integer, nullable=False, default=1)
    reason = Column(Text)
    status = Column(String(30), default="pending")
    level1_approver = Column(String(100))
    level1_approved_at = Column(DateTime)
    level2_approver = Column(String(100))
    level2_approved_at = Column(DateTime)
    completed_at = Column(DateTime)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    part = relationship("Part", back_populates="requisitions")

    @property
    def needs_level2_approval(self):
        if self.part:
            total_value = self.quantity * self.part.unit_price
            return self.part.is_controllable or total_value > 100000
        return False


class Repair(Base):
    __tablename__ = "repairs"

    id = Column(Integer, primary_key=True, index=True)
    part_id = Column(Integer, ForeignKey("parts.id"), nullable=False)
    quantity = Column(Integer, nullable=False, default=1)
    repair_vendor = Column(String(200))
    repair_cost = Column(Float, default=0.0)
    status = Column(String(20), default="in_repair")
    last_updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    created_at = Column(DateTime, default=datetime.utcnow)
    completed_at = Column(DateTime)

    part = relationship("Part", back_populates="repairs")

    @property
    def is_overdue(self):
        if self.status == "in_repair" and self.last_updated_at:
            return (datetime.utcnow() - self.last_updated_at) > timedelta(days=30)
        return False

    @property
    def should_scrap(self):
        if self.part and self.repair_cost > 0:
            return self.repair_cost > self.part.unit_price * 0.6
        return False


class Purchase(Base):
    __tablename__ = "purchases"

    id = Column(Integer, primary_key=True, index=True)
    part_id = Column(Integer, ForeignKey("parts.id"), nullable=False)
    quantity = Column(Integer, nullable=False, default=1)
    unit_price = Column(Float, nullable=False, default=0.0)
    urgency = Column(String(20), default="normal")
    status = Column(String(30), default="suggested")
    approver = Column(String(100))
    approved_at = Column(DateTime)
    order_date = Column(DateTime)
    receive_date = Column(DateTime)
    reason = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    part = relationship("Part", back_populates="purchases")
    purchase_histories = relationship("PurchaseHistory", back_populates="purchase")

    @classmethod
    def determine_urgency(cls, part: Part):
        if part.available_quantity <= 0:
            return "critical"
        elif part.available_quantity < part.minimum_stock * 0.5:
            return "urgent"
        return "normal"


class PurchaseHistory(Base):
    __tablename__ = "purchase_histories"

    id = Column(Integer, primary_key=True, index=True)
    part_id = Column(Integer, ForeignKey("parts.id"), nullable=False)
    purchase_id = Column(Integer, ForeignKey("purchases.id"))
    quantity = Column(Integer, nullable=False)
    unit_price = Column(Float, nullable=False)
    total_price = Column(Float, nullable=False)
    received_at = Column(DateTime, default=datetime.utcnow)

    part = relationship("Part", back_populates="purchase_histories")
    purchase = relationship("Purchase", back_populates="purchase_histories")


class InventoryCount(Base):
    __tablename__ = "inventory_counts"

    id = Column(Integer, primary_key=True, index=True)
    part_id = Column(Integer, ForeignKey("parts.id"), nullable=False)
    counted_quantity = Column(Integer, nullable=False)
    system_quantity = Column(Integer, nullable=False)
    difference = Column(Integer, nullable=False)
    reason = Column(Text)
    adjuster = Column(String(100), nullable=False)
    adjusted_at = Column(DateTime, default=datetime.utcnow)
    created_at = Column(DateTime, default=datetime.utcnow)

    part = relationship("Part", back_populates="inventory_counts")
