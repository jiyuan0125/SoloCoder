from datetime import datetime
from typing import List, Optional
from sqlalchemy.orm import Session

from app.models.training import Training, TrainingStatus, Registration, RegistrationStatus, Attendance
from app.schemas.training import TrainingCreate, TrainingUpdate, RegistrationCreate, AttendanceCreate
from app.utils.notifications import send_notification, write_system_log


def get_training(db: Session, training_id: int) -> Optional[Training]:
    return db.query(Training).filter(Training.id == training_id).first()


def get_trainings(db: Session, skip: int = 0, limit: int = 100) -> List[Training]:
    return db.query(Training).offset(skip).limit(limit).all()


def create_training(db: Session, training: TrainingCreate) -> Training:
    db_training = Training(**training.model_dump())
    db.add(db_training)
    db.commit()
    db.refresh(db_training)
    return db_training


def update_training(db: Session, training_id: int, training_update: TrainingUpdate) -> Optional[Training]:
    db_training = get_training(db, training_id)
    if not db_training:
        return None
    
    for key, value in training_update.model_dump(exclude_unset=True).items():
        setattr(db_training, key, value)
    
    db.commit()
    db.refresh(db_training)
    return db_training


def delete_training(db: Session, training_id: int) -> bool:
    db_training = get_training(db, training_id)
    if not db_training:
        return False
    
    db.delete(db_training)
    db.commit()
    return True


def get_confirmed_registrations_count(db: Session, training_id: int) -> int:
    return db.query(Registration).filter(
        Registration.training_id == training_id,
        Registration.status == RegistrationStatus.CONFIRMED
    ).count()


def get_next_waitlist_order(db: Session, training_id: int) -> int:
    max_order = db.query(Registration).filter(
        Registration.training_id == training_id,
        Registration.status == RegistrationStatus.WAITLIST
    ).count()
    return max_order + 1


def create_registration(db: Session, registration: RegistrationCreate) -> Optional[Registration]:
    training = get_training(db, registration.training_id)
    if not training:
        return None
    
    confirmed_count = get_confirmed_registrations_count(db, registration.training_id)
    
    if confirmed_count < training.max_participants:
        status = RegistrationStatus.CONFIRMED
        waitlist_order = None
        confirmed_at = datetime.utcnow()
    else:
        status = RegistrationStatus.WAITLIST
        waitlist_order = get_next_waitlist_order(db, registration.training_id)
        confirmed_at = None
    
    db_registration = Registration(
        training_id=registration.training_id,
        student_name=registration.student_name,
        student_phone=registration.student_phone,
        student_email=registration.student_email,
        status=status,
        waitlist_order=waitlist_order,
        consecutive_absences=0,
        confirmed_at=confirmed_at
    )
    db.add(db_registration)
    db.commit()
    db.refresh(db_registration)
    return db_registration


def get_registration(db: Session, registration_id: int) -> Optional[Registration]:
    return db.query(Registration).filter(Registration.id == registration_id).first()


def get_registrations_by_training(db: Session, training_id: int) -> List[Registration]:
    return db.query(Registration).filter(Registration.training_id == training_id).all()


def get_waitlist_registrations(db: Session, training_id: int) -> List[Registration]:
    return db.query(Registration).filter(
        Registration.training_id == training_id,
        Registration.status == RegistrationStatus.WAITLIST
    ).order_by(Registration.waitlist_order).all()


def process_waitlist_notifications(db: Session, training_id: int) -> dict:
    waitlist = get_waitlist_registrations(db, training_id)
    if not waitlist:
        return {"notified": [], "skipped": [], "log_entry": None}
    
    training = get_training(db, training_id)
    if not training:
        return {"notified": [], "skipped": [], "log_entry": None}
    
    confirmed_count = get_confirmed_registrations_count(db, training_id)
    available_slots = training.max_participants - confirmed_count
    
    if available_slots <= 0:
        return {"notified": [], "skipped": [], "log_entry": None}
    
    notified = []
    skipped = []
    slots_filled = 0
    
    for reg in waitlist:
        if slots_filled >= available_slots:
            break
        
        has_contact = reg.student_phone or reg.student_email
        
        if has_contact:
            message = f"您已成功递补到培训《{training.name}》"
            if reg.student_phone:
                send_notification(reg.student_phone, message)
            
            reg.status = RegistrationStatus.CONFIRMED
            reg.waitlist_order = None
            reg.confirmed_at = datetime.utcnow()
            
            notified.append(reg.id)
            slots_filled += 1
        else:
            skipped.append(reg.id)
    
    for i, reg in enumerate(waitlist):
        if reg.status == RegistrationStatus.WAITLIST:
            reg.waitlist_order = i + 1 - slots_filled
    
    db.commit()
    
    if not notified and skipped:
        log_msg = f"培训{training_id}候补递补：所有候补学员联系方式为空，跳过通知"
        write_system_log(log_msg)
        return {
            "notified": [],
            "skipped": skipped,
            "log_entry": log_msg
        }
    
    return {
        "notified": notified,
        "skipped": skipped,
        "log_entry": None
    }


def cancel_registration(db: Session, registration_id: int) -> Optional[Registration]:
    db_reg = get_registration(db, registration_id)
    if not db_reg:
        return None
    
    if db_reg.status == RegistrationStatus.CANCELLED:
        return None
    
    db_reg.status = RegistrationStatus.CANCELLED
    db_reg.cancelled_at = datetime.utcnow()
    
    db.commit()
    
    process_waitlist_notifications(db, db_reg.training_id)
    
    db.refresh(db_reg)
    return db_reg


def create_attendance(db: Session, attendance: AttendanceCreate) -> Optional[Attendance]:
    registration = get_registration(db, attendance.registration_id)
    if not registration:
        return None
    
    if registration.status != RegistrationStatus.CONFIRMED:
        return None
    
    db_attendance = Attendance(
        registration_id=attendance.registration_id,
        session_date=attendance.session_date,
        is_present=attendance.is_present
    )
    db.add(db_attendance)
    
    if attendance.is_present == 1:
        registration.consecutive_absences = 0
    else:
        registration.consecutive_absences += 1
    
    if registration.consecutive_absences >= 3:
        registration.status = RegistrationStatus.CANCELLED
        registration.cancelled_at = datetime.utcnow()
        process_waitlist_notifications(db, registration.training_id)
    
    db.commit()
    db.refresh(db_attendance)
    return db_attendance


def get_attendance(db: Session, registration_id: int) -> List[Attendance]:
    return db.query(Attendance).filter(Attendance.registration_id == registration_id).all()
