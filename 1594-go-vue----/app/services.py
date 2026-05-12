from datetime import datetime, date, time, timedelta
from decimal import Decimal, ROUND_HALF_UP
from typing import Optional, List, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import and_

from app.models import (
    Reader,
    Book,
    BookCopy,
    Borrowing,
    BookReservation,
    Seat,
    SeatReservation,
    Event,
    EventRegistration,
    Notification,
    ReaderType,
    CopyStatus,
    ReservationStatus,
    SeatReservationStatus,
    EventRegistrationStatus,
    READER_RULES,
    SEAT_TIME_SLOTS,
)
from app.schemas import (
    BorrowingCreate,
    BorrowingReturn,
    BorrowingRenew,
    BookReservationCreate,
    SeatReservationCreate,
    EventRegistrationCreate,
    EventRegistrationCancel,
    LateFeeInfo,
)


DAILY_LATE_FEE = Decimal("0.10")
MAX_LATE_FEE_RATIO = Decimal("0.50")
CHECKIN_GRACE_MINUTES = 15


def calculate_late_fee(due_date: date, return_date: date, book_price: Decimal) -> LateFeeInfo:
    if return_date <= due_date:
        return LateFeeInfo(
            overdue_days=0,
            daily_rate=DAILY_LATE_FEE,
            max_fee=book_price * MAX_LATE_FEE_RATIO,
            calculated_fee=Decimal("0.00"),
            final_fee=Decimal("0.00"),
        )

    overdue_days = (return_date - due_date).days
    max_fee = (book_price * MAX_LATE_FEE_RATIO).quantize(Decimal("0.01"), rounding=ROUND_HALF_UP)
    calculated_fee = (DAILY_LATE_FEE * overdue_days).quantize(Decimal("0.01"), rounding=ROUND_HALF_UP)
    final_fee = min(calculated_fee, max_fee)

    return LateFeeInfo(
        overdue_days=overdue_days,
        daily_rate=DAILY_LATE_FEE,
        max_fee=max_fee,
        calculated_fee=calculated_fee,
        final_fee=final_fee,
    )


def get_reader_rules(reader_type: str) -> dict:
    return READER_RULES.get(ReaderType(reader_type), READER_RULES[ReaderType.NORMAL])


def count_active_borrowings(db: Session, reader_id: int) -> int:
    return db.query(Borrowing).filter(
        Borrowing.reader_id == reader_id,
        Borrowing.is_returned == False,
    ).count()


def has_overdue_borrowings(db: Session, reader_id: int) -> bool:
    today = date.today()
    return db.query(Borrowing).filter(
        Borrowing.reader_id == reader_id,
        Borrowing.is_returned == False,
        Borrowing.due_date < today,
    ).first() is not None


def can_borrow(db: Session, reader_id: int, copy_id: int) -> Tuple[bool, str]:
    reader = db.query(Reader).filter(Reader.id == reader_id).first()
    if not reader:
        return False, "读者不存在"
    if not reader.is_active:
        return False, "读者账户已停用"

    copy = db.query(BookCopy).filter(BookCopy.id == copy_id).first()
    if not copy:
        return False, "图书副本不存在"
    if copy.status != CopyStatus.AVAILABLE.value:
        return False, f"图书副本状态为 {copy.status}，不可借阅"

    rules = get_reader_rules(reader.reader_type)
    active_count = count_active_borrowings(db, reader_id)
    if active_count >= rules["max_books"]:
        return False, f"已达到最大借阅数量 {rules['max_books']} 本"

    if has_overdue_borrowings(db, reader_id):
        return False, "存在逾期未还图书，请先归还"

    existing_borrowing = db.query(Borrowing).filter(
        Borrowing.reader_id == reader_id,
        Borrowing.copy_id == copy_id,
        Borrowing.is_returned == False,
    ).first()
    if existing_borrowing:
        return False, "您已借阅该副本"

    existing_reservation = db.query(BookReservation).filter(
        BookReservation.copy_id == copy_id,
        BookReservation.status == ReservationStatus.PENDING.value,
    ).first()
    if existing_reservation and existing_reservation.reader_id != reader_id:
        return False, "该副本已被其他读者预约"

    return True, ""


