from datetime import datetime, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from fastapi.responses import StreamingResponse, Response
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_

from core.models import (
    Mine,
    MonitorZone,
    ThresholdConfig,
    SensorReading,
    AlarmEvent,
    ShiftStats,
    DailyReport,
    ShiftType,
    MetricType,
)
from core.shifts import get_shift_info
from core.alarms import check_alarms
from core.aggregation import aggregate_readings_by_shift, aggregate_shift_stats_by_day
from core.export import (
    export_readings_to_csv,
    export_shift_stats_to_csv,
    export_daily_reports_to_csv,
    export_alarms_to_csv,
)
from server.database import (
    get_db,
    MineDB,
    MonitorZoneDB,
    ThresholdConfigDB,
    SensorReadingDB,
    AlarmEventDB,
    ShiftStatsDB,
    DailyReportDB,
)
from server.models import (
    mine_to_domain, mine_to_db,
    zone_to_domain, zone_to_db,
    threshold_to_domain, threshold_to_db,
    reading_to_domain, reading_to_db,
    alarm_to_domain, alarm_to_db,
    shift_stats_to_domain, shift_stats_to_db,
    daily_report_to_domain, daily_report_to_db,
)

router = APIRouter()


@router.post("/mines", response_model=Mine)
def create_mine(mine: Mine, db: Session = Depends(get_db)):
    db_mine = mine_to_db(mine)
    db.add(db_mine)
    db.commit()
    db.refresh(db_mine)
    return mine_to_domain(db_mine)


