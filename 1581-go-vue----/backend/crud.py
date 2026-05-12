from datetime import datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from . import models, schemas
from .models import (
    TrainStatus,
    SignalStatus,
    PowerStatus,
    LogLevel,
)


def get_line(db: Session, line_id: int):
    return db.query(models.Line).filter(models.Line.id == line_id).first()


def get_line_by_name(db: Session, name: str):
    return db.query(models.Line).filter(models.Line.name == name).first()


def get_lines(db: Session, skip: int = 0, limit: int = 100):
    return db.query(models.Line).offset(skip).limit(limit).all()


def create_line(db: Session, line: schemas.LineCreate):
    db_line = models.Line(**line.model_dump())
    db.add(db_line)
    db.commit()
    db.refresh(db_line)
    return db_line


def update_line(db: Session, line_id: int, line: schemas.LineUpdate):
    db_line = get_line(db, line_id)
    if db_line:
        for key, value in line.model_dump(exclude_unset=True).items():
            setattr(db_line, key, value)
        db.commit()
        db.refresh(db_line)
    return db_line


def get_station(db: Session, station_id: int):
    return db.query(models.Station).filter(models.Station.id == station_id).first()


def get_stations_by_line(db: Session, line_id: int):
    return (
        db.query(models.Station)
        .filter(models.Station.line_id == line_id)
        .order_by(models.Station.sequence)
        .all()
    )


def create_station(db: Session, station: schemas.StationCreate):
    db_station = models.Station(**station.model_dump())
    db.add(db_station)
    db.commit()
    db.refresh(db_station)
    return db_station


def update_station(db: Session, station_id: int, station: schemas.StationUpdate):
    db_station = get_station(db, station_id)
    if db_station:
        for key, value in station.model_dump(exclude_unset=True).items():
            setattr(db_station, key, value)
        db.commit()
        db.refresh(db_station)
    return db_station


def get_train(db: Session, train_id: int):
    return db.query(models.Train).filter(models.Train.id == train_id).first()


def get_train_by_number(db: Session, train_number: str):
    return (
        db.query(models.Train).filter(models.Train.train_number == train_number).first()
    )


def get_trains(db: Session, skip: int = 0, limit: int = 100):
    return db.query(models.Train).offset(skip).limit(limit).all()


def create_train(db: Session, train: schemas.TrainCreate):
    db_train = models.Train(**train.model_dump())
    db.add(db_train)
    db.commit()
    db.refresh(db_train)
    return db_train


def update_train(db: Session, train_id: int, train: schemas.TrainUpdate):
    db_train = get_train(db, train_id)
    if db_train:
        for key, value in train.model_dump(exclude_unset=True).items():
            setattr(db_train, key, value)
        db.commit()
        db.refresh(db_train)
    return db_train


def get_signal(db: Session, signal_id: int):
    return db.query(models.Signal).filter(models.Signal.id == signal_id).first()


def get_signal_by_code(db: Session, signal_code: str):
    return (
        db.query(models.Signal).filter(models.Signal.signal_code == signal_code).first()
    )


def get_signals(db: Session, station_id: Optional[int] = None):
    query = db.query(models.Signal)
    if station_id:
        query = query.filter(models.Signal.station_id == station_id)
    return query.all()


def create_signal(db: Session, signal: schemas.SignalCreate):
    db_signal = models.Signal(**signal.model_dump())
    db.add(db_signal)
    db.commit()
    db.refresh(db_signal)
    return db_signal


def update_signal(db: Session, signal_id: int, signal: schemas.SignalUpdate):
    db_signal = get_signal(db, signal_id)
    if db_signal:
        for key, value in signal.model_dump(exclude_unset=True).items():
            setattr(db_signal, key, value)
        db.commit()
        db.refresh(db_signal)
    return db_signal


def get_power_section(db: Session, section_id: int):
    return (
        db.query(models.PowerSection)
        .filter(models.PowerSection.id == section_id)
        .first()
    )


def get_power_section_by_code(db: Session, section_code: str):
    return (
        db.query(models.PowerSection)
        .filter(models.PowerSection.section_code == section_code)
        .first()
    )


def get_power_sections_by_line(db: Session, line_id: int):
    return (
        db.query(models.PowerSection)
        .filter(models.PowerSection.line_id == line_id)
        .all()
    )


def create_power_section(db: Session, section: schemas.PowerSectionCreate):
    db_section = models.PowerSection(**section.model_dump())
    db.add(db_section)
    db.commit()
    db.refresh(db_section)
    return db_section


def update_power_section(
    db: Session, section_id: int, section: schemas.PowerSectionUpdate
):
    db_section = get_power_section(db, section_id)
    if db_section:
        for key, value in section.model_dump(exclude_unset=True).items():
            setattr(db_section, key, value)
        db.commit()
        db.refresh(db_section)
    return db_section


def get_fault_report(db: Session, report_id: int):
    return (
        db.query(models.FaultReport).filter(models.FaultReport.id == report_id).first()
    )