def create_borrowing(db: Session, borrowing_data: BorrowingCreate) -> Tuple[Optional[Borrowing], str]:
    can_do, msg = can_borrow(db, borrowing_data.reader_id, borrowing_data.copy_id)
    if not can_do:
        return None, msg

    reader = db.query(Reader).filter(Reader.id == borrowing_data.reader_id).first()
    copy = db.query(BookCopy).filter(BookCopy.id == borrowing_data.copy_id).first()

    rules = get_reader_rules(reader.reader_type)
    today = date.today()
    due_date = today + timedelta(days=rules["loan_days"])

    borrowing = Borrowing(
        reader_id=borrowing_data.reader_id,
        copy_id=borrowing_data.copy_id,
        borrow_date=today,
        due_date=due_date,
        renew_count=0,
        max_renew_count=1,
        is_returned=False,
    )

    copy.status = CopyStatus.BORROWED.value

    db.add(borrowing)
    db.commit()
    db.refresh(borrowing)
    db.refresh(copy)

    return borrowing, ""


def can_renew(db: Session, borrowing_id: int) -> Tuple[bool, str]:
    borrowing = db.query(Borrowing).filter(Borrowing.id == borrowing_id).first()
    if not borrowing:
        return False, "借阅记录不存在"
    if borrowing.is_returned:
        return False, "该书已归还"
    if borrowing.renew_count >= borrowing.max_renew_count:
        return False, "已达到最大续借次数"

    today = date.today()
    if today >= borrowing.due_date:
        return False, "已逾期，无法续借"

    pending_reservation = db.query(BookReservation).filter(
        BookReservation.copy_id == borrowing.copy_id,
        BookReservation.status == ReservationStatus.PENDING.value,
    ).first()
    if pending_reservation:
        return False, "该图书已被其他读者预约，无法续借"

    return True, ""


def renew_borrowing(db: Session, renew_data: BorrowingRenew) -> Tuple[Optional[Borrowing], str]:
    can_do, msg = can_renew(db, renew_data.borrowing_id)
    if not can_do:
        return None, msg

    borrowing = db.query(Borrowing).filter(Borrowing.id == renew_data.borrowing_id).first()
    original_days = (borrowing.due_date - borrowing.borrow_date).days
    extend_days = original_days // 2

    borrowing.due_date += timedelta(days=extend_days)
    borrowing.renew_count += 1

    db.commit()
    db.refresh(borrowing)

    return borrowing, ""


def create_notification(db: Session, reader_id: int, title: str, message: str, notification_type: str = None) -> Notification:
    notification = Notification(
        reader_id=reader_id,
        title=title,
        message=message,
        notification_type=notification_type,
        is_read=False,
    )
    db.add(notification)
    db.commit()
    db.refresh(notification)
    return notification


def notify_next_reservation(db: Session, copy_id: int, book_id: int) -> Optional[BookReservation]:
    pending_reservations = db.query(BookReservation).filter(
        BookReservation.book_id == book_id,
        BookReservation.status == ReservationStatus.PENDING.value,
    ).order_by(BookReservation.queue_position.asc()).all()

    if not pending_reservations:
        return None

    next_reservation = pending_reservations[0]
    next_reservation.copy_id = copy_id
    next_reservation.status = ReservationStatus.FULFILLED.value
    next_reservation.fulfilled_at = datetime.utcnow()
    next_reservation.expires_at = datetime.utcnow() + timedelta(days=3)
    next_reservation.notification_sent = True

    copy = db.query(BookCopy).filter(BookCopy.id == copy_id).first()
    if copy:
        copy.status = CopyStatus.RESERVED.value

    create_notification(
        db,
        reader_id=next_reservation.reader_id,
        title="预约图书已到馆",
        message=f"您预约的图书已到馆，请在3天内前往图书馆借阅。",
        notification_type="reservation_fulfilled",
    )

    for i, reservation in enumerate(pending_reservations[1:], start=1):
        reservation.queue_position = i

    db.commit()
    db.refresh(next_reservation)
    if copy:
        db.refresh(copy)

    return next_reservation


