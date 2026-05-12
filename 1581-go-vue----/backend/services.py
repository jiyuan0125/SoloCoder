from datetime import datetime, timedelta
from typing import List, Optional, Tuple, Dict
from sqlalchemy.orm import Session
from . import models, schemas, crud
from .models import (
    TrainStatus,
    SignalStatus,
    PowerStatus,
    LogLevel,
)
from .config import settings


class DispatcherService:
    def __init__(self, db: Session):
        self.db = db

    def calculate_optimal_interval(
        self, line_id: int, time_window: timedelta = timedelta(hours=1)
    ) -> int:
        end_time = datetime.now()
        start_time = end_time - time_window
        passenger_data = crud.get_passenger_data(self.db, line_id, start_time, end_time)

        if not passenger_data:
            return 300

        total_passengers = sum(d.passenger_count for d in passenger_data)
        avg_passengers = total_passengers / max(len(passenger_data), 1)

        if avg_passengers > 1000:
            return 120
        elif avg_passengers > 500:
            return 180
        elif avg_passengers > 200:
            return 240
        else:
            return 300

    def generate_schedule_trains(
        self,
        schedule: models.Schedule,
    ) -> List[models.ScheduleTrain]:
        schedule_trains = []
        current_time = schedule.first_departure
        sequence = 1

        while current_time <= schedule.last_departure:
            travel_duration = timedelta(minutes=30)
            scheduled_arrival = current_time + travel_duration

            st = models.ScheduleTrain(
                schedule_id=schedule.id,
                sequence=sequence,
                scheduled_departure=current_time,
                scheduled_arrival=scheduled_arrival,
            )
            self.db.add(st)
            schedule_trains.append(st)
            sequence += 1
            current_time += timedelta(seconds=schedule.interval_seconds)

        self.db.commit()
        return schedule_trains

    def check_interval_conflict(
        self,
        schedule_id: int,
        current_departure: datetime,
        st_id: Optional[int] = None,
    ) -> List[schemas.ConflictWarning]:
        conflicts = []
        all_trains = crud.get_schedule_trains_by_schedule(self.db, schedule_id)

        for train in all_trains:
            if st_id and train.id == st_id:
                continue
            if train.actual_departure is not None:
                continue

            interval = abs((current_departure - train.scheduled_departure).total_seconds())
            if 0 < interval < settings.MIN_SAFE_INTERVAL:
                warnings = schemas.ConflictWarning(
                    type="interval_conflict",
                    message=f"发车间隔 {int(interval)} 秒小于最小安全间隔 {settings.MIN_SAFE_INTERVAL} 秒",
                    schedule_train_id=train.id,
                    schedule_train_sequence=train.sequence,
                    previous_departure=train.scheduled_departure,
                    current_departure=current_departure,
                    interval_seconds=int(interval),
                    min_safe_interval=settings.MIN_SAFE_INTERVAL,
                )
                conflicts.append(warnings)

        return conflicts

    def reschedule_upcoming_trains(
        self,
        schedule_id: int,
        new_interval_seconds: Optional[int] = None,
    ) -> List[models.ScheduleTrain]:
        schedule = crud.get_schedule(self.db, schedule_id)
        if not schedule:
            return []

        current_time = datetime.now()
        upcoming_trains = crud.get_upcoming_schedule_trains(
            self.db, schedule_id, current_time
        )

        if not upcoming_trains:
            return []

        interval = new_interval_seconds or schedule.interval_seconds
        base_time = upcoming_trains[0].scheduled_departure

        for i, train in enumerate(upcoming_trains):
            new_departure = base_time + timedelta(seconds=i * interval)
            if train.scheduled_arrival:
                original_duration = train.scheduled_arrival - train.scheduled_departure
            else:
                original_duration = timedelta(minutes=30)
            train.scheduled_departure = new_departure
            train.scheduled_arrival = new_departure + original_duration

        self.db.commit()

        crud.create_dispatch_log(
            self.db,
            schemas.DispatchLogCreate(
                level=LogLevel.INFO,
                message=f"运行计划 {schedule.name} 已重新计算，新间隔 {interval} 秒",
                entity_type="schedule",
                entity_id=schedule_id,
            ),
        )

        return upcoming_trains

    def update_schedule_train_departure(
        self,
        st_id: int,
        new_departure: datetime,
    ) -> Tuple[models.ScheduleTrain, List[schemas.ConflictWarning]]:
        st = crud.get_schedule_train(self.db, st_id)
        if not st:
            raise ValueError(f"Schedule train {st_id} not found")

        if st.actual_departure is not None:
            raise ValueError("Cannot adjust already departed train")

        conflicts = self.check_interval_conflict(
            st.schedule_id, new_departure, st_id=st_id
        )

        st.scheduled_departure = new_departure
        if st.scheduled_arrival:
            original_duration = (
                st.scheduled_arrival - st.scheduled_departure
                if st.scheduled_arrival and st.scheduled_departure
                else timedelta(minutes=30)
            )
            st.scheduled_arrival = new_departure + original_duration

        self.db.commit()

        crud.create_dispatch_log(
            self.db,
            schemas.DispatchLogCreate(
                level=LogLevel.WARNING if conflicts else LogLevel.INFO,
                message=f"班次 {st.sequence} 发车时间调整为 {new_departure}"
                + (f" - 存在 {len(conflicts)} 个冲突" if conflicts else ""),
                entity_type="schedule_train",
                entity_id=st_id,
            ),
        )

        return st, conflicts

    def complete_schedule_train(
        self,
        st_id: int,
        actual_departure: datetime,
        actual_arrival: datetime,
    ) -> models.ScheduleTrain:
        st = crud.get_schedule_train(self.db, st_id)
        if not st:
            raise ValueError(f"Schedule train {st_id} not found")

        st.actual_departure = actual_departure
        st.actual_arrival = actual_arrival
        st.is_completed = True

        all_trains = crud.get_schedule_trains_by_schedule(self.db, st.schedule_id)
        idx = next((i for i, t in enumerate(all_trains) if t.id == st_id), -1)

        if idx > 0:
            prev_train = all_trains[idx - 1]
            if prev_train.actual_arrival:
                turnaround_time = int(
                    (actual_departure - prev_train.actual_arrival).total_seconds()
                )
                st.turnaround_time = turnaround_time

                if turnaround_time > settings.TURNAROUND_TIMEOUT:
                    crud.create_dispatch_log(
                        self.db,
                        schemas.DispatchLogCreate(
                            level=LogLevel.WARNING,
                            message=f"列车折返时间超时: {turnaround_time} 秒 (超过 {settings.TURNAROUND_TIMEOUT} 秒)",
                            entity_type="schedule_train",
                            entity_id=st_id,
                        ),
                    )

        self.db.commit()
        return st


