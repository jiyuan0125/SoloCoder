from datetime import datetime
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey
from sqlalchemy.orm import relationship
from database import Base


class Config(Base):
    __tablename__ = "configs"

    id = Column(Integer, primary_key=True, index=True)
    app_name = Column(String, index=True, nullable=False)
    key = Column(String, index=True, nullable=False)
    value = Column(String, nullable=False)
    version = Column(Integer, default=1, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    versions = relationship("ConfigVersion", back_populates="config", cascade="all, delete-orphan")
    notifications = relationship("Notification", back_populates="config", cascade="all, delete-orphan")

    __table_args__ = (
        {"sqlite_autoincrement": True},
    )


class ConfigVersion(Base):
    __tablename__ = "config_versions"

    id = Column(Integer, primary_key=True, index=True)
    config_id = Column(Integer, ForeignKey("configs.id"), nullable=False)
    app_name = Column(String, index=True, nullable=False)
    key = Column(String, index=True, nullable=False)
    value = Column(String, nullable=False)
    version = Column(Integer, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    config = relationship("Config", back_populates="versions")


class Subscriber(Base):
    __tablename__ = "subscribers"

    id = Column(Integer, primary_key=True, index=True)
    app_name = Column(String, index=True, nullable=False)
    callback_url = Column(String, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)


class Notification(Base):
    __tablename__ = "notifications"

    id = Column(Integer, primary_key=True, index=True)
    config_id = Column(Integer, ForeignKey("configs.id"), nullable=False)
    app_name = Column(String, index=True, nullable=False)
    key = Column(String, index=True, nullable=False)
    callback_url = Column(String, nullable=False)
    version = Column(Integer, nullable=False)
    status = Column(String, default="pending")  # pending, success, failed
    retry_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    completed_at = Column(DateTime, nullable=True)

    config = relationship("Config", back_populates="notifications")
