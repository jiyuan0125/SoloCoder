from typing import List, Optional
from datetime import datetime

from core.models import (
    Mine,
    MonitorZone,
    ThresholdConfig,
    SensorReading,
    AlarmEvent,
    ShiftStats,
    DailyReport,
)
from server.database import (
    MineDB,
    MonitorZoneDB,
    ThresholdConfigDB,
    SensorReadingDB,
    AlarmEventDB,
    ShiftStatsDB,
    DailyReportDB,
)


def mine_to_domain(db_obj: MineDB) -> Mine:
    return Mine(
        id=db_obj.id,
        name=db_obj.name,
        description=db_obj.description,
        created_at=db_obj.created_at,
    )


def mine_to_db(domain: Mine) -> MineDB:
    return MineDB(
        id=domain.id,
        name=domain.name,
        description=domain.description,
        created_at=domain.created_at,
    )


def zone_to_domain(db_obj: MonitorZoneDB) -> MonitorZone:
    return MonitorZone(
        id=db_obj.id,
        mine_id=db_obj.mine_id,
        name=db_obj.name,
        description=db_obj.description,
        created_at=db_obj.created_at,
    )


def zone_to_db(domain: MonitorZone) -> MonitorZoneDB:
    return MonitorZoneDB(
        id=domain.id,
        mine_id=domain.mine_id,
        name=domain.name,
        description=domain.description,
        created_at=domain.created_at,
    )


def threshold_to_domain(db_obj: ThresholdConfigDB) -> ThresholdConfig:
    return ThresholdConfig(
        id=db_obj.id,
        zone_id=db_obj.zone_id,
        metric_type=db_obj.metric_type,
        level1_threshold=db_obj.level1_threshold,
        level2_threshold=db_obj.level2_threshold,
        level3_threshold=db_obj.level3_threshold,
        created_at=db_obj.created_at,
    )


def threshold_to_db(domain: ThresholdConfig) -> ThresholdConfigDB:
    return ThresholdConfigDB(
        id=domain.id,
        zone_id=domain.zone_id,
        metric_type=domain.metric_type,
        level1_threshold=domain.level1_threshold,
        level2_threshold=domain.level2_threshold,
        level3_threshold=domain.level3_threshold,
        created_at=domain.created_at,
    )


def reading_to_domain(db_obj: SensorReadingDB) -> SensorReading:
    return SensorReading(
        id=db_obj.id,
        zone_id=db_obj.zone_id,
        timestamp=db_obj.timestamp,
        methane=db_obj.methane,
        co=db_obj.co,
        wind_speed=db_obj.wind_speed,
        temperature=db_obj.temperature,
        dust=db_obj.dust,
        shift_type=db_obj.shift_type,
    )


def reading_to_db(domain: SensorReading) -> SensorReadingDB:
    return SensorReadingDB(
        id=domain.id,
        zone_id=domain.zone_id,
        timestamp=domain.timestamp,
        methane=domain.methane,
        co=domain.co,
        wind_speed=domain.wind_speed,
        temperature=domain.temperature,
        dust=domain.dust,
        shift_type=domain.shift_type,
    )


def alarm_to_domain(db_obj: AlarmEventDB) -> AlarmEvent:
    return AlarmEvent(
        id=db_obj.id,
        zone_id=db_obj.zone_id,
        metric_type=db_obj.metric_type,
        alarm_level=db_obj.alarm_level,
        value=db_obj.value,
        threshold=db_obj.threshold,
        start_time=db_obj.start_time,
        end_time=db_obj.end_time,
        acknowledged=db_obj.acknowledged,
        acknowledged_at=db_obj.acknowledged_at,
        acknowledged_by=db_obj.acknowledged_by,
    )


def alarm_to_db(domain: AlarmEvent) -> AlarmEventDB:
    return AlarmEventDB(
        id=domain.id,
        zone_id=domain.zone_id,
        metric_type=domain.metric_type,
        alarm_level=domain.alarm_level,
        value=domain.value,
        threshold=domain.threshold,
        start_time=domain.start_time,
        end_time=domain.end_time,
        acknowledged=domain.acknowledged,
        acknowledged_at=domain.acknowledged_at,
        acknowledged_by=domain.acknowledged_by,
    )


