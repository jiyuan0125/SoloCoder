from sqlalchemy import Column, Float, Integer, String

from server.core.database import Base

from .base import BaseModel


class Lighthouse(Base, BaseModel):
    __tablename__ = "lighthouses"

    name = Column(String(100), nullable=False, unique=True)
    location = Column(String(255), nullable=False)
    rated_illuminance = Column(Float, nullable=False)
    battery_capacity = Column(Float, nullable=False)
    solar_power = Column(Float, nullable=False)
    last_overhaul_date = Column(String(20), nullable=True)
    status = Column(String(20), default="normal")
    operator = Column(String(100), nullable=True)