def get_fault_reports_by_signal(db: Session, signal_id: int):
    return (
        db.query(models.FaultReport)
        .filter(models.FaultReport.signal_id == signal_id)
        .order_by(models.FaultReport.report_time.desc())
        .all()
    )


def create_fault_report(db: Session, report: schemas.FaultReportCreate):
    db_report = models.FaultReport(**report.model_dump())
    db.add(db_report)
    db.commit()
    db.refresh(db_report)
    return db_report


def resolve_fault_report(db: Session, report_id: int, resolved_by: str):
    db_report = get_fault_report(db, report_id)
    if db_report:
        db_report.resolved_by = resolved_by
        db_report.resolved_at = datetime.now()
        db.commit()
        db.refresh(db_report)
    return db_report


def get_schedule(db: Session, schedule_id: int):
    return db.query(models.Schedule).filter(models.Schedule.id == schedule_id).first()


def get_schedules_by_line(db: Session, line_id: int, active_only: bool = False):
    query = db.query(models.Schedule).filter(models.Schedule.line_id == line_id)
    if active_only:
        query = query.filter(models.Schedule.is_active == True)
    return query.all()


def create_schedule(db: Session, schedule: schemas.ScheduleCreate):
    schedule_data = schedule.model_dump(exclude={"auto_generate_trains"})
    db_schedule = models.Schedule(**schedule_data)
    db.add(db_schedule)
    db.commit()
    db.refresh(db_schedule)
    return db_schedule


def update_schedule(db: Session, schedule_id: int, schedule: schemas.ScheduleUpdate):
    db_schedule = get_schedule(db, schedule_id)
    if db_schedule:
        for key, value in schedule.model_dump(exclude_unset=True).items():
            setattr(db_schedule, key, value)
        db.commit()
        db.refresh(db_schedule)
    return db_schedule


def get_schedule_train(db: Session, st_id: int):
    return (
        db.query(models.ScheduleTrain)
        .filter(models.ScheduleTrain.id == st_id)
        .first()
    )


def get_schedule_trains_by_schedule(
    db: Session, schedule_id: int, sequence_order: bool = True
):
    query = db.query(models.ScheduleTrain).filter(
        models.ScheduleTrain.schedule_id == schedule_id
    )
    if sequence_order:
        query = query.order_by(models.ScheduleTrain.sequence)
    return query.all()


def get_upcoming_schedule_trains(db: Session, schedule_id: int, current_time: datetime):
    return (
        db.query(models.ScheduleTrain)
        .filter(
            models.ScheduleTrain.schedule_id == schedule_id,
            models.ScheduleTrain.scheduled_departure > current_time,
            models.ScheduleTrain.actual_departure.is_(None),
        )
        .order_by(models.ScheduleTrain.sequence)
        .all()
    )


def create_schedule_train(db: Session, st: schemas.ScheduleTrainCreate):
    db_st = models.ScheduleTrain(**st.model_dump())
    db.add(db_st)
    db.commit()
    db.refresh(db_st)
    return db_st


def update_schedule_train(db: Session, st_id: int, st: schemas.ScheduleTrainUpdate):
    db_st = get_schedule_train(db, st_id)
    if db_st:
        for key, value in st.model_dump(exclude_unset=True).items():
            setattr(db_st, key, value)
        db.commit()
        db.refresh(db_st)
    return db_st


def get_passenger_data(db: Session, line_id: int, start_time: datetime, end_time: datetime):
    return (
        db.query(models.PassengerData)
        .filter(
            models.PassengerData.line_id == line_id,
            models.PassengerData.timestamp >= start_time,
            models.PassengerData.timestamp <= end_time,
        )
        .all()
    )


def create_passenger_data(db: Session, data: schemas.PassengerDataCreate):
    db_data = models.PassengerData(**data.model_dump())
    db.add(db_data)
    db.commit()
    db.refresh(db_data)
    return db_data


def get_dispatch_logs(
    db: Session,
    entity_type: Optional[str] = None,
    level: Optional[LogLevel] = None,
    limit: int = 100,
):
    query = db.query(models.DispatchLog).order_by(
        models.DispatchLog.created_at.desc()
    )
    if entity_type:
        query = query.filter(models.DispatchLog.entity_type == entity_type)
    if level:
        query = query.filter(models.DispatchLog.level == level)
    return query.limit(limit).all()


def create_dispatch_log(db: Session, log: schemas.DispatchLogCreate):
    db_log = models.DispatchLog(**log.model_dump())
    db.add(db_log)
    db.commit()
    db.refresh(db_log)
    return db_log


def get_trains_in_section(db: Session, section_id: int):
    return (
        db.query(models.Train)
        .filter(models.Train.current_section_id == section_id)
        .all()
    )


def get_signals_in_section(db: Session, section_id: int):
    return (
        db.query(models.Signal)
        .filter(models.Signal.section_id == section_id)
        .all()
    )


def get_completed_schedule_trains(db: Session, schedule_id: int):
    return (
        db.query(models.ScheduleTrain)
        .filter(
            models.ScheduleTrain.schedule_id == schedule_id,
            models.ScheduleTrain.is_completed == True,
        )
        .order_by(models.ScheduleTrain.sequence)
        .all()
    )
