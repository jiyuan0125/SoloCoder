from datetime import datetime, timezone
from typing import Any
from dateutil import parser as dateutil_parser


def parse_timestamp(value: Any) -> float:
    if value is None:
        raise ValueError("timestamp is None")

    if isinstance(value, (int, float)):
        ts = float(value)
        if ts > 1e12:
            ts = ts / 1000.0
        if ts < 0:
            raise ValueError(f"Invalid timestamp: {value}")
        return ts

    if isinstance(value, str):
        value = value.strip()
        if not value:
            raise ValueError("Empty timestamp string")

        try:
            ts = float(value)
            if ts > 1e12:
                ts = ts / 1000.0
            return ts
        except ValueError:
            pass

        try:
            dt = dateutil_parser.parse(value)
            if dt.tzinfo is None:
                dt = dt.replace(tzinfo=timezone.utc)
            else:
                dt = dt.astimezone(timezone.utc)
            return dt.timestamp()
        except Exception:
            raise ValueError(f"Cannot parse timestamp: {value}")

    raise ValueError(f"Unsupported timestamp type: {type(value)}")
