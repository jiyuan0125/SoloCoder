from datetime import date, datetime, timedelta
from typing import Set

PUBLIC_HOLIDAYS: Set[date] = {
    date(2026, 1, 1),
    date(2026, 1, 29),
    date(2026, 1, 30),
    date(2026, 1, 31),
    date(2026, 2, 1),
    date(2026, 2, 2),
    date(2026, 4, 4),
    date(2026, 4, 5),
    date(2026, 4, 6),
    date(2026, 5, 1),
    date(2026, 5, 2),
    date(2026, 5, 3),
    date(2026, 6, 19),
    date(2026, 10, 1),
    date(2026, 10, 2),
    date(2026, 10, 3),
    date(2026, 10, 4),
    date(2026, 10, 5),
    date(2026, 10, 6),
    date(2026, 10, 7),
}


def is_weekend(d: date) -> bool:
    return d.weekday() >= 5


def is_public_holiday(d: date) -> bool:
    return d in PUBLIC_HOLIDAYS


def is_working_day(d: date) -> bool:
    return not is_weekend(d) and not is_public_holiday(d)


def add_working_days(start_date: date, days: int) -> date:
    current = start_date
    added = 0
    while added < days:
        current += timedelta(days=1)
        if is_working_day(current):
            added += 1
    return current


def calculate_working_days_between(start_date: date, end_date: date) -> int:
    if end_date < start_date:
        return 0
    working_days = 0
    current = start_date + timedelta(days=1)
    while current <= end_date:
        if is_working_day(current):
            working_days += 1
        current += timedelta(days=1)
    return working_days


def get_public_opinion_deadline(received_date: date) -> date:
    return add_working_days(received_date, 10)


def get_publicity_deadline(start_date: date) -> date:
    return add_working_days(start_date, 7)
