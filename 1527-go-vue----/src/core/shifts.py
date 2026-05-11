from datetime import datetime, timedelta
from typing import Tuple

from .models import ShiftType


MORNING_START_HOUR = 6
AFTERNOON_START_HOUR = 14
NIGHT_START_HOUR = 22


def get_shift_type(dt: datetime) -> ShiftType:
    hour = dt.hour
    if MORNING_START_HOUR <= hour < AFTERNOON_START_HOUR:
        return ShiftType.MORNING
    elif AFTERNOON_START_HOUR <= hour < NIGHT_START_HOUR:
        return ShiftType.AFTERNOON
    else:
        return ShiftType.NIGHT


def get_shift_date(dt: datetime, shift_type: ShiftType) -> str:
    if shift_type == ShiftType.NIGHT:
        hour = dt.hour
        if hour >= NIGHT_START_HOUR:
            return dt.strftime("%Y-%m-%d")
        else:
            return (dt - timedelta(days=1)).strftime("%Y-%m-%d")
    else:
        return dt.strftime("%Y-%m-%d")


def get_shift_time_range(shift_date: str, shift_type: ShiftType) -> Tuple[datetime, datetime]:
    date_part = datetime.strptime(shift_date, "%Y-%m-%d").date()
    if shift_type == ShiftType.MORNING:
        start = datetime.combine(date_part, datetime.min.time().replace(hour=MORNING_START_HOUR))
        end = datetime.combine(date_part, datetime.min.time().replace(hour=AFTERNOON_START_HOUR))
    elif shift_type == ShiftType.AFTERNOON:
        start = datetime.combine(date_part, datetime.min.time().replace(hour=AFTERNOON_START_HOUR))
        end = datetime.combine(date_part, datetime.min.time().replace(hour=NIGHT_START_HOUR))
    else:
        start = datetime.combine(date_part, datetime.min.time().replace(hour=NIGHT_START_HOUR))
        end = datetime.combine(date_part + timedelta(days=1), datetime.min.time().replace(hour=MORNING_START_HOUR))
    return start, end


def get_shift_info(dt: datetime) -> Tuple[ShiftType, str]:
    shift_type = get_shift_type(dt)
    shift_date = get_shift_date(dt, shift_type)
    return shift_type, shift_date


def get_shift_display_name(shift_type: ShiftType) -> str:
    names = {
        ShiftType.MORNING: "早班",
        ShiftType.AFTERNOON: "中班",
        ShiftType.NIGHT: "夜班"
    }
    return names.get(shift_type, "未知")


def get_shift_time_display(shift_type: ShiftType) -> str:
    times = {
        ShiftType.MORNING: "06:00-14:00",
        ShiftType.AFTERNOON: "14:00-22:00",
        ShiftType.NIGHT: "22:00-06:00(次日)"
    }
    return times.get(shift_type, "")
