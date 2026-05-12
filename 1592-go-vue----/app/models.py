from sqlalchemy import Column, Integer, String, Float, DateTime, Boolean, ForeignKey, Text, Date, Time
from sqlalchemy.orm import relationship
from datetime import datetime, timedelta
from app.database import Base


class Venue(Base):
    __tablename__ = "venues"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    location = Column(String(200))
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    halls = relationship("Hall", back_populates="venue", cascade="all, delete-orphan")


class Hall(Base):
    __tablename__ = "halls"

    id = Column(Integer, primary_key=True, index=True)
    venue_id = Column(Integer, ForeignKey("venues.id"), nullable=False)
    name = Column(String(100), nullable=False)
    floor = Column(Integer)
    total_area = Column(Float)
    max_exhibitors_per_slot = Column(Integer, default=5)
    description = Column(Text)

    venue = relationship("Venue", back_populates="halls")
    booths = relationship("Booth", back_populates="hall", cascade="all, delete-orphan")
    exhibition_halls = relationship("ExhibitionHall", back_populates="hall", cascade="all, delete-orphan")
    setup_schedules = relationship("SetupSchedule", back_populates="hall")


class Booth(Base):
    __tablename__ = "booths"

    id = Column(Integer, primary_key=True, index=True)
    hall_id = Column(Integer, ForeignKey("halls.id"), nullable=False)
    booth_number = Column(String(50), nullable=False)
    area = Column(Float, nullable=False)
    is_special = Column(Boolean, default=False)
    base_price = Column(Float, nullable=False)
    position = Column(String(100))
    status = Column(String(20), default="available")
    description = Column(Text)

    hall = relationship("Hall", back_populates="booths")
    selections = relationship("BoothSelection", back_populates="booth")


class Exhibition(Base):
    __tablename__ = "exhibitions"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    organizer = Column(String(100))
    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    exhibition_halls = relationship("ExhibitionHall", back_populates="exhibition", cascade="all, delete-orphan")
    booth_selections = relationship("BoothSelection", back_populates="exhibition")


class ExhibitionHall(Base):
    __tablename__ = "exhibition_halls"

    id = Column(Integer, primary_key=True, index=True)
    exhibition_id = Column(Integer, ForeignKey("exhibitions.id"), nullable=False)
    hall_id = Column(Integer, ForeignKey("halls.id"), nullable=False)
    exhibition_date = Column(Date, nullable=False)

    exhibition = relationship("Exhibition", back_populates="exhibition_halls")
    hall = relationship("Hall", back_populates="exhibition_halls")


class BoothSelection(Base):
    __tablename__ = "booth_selections"

    id = Column(Integer, primary_key=True, index=True)
    booth_id = Column(Integer, ForeignKey("booths.id"), nullable=False)
    exhibition_id = Column(Integer, ForeignKey("exhibitions.id"), nullable=False)
    exhibitor_name = Column(String(100), nullable=False)
    contact_person = Column(String(100))
    contact_phone = Column(String(50))
    selected_at = Column(DateTime, default=datetime.utcnow)
    expires_at = Column(DateTime)
    status = Column(String(20), default="selected")
    final_price = Column(Float)
    is_paid = Column(Boolean, default=False)
    paid_at = Column(DateTime)
    notes = Column(Text)

    booth = relationship("Booth", back_populates="selections")
    exhibition = relationship("Exhibition", back_populates="booth_selections")
    setup_schedule = relationship("SetupSchedule", back_populates="selection", uselist=False)

    @property
    def is_expired(self):
        if self.status != "selected":
            return False
        if self.expires_at is None:
            return False
        return datetime.utcnow() > self.expires_at

    def set_expiration(self, hours: int = 48):
        self.expires_at = datetime.utcnow() + timedelta(hours=hours)


class SetupSchedule(Base):
    __tablename__ = "setup_schedules"

    id = Column(Integer, primary_key=True, index=True)
    hall_id = Column(Integer, ForeignKey("halls.id"), nullable=False)
    selection_id = Column(Integer, ForeignKey("booth_selections.id"), nullable=False)
    setup_date = Column(Date, nullable=False)
    start_time = Column(Time, nullable=False)
    end_time = Column(Time, nullable=False)
    actual_hours = Column(Float)
    billed_hours = Column(Integer)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    hall = relationship("Hall", back_populates="setup_schedules")
    selection = relationship("BoothSelection", back_populates="setup_schedule")


class VisitReservation(Base):
    __tablename__ = "visit_reservations"

    id = Column(Integer, primary_key=True, index=True)
    exhibition_id = Column(Integer, ForeignKey("exhibitions.id"), nullable=False)
    visitor_name = Column(String(100), nullable=False)
    visitor_phone = Column(String(50))
    visitor_email = Column(String(100))
    visit_date = Column(Date, nullable=False)
    party_size = Column(Integer, default=1)
    status = Column(String(20), default="confirmed")
    queue_position = Column(Integer)
    created_at = Column(DateTime, default=datetime.utcnow)

    exhibition = relationship("Exhibition")


class DailyCapacity(Base):
    __tablename__ = "daily_capacities"

    id = Column(Integer, primary_key=True, index=True)
    exhibition_id = Column(Integer, ForeignKey("exhibitions.id"), nullable=False)
    date = Column(Date, nullable=False)
    max_visitors = Column(Integer, nullable=False)
    current_confirmed = Column(Integer, default=0)

    exhibition = relationship("Exhibition")
