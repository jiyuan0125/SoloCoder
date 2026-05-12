from datetime import date, datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session

from .models import Pilot, Qualification, Medical, FlightSchedule
from .schemas import ValidationResult


MONTHLY_LIMIT = 100.0
YEARLY_LIMIT = 1000.0
WEEKLY_LIMIT = 35.0
YELLOW_WARNING = 80.0
RED_WARNING = 95.0
INTERNATIONAL_ENGLISH_LEVEL = 5


def get_pilot_age(pilot: Pilot, check_date: Optional[date] = None) -> int:
    if check_date is None:
        check_date = date.today()
    age = check_date.year - pilot.birth_date.year
    if (check_date.month, check_date.day) < (pilot.birth_date.month, pilot.birth_date.day):
        age -= 1
    return age


def check_medical_validity(db: Session, pilot: Pilot, check_date: Optional[date] = None) -> tuple[bool, List[str]]:
    if check_date is None:
        check_date = date.today()
    
    latest_medical = db.query(Medical).filter(
        Medical.pilot_id == pilot.id
    ).order_by(Medical.examination_date.desc()).first()
    
    messages = []
    
    if not latest_medical:
        messages.append("飞行员没有体检记录")
        return False, messages
    
    if latest_medical.result != "qualified":
        messages.append(f"最近一次体检结果为: {latest_medical.result}")
        return False, messages
    
    if latest_medical.expiry_date < check_date:
        messages.append(f"体检已过期，过期日期: {latest_medical.expiry_date}")
        return False, messages
    
    age = get_pilot_age(pilot, check_date)
    max_validity_years = 1 if age >= 40 else 2
    expected_expiry = latest_medical.examination_date + timedelta(days=365 * max_validity_years)
    
    if latest_medical.expiry_date > expected_expiry:
        messages.append(f"体检有效期超过规定年限(年龄{age}岁, 最多{max_validity_years}年)")
        return False, messages
    
    return True, messages


def check_qualifications(db: Session, pilot: Pilot, aircraft_type: str, check_date: Optional[date] = None) -> tuple[bool, List[str]]:
    if check_date is None:
        check_date = date.today()
    
    messages = []
    
    qualifications = db.query(Qualification).filter(
        Qualification.pilot_id == pilot.id,
        Qualification.is_valid == True
    ).all()
    
    mandatory_quals = [q for q in qualifications if q.is_mandatory]
    expired_mandatory = [
        q for q in mandatory_quals 
        if q.expiry_date < check_date
    ]
    
    if expired_mandatory:
        for q in expired_mandatory:
            messages.append(f"必需资质已过期: {q.qualification_type}, 过期日期: {q.expiry_date}")
        return False, messages
    
    aircraft_quals = [
        q for q in qualifications 
        if q.aircraft_type == aircraft_type and q.expiry_date >= check_date
    ]
    
    if not aircraft_quals:
        messages.append(f"飞行员没有该机型({aircraft_type})的有效资质")
        return False, messages
    
    return True, messages


def check_english_level(pilot: Pilot, is_international: bool) -> tuple[bool, List[str]]:
    messages = []
    
    if is_international and pilot.english_level < INTERNATIONAL_ENGLISH_LEVEL:
        messages.append(f"国际航线要求英语等级{INTERNATIONAL_ENGLISH_LEVEL}级, 当前等级: {pilot.english_level}")
        return False, messages
    
    return True, messages


def get_pilot_7day_hours(db: Session, pilot_id: int, check_start: datetime, check_end: datetime) -> float:
    week_start = check_start - timedelta(days=7)
    week_end = check_end
    
    schedules = db.query(FlightSchedule).filter(
        FlightSchedule.pilot_id == pilot_id,
        FlightSchedule.status.in_(["confirmed", "completed", "in_flight"]),
        FlightSchedule.departure_time >= week_start,
        FlightSchedule.arrival_time <= week_end
    ).all()
    
    total = sum(s.flight_duration for s in schedules)
    return total


def get_pilot_monthly_hours(db: Session, pilot_id: int, check_date: datetime) -> float:
    month_start = check_date.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
    
    schedules = db.query(FlightSchedule).filter(
        FlightSchedule.pilot_id == pilot_id,
        FlightSchedule.status.in_(["confirmed", "completed", "in_flight"]),
        FlightSchedule.departure_time >= month_start
    ).all()
    
    total = sum(s.flight_duration for s in schedules)
    return total


def get_pilot_yearly_hours(db: Session, pilot_id: int, check_date: datetime) -> float:
    year_start = check_date.replace(month=1, day=1, hour=0, minute=0, second=0, microsecond=0)
    
    schedules = db.query(FlightSchedule).filter(
        FlightSchedule.pilot_id == pilot_id,
        FlightSchedule.status.in_(["confirmed", "completed", "in_flight"]),
        FlightSchedule.departure_time >= year_start
    ).all()
    
    total = sum(s.flight_duration for s in schedules)
    return total