class SignalService:
    def __init__(self, db: Session):
        self.db = db

    def report_signal_fault(
        self,
        signal_id: int,
        description: Optional[str] = None,
    ) -> Tuple[models.Signal, models.FaultReport]:
        signal = crud.get_signal(self.db, signal_id)
        if not signal:
            raise ValueError(f"Signal {signal_id} not found")

        if signal.status == SignalStatus.FAULT:
            raise ValueError("Signal is already in fault state")

        signal.status = SignalStatus.FAULT
        signal.fault_time = datetime.now()
        signal.fault_description = description

        report = crud.create_fault_report(
            self.db,
            schemas.FaultReportCreate(
                signal_id=signal_id,
                fault_description=description,
            ),
        )

        if signal.section_id:
            self._handle_affected_section(signal.section_id)

        self.db.commit()

        crud.create_dispatch_log(
            self.db,
            schemas.DispatchLogCreate(
                level=LogLevel.ERROR,
                message=f"信号机 {signal.signal_code} 发生故障: {description or '未知原因'}",
                entity_type="signal",
                entity_id=signal_id,
            ),
        )

        return signal, report

    def confirm_signal_recovery(
        self,
        signal_id: int,
        operator: str,
    ) -> models.Signal:
        signal = crud.get_signal(self.db, signal_id)
        if not signal:
            raise ValueError(f"Signal {signal_id} not found")

        if signal.status == SignalStatus.NORMAL:
            raise ValueError("Signal is already normal")

        if signal.status == SignalStatus.FAULT:
            signal.status = SignalStatus.RECOVERED
            signal.recovery_time = datetime.now()

            reports = crud.get_fault_reports_by_signal(self.db, signal_id)
            for report in reports:
                if not report.resolved_at:
                    crud.resolve_fault_report(self.db, report.id, operator)

            self.db.commit()

            crud.create_dispatch_log(
                self.db,
                schemas.DispatchLogCreate(
                    level=LogLevel.INFO,
                    message=f"信号机 {signal.signal_code} 已恢复，由 {operator} 确认",
                    entity_type="signal",
                    entity_id=signal_id,
                    operator=operator,
                ),
            )

        return signal

    def restore_signal_to_normal(
        self,
        signal_id: int,
    ) -> models.Signal:
        signal = crud.get_signal(self.db, signal_id)
        if not signal:
            raise ValueError(f"Signal {signal_id} not found")

        if signal.status != SignalStatus.RECOVERED:
            raise ValueError(
                "Signal can only be restored to normal from RECOVERED state"
            )

        signal.status = SignalStatus.NORMAL

        if signal.section_id:
            self._restore_section_trains(signal.section_id)

        self.db.commit()

        crud.create_dispatch_log(
            self.db,
            schemas.DispatchLogCreate(
                level=LogLevel.INFO,
                message=f"信号机 {signal.signal_code} 已恢复正常运行",
                entity_type="signal",
                entity_id=signal_id,
            ),
        )

        return signal

    def _handle_affected_section(self, section_id: int):
        trains = crud.get_trains_in_section(self.db, section_id)
        for train in trains:
            if train.status == TrainStatus.RUNNING:
                train.speed_limit = 20.0
                train.status = TrainStatus.DELAYED
                crud.create_dispatch_log(
                    self.db,
                    schemas.DispatchLogCreate(
                        level=LogLevel.WARNING,
                        message=f"列车 {train.train_number} 因信号故障自动限速至 20 km/h",
                        entity_type="train",
                        entity_id=train.id,
                    ),
                )

    def _restore_section_trains(self, section_id: int):
        trains = crud.get_trains_in_section(self.db, section_id)
        for train in trains:
            train.speed_limit = 60.0
            if train.status == TrainStatus.DELAYED:
                train.status = TrainStatus.RUNNING


