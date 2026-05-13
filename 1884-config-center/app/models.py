from datetime import datetime
from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey, Index
from sqlalchemy.orm import relationship
from app.database import Base


class Config(Base):
    __tablename__ = "configs"

    id = Column(Integer, primary_key=True, index=True)
    project = Column(String(100), nullable=False)
    env = Column(String(50), nullable=False)
    key = Column(String(200), nullable=False)
    current_value = Column(Text, nullable=True)
    current_version = Column(Integer, default=1)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    __table_args__ = (
        Index('ix_config_project_env_key', 'project', 'env', 'key', unique=True),
    )


class ConfigVersion(Base):
    __tablename__ = "config_versions"

    id = Column(Integer, primary_key=True, index=True)
    config_id = Column(Integer, ForeignKey("configs.id"), nullable=False)
    version = Column(Integer, nullable=False)
    value = Column(Text, nullable=True)
    operation = Column(String(50), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    config = relationship("Config", backref="versions")


class Watch(Base):
    __tablename__ = "watches"

    id = Column(Integer, primary_key=True, index=True)
    project = Column(String(100), nullable=False)
    env = Column(String(50), nullable=False)
    key = Column(String(200), nullable=False)
    callback_url = Column(String(500), nullable=False)
    status = Column(String(20), default="active")
    failed_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    __table_args__ = (
        Index('ix_watch_project_env_key', 'project', 'env', 'key'),
    )
