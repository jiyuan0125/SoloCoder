from datetime import datetime, date, time
from sqlalchemy import (
    Column, Integer, String, Float, DateTime, Date, Time, Boolean, Text,
    ForeignKey, UniqueConstraint, Enum as SQLEnum
)
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
import enum
from app.database import Base


class BookingType(str, enum.Enum):
    INDIVIDUAL = "individual"
    MATCH = "match"
    CLASS = "class"
    MAINTENANCE = "maintenance"


class PaymentStatus(str, enum.Enum):
    PENDING = "pending"
    PAID = "paid"
    REFUNDED = "refunded"
    CANCELLED = "cancelled"


class BookingStatus(str, enum.Enum):
    PENDING_PAYMENT = "pending_payment"
    CONFIRMED = "confirmed"
    CANCELLED = "cancelled"
    COMPLETED = "completed"
    RESCHEDULED = "rescheduled"


class ChargeType(str, enum.Enum):
    PER_HOUR = "per_hour"
    PER_SESSION = "per_session"


class MatchStatus(str, enum.Enum):
    SCHEDULED = "scheduled"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"
    FORFEIT = "forfeit"
    WALK_OVER = "walk_over"


class TournamentStatus(str, enum.Enum):
    DRAFT = "draft"
    SCHEDULED = "scheduled"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"


class ClassStatus(str, enum.Enum):
    PENDING = "pending"
    CONFIRMED = "confirmed"
    CANCELLED = "cancelled"
    COMPLETED = "completed"


class VenueType(Base):
    __tablename__ = "venue_types"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False)
    charge_type = Column(SQLEnum(ChargeType), nullable=False)
    price = Column(Float, nullable=False)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    
    venues = relationship("Venue", back_populates="venue_type", cascade="all, delete-orphan")


class Venue(Base):
    __tablename__ = "venues"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    venue_type_id = Column(Integer, ForeignKey("venue_types.id"), nullable=False)
    capacity = Column(Integer, default=10)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    
    venue_type = relationship("VenueType", back_populates="venues")
    bookings = relationship("Booking", back_populates="venue", cascade="all, delete-orphan")
    maintenances = relationship("Maintenance", back_populates="venue", cascade="all, delete-orphan")


class User(Base):
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True, index=True)
    username = Column(String(50), unique=True, nullable=False)
    phone = Column(String(20), unique=True, nullable=False)
    is_admin = Column(Boolean, default=False)
    created_at = Column(DateTime, default=func.now())
    
    bookings = relationship("Booking", back_populates="user", cascade="all, delete-orphan")
    payments = relationship("Payment", back_populates="user", cascade="all, delete-orphan")
    class_registrations = relationship("ClassRegistration", back_populates="user", cascade="all, delete-orphan")


class Maintenance(Base):
    __tablename__ = "maintenances"
    
    id = Column(Integer, primary_key=True, index=True)
    venue_id = Column(Integer, ForeignKey("venues.id"), nullable=False)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=False)
    reason = Column(Text, nullable=True)
    created_at = Column(DateTime, default=func.now())
    
    venue = relationship("Venue", back_populates="maintenances")


