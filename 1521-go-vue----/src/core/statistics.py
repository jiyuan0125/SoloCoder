from datetime import date, timedelta
from typing import Optional

from .models import Statistics, TodoStatus
from .repository import Repository


def get_monthly_statistics(
    repo: Repository,
    year: Optional[int] = None,
    month: Optional[int] = None,
) -> Statistics:
    today = date.today()
    if year is None:
        year = today.year
    if month is None:
        month = today.month

    if month == 12:
        next_month_start = date(year + 1, 1, 1)
    else:
        next_month_start = date(year, month + 1, 1)
    current_month_start = date(year, month, 1)

    total_output = 0.0
    total_planned = 0.0

    for op in repo.list_mining_operations():
        if current_month_start <= op.operation_date < next_month_start:
            total_output += op.actual_output
            total_planned += op.planned_output

    completion_rate = (total_output / total_planned * 100) if total_planned > 0 else 0.0

    transport_count = 0
    total_duration_minutes = 0.0
    for t in repo.list_transports():
        if current_month_start <= t.transport_date < next_month_start:
            transport_count += 1
            duration = t.arrival_time - t.departure_time
            total_duration_minutes += duration.total_seconds() / 60

    avg_transport_duration = (
        total_duration_minutes / transport_count if transport_count > 0 else 0.0
    )

    safety_check_count = 0
    for c in repo.list_safety_checks():
        if current_month_start <= c.check_date < next_month_start:
            safety_check_count += 1

    total_todos = 0
    completed_todos = 0
    for todo in repo.list_todos():
        todo_month_start = date(todo.deadline.year, todo.deadline.month, 1)
        if current_month_start <= todo_month_start < next_month_start:
            total_todos += 1
            if todo.status == TodoStatus.COMPLETED:
                completed_todos += 1

    risk_remediation_rate = (
        completed_todos / total_todos * 100 if total_todos > 0 else 0.0
    )

    return Statistics(
        month=f"{year}-{month:02d}",
        total_output=round(total_output, 2),
        completion_rate=round(completion_rate, 2),
        transport_count=transport_count,
        avg_transport_duration=round(avg_transport_duration, 2),
        safety_check_count=safety_check_count,
        risk_remediation_rate=round(risk_remediation_rate, 2),
    )
