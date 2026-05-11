from datetime import date, timedelta
from typing import List

from .models import (
    MiningOperation,
    Transport,
    SafetyCheck,
    Todo,
    SafetyCheckStatus,
)


class ValidationError(Exception):
    pass


def validate_mining_operation(
    operation: MiningOperation,
    existing_operations: List[MiningOperation] = None,
) -> MiningOperation:
    if operation.actual_output > operation.planned_output * 1.2:
        raise ValidationError(
            f"实际产量不能超过计划产量的120%。实际: {operation.actual_output}, 最大允许: {operation.planned_output * 1.2}"
        )

    existing_operations = existing_operations or []
    for existing in existing_operations:
        if (
            existing.mine_id == operation.mine_id
            and existing.operation_date == operation.operation_date
            and existing.id != operation.id
        ):
            raise ValidationError(
                f"矿区 ID {operation.mine_id} 在 {operation.operation_date} 已存在采矿作业记录"
            )

    return operation


def validate_transport(transport: Transport) -> Transport:
    if transport.arrival_time < transport.departure_time:
        raise ValidationError(
            f"到达时间不能早于出发时间。到达: {transport.arrival_time}, 出发: {transport.departure_time}"
        )
    return transport


def validate_safety_check(check: SafetyCheck) -> SafetyCheck:
    has_abnormal = any(item.is_abnormal for item in check.items)
    if has_abnormal:
        check.status = SafetyCheckStatus.UNQUALIFIED
    else:
        check.status = SafetyCheckStatus.QUALIFIED
    return check


def validate_todo(todo: Todo) -> Todo:
    max_deadline = date.today() + timedelta(days=30)
    if todo.deadline > max_deadline:
        raise ValidationError(
            f"整改期限不能超过30天。最大允许: {max_deadline}, 提交: {todo.deadline}"
        )
    return todo