def return_borrowing(db: Session, return_data: BorrowingReturn) -> Tuple[Optional[Borrowing], str]:
    borrowing = db.query(Borrowing).filter(Borrowing.id == return_data.borrowing_id).first()
    if not borrowing:
        return None, "借阅记录不存在"
    if borrowing.is_returned:
        return None, "该书已归还"

    today = date.today()
    copy = db.query(BookCopy).filter(BookCopy.id == borrowing.copy_id).first()
    book = db.query(Book).filter(Book.id == copy.book_id).first()

    fee_info = calculate_late_fee(borrowing.due_date, today, book.price)

    borrowing.return_date = today
    borrowing.is_returned = True
    borrowing.late_fee = fee_info.final_fee

    has_pending = notify_next_reservation(db, borrowing.copy_id, copy.book_id)
    if not has_pending:
        copy.status = CopyStatus.AVAILABLE.value
    else:
        copy.status = CopyStatus.RESERVED.value

    db.commit()
    db.refresh(borrowing)
    db.refresh(copy)

    return borrowing, ""


def create_book_reservation(db: Session, reservation_data: BookReservationCreate) -> Tuple[Optional[BookReservation], str]:
    reader = db.query(Reader).filter(Reader.id == reservation_data.reader_id).first()
    if not reader or not reader.is_active:
        return None, "读者不存在或已停用"

    book = db.query(Book).filter(Book.id == reservation_data.book_id).first()
    if not book:
        return None, "图书不存在"

    available_copy = db.query(BookCopy).filter(
        BookCopy.book_id == book.id,
        BookCopy.status == CopyStatus.AVAILABLE.value,
    ).first()
    if available_copy:
        return None, "该图书有可借副本，无需预约"

    existing_reservation = db.query(BookReservation).filter(
        BookReservation.reader_id == reservation_data.reader_id,
        BookReservation.book_id == reservation_data.book_id,
        BookReservation.status.in_([
            ReservationStatus.PENDING.value,
            ReservationStatus.FULFILLED.value,
        ]),
    ).first()
    if existing_reservation:
        return None, "您已预约该图书"

    pending_count = db.query(BookReservation).filter(
        BookReservation.book_id == book.id,
        BookReservation.status == ReservationStatus.PENDING.value,
    ).count()

    reservation = BookReservation(
        reader_id=reservation_data.reader_id,
        book_id=reservation_data.book_id,
        status=ReservationStatus.PENDING.value,
        queue_position=pending_count + 1,
    )

    db.add(reservation)
    db.commit()
    db.refresh(reservation)

    return reservation, ""


def is_valid_time_slot(start_time: time, end_time: time) -> bool:
    return (start_time, end_time) in SEAT_TIME_SLOTS


def has_conflicting_reservation(db: Session, reader_id: int, reservation_date: date, start_time: time, end_time: time) -> bool:
    conflict = db.query(SeatReservation).filter(
        SeatReservation.reader_id == reader_id,
        SeatReservation.reservation_date == reservation_date,
        SeatReservation.status.in_([
            SeatReservationStatus.RESERVED.value,
            SeatReservationStatus.CHECKED_IN.value,
        ]),
        SeatReservation.start_time < end_time,
        SeatReservation.end_time > start_time,
    ).first()
    return conflict is not None


def is_seat_available(db: Session, seat_id: int, reservation_date: date, start_time: time, end_time: time) -> bool:
    conflict = db.query(SeatReservation).filter(
        SeatReservation.seat_id == seat_id,
        SeatReservation.reservation_date == reservation_date,
        SeatReservation.status.in_([
            SeatReservationStatus.RESERVED.value,
            SeatReservationStatus.CHECKED_IN.value,
        ]),
        SeatReservation.start_time < end_time,
        SeatReservation.end_time > start_time,
    ).first()
    return conflict is None