@router.get("/mines", response_model=List[Mine])
def list_mines(
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    mines = db.query(MineDB).offset(skip).limit(limit).all()
    return [mine_to_domain(m) for m in mines]


@router.get("/mines/{mine_id}", response_model=Mine)
def get_mine(mine_id: int, db: Session = Depends(get_db)):
    mine = db.query(MineDB).filter(MineDB.id == mine_id).first()
    if mine is None:
        raise HTTPException(status_code=404, detail="Mine not found")
    return mine_to_domain(mine)


@router.put("/mines/{mine_id}", response_model=Mine)
def update_mine(mine_id: int, mine: Mine, db: Session = Depends(get_db)):
    db_mine = db.query(MineDB).filter(MineDB.id == mine_id).first()
    if db_mine is None:
        raise HTTPException(status_code=404, detail="Mine not found")
    
    db_mine.name = mine.name
    db_mine.description = mine.description
    db.commit()
    db.refresh(db_mine)
    return mine_to_domain(db_mine)


@router.delete("/mines/{mine_id}")
def delete_mine(mine_id: int, db: Session = Depends(get_db)):
    mine = db.query(MineDB).filter(MineDB.id == mine_id).first()
    if mine is None:
        raise HTTPException(status_code=404, detail="Mine not found")
    db.delete(mine)
    db.commit()
    return {"message": "Deleted successfully"}


@router.post("/zones", response_model=MonitorZone)
def create_zone(zone: MonitorZone, db: Session = Depends(get_db)):
    mine = db.query(MineDB).filter(MineDB.id == zone.mine_id).first()
    if mine is None:
        raise HTTPException(status_code=404, detail="Mine not found")
    
    db_zone = zone_to_db(zone)
    db.add(db_zone)
    db.commit()
    db.refresh(db_zone)
    return zone_to_domain(db_zone)


@router.get("/zones", response_model=List[MonitorZone])
def list_zones(
    mine_id: Optional[int] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(MonitorZoneDB)
    if mine_id is not None:
        query = query.filter(MonitorZoneDB.mine_id == mine_id)
    zones = query.offset(skip).limit(limit).all()
    return [zone_to_domain(z) for z in zones]


@router.get("/zones/{zone_id}", response_model=MonitorZone)
def get_zone(zone_id: int, db: Session = Depends(get_db)):
    zone = db.query(MonitorZoneDB).filter(MonitorZoneDB.id == zone_id).first()
    if zone is None:
        raise HTTPException(status_code=404, detail="Zone not found")
    return zone_to_domain(zone)


@router.put("/zones/{zone_id}", response_model=MonitorZone)
def update_zone(zone_id: int, zone: MonitorZone, db: Session = Depends(get_db)):
    db_zone = db.query(MonitorZoneDB).filter(MonitorZoneDB.id == zone_id).first()
    if db_zone is None:
        raise HTTPException(status_code=404, detail="Zone not found")
    
    db_zone.name = zone.name
    db_zone.description = zone.description
    db.commit()
    db.refresh(db_zone)
    return zone_to_domain(db_zone)


@router.delete("/zones/{zone_id}")
def delete_zone(zone_id: int, db: Session = Depends(get_db)):
    zone = db.query(MonitorZoneDB).filter(MonitorZoneDB.id == zone_id).first()
    if zone is None:
        raise HTTPException(status_code=404, detail="Zone not found")
    db.delete(zone)
    db.commit()
    return {"message": "Deleted successfully"}


@router.post("/thresholds", response_model=ThresholdConfig)
def create_threshold(threshold: ThresholdConfig, db: Session = Depends(get_db)):
    zone = db.query(MonitorZoneDB).filter(MonitorZoneDB.id == threshold.zone_id).first()
    if zone is None:
        raise HTTPException(status_code=404, detail="Zone not found")
    
    existing = db.query(ThresholdConfigDB).filter(
        and_(
            ThresholdConfigDB.zone_id == threshold.zone_id,
            ThresholdConfigDB.metric_type == threshold.metric_type,
        )
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="Threshold already exists for this metric")
    
    db_threshold = threshold_to_db(threshold)
    db.add(db_threshold)
    db.commit()
    db.refresh(db_threshold)
    return threshold_to_domain(db_threshold)


@router.get("/thresholds", response_model=List[ThresholdConfig])
def list_thresholds(
    zone_id: Optional[int] = None,
    db: Session = Depends(get_db),
):
    query = db.query(ThresholdConfigDB)
    if zone_id is not None:
        query = query.filter(ThresholdConfigDB.zone_id == zone_id)
    thresholds = query.all()
    return [threshold_to_domain(t) for t in thresholds]


@router.put("/thresholds/{threshold_id}", response_model=ThresholdConfig)
def update_threshold(
    threshold_id: int,
    threshold: ThresholdConfig,
    db: Session = Depends(get_db),
):
    db_threshold = db.query(ThresholdConfigDB).filter(
        ThresholdConfigDB.id == threshold_id
    ).first()
    if db_threshold is None:
        raise HTTPException(status_code=404, detail="Threshold not found")
    
    db_threshold.level1_threshold = threshold.level1_threshold
    db_threshold.level2_threshold = threshold.level2_threshold
    db_threshold.level3_threshold = threshold.level3_threshold
    db.commit()
    db.refresh(db_threshold)
    return threshold_to_domain(db_threshold)


@router.post("/readings", response_model=SensorReading)
def create_reading(reading: SensorReading, db: Session = Depends(get_db)):
    zone = db.query(MonitorZoneDB).filter(MonitorZoneDB.id == reading.zone_id).first()
    if zone is None:
        raise HTTPException(status_code=404, detail="Zone not found")
    
    shift_type, _ = get_shift_info(reading.timestamp)
    reading.shift_type = shift_type
    
    db_reading = reading_to_db(reading)
    db.add(db_reading)
    db.flush()
    
    thresholds = db.query(ThresholdConfigDB).filter(
        ThresholdConfigDB.zone_id == reading.zone_id
    ).all()
    thresholds_domain = [threshold_to_domain(t) for t in thresholds]
    
    existing_alarms = db.query(AlarmEventDB).filter(
        and_(
            AlarmEventDB.zone_id == reading.zone_id,
            AlarmEventDB.end_time == None,
        )
    ).all()
    existing_alarms_domain = [alarm_to_domain(a) for a in existing_alarms]
    
    new_alarms, updated_alarms = check_alarms(
        reading,
        thresholds_domain,
        existing_alarms_domain,
    )
    
    for alarm in new_alarms:
        db.add(alarm_to_db(alarm))
    
    for alarm in updated_alarms:
        db_alarm = db.query(AlarmEventDB).filter(
            AlarmEventDB.id == alarm.id
        ).first()
        if db_alarm:
            db_alarm.value = alarm.value
            db_alarm.alarm_level = alarm.alarm_level
            db_alarm.threshold = alarm.threshold
            db_alarm.end_time = alarm.end_time
    
    db.commit()
    db.refresh(db_reading)
    
    return reading_to_domain(db_reading)


@router.get("/readings", response_model=List[SensorReading])
def list_readings(
    zone_id: Optional[int] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    skip: int = 0,
    limit: int = 1000,
    db: Session = Depends(get_db),
):
    query = db.query(SensorReadingDB).order_by(SensorReadingDB.timestamp.desc())
    if zone_id is not None:
        query = query.filter(SensorReadingDB.zone_id == zone_id)
    if start_time is not None:
        query = query.filter(SensorReadingDB.timestamp >= start_time)
    if end_time is not None:
        query = query.filter(SensorReadingDB.timestamp <= end_time)
    
    readings = query.offset(skip).limit(limit).all()
    return [reading_to_domain(r) for r in readings]


@router.get("/readings/export")
def export_readings(
    zone_id: Optional[int] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    db: Session = Depends(get_db),
):
    query = db.query(SensorReadingDB).order_by(SensorReadingDB.timestamp.asc())
    if zone_id is not None:
        query = query.filter(SensorReadingDB.zone_id == zone_id)
    if start_time is not None:
        query = query.filter(SensorReadingDB.timestamp >= start_time)
    if end_time is not None:
        query = query.filter(SensorReadingDB.timestamp <= end_time)
    
    readings = query.all()
    domain_readings = [reading_to_domain(r) for r in readings]
    csv_data = export_readings_to_csv(domain_readings)
    
    filename = f"readings_{datetime.now().strftime('%Y%m%d%H%M%S')}.csv"
    return Response(
        content=csv_data,
        media_type="text/csv; charset=utf-8-sig",
        headers={
            "Content-Disposition": f'attachment; filename="{filename}"',
            "Content-Type": "text/csv; charset=utf-8-sig",
        },
    )


@router.get("/alarms", response_model=List[AlarmEvent])
def list_alarms(
    zone_id: Optional[int] = None,
    metric_type: Optional[MetricType] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    acknowledged: Optional[bool] = None,
    skip: int = 0,
    limit: int = 1000,
    db: Session = Depends(get_db),
):
    query = db.query(AlarmEventDB).order_by(AlarmEventDB.start_time.desc())
    if zone_id is not None:
        query = query.filter(AlarmEventDB.zone_id == zone_id)
    if metric_type is not None:
        query = query.filter(AlarmEventDB.metric_type == metric_type)
    if start_time is not None:
        query = query.filter(AlarmEventDB.start_time >= start_time)
    if end_time is not None:
        query = query.filter(
            or_(
                AlarmEventDB.end_time <= end_time,
                AlarmEventDB.end_time == None,
            )
        )
    if acknowledged is not None:
        query = query.filter(AlarmEventDB.acknowledged == acknowledged)
    
    alarms = query.offset(skip).limit(limit).all()
    return [alarm_to_domain(a) for a in alarms]


@router.post("/alarms/{alarm_id}/acknowledge", response_model=AlarmEvent)
def acknowledge_alarm(
    alarm_id: int,
    acknowledged_by: Optional[str] = None,
    db: Session = Depends(get_db),
):
    db_alarm = db.query(AlarmEventDB).filter(AlarmEventDB.id == alarm_id).first()
    if db_alarm is None:
        raise HTTPException(status_code=404, detail="Alarm not found")
    
    db_alarm.acknowledged = True
    db_alarm.acknowledged_at = datetime.utcnow()
    db_alarm.acknowledged_by = acknowledged_by
    db.commit()
    db.refresh(db_alarm)
    return alarm_to_domain(db_alarm)


@router.get("/alarms/export")
def export_alarms(
    zone_id: Optional[int] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    db: Session = Depends(get_db),
):
    query = db.query(AlarmEventDB).order_by(AlarmEventDB.start_time.asc())
    if zone_id is not None:
        query = query.filter(AlarmEventDB.zone_id == zone_id)
    if start_time is not None:
        query = query.filter(AlarmEventDB.start_time >= start_time)
    if end_time is not None:
        query = query.filter(
            or_(
                AlarmEventDB.end_time <= end_time,
                AlarmEventDB.end_time == None,
            )
        )
    
    alarms = query.all()
    domain_alarms = [alarm_to_domain(a) for a in alarms]
    csv_data = export_alarms_to_csv(domain_alarms)
    
    filename = f"alarms_{datetime.now().strftime('%Y%m%d%H%M%S')}.csv"
    return Response(
        content=csv_data,
        media_type="text/csv; charset=utf-8-sig",
        headers={
            "Content-Disposition": f'attachment; filename="{filename}"',
            "Content-Type": "text/csv; charset=utf-8-sig",
        },
    )


@router.post("/stats/generate-shift")
def generate_shift_stats(
    zone_id: Optional[int] = None,
    db: Session = Depends(get_db),
):
    query = db.query(SensorReadingDB)
    if zone_id is not None:
        query = query.filter(SensorReadingDB.zone_id == zone_id)
    
    readings = query.all()
    if not readings:
        return {"message": "No readings to process", "stats_created": 0}
    
    thresholds = db.query(ThresholdConfigDB).all()
    thresholds_domain = [threshold_to_domain(t) for t in thresholds]
    readings_domain = [reading_to_domain(r) for r in readings]
    
    stats_list = aggregate_readings_by_shift(readings_domain, thresholds_domain)
    
    created_count = 0
    for stats in stats_list:
        existing = db.query(ShiftStatsDB).filter(
            and_(
                ShiftStatsDB.zone_id == stats.zone_id,
                ShiftStatsDB.shift_date == stats.shift_date,
                ShiftStatsDB.shift_type == stats.shift_type,
            )
        ).first()
        
        if existing:
            existing.methane_avg = stats.methane_avg
            existing.methane_max = stats.methane_max
            existing.methane_over_count = stats.methane_over_count
            existing.co_avg = stats.co_avg
            existing.co_max = stats.co_max
            existing.co_over_count = stats.co_over_count
            existing.wind_speed_avg = stats.wind_speed_avg
            existing.wind_speed_max = stats.wind_speed_max
            existing.wind_speed_over_count = stats.wind_speed_over_count
            existing.temperature_avg = stats.temperature_avg
            existing.temperature_max = stats.temperature_max
            existing.temperature_over_count = stats.temperature_over_count
            existing.dust_avg = stats.dust_avg
            existing.dust_max = stats.dust_max
            existing.dust_over_count = stats.dust_over_count
            existing.reading_count = stats.reading_count
        else:
            db.add(shift_stats_to_db(stats))
            created_count += 1
    
    db.commit()
    return {"message": "Shift stats generated", "stats_created": created_count}


@router.get("/stats/shift", response_model=List[ShiftStats])
def list_shift_stats(
    zone_id: Optional[int] = None,
    shift_date: Optional[str] = None,
    shift_type: Optional[ShiftType] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(ShiftStatsDB).order_by(
        ShiftStatsDB.shift_date.desc(),
        ShiftStatsDB.shift_type.asc(),
    )
    if zone_id is not None:
        query = query.filter(ShiftStatsDB.zone_id == zone_id)
    if shift_date is not None:
        query = query.filter(ShiftStatsDB.shift_date == shift_date)
    if shift_type is not None:
        query = query.filter(ShiftStatsDB.shift_type == shift_type)
    
    stats = query.offset(skip).limit(limit).all()
    return [shift_stats_to_domain(s) for s in stats]


@router.get("/stats/shift/export")
def export_shift_stats(
    zone_id: Optional[int] = None,
    start_date: Optional[str] = None,
    end_date: Optional[str] = None,
    db: Session = Depends(get_db),
):
    query = db.query(ShiftStatsDB).order_by(
        ShiftStatsDB.shift_date.asc(),
        ShiftStatsDB.shift_type.asc(),
    )
    if zone_id is not None:
        query = query.filter(ShiftStatsDB.zone_id == zone_id)
    if start_date is not None:
        query = query.filter(ShiftStatsDB.shift_date >= start_date)
    if end_date is not None:
        query = query.filter(ShiftStatsDB.shift_date <= end_date)
    
    stats = query.all()
    csv_data = export_shift_stats_to_csv(stats)
    
    filename = f"shift_stats_{datetime.now().strftime('%Y%m%d%H%M%S')}.csv"
    return Response(
        content=csv_data,
        media_type="text/csv; charset=utf-8-sig",
        headers={
            "Content-Disposition": f'attachment; filename="{filename}"',
            "Content-Type": "text/csv; charset=utf-8-sig",
        },
    )


@router.post("/stats/generate-daily")
def generate_daily_reports(
    zone_id: Optional[int] = None,
    db: Session = Depends(get_db),
):
    query = db.query(ShiftStatsDB)
    if zone_id is not None:
        query = query.filter(ShiftStatsDB.zone_id == zone_id)
    
    shift_stats = query.all()
    if not shift_stats:
        return {"message": "No shift stats to process", "reports_created": 0}
    
    shift_stats_domain = [shift_stats_to_domain(s) for s in shift_stats]
    
    alarm_query = db.query(AlarmEventDB)
    if zone_id is not None:
        alarm_query = alarm_query.filter(AlarmEventDB.zone_id == zone_id)
    alarms = alarm_query.all()
    
    alarm_counts: dict = {}
    for alarm in alarms:
        from core.shifts import get_shift_date, get_shift_type
        key = (alarm.zone_id, get_shift_date(alarm.start_time, get_shift_type(alarm.start_time)))
        alarm_counts[key] = alarm_counts.get(key, 0) + 1
    
    reports = aggregate_shift_stats_by_day(shift_stats_domain, alarm_counts)
    
    created_count = 0
    for report in reports:
        existing = db.query(DailyReportDB).filter(
            and_(
                DailyReportDB.zone_id == report.zone_id,
                DailyReportDB.report_date == report.report_date,
            )
        ).first()
        
        if existing:
            existing.methane_avg = report.methane_avg
            existing.methane_max = report.methane_max
            existing.methane_over_count = report.methane_over_count
            existing.co_avg = report.co_avg
            existing.co_max = report.co_max
            existing.co_over_count = report.co_over_count
            existing.wind_speed_avg = report.wind_speed_avg
            existing.wind_speed_max = report.wind_speed_max
            existing.wind_speed_over_count = report.wind_speed_over_count
            existing.temperature_avg = report.temperature_avg
            existing.temperature_max = report.temperature_max
            existing.temperature_over_count = report.temperature_over_count
            existing.dust_avg = report.dust_avg
            existing.dust_max = report.dust_max
            existing.dust_over_count = report.dust_over_count
            existing.alarm_count = report.alarm_count
            existing.reading_count = report.reading_count
        else:
            db.add(daily_report_to_db(report))
            created_count += 1
    
    db.commit()
    return {"message": "Daily reports generated", "reports_created": created_count}


@router.get("/stats/daily", response_model=List[DailyReport])
def list_daily_reports(
    zone_id: Optional[int] = None,
    report_date: Optional[str] = None,
    start_date: Optional[str] = None,
    end_date: Optional[str] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(DailyReportDB).order_by(DailyReportDB.report_date.desc())
    if zone_id is not None:
        query = query.filter(DailyReportDB.zone_id == zone_id)
    if report_date is not None:
        query = query.filter(DailyReportDB.report_date == report_date)
    if start_date is not None:
        query = query.filter(DailyReportDB.report_date >= start_date)
    if end_date is not None:
        query = query.filter(DailyReportDB.report_date <= end_date)
    
    reports = query.offset(skip).limit(limit).all()
    return [daily_report_to_domain(r) for r in reports]


@router.get("/stats/daily/export")
def export_daily_reports(
    zone_id: Optional[int] = None,
    start_date: Optional[str] = None,
    end_date: Optional[str] = None,
    db: Session = Depends(get_db),
):
    query = db.query(DailyReportDB).order_by(DailyReportDB.report_date.asc())
    if zone_id is not None:
        query = query.filter(DailyReportDB.zone_id == zone_id)
    if start_date is not None:
        query = query.filter(DailyReportDB.report_date >= start_date)
    if end_date is not None:
        query = query.filter(DailyReportDB.report_date <= end_date)
    
    reports = query.all()
    csv_data = export_daily_reports_to_csv(reports)
    
    filename = f"daily_reports_{datetime.now().strftime('%Y%m%d%H%M%S')}.csv"
    return Response(
        content=csv_data,
        media_type="text/csv; charset=utf-8-sig",
        headers={
            "Content-Disposition": f'attachment; filename="{filename}"',
            "Content-Type": "text/csv; charset=utf-8-sig",
        },
    )
