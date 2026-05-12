from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime, date, time, timedelta
from app import models, schemas


def get_venue(db: Session, venue_id: int):
    return db.query(models.Venue).filter(models.Venue.id == venue_id).first()


def get_venues(db: Session, skip: int = 0, limit: int = 100):
    return db.query(models.Venue).offset(skip).limit(limit).all()


def create_venue(db: Session, venue: schemas.VenueCreate):
    db_venue = models.Venue(**venue.model_dump())
    db.add(db_venue)
    db.commit()
    db.refresh(db_venue)
    return db_venue


def update_venue(db: Session, venue_id: int, venue: schemas.VenueUpdate):
    db_venue = get_venue(db, venue_id)
    if db_venue:
        for key, value in venue.model_dump(exclude_unset=True).items():
            setattr(db_venue, key, value)
        db.commit()
        db.refresh(db_venue)
    return db_venue


def delete_venue(db: Session, venue_id: int):
    db_venue = get_venue(db, venue_id)
    if db_venue:
        db.delete(db_venue)
        db.commit()
    return db_venue


def get_hall(db: Session, hall_id: int):
    return db.query(models.Hall).filter(models.Hall.id == hall_id).first()


def get_halls_by_venue(db: Session, venue_id: int):
    return db.query(models.Hall).filter(models.Hall.venue_id == venue_id).all()


def create_hall(db: Session, hall: schemas.HallCreate):
    db_hall = models.Hall(**hall.model_dump())
    db.add(db_hall)
    db.commit()
    db.refresh(db_hall)
    return db_hall


def update_hall(db: Session, hall_id: int, hall: schemas.HallUpdate):
    db_hall = get_hall(db, hall_id)
    if db_hall:
        for key, value in hall.model_dump(exclude_unset=True).items():
            setattr(db_hall, key, value)
        db.commit()
        db.refresh(db_hall)
    return db_hall


def delete_hall(db: Session, hall_id: int):
    db_hall = get_hall(db, hall_id)
    if db_hall:
        db.delete(db_hall)
        db.commit()
    return db_hall


def get_booth(db: Session, booth_id: int):
    return db.query(models.Booth).filter(models.Booth.id == booth_id).first()


def get_booths_by_hall(db: Session, hall_id: int):
    return db.query(models.Booth).filter(models.Booth.hall_id == hall_id).all()


def get_available_booths(db: Session, exhibition_id: int, hall_id: Optional[int] = None):
    from app.services.scheduler import get_occupied_booths_for_exhibition
    occupied_booth_ids = get_occupied_booths_for_exhibition(db, exhibition_id)

    query = db.query(models.Booth)
    if hall_id:
        query = query.filter(models.Booth.hall_id == hall_id)

    booths = query.all()
    return [b for b in booths if b.id not in occupied_booth_ids and b.status == "available"]


def create_booth(db: Session, booth: schemas.BoothCreate):
    db_booth = models.Booth(**booth.model_dump())
    db.add(db_booth)
    db.commit()
    db.refresh(db_booth)
    return db_booth


def update_booth(db: Session, booth_id: int, booth: schemas.BoothUpdate):
    db_booth = get_booth(db, booth_id)
    if db_booth:
        for key, value in booth.model_dump(exclude_unset=True).items():
            setattr(db_booth, key, value)
        db.commit()
        db.refresh(db_booth)
    return db_booth


def update_booth_price(db: Session, booth_id: int, new_price: float):
    from app.services.pricing import sync_unconfirmed_prices

    db_booth = get_booth(db, booth_id)
    if db_booth:
        old_price = db_booth.base_price
        db_booth.base_price = new_price
        db.commit()
        db.refresh(db_booth)

        if old_price != new_price:
            sync_unconfirmed_prices(db, booth_id, new_price)

    return db_booth


def delete_booth(db: Session, booth_id: int):
    db_booth = get_booth(db, booth_id)
    if db_booth:
        db.delete(db_booth)
        db.commit()
    return db_booth


def get_exhibition(db: Session, exhibition_id: int):
    return db.query(models.Exhibition).filter(models.Exhibition.id == exhibition_id).first()


