from datetime import datetime, timedelta
from typing import List, Optional, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import and_, desc
from server.models import (
    Reservoir, FloodFacility, Gate, WaterLevelRecord, InflowRecord,
    DispatchCommand, FloodOperationRecord, Alert,
    FacilityType, FacilityStatus, AlertLevel, CommandStatus, FloodModeStatus
)
from server.schemas import (
    ReservoirCreate, ReservoirUpdate, FloodFacilityCreate, FloodFacilityUpdate,
    GateCreate, GateUpdate, WaterLevelRecordCreate, InflowRecordCreate,
    DispatchCommandCreate, DispatchCommandConfirm, DispatchCommandAuthorize
)
from server.config import settings


class ReservoirService:
    @staticmethod
    def create(db: Session, data: ReservoirCreate) -> Reservoir:
        reservoir = Reservoir(**data.model_dump())
        db.add(reservoir)
        db.commit()
        db.refresh(reservoir)
        return reservoir

    @staticmethod
    def get_by_id(db: Session, reservoir_id: int) -> Optional[Reservoir]:
        return db.query(Reservoir).filter(Reservoir.id == reservoir_id).first()

    @staticmethod
    def get_all(db: Session, skip: int = 0, limit: int = 100) -> List[Reservoir]:
        return db.query(Reservoir).offset(skip).limit(limit).all()

    @staticmethod
    def update(db: Session, reservoir_id: int, data: ReservoirUpdate) -> Optional[Reservoir]:
        reservoir = ReservoirService.get_by_id(db, reservoir_id)
        if not reservoir:
            return None
        for key, value in data.model_dump(exclude_unset=True).items():
            setattr(reservoir, key, value)
        db.commit()
        db.refresh(reservoir)
        return reservoir

    @staticmethod
    def is_main_flood_season(reservoir: Reservoir, current_month: Optional[int] = None) -> bool:
        month = current_month if current_month else datetime.now().month
        start = reservoir.main_flood_season_start
        end = reservoir.main_flood_season_end
        if start <= end:
            return start <= month <= end
        return month >= start or month <= end

    @staticmethod
    def get_flood_limit_level(reservoir: Reservoir) -> float:
        is_main = ReservoirService.is_main_flood_season(reservoir)
        reduction = 2.0 if is_main else 1.0
        return reservoir.normal_storage_level - reduction


class FloodFacilityService:
    @staticmethod
    def create(db: Session, data: FloodFacilityCreate) -> FloodFacility:
        facility = FloodFacility(**data.model_dump())
        db.add(facility)
        db.commit()
        db.refresh(facility)
        return facility

    @staticmethod
    def get_by_id(db: Session, facility_id: int) -> Optional[FloodFacility]:
        return db.query(FloodFacility).filter(FloodFacility.id == facility_id).first()

    @staticmethod
    def get_by_reservoir(db: Session, reservoir_id: int) -> List[FloodFacility]:
        return db.query(FloodFacility).filter(
            FloodFacility.reservoir_id == reservoir_id
        ).order_by(desc(FloodFacility.priority)).all()

    @staticmethod
    def update(db: Session, facility_id: int, data: FloodFacilityUpdate) -> Optional[FloodFacility]:
        facility = FloodFacilityService.get_by_id(db, facility_id)
        if not facility:
            return None
        for key, value in data.model_dump(exclude_unset=True).items():
            setattr(facility, key, value)
        db.commit()
        db.refresh(facility)
        return facility

    @staticmethod
    def update_status(db: Session, facility: FloodFacility, new_status: FacilityStatus) -> None:
        now = datetime.utcnow()
        if facility.status == FacilityStatus.RUNNING and new_status != FacilityStatus.RUNNING:
            if facility.run_start_time:
                elapsed = (now - facility.run_start_time).total_seconds() / 3600.0
                facility.total_run_hours += elapsed
                facility.run_start_time = None
        elif facility.status != FacilityStatus.RUNNING and new_status == FacilityStatus.RUNNING:
            facility.run_start_time = now
        facility.status = new_status
        facility.updated_at = now
        db.commit()

    @staticmethod
    def get_facilities_ordered(db: Session, reservoir_id: int) -> List[FloodFacility]:
        type_priority = {
            FacilityType.SPILLWAY_GATE: 3,
            FacilityType.OVERFLOW_CHUTE: 2,
            FacilityType.DRAINAGE_TUNNEL: 1
        }
        facilities = FloodFacilityService.get_by_reservoir(db, reservoir_id)
        return sorted(
            facilities,
            key=lambda f: (type_priority.get(f.facility_type, 0), f.priority),
            reverse=True
        )


