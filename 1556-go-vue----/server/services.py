from datetime import date, datetime, timedelta
from typing import List, Optional, Tuple
from sqlalchemy.orm import Session
from dateutil.relativedelta import relativedelta

from .models import (
    Crew, CrewStatus, Certificate, TrainingRecord, Ship, Assignment,
    AttendanceRecord, AttendanceType, SalaryRecord, CertificateStatus
)


def calculate_certificate_status(expiry_date: date, current_date: Optional[date] = None) -> Tuple[str, int]:
    if current_date is None:
        current_date = date.today()
    
    days_until_expiry = (expiry_date - current_date).days
    
    if days_until_expiry < 0:
        return CertificateStatus.EXPIRED.value, days_until_expiry
    elif days_until_expiry <= 30:
        return CertificateStatus.WARNING_ORANGE.value, days_until_expiry
    elif days_until_expiry <= 90:
        return CertificateStatus.WARNING_YELLOW.value, days_until_expiry
    else:
        return CertificateStatus.VALID.value, days_until_expiry


def check_crew_certificates_valid(db: Session, crew_id: int) -> Tuple[bool, List[dict]]:
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        return False, []
    
    has_expired = False
    certificate_summary = []
    
    for cert in crew.certificates:
        status, days_until = calculate_certificate_status(cert.expiry_date)
        if status == CertificateStatus.EXPIRED.value:
            has_expired = True
        certificate_summary.append({
            "certificate_id": cert.id,
            "certificate_type": cert.certificate_type,
            "status": status,
            "days_until_expiry": days_until
        })
    
    return not has_expired, certificate_summary


def check_training_interval(db: Session, crew_id: int, training_type: str, 
                            training_date: date, certificate: Optional[Certificate] = None) -> bool:
    if not certificate:
        certificates = db.query(Certificate).filter(
            Certificate.crew_id == crew_id,
            Certificate.certificate_type == training_type
        ).all()
        if not certificates:
            return True
        certificate = certificates[0]
    
    cert_validity_days = (certificate.expiry_date - certificate.issue_date).days
    max_interval_days = cert_validity_days // 2
    
    recent_trainings = db.query(TrainingRecord).filter(
        TrainingRecord.crew_id == crew_id,
        TrainingRecord.training_type == training_type,
        TrainingRecord.training_date < training_date
    ).order_by(TrainingRecord.training_date.desc()).first()
    
    if recent_trainings:
        days_since_last_training = (training_date - recent_trainings.training_date).days
        if days_since_last_training > max_interval_days:
            return False
    
    return True


def check_ship_capacity(db: Session, ship_id: int) -> bool:
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        return False
    
    active_assignments = db.query(Assignment).filter(
        Assignment.ship_id == ship_id,
        Assignment.is_active == True
    ).count()
    
    return active_assignments < ship.capacity


def validate_assignment(db: Session, crew_id: int, ship_id: int, is_watch_keeper: bool,
                        start_date: date) -> Tuple[bool, List[str]]:
    errors = []
    
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        errors.append("船员不存在")
        return False, errors
    
    if crew.status != CrewStatus.AVAILABLE.value:
        errors.append(f"船员状态为'{crew.status}'，不是待派状态，无法分配")
    
    if not check_ship_capacity(db, ship_id):
        errors.append("船舶已满员，没有空缺")
    
    certs_valid, _ = check_crew_certificates_valid(db, crew_id)
    if not certs_valid:
        errors.append("船员有过期证书，无法上船")
    
    if is_watch_keeper and crew.is_intern:
        errors.append("实习生不能单独分配为值班人员")
    
    return len(errors) == 0, errors


def create_on_board_attendance(db: Session, crew_id: int, ship_id: int, 
                               assignment_id: int, start_date: date):
    current = start_date
    today = date.today()
    
    while current <= today:
        existing = db.query(AttendanceRecord).filter(
            AttendanceRecord.crew_id == crew_id,
            AttendanceRecord.attendance_date == current
        ).first()
        
        if not existing:
            record = AttendanceRecord(
                crew_id=crew_id,
                attendance_date=current,
                attendance_type=AttendanceType.NORMAL.value,
                assignment_id=assignment_id,
                ship_id=ship_id,
                is_half_day=False
            )
            db.add(record)
        
        current += timedelta(days=1)
    
    db.commit()


def update_on_board_status(db: Session, crew_id: int, assignment_id: int, start_date: date):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        return
    
    crew.status = CrewStatus.ON_BOARD.value
    db.commit()
    
    create_on_board_attendance(db, crew_id, None, assignment_id, start_date)


def calculate_salary_coefficient(attendance_type: str) -> float:
    coefficients = {
        AttendanceType.NORMAL.value: 1.0,
        AttendanceType.OVERTIME.value: 1.5,
        AttendanceType.SICK_LEAVE.value: 0.6,
        AttendanceType.PERSONAL_LEAVE.value: 0.0
    }
    return coefficients.get(attendance_type, 0.0)


