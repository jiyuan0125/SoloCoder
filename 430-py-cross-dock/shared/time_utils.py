from datetime import datetime, timezone


def utc_now() -> datetime:
    return datetime.now(timezone.utc)


def to_utc(dt: datetime) -> datetime:
    if dt.tzinfo is None:
        return dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


def is_same_day_utc(dt1: datetime, dt2: datetime) -> bool:
    utc1 = to_utc(dt1)
    utc2 = to_utc(dt2)
    return utc1.date() == utc2.date()
