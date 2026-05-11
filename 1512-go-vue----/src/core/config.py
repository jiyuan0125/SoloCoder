from datetime import datetime, timedelta

GRAIN_VARIETIES = {
    "小麦": 18,
    "水稻": 24,
    "玉米": 12,
    "大豆": 12,
    "花生": 12,
    "菜籽": 12,
    "芝麻": 18,
    "高粱": 12,
    "谷子": 18,
}

DEFAULT_STORAGE_MONTHS = 12


def get_storage_months(variety: str) -> int:
    return GRAIN_VARIETIES.get(variety, DEFAULT_STORAGE_MONTHS)


def calculate_expiry_date(inbound_date: datetime, variety: str) -> datetime:
    months = get_storage_months(variety)
    return inbound_date + timedelta(days=months * 30)


def get_minute_key(dt: datetime) -> str:
    return dt.strftime("%Y%m%d%H%M")


def is_pest_density_exceeded(count_per_kg: float) -> bool:
    return count_per_kg > 10