def create_seat_reservation(db: Session, reservation_data: SeatReservationCreate) -> Tuple[Optional[SeatReservation], str]:
    reader = db.query(Reader).filter(Reader.id == reservation_data.reader_id).first()
    if not reader or not reader.is_active:
        return None, "读者不存在或已停用"

    seat = db.query(Seat).filter(Seat.id == reservation_data.seat_id).first()
    if not seat or not seat.is_active:
        return None, "座位不存在或不可用"

    if not is_valid_time_slot(reservation_data.start_time, reservation_data.end_time):
        return None, "无效的时段，请选择2小时档（8-10, 10-12等）"

    today = date.today()
    if reservation_data.reservation_date < today:
        return None, "不能预约过去的日期"

    if has_conflicting_reservation(
        db,
        reservation_data.reader_id,
        reservation_data.reservation_date,
        reservation_data.start_time,
        reservation_data.end_time,
    ):
        return None, "同一时段您已有座位预约"

    if not is_seat_available(
        db,
        reservation_data.seat_id,
        reservation_data.reservation_date,
        reservation_data.start_time,
        reservation_data.end_time,
    ):
        return None, "该时段座位已被预约"

    reservation = SeatReservation(
        reader_id=reservation_data.reader_id,
        seat_id=reservation_data.seat_id,
        reservation_date=reservation_data.reservation_date,
        start_time=reservation_data.start_time,
        end_time=reservation_data.end_time,
        status=SeatReservationStatus.RESERVED.value,
    )

    db.add(reservation)
    db.commit()
    db.refresh(reservation)

    return reservation, ""


def check_in_seat(db: Session, reservation_id: int) -> Tuple[Optional[SeatReservation], str]:
    reservation = db.query(SeatReservation).filter(SeatReservation.id == reservation_id).first()
    if not reservation:
        return None, "预约记录不存在"

    if reservation.status == SeatReservationStatus.CANCELLED.value:
        return None, "预约已取消"
    if reservation.status == SeatReservationStatus.EXPIRED.value:
        return None, "预约已过期"
    if reservation.status == SeatReservationStatus.CHECKED_IN.value:
        return None, "已签到"

    now = datetime.now()
    reservation_start = datetime.combine(reservation.reservation_date, reservation.start_time)
    grace_end = reservation_start + timedelta(minutes=CHECKIN_GRACE_MINUTES)

    if now < reservation_start:
        return None, "签到时间未到"

    if now > grace_end:
        reservation.status = SeatReservationStatus.EXPIRED.value
        db.commit()
        db.refresh(reservation)
        return None, "已超过15分钟签到期限，预约已取消"

    reservation.status = SeatReservationStatus.CHECKED_IN.value
    reservation.check_in_time = now

    db.commit()
    db.refresh(reservation)

    return reservation, ""


def cancel_seat_reservation(db: Session, reservation_id: int) -> Tuple[Optional[SeatReservation], str]:
    reservation = db.query(SeatReservation).filter(SeatReservation.id == reservation_id).first()
    if not reservation:
        return None, "预约记录不存在"

    if reservation.status != SeatReservationStatus.RESERVED.value:
        return None, "该预约状态不可取消"

    reservation.status = SeatReservationStatus.CANCELLED.value
    db.commit()
    db.refresh(reservation)

    return reservation, ""


def check_expired_seat_reservations(db: Session):
    now = datetime.now()
    today = date.today()
    current_time = now.time()

    expired_reservations = db.query(SeatReservation).filter(
        SeatReservation.status == SeatReservationStatus.RESERVED.value,
        and_(
            SeatReservation.reservation_date == today,
            SeatReservation.start_time <= current_time,
        ),
    ).all()

    for reservation in expired_reservations:
        reservation_start = datetime.combine(reservation.reservation_date, reservation.start_time)
        grace_end = reservation_start + timedelta(minutes=CHECKIN_GRACE_MINUTES)
        if now > grace_end:
            reservation.status = SeatReservationStatus.EXPIRED.value

    db.commit()