def calculate_monthly_salary(db: Session, crew_id: int, year: int, month: int,
                             base_salary: float = 0.0, ship_id: Optional[int] = None,
                             assignment_id: Optional[int] = None, 
                             end_date: Optional[date] = None) -> SalaryRecord:
    if end_date is None:
        if month == 12:
            end_date = date(year + 1, 1, 1) - timedelta(days=1)
        else:
            end_date = date(year, month + 1, 1) - timedelta(days=1)
    
    start_date = date(year, month, 1)
    if end_date > date.today():
        end_date = date.today()
    
    records = db.query(AttendanceRecord).filter(
        AttendanceRecord.crew_id == crew_id,
        AttendanceRecord.attendance_date >= start_date,
        AttendanceRecord.attendance_date <= end_date
    ).all()
    
    normal_days = 0.0
    overtime_days = 0.0
    sick_leave_days = 0.0
    personal_leave_days = 0.0
    total_amount = 0.0
    
    daily_rate = base_salary / 30.0 if base_salary > 0 else 1.0
    
    for record in records:
        day_value = 0.5 if record.is_half_day else 1.0
        coefficient = calculate_salary_coefficient(record.attendance_type)
        
        if record.attendance_type == AttendanceType.NORMAL.value:
            normal_days += day_value
            total_amount += daily_rate * day_value * coefficient
        elif record.attendance_type == AttendanceType.OVERTIME.value:
            overtime_days += day_value
            total_amount += daily_rate * day_value * coefficient
        elif record.attendance_type == AttendanceType.SICK_LEAVE.value:
            sick_leave_days += day_value
            total_amount += daily_rate * day_value * coefficient
        elif record.attendance_type == AttendanceType.PERSONAL_LEAVE.value:
            personal_leave_days += day_value
    
    existing = db.query(SalaryRecord).filter(
        SalaryRecord.crew_id == crew_id,
        SalaryRecord.year == year,
        SalaryRecord.month == month
    ).first()
    
    if existing:
        existing.base_salary = base_salary
        existing.normal_days = normal_days
        existing.overtime_days = overtime_days
        existing.sick_leave_days = sick_leave_days
        existing.personal_leave_days = personal_leave_days
        existing.total_amount = total_amount
        existing.ship_id = ship_id
        existing.assignment_id = assignment_id
        salary_record = existing
    else:
        salary_record = SalaryRecord(
            crew_id=crew_id,
            year=year,
            month=month,
            base_salary=base_salary,
            normal_days=normal_days,
            overtime_days=overtime_days,
            sick_leave_days=sick_leave_days,
            personal_leave_days=personal_leave_days,
            total_amount=total_amount,
            ship_id=ship_id,
            assignment_id=assignment_id
        )
        db.add(salary_record)
    
    db.commit()
    db.refresh(salary_record)
    return salary_record


def off_board_crew(db: Session, crew_id: int, end_date: date, base_salary: float = 0.0):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        raise ValueError("船员不存在")
    
    active_assignment = db.query(Assignment).filter(
        Assignment.crew_id == crew_id,
        Assignment.is_active == True
    ).first()
    
    if not active_assignment:
        raise ValueError("船员没有在船分配记录")
    
    today_record = db.query(AttendanceRecord).filter(
        AttendanceRecord.crew_id == crew_id,
        AttendanceRecord.attendance_date == end_date
    ).first()
    
    if not today_record:
        today_record = AttendanceRecord(
            crew_id=crew_id,
            attendance_date=end_date,
            attendance_type=AttendanceType.NORMAL.value,
            assignment_id=active_assignment.id,
            ship_id=active_assignment.ship_id,
            is_half_day=True
        )
        db.add(today_record)
    else:
        today_record.is_half_day = True
    
    active_assignment.is_active = False
    active_assignment.end_date = end_date
    
    crew.status = CrewStatus.OFF_BOARD.value
    
    calculate_monthly_salary(
        db=db,
        crew_id=crew_id,
        year=end_date.year,
        month=end_date.month,
        base_salary=base_salary,
        ship_id=active_assignment.ship_id,
        assignment_id=active_assignment.id,
        end_date=end_date
    )
    
    db.commit()


def get_ship_crew_list(db: Session, ship_id: int) -> List[Assignment]:
    return db.query(Assignment).filter(
        Assignment.ship_id == ship_id,
        Assignment.is_active == True
    ).all()


def get_monthly_attendance_summary(db: Session, crew_id: int, year: int, month: int) -> dict:
    start_date = date(year, month, 1)
    if month == 12:
        end_date = date(year + 1, 1, 1) - timedelta(days=1)
    else:
        end_date = date(year, month + 1, 1) - timedelta(days=1)
    
    records = db.query(AttendanceRecord).filter(
        AttendanceRecord.crew_id == crew_id,
        AttendanceRecord.attendance_date >= start_date,
        AttendanceRecord.attendance_date <= end_date
    ).all()
    
    normal = 0.0
    overtime = 0.0
    sick = 0.0
    personal = 0.0
    
    for record in records:
        day_val = 0.5 if record.is_half_day else 1.0
        if record.attendance_type == AttendanceType.NORMAL.value:
            normal += day_val
        elif record.attendance_type == AttendanceType.OVERTIME.value:
            overtime += day_val
        elif record.attendance_type == AttendanceType.SICK_LEAVE.value:
            sick += day_val
        elif record.attendance_type == AttendanceType.PERSONAL_LEAVE.value:
            personal += day_val
    
    return {
        "year": year,
        "month": month,
        "normal_days": normal,
        "overtime_days": overtime,
        "sick_leave_days": sick,
        "personal_leave_days": personal,
        "total_days": normal + overtime + sick + personal
    }
