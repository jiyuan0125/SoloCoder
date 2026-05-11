from datetime import date, datetime
from typing import Callable, List

from .models import (
    CrushingRecord,
    CrushingRecordCreate,
    Equipment,
    EquipmentStatus,
    EquipmentUpdateStatus,
    FlotationRecord,
    FlotationRecordCreate,
)


class BusinessRuleError(Exception):
    pass


def validate_crushing_record(
    record: CrushingRecordCreate,
    existing_records: List[CrushingRecord],
) -> None:
    if record.product_grade_size >= record.feed_grade_size:
        raise BusinessRuleError("破碎后粒度不能大于破碎前粒度")

    if record.throughput <= 0:
        raise BusinessRuleError("处理量必须大于0")

    for existing in existing_records:
        if (
            existing.crusher_id == record.crusher_id
            and existing.record_date == record.record_date
        ):
            raise BusinessRuleError(
                f"破碎机 {record.crusher_id} 在 {record.record_date} 已有记录，同一破碎机同一天只能一条破碎记录"
            )


def validate_flotation_record(record: FlotationRecordCreate) -> None:
    if record.concentrate_grade <= record.feed_grade:
        raise BusinessRuleError("精矿品位不能低于原矿品位")

    if record.recovery < 0 or record.recovery > 100:
        raise BusinessRuleError("回收率必须在0-100之间")


def validate_equipment_status_transition(
    equipment: Equipment,
    update: EquipmentUpdateStatus,
) -> None:
    if update.status == equipment.status:
        raise BusinessRuleError("设备状态未发生变化")


def calculate_running_hours(
    start_time: datetime,
    end_time: datetime,
) -> float:
    delta = end_time - start_time
    return delta.total_seconds() / 3600.0


def update_equipment_running_hours(
    equipment: Equipment,
    update: EquipmentUpdateStatus,
) -> Equipment:
    if (
        equipment.status == EquipmentStatus.RUNNING
        and update.status != EquipmentStatus.RUNNING
        and equipment.last_running_time is not None
    ):
        running_hours = calculate_running_hours(
            equipment.last_running_time,
            update.timestamp,
        )
        equipment.total_running_hours += running_hours

    if update.status == EquipmentStatus.RUNNING:
        equipment.last_running_time = update.timestamp
    else:
        equipment.last_running_time = None

    equipment.status = update.status
    equipment.updated_at = update.timestamp
    return equipment