def check_flight_hours(db: Session, pilot: Pilot, departure_time: datetime, arrival_time: datetime, flight_duration: float) -> tuple[bool, List[str], List[str]]:
    messages = []
    warnings = []
    
    current_monthly = get_pilot_monthly_hours(db, pilot.id, departure_time)
    current_yearly = get_pilot_yearly_hours(db, pilot.id, departure_time)
    current_7day = get_pilot_7day_hours(db, pilot.id, departure_time, arrival_time)
    
    new_monthly = current_monthly + flight_duration
    new_yearly = current_yearly + flight_duration
    new_7day = current_7day + flight_duration
    
    if new_7day > WEEKLY_LIMIT:
        messages.append(f"连续7天飞行时间将超限: {new_7day:.1f}小时 / 限制{WEEKLY_LIMIT}小时")
    
    if new_monthly > MONTHLY_LIMIT:
        messages.append(f"月度飞行时间将超限: {new_monthly:.1f}小时 / 限制{MONTHLY_LIMIT}小时")
    
    if new_yearly > YEARLY_LIMIT:
        messages.append(f"年度飞行时间将超限: {new_yearly:.1f}小时 / 限制{YEARLY_LIMIT}小时")
    
    if new_monthly >= YELLOW_WARNING and new_monthly < RED_WARNING:
        warnings.append(f"月度飞行时间接近超限: {new_monthly:.1f}小时 / 黄色预警{YELLOW_WARNING}小时")
    elif new_monthly >= RED_WARNING:
        warnings.append(f"月度飞行时间即将超限: {new_monthly:.1f}小时 / 红色预警{RED_WARNING}小时")
    
    return len(messages) == 0, messages, warnings


def check_pilot_status(pilot: Pilot, departure_time: datetime, arrival_time: datetime) -> tuple[bool, List[str]]:
    messages = []
    
    if pilot.status == "grounded":
        messages.append("飞行员已停飞")
        return False, messages
    
    if pilot.status == "in_flight":
        messages.append("飞行员正在执飞")
        return False, messages
    
    if pilot.status == "flying":
        messages.append("飞行员正在执飞任务中")
        return False, messages
    
    return True, messages


def check_status_transition(db: Session, pilot: Pilot, departure_time: datetime) -> tuple[bool, List[str]]:
    messages = []
    
    last_schedule = db.query(FlightSchedule).filter(
        FlightSchedule.pilot_id == pilot.id,
        FlightSchedule.status.in_(["confirmed", "completed", "in_flight"])
    ).order_by(FlightSchedule.arrival_time.desc()).first()
    
    if last_schedule and last_schedule.status in ["in_flight", "confirmed"]:
        if last_schedule.arrival_time > departure_time:
            messages.append("飞行员当前任务尚未结束，无法安排新任务")
            return False, messages
    
    return True, messages


def validate_flight_assignment(
    db: Session,
    pilot: Pilot,
    aircraft_type: str,
    is_international: bool,
    departure_time: datetime,
    arrival_time: datetime,
    flight_duration: float
) -> ValidationResult:
    all_messages: List[str] = []
    all_warnings: List[str] = []
    is_valid = True
    
    check_date = departure_time.date()
    
    status_ok, status_msgs = check_pilot_status(pilot, departure_time, arrival_time)
    all_messages.extend(status_msgs)
    if not status_ok:
        is_valid = False
    
    transition_ok, transition_msgs = check_status_transition(db, pilot, departure_time)
    all_messages.extend(transition_msgs)
    if not transition_ok:
        is_valid = False
    
    medical_ok, medical_msgs = check_medical_validity(db, pilot, check_date)
    all_messages.extend(medical_msgs)
    if not medical_ok:
        is_valid = False
    
    quals_ok, quals_msgs = check_qualifications(db, pilot, aircraft_type, check_date)
    all_messages.extend(quals_msgs)
    if not quals_ok:
        is_valid = False
    
    english_ok, english_msgs = check_english_level(pilot, is_international)
    all_messages.extend(english_msgs)
    if not english_ok:
        is_valid = False
    
    hours_ok, hours_msgs, hours_warnings = check_flight_hours(
        db, pilot, departure_time, arrival_time, flight_duration
    )
    all_messages.extend(hours_msgs)
    all_warnings.extend(hours_warnings)
    if not hours_ok:
        is_valid = False
    
    return ValidationResult(
        valid=is_valid,
        messages=all_messages,
        warnings=all_warnings
    )


def get_medical_expiry_date(birth_date: date, examination_date: date) -> date:
    check_date = examination_date
    age = check_date.year - birth_date.year
    if (check_date.month, check_date.day) < (birth_date.month, birth_date.day):
        age -= 1
    
    validity_years = 1 if age >= 40 else 2
    return examination_date + timedelta(days=365 * validity_years)