class GateService:
    @staticmethod
    def create(db: Session, data: GateCreate) -> Gate:
        gate = Gate(**data.model_dump())
        db.add(gate)
        db.commit()
        db.refresh(gate)
        return gate

    @staticmethod
    def get_by_id(db: Session, gate_id: int) -> Optional[Gate]:
        return db.query(Gate).filter(Gate.id == gate_id).first()

    @staticmethod
    def get_by_facility(db: Session, facility_id: int) -> List[Gate]:
        return db.query(Gate).filter(Gate.facility_id == facility_id).all()

    @staticmethod
    def get_by_reservoir(db: Session, reservoir_id: int) -> List[Gate]:
        return db.query(Gate).join(FloodFacility).filter(
            FloodFacility.reservoir_id == reservoir_id
        ).all()

    @staticmethod
    def update(db: Session, gate_id: int, data: GateUpdate) -> Optional[Gate]:
        gate = GateService.get_by_id(db, gate_id)
        if not gate:
            return None
        for key, value in data.model_dump(exclude_unset=True).items():
            setattr(gate, key, value)
        db.commit()
        db.refresh(gate)
        return gate

    @staticmethod
    def can_adjust_open(db: Session, gate: Gate, current_time: Optional[datetime] = None) -> bool:
        if gate.last_operation_time is None:
            return True
        now = current_time if current_time else datetime.utcnow()
        elapsed = (now - gate.last_operation_time).total_seconds() / 60.0
        return elapsed >= settings.gate_open_interval_minutes

    @staticmethod
    def calculate_next_step(
        db: Session,
        gate: Gate,
        increase: bool = True
    ) -> Optional[float]:
        step = settings.gate_step_open_percent
        current = gate.current_open_percent
        if increase:
            if current >= 100:
                return None
            return min(current + step, 100)
        else:
            if current <= 0:
                return None
            return max(current - step, 0)


class WaterLevelService:
    @staticmethod
    def create(db: Session, data: WaterLevelRecordCreate) -> WaterLevelRecord:
        record = WaterLevelRecord(**data.model_dump())
        db.add(record)
        db.commit()
        db.refresh(record)
        return record

    @staticmethod
    def get_latest(db: Session, reservoir_id: int) -> Optional[WaterLevelRecord]:
        return db.query(WaterLevelRecord).filter(
            WaterLevelRecord.reservoir_id == reservoir_id,
            WaterLevelRecord.is_aggregated == False
        ).order_by(desc(WaterLevelRecord.record_time)).first()

    @staticmethod
    def get_by_range(
        db: Session,
        reservoir_id: int,
        start_time: datetime,
        end_time: datetime,
        aggregated_only: bool = False
    ) -> List[WaterLevelRecord]:
        filters = [
            WaterLevelRecord.reservoir_id == reservoir_id,
            WaterLevelRecord.record_time >= start_time,
            WaterLevelRecord.record_time <= end_time
        ]
        if aggregated_only:
            filters.append(WaterLevelRecord.is_aggregated == True)
        return db.query(WaterLevelRecord).filter(
            and_(*filters)
        ).order_by(WaterLevelRecord.record_time).all()

    @staticmethod
    def aggregate_by_hour(db: Session, reservoir_id: int) -> List[WaterLevelRecord]:
        now = datetime.utcnow()
        hour_start = now.replace(minute=0, second=0, microsecond=0)
        last_hour = hour_start - timedelta(hours=1)

        records = db.query(WaterLevelRecord).filter(
            WaterLevelRecord.reservoir_id == reservoir_id,
            WaterLevelRecord.is_aggregated == False,
            WaterLevelRecord.record_time >= last_hour,
            WaterLevelRecord.record_time < hour_start
        ).all()

        if not records:
            return []

        levels = [r.level for r in records]
        avg_level = sum(levels) / len(levels)

        aggregated = WaterLevelRecord(
            reservoir_id=reservoir_id,
            level=avg_level,
            record_time=last_hour,
            is_aggregated=True,
            aggregate_hour=last_hour.hour,
            aggregate_date=last_hour.date()
        )
        db.add(aggregated)
        db.commit()
        db.refresh(aggregated)
        return [aggregated]