def get_exhibitions(db: Session, skip: int = 0, limit: int = 100):
    return db.query(models.Exhibition).offset(skip).limit(limit).all()


def get_exhibition_halls(db: Session, exhibition_id: int):
    return db.query(models.ExhibitionHall).filter(
        models.ExhibitionHall.exhibition_id == exhibition_id
    ).all()


def create_exhibition(db: Session, exhibition: schemas.ExhibitionCreate):
    from app.services.scheduler import validate_exhibition_halls

    hall_ids = exhibition.hall_ids
    validate_exhibition_halls(db, hall_ids, exhibition.start_date, exhibition.end_date)

    db_exhibition = models.Exhibition(
        name=exhibition.name,
        organizer=exhibition.organizer,
        start_date=exhibition.start_date,
        end_date=exhibition.end_date,
        description=exhibition.description
    )
    db.add(db_exhibition)
    db.commit()
    db.refresh(db_exhibition)

    from datetime import timedelta
    current_date = exhibition.start_date
    while current_date <= exhibition.end_date:
        for hall_id in hall_ids:
            db_exhibition_hall = models.ExhibitionHall(
                exhibition_id=db_exhibition.id,
                hall_id=hall_id,
                exhibition_date=current_date
            )
            db.add(db_exhibition_hall)
        current_date += timedelta(days=1)

    db.commit()
    return db_exhibition


def update_exhibition(db: Session, exhibition_id: int, exhibition: schemas.ExhibitionUpdate):
    db_exhibition = get_exhibition(db, exhibition_id)
    if db_exhibition:
        for key, value in exhibition.model_dump(exclude_unset=True).items():
            setattr(db_exhibition, key, value)
        db.commit()
        db.refresh(db_exhibition)
    return db_exhibition


def delete_exhibition(db: Session, exhibition_id: int):
    db_exhibition = get_exhibition(db, exhibition_id)
    if db_exhibition:
        db.delete(db_exhibition)
        db.commit()
    return db_exhibition


def get_booth_selection(db: Session, selection_id: int):
    return db.query(models.BoothSelection).filter(models.BoothSelection.id == selection_id).first()


def get_booth_selections_by_exhibition(db: Session, exhibition_id: int):
    return db.query(models.BoothSelection).filter(
        models.BoothSelection.exhibition_id == exhibition_id
    ).all()


def get_expired_selections(db: Session):
    now = datetime.utcnow()
    return db.query(models.BoothSelection).filter(
        models.BoothSelection.status == "selected",
        models.BoothSelection.expires_at < now
    ).all()


def create_booth_selection(db: Session, selection: schemas.BoothSelectionCreate):
    from app.services.scheduler import validate_booth_selection

    validate_booth_selection(db, selection.booth_id, selection.exhibition_id)

    db_booth = get_booth(db, selection.booth_id)

    db_selection = models.BoothSelection(
        booth_id=selection.booth_id,
        exhibition_id=selection.exhibition_id,
        exhibitor_name=selection.exhibitor_name,
        contact_person=selection.contact_person,
        contact_phone=selection.contact_phone,
        final_price=db_booth.base_price,
        notes=selection.notes
    )
    db_selection.set_expiration(48)

    db.add(db_selection)
    db_booth.status = "selected"
    db.commit()
    db.refresh(db_selection)
    return db_selection


def confirm_booth_selection(db: Session, selection_id: int):
    db_selection = get_booth_selection(db, selection_id)
    if db_selection and db_selection.status == "selected":
        if db_selection.is_expired:
            raise ValueError("Selection has expired")
        db_selection.status = "confirmed"
        db_selection.is_paid = True
        db_selection.paid_at = datetime.utcnow()

        db_booth = get_booth(db, db_selection.booth_id)
        if db_booth:
            db_booth.status = "occupied"

        db.commit()
        db.refresh(db_selection)
    return db_selection


