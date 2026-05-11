from datetime import datetime, timedelta
from typing import List, Tuple
from decimal import Decimal


def check_overdue(created_at: datetime, current_time: datetime = None) -> bool:
    current_time = current_time or datetime.now()
    threshold = created_at + timedelta(days=3)
    return current_time > threshold


def split_date_range(start_date: datetime, end_date: datetime, max_span_days: int = 90) -> List[Tuple[datetime, datetime]]:
    if (end_date - start_date).days <= max_span_days:
        return [(start_date, end_date)]
    
    segments = []
    current_start = start_date
    
    while current_start < end_date:
        current_end = current_start.replace(day=1, month=current_start.month + 1)
        if current_end.month == 1:
            current_end = current_end.replace(year=current_end.year + 1, month=1)
        
        if current_end > end_date:
            current_end = end_date
        
        segments.append((current_start, current_end))
        current_start = current_end
    
    return segments


def format_as_text_table(headers: List[str], rows: List[List[str]]) -> str:
    if not rows and not headers:
        return ""
    
    all_rows = [headers] + rows if headers else rows
    col_widths = [max(len(str(cell)) for cell in col) for col in zip(*all_rows)]
    
    format_str = " | ".join(f"{{:<{w}}}" for w in col_widths)
    separator = "-+-".join("-" * w for w in col_widths)
    
    lines = []
    if headers:
        lines.append(format_str.format(*headers))
        lines.append(separator)
    
    for row in rows:
        lines.append(format_str.format(*[str(cell) for cell in row]))
    
    return "\n".join(lines)


def decimal_to_str(value) -> str:
    if value is None:
        return ""
    if isinstance(value, Decimal):
        if value == value.to_integral_value():
            return f"{int(value)}"
        return f"{float(value):g}"
    return str(value)