def shift_stats_to_domain(db_obj: ShiftStatsDB) -> ShiftStats:
    return ShiftStats(
        id=db_obj.id,
        zone_id=db_obj.zone_id,
        shift_date=db_obj.shift_date,
        shift_type=db_obj.shift_type,
        methane_avg=db_obj.methane_avg,
        methane_max=db_obj.methane_max,
        methane_over_count=db_obj.methane_over_count,
        co_avg=db_obj.co_avg,
        co_max=db_obj.co_max,
        co_over_count=db_obj.co_over_count,
        wind_speed_avg=db_obj.wind_speed_avg,
        wind_speed_max=db_obj.wind_speed_max,
        wind_speed_over_count=db_obj.wind_speed_over_count,
        temperature_avg=db_obj.temperature_avg,
        temperature_max=db_obj.temperature_max,
        temperature_over_count=db_obj.temperature_over_count,
        dust_avg=db_obj.dust_avg,
        dust_max=db_obj.dust_max,
        dust_over_count=db_obj.dust_over_count,
        reading_count=db_obj.reading_count,
        created_at=db_obj.created_at,
    )


def shift_stats_to_db(domain: ShiftStats) -> ShiftStatsDB:
    return ShiftStatsDB(
        id=domain.id,
        zone_id=domain.zone_id,
        shift_date=domain.shift_date,
        shift_type=domain.shift_type,
        methane_avg=domain.methane_avg,
        methane_max=domain.methane_max,
        methane_over_count=domain.methane_over_count,
        co_avg=domain.co_avg,
        co_max=domain.co_max,
        co_over_count=domain.co_over_count,
        wind_speed_avg=domain.wind_speed_avg,
        wind_speed_max=domain.wind_speed_max,
        wind_speed_over_count=domain.wind_speed_over_count,
        temperature_avg=domain.temperature_avg,
        temperature_max=domain.temperature_max,
        temperature_over_count=domain.temperature_over_count,
        dust_avg=domain.dust_avg,
        dust_max=domain.dust_max,
        dust_over_count=domain.dust_over_count,
        reading_count=domain.reading_count,
        created_at=domain.created_at,
    )


def daily_report_to_domain(db_obj: DailyReportDB) -> DailyReport:
    return DailyReport(
        id=db_obj.id,
        zone_id=db_obj.zone_id,
        report_date=db_obj.report_date,
        methane_avg=db_obj.methane_avg,
        methane_max=db_obj.methane_max,
        methane_over_count=db_obj.methane_over_count,
        co_avg=db_obj.co_avg,
        co_max=db_obj.co_max,
        co_over_count=db_obj.co_over_count,
        wind_speed_avg=db_obj.wind_speed_avg,
        wind_speed_max=db_obj.wind_speed_max,
        wind_speed_over_count=db_obj.wind_speed_over_count,
        temperature_avg=db_obj.temperature_avg,
        temperature_max=db_obj.temperature_max,
        temperature_over_count=db_obj.temperature_over_count,
        dust_avg=db_obj.dust_avg,
        dust_max=db_obj.dust_max,
        dust_over_count=db_obj.dust_over_count,
        alarm_count=db_obj.alarm_count,
        reading_count=db_obj.reading_count,
        created_at=db_obj.created_at,
    )


def daily_report_to_db(domain: DailyReport) -> DailyReportDB:
    return DailyReportDB(
        id=domain.id,
        zone_id=domain.zone_id,
        report_date=domain.report_date,
        methane_avg=domain.methane_avg,
        methane_max=domain.methane_max,
        methane_over_count=domain.methane_over_count,
        co_avg=domain.co_avg,
        co_max=domain.co_max,
        co_over_count=domain.co_over_count,
        wind_speed_avg=domain.wind_speed_avg,
        wind_speed_max=domain.wind_speed_max,
        wind_speed_over_count=domain.wind_speed_over_count,
        temperature_avg=domain.temperature_avg,
        temperature_max=domain.temperature_max,
        temperature_over_count=domain.temperature_over_count,
        dust_avg=domain.dust_avg,
        dust_max=domain.dust_max,
        dust_over_count=domain.dust_over_count,
        alarm_count=domain.alarm_count,
        reading_count=domain.reading_count,
        created_at=domain.created_at,
    )
