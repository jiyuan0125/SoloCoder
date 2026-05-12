from sqlalchemy.orm import Session
from typing import List, Set, Tuple
from datetime import datetime, date, time, timedelta
from app import models


def time_overlaps(start1: time, end1: time, start2: time, end2: time) -> bool:
    return start1 < end2 and start2 < end1


def validate_exhibition_halls(db: Session, hall_ids: List[int],
                               start_date: date, end_date: date):
    for hall_id in hall_ids:
        current_date = start_date
        while current_date <= end_date:
            existing = db.query(models.ExhibitionHall).filter(
                models.ExhibitionHall.hall_id == hall_id,
                models.ExhibitionHall.exhibition_date == current_date
            ).first()

            if existing:
                raise ValueError(
                    f"Hall {hall_id} is already occupied on {current_date}"
                )
            current_date += timedelta(days=1)


def get_occupied_booths_for_exhibition(db: Session, exhibition_id: int) -> Set[int]:
    selections = db.query(models.BoothSelection).filter(
        models.BoothSelection.exhibition_id == exhibition_id,
        models.BoothSelection.status.in_(["selected", "confirmed"])
    ).all()
    return {s.booth_id for s in selections}


def validate_booth_selection(db: Session, booth_id: int, exhibition_id: int):
    occupied = get_occupied_booths_for_exhibition(db, exhibition_id)
    if booth_id in occupied:
        raise ValueError("Booth is already occupied for this exhibition")

    booth = db.query(models.Booth).filter(models.Booth.id == booth_id).first()
    if not booth:
        raise ValueError("Booth not found")

    if booth.status != "available":
        raise ValueError("Booth is not available")


def get_concentrated_booths(db: Session, exhibition_id: int, hall_ids: List[int],
                             requested_count: int) -> List[models.Booth]:
    from app import crud

    hall_booth_map = {}
    for hall_id in hall_ids:
        available_booths = crud.get_available_booths(db, exhibition_id, hall_id)
        if available_booths:
            hall_booth_map[hall_id] = sorted(
                available_booths,
                key=lambda b: (
                    int(''.join(c for c in b.booth_number if c.isdigit()) or '0'),
                    b.booth_number
                )
            )

    for hall_id, booths in hall_booth_map.items():
        if len(booths) >= requested_count:
            return booths[:requested_count]

    result = []
    sorted_halls = sorted(
        hall_booth_map.items(),
        key=lambda x: len(x[1]),
        reverse=True
    )
    for hall_id, booths in sorted_halls:
        result.extend(booths)
        if len(result) >= requested_count:
            break

    return result[:requested_count]


def validate_setup_schedule(db: Session, hall_id: int, setup_date: date,
                            start_time: time, end_time: time,
                            is_special: bool, booth_area: float):
    hall = db.query(models.Hall).filter(models.Hall.id == hall_id).first()
    if not hall:
        raise ValueError("Hall not found")

    is_exclusive = is_special and booth_area > 36

    existing_schedules = db.query(models.SetupSchedule).filter(
        models.SetupSchedule.hall_id == hall_id,
        models.SetupSchedule.setup_date == setup_date
    ).all()

    if is_exclusive:
        if existing_schedules:
            raise ValueError(
                "Special booth over 36 sqm requires exclusive access; "
                "hall already has schedules for this date"
            )
        return

    overlapping_schedules = []
    for schedule in existing_schedules:
        if time_overlaps(start_time, end_time, schedule.start_time, schedule.end_time):
            selection = db.query(models.BoothSelection).filter(
                models.BoothSelection.id == schedule.selection_id
            ).first()
            if selection:
                booth = db.query(models.Booth).filter(
                    models.Booth.id == selection.booth_id
                ).first()
                if booth and booth.is_special and booth.area > 36:
                    raise ValueError(
                        "Time slot overlaps with an exclusive special booth schedule"
                    )
            overlapping_schedules.append(schedule)

    if len(overlapping_schedules) >= hall.max_exhibitors_per_slot:
        raise ValueError(
            f"Maximum {hall.max_exhibitors_per_slot} exhibitors "
            "per time slot reached for this hall"
        )


def calculate_billed_hours(start_time: time, end_time: time) -> Tuple[float, int]:
    start_dt = datetime.combine(date.min, start_time)
    end_dt = datetime.combine(date.min, end_time)

    if end_dt <= start_dt:
        end_dt += timedelta(days=1)

    delta = end_dt - start_dt
    actual_hours = delta.total_seconds() / 3600
    billed_hours = int(actual_hours) if actual_hours == int(actual_hours) else int(actual_hours) + 1

    return actual_hours, billed_hours


def validate_visit_reservation(db: Session, exhibition_id: int,
                                visit_date: date, party_size: int) -> int:
    exhibition = db.query(models.Exhibition).filter(
        models.Exhibition.id == exhibition_id
    ).first()
    if not exhibition:
        raise ValueError("Exhibition not found")

    if not (exhibition.start_date <= visit_date <= exhibition.end_date):
        raise ValueError("Visit date is outside exhibition dates")

    capacity = db.query(models.DailyCapacity).filter(
        models.DailyCapacity.exhibition_id == exhibition_id,
        models.DailyCapacity.date == visit_date
    ).first()

    if not capacity:
        raise ValueError(f"No capacity configured for {visit_date}")

    confirmed = db.query(models.VisitReservation).filter(
        models.VisitReservation.exhibition_id == exhibition_id,
        models.VisitReservation.visit_date == visit_date,
        models.VisitReservation.status == "confirmed"
    ).all()

    total_confirmed = sum(r.party_size for r in confirmed)
    remaining = capacity.max_visitors - total_confirmed

    if remaining >= party_size:
        return 0

    queued = db.query(models.VisitReservation).filter(
        models.VisitReservation.exhibition_id == exhibition_id,
        models.VisitReservation.visit_date == visit_date,
        models.VisitReservation.status == "queued"
    ).order_by(models.VisitReservation.queue_position.asc()).all()

    last_position = max((r.queue_position for r in queued), default=0)
    return last_position + 1


def release_expired_selections(db: Session):
    from app import crud

    expired = crud.get_expired_selections(db)
    for selection in expired:
        crud.release_expired_selection(db, selection.id)

    return len(expired)