class Booking(Base):
    __tablename__ = "bookings"
    
    id = Column(Integer, primary_key=True, index=True)
    booking_no = Column(String(20), unique=True, nullable=False, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    venue_id = Column(Integer, ForeignKey("venues.id"), nullable=False)
    booking_type = Column(SQLEnum(BookingType), default=BookingType.INDIVIDUAL)
    booking_date = Column(Date, nullable=False)
    start_time = Column(Time, nullable=False)
    end_time = Column(Time, nullable=False)
    hours = Column(Float, nullable=False)
    original_amount = Column(Float, nullable=False)
    discount_amount = Column(Float, default=0.0)
    final_amount = Column(Float, nullable=False)
    status = Column(SQLEnum(BookingStatus), default=BookingStatus.PENDING_PAYMENT)
    qr_token = Column(String(255), nullable=True)
    is_continuous = Column(Boolean, default=False)
    parent_booking_id = Column(Integer, ForeignKey("bookings.id"), nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    paid_at = Column(DateTime, nullable=True)
    cancelled_at = Column(DateTime, nullable=True)
    
    user = relationship("User", back_populates="bookings")
    venue = relationship("Venue", back_populates="bookings")
    payment = relationship("Payment", back_populates="booking", uselist=False, cascade="all, delete-orphan")
    refund = relationship("Refund", back_populates="booking", uselist=False, cascade="all, delete-orphan")
    match = relationship("Match", back_populates="booking", uselist=False)
    training_class = relationship("TrainingClass", back_populates="booking", uselist=False)
    child_bookings = relationship("Booking", remote_side=[id])
    
    __table_args__ = (
        UniqueConstraint('venue_id', 'booking_date', 'start_time', 'end_time', name='uix_venue_time'),
    )


class Payment(Base):
    __tablename__ = "payments"
    
    id = Column(Integer, primary_key=True, index=True)
    payment_no = Column(String(30), unique=True, nullable=False)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    booking_id = Column(Integer, ForeignKey("bookings.id"), nullable=False, unique=True)
    amount = Column(Float, nullable=False)
    status = Column(SQLEnum(PaymentStatus), default=PaymentStatus.PENDING)
    payment_method = Column(String(50), nullable=True)
    transaction_id = Column(String(100), nullable=True)
    created_at = Column(DateTime, default=func.now())
    paid_at = Column(DateTime, nullable=True)
    
    user = relationship("User", back_populates="payments")
    booking = relationship("Booking", back_populates="payment")
    refunds = relationship("Refund", back_populates="payment", cascade="all, delete-orphan")


class Refund(Base):
    __tablename__ = "refunds"
    
    id = Column(Integer, primary_key=True, index=True)
    refund_no = Column(String(30), unique=True, nullable=False)
    booking_id = Column(Integer, ForeignKey("bookings.id"), nullable=False, unique=True)
    payment_id = Column(Integer, ForeignKey("payments.id"), nullable=False)
    refund_amount = Column(Float, nullable=False)
    refund_rate = Column(Float, nullable=False)
    reason = Column(Text, nullable=True)
    status = Column(SQLEnum(PaymentStatus), default=PaymentStatus.PENDING)
    created_at = Column(DateTime, default=func.now())
    processed_at = Column(DateTime, nullable=True)
    
    booking = relationship("Booking", back_populates="refund")
    payment = relationship("Payment", back_populates="refunds")


class Tournament(Base):
    __tablename__ = "tournaments"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False)
    venue_type_id = Column(Integer, ForeignKey("venue_types.id"), nullable=False)
    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=False)
    status = Column(SQLEnum(TournamentStatus), default=TournamentStatus.DRAFT)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    
    teams = relationship("Team", back_populates="tournament", cascade="all, delete-orphan")
    matches = relationship("Match", back_populates="tournament", cascade="all, delete-orphan")
    statistics = relationship("TeamStatistics", back_populates="tournament", cascade="all, delete-orphan")


class Team(Base):
    __tablename__ = "teams"
    
    id = Column(Integer, primary_key=True, index=True)
    tournament_id = Column(Integer, ForeignKey("tournaments.id"), nullable=False)
    name = Column(String(100), nullable=False)
    captain_name = Column(String(50), nullable=True)
    phone = Column(String(20), nullable=True)
    is_forfeited = Column(Boolean, default=False)
    is_withdrawn = Column(Boolean, default=False)
    created_at = Column(DateTime, default=func.now())
    
    tournament = relationship("Tournament", back_populates="teams")
    statistics = relationship("TeamStatistics", back_populates="team", uselist=False, cascade="all, delete-orphan")
    home_matches = relationship("Match", foreign_keys="Match.home_team_id", back_populates="home_team")
    away_matches = relationship("Match", foreign_keys="Match.away_team_id", back_populates="away_team")


