from datetime import datetime, timedelta, date
from typing import List, Optional, Tuple
from uuid import uuid4

from sqlalchemy.orm import Session
from sqlalchemy import func, and_, or_

from src.core.models import (
    Unit, Qualification, WasteLedger, Waybill, MonthlyLedger, Alert
)
from src.core.enums import WasteType, WaybillStatus, UnitType
from src.core.config import (
    GENERAL_WASTE_MAX_STORAGE_DAYS,
    HAZARDOUS_WASTE_MAX_STORAGE_DAYS,
    MONTHLY_BALANCE_TOLERANCE
)


def generate_waybill_code() -> str:
    return f"WB{datetime.now().strftime('%Y%m%d%H%M%S')}{uuid4().hex[:4].upper()}"


def check_disposer_qualification(
    db: Session, disposer_id: int, waste_type: WasteType, hw_code: Optional[str]
) -> Tuple[bool, str]:
    if waste_type == WasteType.GENERAL:
        return True, ""

    if not hw_code:
        return False, "危废必须提供HW编号"

    disposer = db.query(Unit).filter(
        Unit.id == disposer_id,
        Unit.unit_type == UnitType.DISPOSER
    ).first()

    if not disposer:
        return False, "处置单位不存在或不是处置单位"

    today = date.today()
    qualification = db.query(Qualification).filter(
        Qualification.disposer_id == disposer_id,
        Qualification.is_active == True,
        Qualification.valid_from <= today,
        Qualification.valid_until >= today
    ).first()

    if not qualification:
        return False, "处置单位无有效资质"

    allowed_hw_codes = [code.strip() for code in qualification.hw_codes.split(",")]
    if hw_code not in allowed_hw_codes:
        return False, f"处置单位无HW{hw_code}的处置资质"

    return True, ""


def create_waybill(
    db: Session,
    producer_id: int,
    disposer_id: int,
    waste_type: WasteType,
    waste_name: str,
    quantity: float,
    hw_code: Optional[str] = None,
    unit: str = "吨"
) -> Waybill:
    is_qualified, msg = check_disposer_qualification(db, disposer_id, waste_type, hw_code)
    if not is_qualified:
        raise ValueError(msg)

    waybill = Waybill(
        waybill_code=generate_waybill_code(),
        producer_id=producer_id,
        disposer_id=disposer_id,
        waste_type=waste_type.value,
        hw_code=hw_code,
        waste_name=waste_name,
        quantity=quantity,
        unit=unit,
        status=WaybillStatus.PENDING_SHIPMENT.value
    )

    db.add(waybill)
    db.commit()
    db.refresh(waybill)
    return waybill


def update_waybill_status(
    db: Session,
    waybill_id: int,
    new_status: WaybillStatus,
    reject_reason: Optional[str] = None
) -> Waybill:
    waybill = db.query(Waybill).filter(Waybill.id == waybill_id).first()
    if not waybill:
        raise ValueError("联单不存在")

    current_status = waybill.status

    valid_transitions = {
        WaybillStatus.PENDING_SHIPMENT.value: [WaybillStatus.IN_TRANSIT.value],
        WaybillStatus.IN_TRANSIT.value: [WaybillStatus.ARRIVED.value, WaybillStatus.REJECTED.value],
        WaybillStatus.ARRIVED.value: [WaybillStatus.DISPOSED.value],
        WaybillStatus.REJECTED.value: [],
        WaybillStatus.DISPOSED.value: []
    }

    if new_status.value not in valid_transitions.get(current_status, []):
        raise ValueError(f"无效的状态流转: {current_status} -> {new_status.value}")

    if new_status == WaybillStatus.REJECTED and not reject_reason:
        raise ValueError("拒收时必须提供拒收原因")

    waybill.status = new_status.value
    if reject_reason:
        waybill.reject_reason = reject_reason

    db.commit()
    db.refresh(waybill)
    return waybill


def check_storage_overdue(db: Session) -> List[Alert]:
    today = date.today()
    alerts = []

    ledgers = db.query(WasteLedger).all()

    for ledger in ledgers:
        max_days = (
            GENERAL_WASTE_MAX_STORAGE_DAYS
            if ledger.waste_type == WasteType.GENERAL.value
            else HAZARDOUS_WASTE_MAX_STORAGE_DAYS
        )

        overdue_days = (today - ledger.production_date).days - max_days
        if overdue_days > 0:
            existing = db.query(Alert).filter(
                Alert.unit_id == ledger.producer_id,
                Alert.alert_type == "storage_overdue",
                Alert.message.like(f"%{ledger.id}%"),
                Alert.is_resolved == False
            ).first()

            if not existing:
                alert = Alert(
                    unit_id=ledger.producer_id,
                    alert_type="storage_overdue",
                    message=(
                        f"固废台账ID:{ledger.id} "
                        f"{ledger.waste_name}({ledger.waste_type}) "
                        f"已超期{overdue_days}天"
                    )
                )
                db.add(alert)
                alerts.append(alert)

    db.commit()
    return alerts


