import re
from datetime import datetime, date
from dateutil.relativedelta import relativedelta
from app.models import BridgeGrade, InspectionType

def parse_mileage(mileage: str) -> tuple:
    match = re.match(r'K(\d+)\+(\d+)', mileage)
    if match:
        km = int(match.group(1))
        m = int(match.group(2))
        return (km, m)
    return (0, 0)

def sort_devices_by_mileage(devices):
    return sorted(devices, key=lambda d: parse_mileage(d.mileage))

def calculate_bridge_grade(length: float) -> BridgeGrade:
    if length < 20:
        return BridgeGrade.SMALL
    elif length < 100:
        return BridgeGrade.MEDIUM
    elif length < 1000:
        return BridgeGrade.LARGE
    else:
        return BridgeGrade.EXTRA_LARGE

def get_inspection_frequency(grade: BridgeGrade, inspection_type: InspectionType) -> int:
    if inspection_type == InspectionType.FLOOD:
        return 4
    frequencies = {
        BridgeGrade.SMALL: 1,
        BridgeGrade.MEDIUM: 1,
        BridgeGrade.LARGE: 1,
        BridgeGrade.EXTRA_LARGE: 2,
    }
    return frequencies.get(grade, 1)

def is_flood_season(current_date: date = None) -> bool:
    if current_date is None:
        current_date = date.today()
    month = current_date.month
    return 4 <= month <= 10

def get_next_month(current_date: date = None) -> date:
    if current_date is None:
        current_date = date.today()
    return (current_date + relativedelta(months=1)).replace(day=1)