class InflowService:
    @staticmethod
    def create(db: Session, data: InflowRecordCreate) -> InflowRecord:
        record = InflowRecord(**data.model_dump())
        db.add(record)
        db.commit()
        db.refresh(record)
        return record

    @staticmethod
    def get_latest(db: Session, reservoir_id: int) -> Optional[InflowRecord]:
        return db.query(InflowRecord).filter(
            InflowRecord.reservoir_id == reservoir_id
        ).order_by(desc(InflowRecord.record_time)).first()

    @staticmethod
    def get_recent_records(db: Session, reservoir_id: int, count: int = 3) -> List[InflowRecord]:
        return db.query(InflowRecord).filter(
            InflowRecord.reservoir_id == reservoir_id
        ).order_by(desc(InflowRecord.record_time)).limit(count).all()

    @staticmethod
    def is_inflow_increasing(db: Session, reservoir_id: int) -> bool:
        records = InflowService.get_recent_records(db, reservoir_id, count=3)
        if len(records) < 2:
            return False
        for i in range(len(records) - 1):
            if records[i].inflow_rate <= records[i + 1].inflow_rate:
                return False
        return True


class FloodControlService:
    @staticmethod
    def is_flood_mode(db: Session, reservoir: Reservoir) -> bool:
        return reservoir.current_flood_mode == FloodModeStatus.FLOOD_CONTROL

    @staticmethod
    def should_enter_flood_mode(
        db: Session,
        reservoir: Reservoir,
        current_level: Optional[float] = None
    ) -> bool:
        if current_level is None:
            latest = WaterLevelService.get_latest(db, reservoir.id)
            if latest is None:
                return False
            current_level = latest.level
        flood_limit = ReservoirService.get_flood_limit_level(reservoir)
        return current_level > flood_limit

    @staticmethod
    def should_exit_flood_mode(
        db: Session,
        reservoir: Reservoir,
        current_level: Optional[float] = None
    ) -> bool:
        if not FloodControlService.is_flood_mode(db, reservoir):
            return False
        if reservoir.flood_mode_start_time is None:
            return False
        if current_level is None:
            latest = WaterLevelService.get_latest(db, reservoir.id)
            if latest is None:
                return False
            current_level = latest.level
        flood_limit = ReservoirService.get_flood_limit_level(reservoir)
        if current_level >= flood_limit:
            return False
        records = db.query(WaterLevelRecord).filter(
            WaterLevelRecord.reservoir_id == reservoir.id,
            WaterLevelRecord.record_time >= datetime.utcnow() - timedelta(hours=settings.exit_flood_mode_hours)
        ).order_by(desc(WaterLevelRecord.record_time)).all()
        if not records:
            return False
        return all(r.level < flood_limit for r in records)

    @staticmethod
    def is_red_alert(
        db: Session,
        reservoir: Reservoir,
        current_level: Optional[float] = None
    ) -> bool:
        if current_level is None:
            latest = WaterLevelService.get_latest(db, reservoir.id)
            if latest is None:
                return False
            current_level = latest.level
        threshold = reservoir.design_flood_level * settings.red_alert_threshold
        return current_level >= threshold

    @staticmethod
    def enter_flood_mode(db: Session, reservoir: Reservoir) -> FloodOperationRecord:
        reservoir.current_flood_mode = FloodModeStatus.FLOOD_CONTROL
        reservoir.flood_mode_start_time = datetime.utcnow()
        latest = WaterLevelService.get_latest(db, reservoir.id)
        record = FloodOperationRecord(
            reservoir_id=reservoir.id,
            start_time=datetime.utcnow(),
            max_water_level=latest.level if latest else None,
            max_water_level_time=latest.record_time if latest else None
        )
        db.add(record)
        db.commit()
        db.refresh(reservoir)
        db.refresh(record)
        return record

    @staticmethod
    def exit_flood_mode(db: Session, reservoir: Reservoir) -> FloodOperationRecord:
        record = db.query(FloodOperationRecord).filter(
            FloodOperationRecord.reservoir_id == reservoir.id,
            FloodOperationRecord.end_time.is_(None)
        ).order_by(desc(FloodOperationRecord.start_time)).first()
        gates = GateService.get_by_reservoir(db, reservoir.id)
        for gate in gates:
            gate.current_open_percent = 0.0
            gate.last_operation_time = datetime.utcnow()
            if gate.facility:
                FloodFacilityService.update_status(
                    db, gate.facility, FacilityStatus.IDLE
                )
        commands_count = db.query(DispatchCommand).filter(
            DispatchCommand.reservoir_id == reservoir.id,
            DispatchCommand.created_at >= reservoir.flood_mode_start_time
        ).count()
        if record:
            record.end_time = datetime.utcnow()
            record.commands_count = commands_count
            record.summary = (
                f"防汛模式结束。持续时间: {(record.end_time - record.start_time).total_seconds() / 3600:.1f}小时。"
                f"最高水位: {record.max_water_level}米。共发布{commands_count}条调度指令。"
            )
        reservoir.current_flood_mode = FloodModeStatus.NORMAL
        reservoir.flood_mode_start_time = None
        db.commit()
        db.refresh(reservoir)
        if record:
            db.refresh(record)
        return record

    @staticmethod
    def create_red_alert(db: Session, reservoir: Reservoir) -> Alert:
        threshold = reservoir.design_flood_level * settings.red_alert_threshold
        latest = WaterLevelService.get_latest(db, reservoir.id)
        alert = Alert(
            reservoir_id=reservoir.id,
            alert_level=AlertLevel.RED,
            alert_type="water_level_exceeded",
            message=(
                f"库水位{latest.level if latest else 'N/A'}米超过设计洪水位的95% "
                f"({threshold}米)，发布红色预警。所有操作需防汛指挥部授权。"
            ),
            trigger_value=latest.level if latest else None,
            threshold_value=threshold
        )
        db.add(alert)
        db.commit()
        db.refresh(alert)
        return alert

    @staticmethod
    def get_active_flood_record(db: Session, reservoir_id: int) -> Optional[FloodOperationRecord]:
        return db.query(FloodOperationRecord).filter(
            FloodOperationRecord.reservoir_id == reservoir_id,
            FloodOperationRecord.end_time.is_(None)
        ).order_by(desc(FloodOperationRecord.start_time)).first()


