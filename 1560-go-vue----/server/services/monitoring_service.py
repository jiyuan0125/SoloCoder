from datetime import datetime, timedelta
from typing import List, Optional

from sqlalchemy import desc
from sqlalchemy.orm import Session

from server.core.config import settings
from server.models import Alarm, EnergyRecord, Lighthouse, LightRecord, WorkOrder
from server.schemas import EnergyRecordCreate, LightRecordCreate


class MonitoringService:
    def __init__(self, db: Session):
        self.db = db

    def add_light_record(self, data: LightRecordCreate) -> LightRecord:
        lighthouse = self.db.query(Lighthouse).filter(Lighthouse.id == data.lighthouse_id).first()
        if not lighthouse:
            raise ValueError(f"Lighthouse {data.lighthouse_id} not found")

        record = LightRecord(**data.model_dump())

        threshold = lighthouse.rated_illuminance * settings.LIGHT_LOW_THRESHOLD
        if data.is_on == 0:
            record.status = "off"
        elif data.illuminance < threshold:
            record.status = "low"
        else:
            record.status = "normal"

        self.db.add(record)
        self.db.commit()
        self.db.refresh(record)

        self._check_light_alarms(lighthouse, record)

        return record

    def _check_light_alarms(self, lighthouse: Lighthouse, current_record: LightRecord):
        recent_records = (
            self.db.query(LightRecord)
            .filter(LightRecord.lighthouse_id == lighthouse.id)
            .order_by(desc(LightRecord.timestamp))
            .limit(settings.LIGHT_OFF_COUNT + 1)
            .all()
        )

        if len(recent_records) >= settings.LIGHT_OFF_COUNT:
            recent_off = [r for r in recent_records[:settings.LIGHT_OFF_COUNT] if r.is_on == 0]
            if len(recent_off) >= settings.LIGHT_OFF_COUNT:
                existing = (
                    self.db.query(Alarm)
                    .filter(
                        Alarm.lighthouse_id == lighthouse.id,
                        Alarm.alarm_type == "light_off",
                        Alarm.status == "open",
                    )
                    .first()
                )
                if not existing:
                    first_off_time = recent_records[settings.LIGHT_OFF_COUNT - 1].timestamp
                    alarm = Alarm(
                        lighthouse_id=lighthouse.id,
                        alarm_type="light_off",
                        severity="critical",
                        message=f"灯塔{lighthouse.name}连续{settings.LIGHT_OFF_COUNT}条记录灯光熄灭，请立即处理",
                        status="open",
                    )
                    self.db.add(alarm)
                    self.db.commit()

        if current_record.status == "low":
            existing = (
                self.db.query(Alarm)
                .filter(
                    Alarm.lighthouse_id == lighthouse.id,
                    Alarm.alarm_type == "light_low",
                    Alarm.status == "open",
                )
                .first()
            )
            if not existing:
                threshold_pct = settings.LIGHT_LOW_THRESHOLD * 100
                alarm = Alarm(
                    lighthouse_id=lighthouse.id,
                    alarm_type="light_low",
                    severity="warning",
                    message=f"灯塔{lighthouse.name}光照度低于额定值{threshold_pct:.0f}%，当前: {current_record.illuminance:.2f}",
                    status="open",
                )
                self.db.add(alarm)
                self.db.commit()

        self._check_critical_fault(lighthouse)

    def _check_critical_fault(self, lighthouse: Lighthouse):
        critical_hours = settings.LIGHT_OFF_CRITICAL_HOURS

        open_alarm = (
            self.db.query(Alarm)
            .filter(
                Alarm.lighthouse_id == lighthouse.id,
                Alarm.alarm_type == "light_off",
                Alarm.status == "open",
            )
            .first()
        )

        if not open_alarm:
            return

        if datetime.utcnow() - open_alarm.created_at >= timedelta(hours=critical_hours):
            existing_order = (
                self.db.query(WorkOrder)
                .filter(
                    WorkOrder.lighthouse_id == lighthouse.id,
                    WorkOrder.order_type == "emergency_repair",
                    WorkOrder.status.in_(["created", "assigned", "executed"]),
                )
                .first()
            )
            if not existing_order:
                order = WorkOrder(
                    lighthouse_id=lighthouse.id,
                    order_type="emergency_repair",
                    priority="critical",
                    title=f"灯塔{lighthouse.name}严重故障紧急维修",
                    description=f"灯光熄灭已超过{critical_hours}小时，需要立即维修",
                    status="created",
                )
                self.db.add(order)
                self.db.commit()

    def add_energy_record(self, data: EnergyRecordCreate) -> EnergyRecord:
        lighthouse = self.db.query(Lighthouse).filter(Lighthouse.id == data.lighthouse_id).first()
        if not lighthouse:
            raise ValueError(f"Lighthouse {data.lighthouse_id} not found")

        record = EnergyRecord(**data.model_dump())
        self.db.add(record)
        self.db.commit()
        self.db.refresh(record)

        solar_threshold = lighthouse.solar_power * settings.ENERGY_LOW_THRESHOLD
        if data.solar_generation < solar_threshold:
            existing = (
                self.db.query(Alarm)
                .filter(
                    Alarm.lighthouse_id == lighthouse.id,
                    Alarm.alarm_type == "solar_low",
                    Alarm.status == "open",
                )
                .first()
            )
            if not existing:
                alarm = Alarm(
                    lighthouse_id=lighthouse.id,
                    alarm_type="solar_low",
                    severity="warning",
                    message=f"灯塔{lighthouse.name}太阳能发电量低于额定值20%，当前: {data.solar_generation:.2f}，额定: {lighthouse.solar_power}",
                    status="open",
                )
                self.db.add(alarm)
                self.db.commit()

        if data.battery_percent < settings.ENERGY_LOW_THRESHOLD * 100:
            existing = (
                self.db.query(Alarm)
                .filter(
                    Alarm.lighthouse_id == lighthouse.id,
                    Alarm.alarm_type == "battery_low",
                    Alarm.status == "open",
                )
                .first()
            )
            if not existing:
                alarm = Alarm(
                    lighthouse_id=lighthouse.id,
                    alarm_type="battery_low",
                    severity="warning",
                    message=f"灯塔{lighthouse.name}蓄电池电量低于20%，当前: {data.battery_percent:.2f}%",
                    status="open",
                )
                self.db.add(alarm)
                self.db.commit()

        return record

    def get_light_records(self, lighthouse_id: int, limit: int = 100) -> List[LightRecord]:
        return (
            self.db.query(LightRecord)
            .filter(LightRecord.lighthouse_id == lighthouse_id)
            .order_by(desc(LightRecord.timestamp))
            .limit(limit)
            .all()
        )

    def get_energy_records(self, lighthouse_id: int, limit: int = 100) -> List[EnergyRecord]:
        return (
            self.db.query(EnergyRecord)
            .filter(EnergyRecord.lighthouse_id == lighthouse_id)
            .order_by(desc(EnergyRecord.timestamp))
            .limit(limit)
            .all()
        )

    def get_open_alarms(self) -> List[Alarm]:
        return self.db.query(Alarm).filter(Alarm.status == "open").order_by(Alarm.created_at.desc()).all()

    def resolve_alarm(self, alarm_id: int, resolved_by: str) -> Optional[Alarm]:
        alarm = self.db.query(Alarm).filter(Alarm.id == alarm_id).first()
        if not alarm:
            return None

        alarm.status = "resolved"
        alarm.resolved_at = datetime.utcnow()
        alarm.resolved_by = resolved_by
        self.db.commit()
        self.db.refresh(alarm)
        return alarm