class Match(Base):
    __tablename__ = "matches"
    
    id = Column(Integer, primary_key=True, index=True)
    tournament_id = Column(Integer, ForeignKey("tournaments.id"), nullable=False)
    match_no = Column(String(20), nullable=False)
    round_no = Column(Integer, default=1)
    booking_id = Column(Integer, ForeignKey("bookings.id"), nullable=True, unique=True)
    home_team_id = Column(Integer, ForeignKey("teams.id"), nullable=True)
    away_team_id = Column(Integer, ForeignKey("teams.id"), nullable=True)
    match_date = Column(Date, nullable=False)
    start_time = Column(Time, nullable=False)
    end_time = Column(Time, nullable=False)
    home_score = Column(Integer, nullable=True)
    away_score = Column(Integer, nullable=True)
    status = Column(SQLEnum(MatchStatus), default=MatchStatus.SCHEDULED)
    winner_id = Column(Integer, ForeignKey("teams.id"), nullable=True)
    is_forfeit = Column(Boolean, default=False)
    forfeit_team_id = Column(Integer, ForeignKey("teams.id"), nullable=True)
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    
    tournament = relationship("Tournament", back_populates="matches")
    booking = relationship("Booking", back_populates="match")
    home_team = relationship("Team", foreign_keys=[home_team_id], back_populates="home_matches")
    away_team = relationship("Team", foreign_keys=[away_team_id], back_populates="away_matches")
    winner = relationship("Team", foreign_keys=[winner_id])
    forfeit_team = relationship("Team", foreign_keys=[forfeit_team_id])


class TeamStatistics(Base):
    __tablename__ = "team_statistics"
    
    id = Column(Integer, primary_key=True, index=True)
    tournament_id = Column(Integer, ForeignKey("tournaments.id"), nullable=False)
    team_id = Column(Integer, ForeignKey("teams.id"), nullable=False, unique=True)
    matches_played = Column(Integer, nullable=True)
    wins = Column(Integer, nullable=True)
    losses = Column(Integer, nullable=True)
    draws = Column(Integer, nullable=True)
    goals_for = Column(Integer, nullable=True)
    goals_against = Column(Integer, nullable=True)
    goal_difference = Column(Integer, nullable=True)
    points = Column(Integer, nullable=True)
    rank = Column(Integer, nullable=True)
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    
    tournament = relationship("Tournament", back_populates="statistics")
    team = relationship("Team", back_populates="statistics")


class TrainingClass(Base):
    __tablename__ = "training_classes"
    
    id = Column(Integer, primary_key=True, index=True)
    class_no = Column(String(20), unique=True, nullable=False)
    name = Column(String(200), nullable=False)
    instructor = Column(String(100), nullable=True)
    venue_type_id = Column(Integer, ForeignKey("venue_types.id"), nullable=False)
    booking_id = Column(Integer, ForeignKey("bookings.id"), nullable=True, unique=True)
    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=False)
    class_time = Column(Time, nullable=False)
    duration_hours = Column(Float, default=2.0)
    min_students = Column(Integer, default=5)
    max_students = Column(Integer, default=20)
    price_per_student = Column(Float, nullable=False)
    status = Column(SQLEnum(ClassStatus), default=ClassStatus.PENDING)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    
    booking = relationship("Booking", back_populates="training_class")
    registrations = relationship("ClassRegistration", back_populates="training_class", cascade="all, delete-orphan")


class ClassRegistration(Base):
    __tablename__ = "class_registrations"
    
    id = Column(Integer, primary_key=True, index=True)
    registration_no = Column(String(30), unique=True, nullable=False)
    class_id = Column(Integer, ForeignKey("training_classes.id"), nullable=False)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    student_name = Column(String(100), nullable=False)
    phone = Column(String(20), nullable=True)
    amount_paid = Column(Float, nullable=False)
    is_refunded = Column(Boolean, default=False)
    refund_amount = Column(Float, default=0.0)
    registered_at = Column(DateTime, default=func.now())
    
    training_class = relationship("TrainingClass", back_populates="registrations")
    user = relationship("User", back_populates="class_registrations")
    
    __table_args__ = (
        UniqueConstraint('class_id', 'user_id', name='uix_class_user'),
    )


class Notification(Base):
    __tablename__ = "notifications"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=True)
    booking_id = Column(Integer, ForeignKey("bookings.id"), nullable=True)
    message_type = Column(String(50), nullable=False)
    title = Column(String(200), nullable=False)
    content = Column(Text, nullable=False)
    is_read = Column(Boolean, default=False)
    sent_at = Column(DateTime, default=func.now())