class DispatchCommandService:
    @staticmethod
    def create(db: Session, data: DispatchCommandCreate) -> DispatchCommand:
        reservoir = ReservoirService.get_by_id(db, data.reservoir_id)
        is_flood = FloodControlService.is_flood_mode(db, reservoir)
        is_red = FloodControlService.is_red_alert(db, reservoir)
        command = DispatchCommand(
            **data.model_dump(),
            need_dual_confirm=is_flood,
            need_headquarters_auth=is_red
        )
        if not command.need_dual_confirm:
            command.status = CommandStatus.APPROVED
        if command.need_headquarters_auth:
            command.status = CommandStatus.PENDING_AUTHORIZATION
        db.add(command)
        db.commit()
        db.refresh(command)
        return command

    @staticmethod
    def get_by_id(db: Session, command_id: int) -> Optional[DispatchCommand]:
        return db.query(DispatchCommand).filter(DispatchCommand.id == command_id).first()

    @staticmethod
    def confirm(db: Session, command_id: int, data: DispatchCommandConfirm) -> Optional[DispatchCommand]:
        command = DispatchCommandService.get_by_id(db, command_id)
        if not command:
            return None
        if command.status != CommandStatus.PENDING_CONFIRM:
            return None
        command.confirmer_id = data.confirmer_id
        command.confirmer_name = data.confirmer_name
        if command.need_headquarters_auth:
            command.status = CommandStatus.PENDING_AUTHORIZATION
        else:
            command.status = CommandStatus.APPROVED
        db.commit()
        db.refresh(command)
        return command

    @staticmethod
    def authorize(db: Session, command_id: int, data: DispatchCommandAuthorize) -> Optional[DispatchCommand]:
        command = DispatchCommandService.get_by_id(db, command_id)
        if not command:
            return None
        if command.status != CommandStatus.PENDING_AUTHORIZATION:
            return None
        command.authorize_id = data.authorize_id
        command.authorize_name = data.authorize_name
        command.status = CommandStatus.APPROVED
        db.commit()
        db.refresh(command)
        return command

    @staticmethod
    def execute(db: Session, command_id: int) -> Tuple[Optional[DispatchCommand], Optional[str]]:
        command = DispatchCommandService.get_by_id(db, command_id)
        if not command:
            return None, "指令不存在"
        if command.status != CommandStatus.APPROVED:
            return None, f"指令状态为{command.status.value}，无法执行"
        gate = GateService.get_by_id(db, command.gate_id)
        if not gate:
            return None, "闸门不存在"
        if command.target_open_percent > gate.current_open_percent:
            if not GateService.can_adjust_open(db, gate):
                return None, "闸门操作间隔不足15分钟"
            step = settings.gate_step_open_percent
            if (command.target_open_percent - gate.current_open_percent) > step:
                return None, f"单次开度增加不能超过{step}%"
        gate.current_open_percent = command.target_open_percent
        gate.last_operation_time = datetime.utcnow()
        facility = gate.facility
        if facility:
            if command.target_open_percent > 0:
                FloodFacilityService.update_status(db, facility, FacilityStatus.RUNNING)
            else:
                FloodFacilityService.update_status(db, facility, FacilityStatus.IDLE)
        command.status = CommandStatus.EXECUTED
        command.executed_time = datetime.utcnow()
        db.commit()
        db.refresh(command)
        db.refresh(gate)
        return command, None

    @staticmethod
    def query_by_reservoir(
        db: Session,
        reservoir_id: int,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        flood_mode_only: bool = False
    ) -> List[DispatchCommand]:
        filters = [DispatchCommand.reservoir_id == reservoir_id]
        if start_time:
            filters.append(DispatchCommand.created_at >= start_time)
        if end_time:
            filters.append(DispatchCommand.created_at <= end_time)
        query = db.query(DispatchCommand).filter(and_(*filters))
        if flood_mode_only:
            reservoir = ReservoirService.get_by_id(db, reservoir_id)
            flood_records = db.query(FloodOperationRecord).filter(
                FloodOperationRecord.reservoir_id == reservoir_id
            ).all()
            flood_periods = [(r.start_time, r.end_time) for r in flood_records if r.end_time]
            if reservoir.flood_mode_start_time:
                flood_periods.append((reservoir.flood_mode_start_time, None))
        return query.order_by(desc(DispatchCommand.created_at)).all()

    @staticmethod
    def query_by_gate(
        db: Session,
        gate_id: int,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None
    ) -> List[DispatchCommand]:
        filters = [DispatchCommand.gate_id == gate_id]
        if start_time:
            filters.append(DispatchCommand.created_at >= start_time)
        if end_time:
            filters.append(DispatchCommand.created_at <= end_time)
        return db.query(DispatchCommand).filter(
            and_(*filters)
        ).order_by(desc(DispatchCommand.created_at)).all()

    @staticmethod
    def query_by_time_range(
        db: Session,
        start_time: datetime,
        end_time: datetime,
        reservoir_id: Optional[int] = None
    ) -> List[DispatchCommand]:
        filters = [
            DispatchCommand.created_at >= start_time,
            DispatchCommand.created_at <= end_time
        ]
        if reservoir_id:
            filters.append(DispatchCommand.reservoir_id == reservoir_id)
        return db.query(DispatchCommand).filter(
            and_(*filters)
        ).order_by(desc(DispatchCommand.created_at)).all()


class AlertService:
    @staticmethod
    def get_by_reservoir(
        db: Session,
        reservoir_id: int,
        unacknowledged_only: bool = False
    ) -> List[Alert]:
        filters = [Alert.reservoir_id == reservoir_id]
        if unacknowledged_only:
            filters.append(Alert.is_acknowledged == False)
        return db.query(Alert).filter(
            and_(*filters)
        ).order_by(desc(Alert.created_at)).all()

    @staticmethod
    def acknowledge(db: Session, alert_id: int) -> Optional[Alert]:
        alert = db.query(Alert).filter(Alert.id == alert_id).first()
        if alert:
            alert.is_acknowledged = True
            alert.acknowledged_time = datetime.utcnow()
            db.commit()
            db.refresh(alert)
        return alert
