from datetime import date, time
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import TrainingClass, ClassRegistration
from app.schemas import (
    TrainingClassCreate, TrainingClassResponse,
    ClassRegistrationCreate, ClassRegistrationResponse
)
from app.services.class_service import (
    create_training_class, book_class_venue,
    register_for_class, cancel_training_class
)

router = APIRouter(prefix="/api/classes", tags=["培训班管理"])


@router.post("/", response_model=TrainingClassResponse)
def create_class_endpoint(data: TrainingClassCreate, db: Session = Depends(get_db)):
    try:
        return create_training_class(
            db, data.name, data.instructor, data.venue_type_id,
            data.start_date, data.end_date, data.class_time,
            data.duration_hours, data.min_students, data.max_students,
            data.price_per_student, data.description
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[TrainingClassResponse])
def list_classes(status: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(TrainingClass)
    if status:
        query = query.filter(TrainingClass.status == status)
    return query.order_by(TrainingClass.created_at.desc()).all()


@router.get("/{class_id}", response_model=TrainingClassResponse)
def get_class(class_id: int, db: Session = Depends(get_db)):
    training_class = db.query(TrainingClass).filter(TrainingClass.id == class_id).first()
    if not training_class:
        raise HTTPException(status_code=404, detail="培训班不存在")
    return training_class


@router.post("/{class_id}/book-venue")
def book_class_venue_endpoint(class_id: int, venue_id: int, admin_user_id: int,
                               db: Session = Depends(get_db)):
    try:
        bookings = book_class_venue(db, class_id, venue_id, admin_user_id)
        return {
            "class_id": class_id,
            "bookings_count": len(bookings),
            "first_booking_id": bookings[0].id if bookings else None,
        }
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/{class_id}/register", response_model=ClassRegistrationResponse)
def register_endpoint(class_id: int, data: ClassRegistrationCreate,
                       db: Session = Depends(get_db)):
    try:
        return register_for_class(
            db, class_id, data.user_id,
            data.student_name, data.phone
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/{class_id}/registrations", response_model=List[ClassRegistrationResponse])
def list_registrations(class_id: int, db: Session = Depends(get_db)):
    return db.query(ClassRegistration).filter(
        ClassRegistration.class_id == class_id
    ).all()


@router.post("/{class_id}/cancel", response_model=TrainingClassResponse)
def cancel_class_endpoint(class_id: int, reason: str = "人数不足",
                           db: Session = Depends(get_db)):
    try:
        return cancel_training_class(db, class_id, reason)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
