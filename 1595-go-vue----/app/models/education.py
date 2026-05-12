from datetime import datetime
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Text
from sqlalchemy.orm import relationship

from app.database import Base


class EducationEvent(Base):
    __tablename__ = "education_events"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False)
    description = Column(Text, nullable=True)
    venue_id = Column(Integer, ForeignKey("venues.id"), nullable=False)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=False)
    min_age_months = Column(Integer, nullable=False)
    max_age_months = Column(Integer, nullable=False)
    max_participants = Column(Integer, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    participants = relationship("EducationParticipant", back_populates="event")


class EducationParticipant(Base):
    __tablename__ = "education_participants"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("education_events.id"), nullable=False)
    participant_name = Column(String(100), nullable=False)
    birth_date = Column(DateTime, nullable=False)
    guardian_name = Column(String(100), nullable=True)
    guardian_phone = Column(String(20), nullable=True)
    registered_at = Column(DateTime, default=datetime.utcnow)

    event = relationship("EducationEvent", back_populates="participants")
