from datetime import datetime
from enum import Enum as PyEnum
from sqlalchemy import Column, Integer, String, Text, DateTime, Enum, ForeignKey
from sqlalchemy.orm import relationship
from app.database import Base


class ConfigStatus(str, PyEnum):
    DRAFT = "draft"
    ACTIVE = "active"
    ARCHIVED = "archived"


class Environment(str, PyEnum):
    DEV = "dev"
    STAGING = "staging"
    PROD = "prod"


class Config(Base):
    __tablename__ = "configs"

    id = Column(Integer, primary_key=True, index=True)
    key = Column(String(255), index=True)
    environment = Column(Enum(Environment), index=True)
    version = Column(Integer, default=1)
    value = Column(Text)
    status = Column(Enum(ConfigStatus), default=ConfigStatus.DRAFT)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class Watch(Base):
    __tablename__ = "watches"

    id = Column(Integer, primary_key=True, index=True)
    key = Column(String(255), index=True)
    callback_url = Column(String(512))
    environment = Column(Enum(Environment), index=True)
    is_active = Column(String(10), default="active")
    failure_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