class PowerService:
    def __init__(self, db: Session):
        self.db = db

    def report_power_outage(
        self,
        section_id: int,
    ) -> models.PowerSection:
        section = crud.get_power_section(self.db, section_id)
        if not section:
            raise ValueError(f"Power section {section_id} not found")

        if section.status == PowerStatus.OUTAGE:
            raise ValueError("Section is already in outage")

        section.status = PowerStatus.OUTAGE
        section.outage_time = datetime.now()

        self._stop_trains_in_section(section_id)

        self.db.commit()

        crud.create_dispatch_log(
            self.db,
            schemas.DispatchLogCreate(
                level=LogLevel.ERROR,
                message=f"供电区间 {section.section_code} 发生停电",
                entity_type="power_section",
                entity_id=section_id,
            ),
        )

        return section

    def restore_power(
        self,
        section_id: int,
    ) -> models.PowerSection:
        section = crud.get_power_section(self.db, section_id)
        if not section:
            raise ValueError(f"Power section {section_id} not found")

        if section.status == PowerStatus.ACTIVE:
            raise ValueError("Section is already active")

        section.status = PowerStatus.ACTIVE
        section.recovery_time = datetime.now()

        self.db.commit()

        crud.create_dispatch_log(
            self.db,
            schemas.DispatchLogCreate(
                level=LogLevel.WARNING,
                message=f"供电区间 {section.section_code} 已恢复供电，需调度员确认运行计划",
                entity_type="power_section",
                entity_id=section_id,
            ),
        )

        return section

    def confirm_operations_after_restoration(
        self,
        section_id: int,
        operator: str,
    ) -> None:
        section = crud.get_power_section(self.db, section_id)
        if not section:
            raise ValueError(f"Power section {section_id} not found")

        if section.status != PowerStatus.ACTIVE:
            raise ValueError("Section power is not active")

        self._resume_trains_in_section(section_id)

        crud.create_dispatch_log(
            self.db,
            schemas.DispatchLogCreate(
                level=LogLevel.INFO,
                message=f"供电区间 {section.section_code} 运行计划已由 {operator} 确认恢复",
                entity_type="power_section",
                entity_id=section_id,
                operator=operator,
            ),
        )

    def _stop_trains_in_section(self, section_id: int):
        trains = crud.get_trains_in_section(self.db, section_id)
        for train in trains:
            train.status = TrainStatus.STOPPED
            crud.create_dispatch_log(
                self.db,
                schemas.DispatchLogCreate(
                    level=LogLevel.ERROR,
                    message=f"列车 {train.train_number} 因区间停电已停运",
                    entity_type="train",
                    entity_id=train.id,
                ),
            )

    def _resume_trains_in_section(self, section_id: int):
        trains = crud.get_trains_in_section(self.db, section_id)
        for train in trains:
            if train.status == TrainStatus.STOPPED:
                train.status = TrainStatus.IDLE

        self.db.commit()


def get_dispatcher_service(db: Session) -> DispatcherService:
    return DispatcherService(db)


def get_signal_service(db: Session) -> SignalService:
    return SignalService(db)


def get_power_service(db: Session) -> PowerService:
    return PowerService(db)