def calculate_monthly_balance(
    db: Session, unit_id: int, year: int, month: int
) -> MonthlyLedger:
    first_day = date(year, month, 1)
    if month == 12:
        next_month = date(year + 1, 1, 1)
    else:
        next_month = date(year, month + 1, 1)

    if month == 1:
        prev_first_day = date(year - 1, 12, 1)
    else:
        prev_first_day = date(year, month - 1, 1)

    prev_month = db.query(MonthlyLedger).filter(
        MonthlyLedger.unit_id == unit_id,
        MonthlyLedger.year == prev_first_day.year,
        MonthlyLedger.month == prev_first_day.month
    ).first()

    opening_balance = prev_month.closing_balance if prev_month else 0.0

    production = db.query(func.sum(WasteLedger.quantity)).filter(
        WasteLedger.producer_id == unit_id,
        WasteLedger.production_date >= first_day,
        WasteLedger.production_date < next_month
    ).scalar() or 0.0

    transfer_out = db.query(func.sum(Waybill.quantity)).filter(
        Waybill.producer_id == unit_id,
        Waybill.status.in_([
            WaybillStatus.IN_TRANSIT.value,
            WaybillStatus.ARRIVED.value,
            WaybillStatus.DISPOSED.value
        ]),
        Waybill.created_at >= datetime.combine(first_day, datetime.min.time()),
        Waybill.created_at < datetime.combine(next_month, datetime.min.time())
    ).scalar() or 0.0

    calculated_closing = opening_balance + production - transfer_out

    existing = db.query(MonthlyLedger).filter(
        MonthlyLedger.unit_id == unit_id,
        MonthlyLedger.year == year,
        MonthlyLedger.month == month
    ).first()

    if existing:
        existing.opening_balance = opening_balance
        existing.production = production
        existing.transfer_out = transfer_out
        existing.calculated_closing = calculated_closing
        existing.difference = abs(existing.closing_balance - calculated_closing)

        base = max(abs(calculated_closing), 1e-9)
        existing.is_approved = (existing.difference / base) <= MONTHLY_BALANCE_TOLERANCE
        ledger = existing
    else:
        ledger = MonthlyLedger(
            unit_id=unit_id,
            year=year,
            month=month,
            opening_balance=opening_balance,
            production=production,
            transfer_out=transfer_out,
            closing_balance=calculated_closing,
            calculated_closing=calculated_closing,
            difference=0.0,
            is_approved=True
        )
        db.add(ledger)

    db.commit()
    db.refresh(ledger)

    if not ledger.is_approved:
        check_consecutive_failures(db, unit_id, year, month)

    return ledger


def check_consecutive_failures(db: Session, unit_id: int, year: int, month: int) -> None:
    current = db.query(MonthlyLedger).filter(
        MonthlyLedger.unit_id == unit_id,
        MonthlyLedger.year == year,
        MonthlyLedger.month == month
    ).first()

    if not current or current.is_approved:
        return

    if month == 1:
        prev_year, prev_month = year - 1, 12
    else:
        prev_year, prev_month = year, month - 1

    prev = db.query(MonthlyLedger).filter(
        MonthlyLedger.unit_id == unit_id,
        MonthlyLedger.year == prev_year,
        MonthlyLedger.month == prev_month
    ).first()

    if prev and not prev.is_approved:
        existing = db.query(Alert).filter(
            Alert.unit_id == unit_id,
            Alert.alert_type == "key_supervision",
            Alert.message.like(f"%{year}年{month}月%"),
            Alert.is_resolved == False
        ).first()

        if not existing:
            alert = Alert(
                unit_id=unit_id,
                alert_type="key_supervision",
                message=f"单位已连续两个月台账校验不通过，标记为重点监管（{year}年{month}月）"
            )
            db.add(alert)
            db.commit()


def export_to_csv(records: List[dict], fields: List[str]) -> str:
    lines = [",".join(fields)]

    for record in records:
        row = []
        for field in fields:
            value = record.get(field, "")
            if isinstance(value, float):
                row.append(f"{value:.2f}")
            elif isinstance(value, (datetime, date)):
                row.append(value.strftime("%Y-%m-%d") if isinstance(value, date) else value.strftime("%Y-%m-%d %H:%M:%S"))
            else:
                row.append(str(value).replace(",", "，"))
        lines.append(",".join(row))

    return "\n".join(lines)
