import csv
import io
from datetime import datetime, timedelta, date
from typing import List, Dict, Any, Callable
from dateutil.relativedelta import relativedelta


def get_now() -> datetime:
    return datetime.now()


def parse_date(date_str: str) -> date:
    return datetime.strptime(date_str, "%Y-%m-%d").date()


def calculate_expiry_date(approval_level: str, issue_date: date) -> date:
    from .config import settings
    months = settings.APPROVAL_VALIDITY.get(approval_level, 12)
    return issue_date + relativedelta(months=months)


def is_expired(expiry_date: date) -> bool:
    return date.today() > expiry_date


def days_until_expiry(expiry_date: date) -> int:
    delta = expiry_date - date.today()
    return max(0, delta.days)


def months_diff(d1: date, d2: date) -> int:
    delta = relativedelta(d1, d2)
    return delta.years * 12 + delta.months


def export_to_csv(data: List[Dict[str, Any]], fieldnames: List[str]) -> str:
    output = io.StringIO()
    writer = csv.DictWriter(output, fieldnames=fieldnames)
    writer.writeheader()
    for row in data:
        writer.writerow(row)
    return output.getvalue()


def add_months(d: date, months: int) -> date:
    return d + relativedelta(months=months)