def release_expired_selection(db: Session, selection_id: int):
    db_selection = get_booth_selection(db, selection_id)
    if db_selection and db_selection.status == "selected":
        db_selection.status = "expired"

        db_booth = get_booth(db, db_selection.booth_id)
        if db_booth:
            db_booth.status = "available"

        db.commit()
    return db_selection


def get_setup_schedule(db: Session, schedule_id: int):
    return db.query(models.SetupSchedule).filter(models.SetupSchedule.id == schedule_id).first()


def get_setup_schedules_by_hall(db: Session, hall_id: int, setup_date: Optional[date] = None):
    query = db.query(models.SetupSchedule).filter(models.SetupSchedule.hall_id == hall_id)
    if setup_date:
        query = query.filter(models.SetupSchedule.setup_date == setup_date)
    return query.all()


def get_setup_schedules_by_selection(db: Session, selection_id: int):
    return db.query(models.SetupSchedule).filter(
        models.SetupSchedule.selection_id == selection_id
    ).first()


def create_setup_schedule(db: Session, schedule: schemas.SetupScheduleCreate):
    from app.services.scheduler import validate_setup_schedule, calculate_billed_hours

    db_selection = get_booth_selection(db, schedule.selection_id)
    if not db_selection:
        raise ValueError("Booth selection not found")

    db_booth = get_booth(db, db_selection.booth_id)
    if not db_booth:
        raise ValueError("Booth not found")

    validate_setup_schedule(db, db_booth.hall_id, schedule.setup_date, schedule.start_time,
                            schedule.end_time, db_booth.is_special, db_booth.area)

    actual_hours, billed_hours = calculate_billed_hours(schedule.start_time, schedule.end_time)

    if db_booth.is_special:
        billed_hours += 4

    db_schedule = models.SetupSchedule(
        hall_id=db_booth.hall_id,
        selection_id=schedule.selection_id,
        setup_date=schedule.setup_date,
        start_time=schedule.start_time,
        end_time=schedule.end_time,
        actual_hours=actual_hours,
        billed_hours=billed_hours,
        notes=schedule.notes
    )

    db.add(db_schedule)
    db.commit()
    db.refresh(db_schedule)
    return db_schedule


def get_visit_reservation(db: Session, reservation_id: int):
    return db.query(models.VisitReservation).filter(models.VisitReservation.id == reservation_id).first()


def get_reservations_by_exhibition(db: Session, exhibition_id: int, visit_date: Optional[date] = None):
    query = db.query(models.VisitReservation).filter(
        models.VisitReservation.exhibition_id == exhibition_id
    )
    if visit_date:
        query = query.filter(models.VisitReservation.visit_date == visit_date)
    return query.all()


def get_daily_capacity(db: Session, exhibition_id: int, visit_date: date):
    return db.query(models.DailyCapacity).filter(
        models.DailyCapacity.exhibition_id == exhibition_id,
        models.DailyCapacity.date == visit_date
    ).first()


def create_daily_capacity(db: Session, capacity: schemas.DailyCapacityCreate):
    db_capacity = models.DailyCapacity(**capacity.model_dump())
    db.add(db_capacity)
    db.commit()
    db.refresh(db_capacity)
    return db_capacity


def create_visit_reservation(db: Session, reservation: schemas.VisitReservationCreate):
    from app.services.scheduler import validate_visit_reservation

    queue_position = validate_visit_reservation(
        db, reservation.exhibition_id, reservation.visit_date, reservation.party_size
    )

    status = "confirmed" if queue_position == 0 else "queued"

    db_reservation = models.VisitReservation(
        exhibition_id=reservation.exhibition_id,
        visitor_name=reservation.visitor_name,
        visitor_phone=reservation.visitor_phone,
        visitor_email=reservation.visitor_email,
        visit_date=reservation.visit_date,
        party_size=reservation.party_size,
        status=status,
        queue_position=queue_position if queue_position > 0 else None
    )

    db.add(db_reservation)

    if status == "confirmed":
        capacity = get_daily_capacity(db, reservation.exhibition_id, reservation.visit_date)
        if capacity:
            capacity.current_confirmed += reservation.party_size

    db.commit()
    db.refresh(db_reservation)
    return db_reservation
