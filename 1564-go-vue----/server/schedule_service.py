from datetime import date, datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session

from .models import Pilot, Qualification, FlightSchedule
from .schemas import FlightScheduleCreate, PilotRecommendation, PilotSummary
from .validation_service import (
    validate_flight_assignment,
    get_pilot_monthly_hours,
    YELLOW_WARNING,
    MONTHLY_LIMIT
)


def update_pilot_after_schedule(db: Session, pilot: Pilot, schedule: FlightSchedule):
    pilot.monthly_hours = get_pilot_monthly_hours(db, pilot.id, schedule.departure_time)
    
    yearly_hours = 0.0
    year_start = schedule.departure_time.replace(
        month=1, day=1, hour=0, minute=0, second=0, microsecond=0
    )
    schedules_year = db.query(FlightSchedule).filter(
        FlightSchedule.pilot_id == pilot.id,
        FlightSchedule.status.in_(["confirmed", "completed", "in_flight"]),
        FlightSchedule.departure_time >= year_start
    ).all()
    yearly_hours = sum(s.flight_duration for s in schedules_year)
    pilot.yearly_hours = yearly_hours
    
    if pilot.monthly_hours >= YELLOW_WARNING:
        hours_above = pilot.monthly_hours - YELLOW_WARNING
        max_reduction = 50
        reduction = min(int(hours_above / 5) * 5, max_reduction)
        pilot.priority_score = max(100 - reduction, 30)
    else:
        pilot.priority_score = 100
    
    aircraft_qual = db.query(Qualification).filter(
        Qualification.pilot_id == pilot.id,
        Qualification.aircraft_type == schedule.aircraft_type
    ).first()
    if aircraft_qual:
        aircraft_qual.last_used_date = schedule.departure_time.date()
    
    db.commit()
    db.refresh(pilot)


def confirm_schedule(db: Session, schedule: FlightSchedule):
    schedule.status = "confirmed"
    db.commit()
    db.refresh(schedule)
    
    pilot = db.query(Pilot).filter(Pilot.id == schedule.pilot_id).first()
    if pilot:
        update_pilot_after_schedule(db, pilot, schedule)
    
    return schedule


def start_flight(db: Session, schedule: FlightSchedule):
    pilot = db.query(Pilot).filter(Pilot.id == schedule.pilot_id).first()
    if pilot:
        if pilot.status == "standby":
            pilot.status = "in_flight"
        else:
            raise ValueError(f"飞行员状态不允许执飞: {pilot.status}")
    
    schedule.status = "in_flight"
    db.commit()
    db.refresh(schedule)
    if pilot:
        db.refresh(pilot)
    
    return schedule


def complete_flight(db: Session, schedule: FlightSchedule):
    schedule.status = "completed"
    db.commit()
    db.refresh(schedule)
    
    pilot = db.query(Pilot).filter(Pilot.id == schedule.pilot_id).first()
    if pilot:
        pilot.status = "rest"
        db.commit()
        db.refresh(pilot)
    
    return schedule


def finish_rest(db: Session, pilot: Pilot):
    if pilot.status == "rest":
        pilot.status = "standby"
        db.commit()
        db.refresh(pilot)
    return pilot


def get_available_pilots(db: Session, exclude_pilot_ids: Optional[List[int]] = None) -> List[Pilot]:
    query = db.query(Pilot).filter(Pilot.status.in_(["standby", "rest"]))
    if exclude_pilot_ids:
        query = query.filter(Pilot.id.notin_(exclude_pilot_ids))
    return query.order_by(Pilot.priority_score.desc()).all()


def recommend_pilots_for_flight(
    db: Session,
    aircraft_type: str,
    is_international: bool,
    departure_time: datetime,
    arrival_time: datetime,
    limit: int = 10
) -> List[PilotRecommendation]:
    flight_duration = (arrival_time - departure_time).total_seconds() / 3600.0
    
    available_pilots = db.query(Pilot).filter(
        Pilot.status.in_(["standby", "rest"])
    ).order_by(Pilot.priority_score.desc()).all()
    
    recommendations: List[PilotRecommendation] = []
    
    for pilot in available_pilots:
        validation = validate_flight_assignment(
            db, pilot, aircraft_type, is_international,
            departure_time, arrival_time, flight_duration
        )
        
        if not validation.valid:
            continue
        
        score = pilot.priority_score
        reasons: List[str] = []
        warnings: List[str] = []
        
        reasons.append(f"优先级分数: {score}")
        reasons.append(f"英语等级: {pilot.english_level}级")
        reasons.append(f"当前月度飞行时间: {pilot.monthly_hours:.1f}小时")
        
        if pilot.monthly_hours >= YELLOW_WARNING:
            warnings.append(f"月度飞行时间已达{pilot.monthly_hours:.1f}小时, 已降低排班优先级")
        
        if validation.warnings:
            warnings.extend(validation.warnings)
        
        if pilot.english_level >= 5:
            score += 5
            reasons.append("英语等级满足国际航线要求")
        elif not is_international:
            score += 2
        
        if pilot.status == "rest":
            score -= 10
            reasons.append("飞行员处于休息状态")
        
        recommendations.append(PilotRecommendation(
            pilot=PilotSummary.model_validate(pilot),
            score=score,
            warnings=warnings,
            reasons=reasons
        ))
    
    recommendations.sort(key=lambda x: x.score, reverse=True)
    return recommendations[:limit]


def calculate_flight_duration(departure: datetime, arrival: datetime) -> float:
    return (arrival - departure).total_seconds() / 3600.0