def process_waitlist_for_event(db: Session, event: Event):
    while event.current_participants < event.max_participants:
        next_waitlist = db.query(EventRegistration).filter(
            EventRegistration.event_id == event.id,
            EventRegistration.status == EventRegistrationStatus.WAITLIST.value,
        ).order_by(EventRegistration.waitlist_position.asc()).first()

        if not next_waitlist:
            break

        next_waitlist.status = EventRegistrationStatus.CONFIRMED.value
        next_waitlist.waitlist_position = None
        event.current_participants += 1

        waitlist_entries = db.query(EventRegistration).filter(
            EventRegistration.event_id == event.id,
            EventRegistration.status == EventRegistrationStatus.WAITLIST.value,
        ).order_by(EventRegistration.waitlist_position.asc()).all()

        for i, entry in enumerate(waitlist_entries, start=1):
            entry.waitlist_position = i

        create_notification(
            db,
            reader_id=next_waitlist.reader_id,
            title="活动报名候补成功",
            message=f"您报名的活动《{event.title}》已从候补转为正式报名。",
            notification_type="event_confirm",
        )

    db.commit()
    db.refresh(event)


def register_for_event(db: Session, registration_data: EventRegistrationCreate) -> Tuple[Optional[EventRegistration], str]:
    reader = db.query(Reader).filter(Reader.id == registration_data.reader_id).first()
    if not reader or not reader.is_active:
        return None, "读者不存在或已停用"

    event = db.query(Event).filter(Event.id == registration_data.event_id).first()
    if not event or not event.is_active:
        return None, "活动不存在或已结束"

    now = datetime.utcnow()
    if event.registration_start and now < event.registration_start:
        return None, "活动报名未开始"
    if event.registration_end and now > event.registration_end:
        return None, "活动报名已结束"

    existing_registration = db.query(EventRegistration).filter(
        EventRegistration.event_id == registration_data.event_id,
        EventRegistration.reader_id == registration_data.reader_id,
        EventRegistration.status.in_([
            EventRegistrationStatus.CONFIRMED.value,
            EventRegistrationStatus.WAITLIST.value,
        ]),
    ).first()
    if existing_registration:
        return None, "您已报名该活动"

    is_confirmed = event.current_participants < event.max_participants

    registration = EventRegistration(
        event_id=registration_data.event_id,
        reader_id=registration_data.reader_id,
        status=EventRegistrationStatus.CONFIRMED.value if is_confirmed else EventRegistrationStatus.WAITLIST.value,
        waitlist_position=None if is_confirmed else (
            db.query(EventRegistration).filter(
                EventRegistration.event_id == registration_data.event_id,
                EventRegistration.status == EventRegistrationStatus.WAITLIST.value,
            ).count() + 1
        ),
    )

    if is_confirmed:
        event.current_participants += 1

    db.add(registration)
    db.commit()
    db.refresh(registration)
    db.refresh(event)

    return registration, ""


def cancel_event_registration(db: Session, cancel_data: EventRegistrationCancel) -> Tuple[Optional[EventRegistration], str]:
    registration = db.query(EventRegistration).filter(EventRegistration.id == cancel_data.registration_id).first()
    if not registration:
        return None, "报名记录不存在"

    if registration.status == EventRegistrationStatus.CANCELLED.value:
        return None, "该报名已取消"

    event = db.query(Event).filter(Event.id == registration.event_id).first()

    was_confirmed = registration.status == EventRegistrationStatus.CONFIRMED.value

    registration.status = EventRegistrationStatus.CANCELLED.value
    registration.waitlist_position = None

    if was_confirmed and event.current_participants > 0:
        event.current_participants -= 1

    db.commit()
    db.refresh(registration)
    db.refresh(event)

    if was_confirmed:
        process_waitlist_for_event(db, event)

    return registration, ""
