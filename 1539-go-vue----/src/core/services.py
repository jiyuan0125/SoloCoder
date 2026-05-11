from datetime import datetime, date, timedelta
from typing import List, Optional, Tuple
import csv
import io

from .models import (
    ZoneType,
    ZoneLimit,
    MonitoringPoint,
    NoiseData,
    TimePeriod,
    AlertEvent,
    AlertType,
    AlertLevel,
    ConstructionPermit,
    DailyStatistics,
)
from .storage import storage


DAYTIME_START = 6
NIGHTTIME_START = 22
CONTINUOUS_EXCEED_HOURS = 2
DATA_MISSING_THRESHOLD_HOURS = 4
CALIBRATION_INTERVAL_MONTHS = 12
CALIBRATION_NOTICE_DAYS = 30


def get_time_period(dt: datetime) -> TimePeriod:
    if DAYTIME_START <= dt.hour < NIGHTTIME_START:
        return TimePeriod.DAYTIME
    else:
        return TimePeriod.NIGHTTIME


def get_zone_limit(zone_type: ZoneType) -> Optional[ZoneLimit]:
    limits = storage.query(ZoneLimit, zone_type=zone_type)
    return limits[0] if limits else None


def has_active_construction_permit(point_id: int, check_date: date) -> bool:
    permits = storage.get_all(ConstructionPermit)
    for permit in permits:
        if (
            permit.is_active 
            and point_id in permit.point_ids 
            and permit.start_date <= check_date <= permit.end_date
        ):
            return True
    return False


def get_effective_limit(
    point: MonitoringPoint, 
    time_period: TimePeriod, 
    check_datetime: datetime
) -> Optional[float]:
    zone_limit = get_zone_limit(point.zone_type)
    if not zone_limit:
        return None
    
    base_limit = zone_limit.daytime_limit if time_period == TimePeriod.DAYTIME else zone_limit.nighttime_limit
    
    if time_period == TimePeriod.NIGHTTIME:
        if has_active_construction_permit(point.id, check_datetime.date()):
            base_limit += 10
    
    return base_limit


def check_exceedance(point: MonitoringPoint, noise_value: float, timestamp: datetime) -> bool:
    time_period = get_time_period(timestamp)
    effective_limit = get_effective_limit(point, time_period, timestamp)
    
    if effective_limit is None:
        return False
    
    return noise_value > effective_limit


def check_continuous_exceedance(point_id: int, latest_timestamp: datetime) -> Tuple[bool, Optional[datetime], Optional[datetime]]:
    noise_data_list = storage.query(NoiseData, point_id=point_id)
    noise_data_list.sort(key=lambda x: x.timestamp)
    
    exceed_start = None
    exceed_end = None
    
    for data in noise_data_list:
        if data.is_exceeded:
            if exceed_start is None:
                exceed_start = data.timestamp
            exceed_end = data.timestamp
        else:
            exceed_start = None
            exceed_end = None
    
    if exceed_start and exceed_end:
        duration = exceed_end - exceed_start
        if duration >= timedelta(hours=CONTINUOUS_EXCEED_HOURS):
            return True, exceed_start, exceed_end
    
    return False, None, None


def create_alert(
    point_id: int, 
    alert_type: AlertType, 
    level: AlertLevel, 
    message: str,
    start_time: datetime,
    end_time: Optional[datetime] = None
) -> AlertEvent:
    existing_alerts = storage.query(
        AlertEvent, 
        point_id=point_id, 
        alert_type=alert_type,
        is_resolved=False
    )
    
    if existing_alerts:
        return existing_alerts[0]
    
    alert = AlertEvent(
        id=0,
        point_id=point_id,
        alert_type=alert_type,
        level=level,
        message=message,
        start_time=start_time,
        end_time=end_time,
    )
    
    return storage.create(alert)


def process_noise_data(
    point_id: int, 
    value: float, 
    timestamp: Optional[datetime] = None
) -> NoiseData:
    if timestamp is None:
        timestamp = datetime.now()
    
    point = storage.get_by_id(MonitoringPoint, point_id)
    if not point:
        raise ValueError(f"Monitoring point {point_id} not found")
    
    if not point.is_active:
        raise ValueError(f"Monitoring point {point_id} is inactive")
    
    time_period = get_time_period(timestamp)
    is_exceeded = check_exceedance(point, value, timestamp)
    
    noise_data = NoiseData(
        id=0,
        point_id=point_id,
        timestamp=timestamp,
        value=value,
        is_exceeded=is_exceeded,
        time_period=time_period,
    )
    
    saved_data = storage.create(noise_data)
    
    if is_exceeded:
        is_continuous, start_time, end_time = check_continuous_exceedance(point_id, timestamp)
        if is_continuous and start_time:
            create_alert(
                point_id=point_id,
                alert_type=AlertType.EXCEED,
                level=AlertLevel.WARNING,
                message=f"连续{CONTINUOUS_EXCEED_HOURS}小时噪声超标",
                start_time=start_time,
                end_time=end_time,
            )
    
    return saved_data


def check_calibration_status() -> List[AlertEvent]:
    now = datetime.now()
    points = storage.get_all(MonitoringPoint)
    alerts = []
    
    for point in points:
        days_until_calibration = (point.next_calibration_date - now.date()).days
        
        if days_until_calibration <= CALIBRATION_NOTICE_DAYS:
            if not point.calibration_due:
                point.calibration_due = True
                storage.update(point)
                
                alert = create_alert(
                    point_id=point.id,
                    alert_type=AlertType.CALIBRATION,
                    level=AlertLevel.INFO if days_until_calibration > 0 else AlertLevel.WARNING,
                    message=f"设备检定到期提醒：还有{days_until_calibration}天" if days_until_calibration > 0 else "设备检定已到期",
                    start_time=now,
                )
                alerts.append(alert)
    
    return alerts


def calculate_daily_statistics(point_id: int, target_date: date) -> DailyStatistics:
    point = storage.get_by_id(MonitoringPoint, point_id)
    if not point:
        raise ValueError(f"Monitoring point {point_id} not found")
    
    start_of_day = datetime.combine(target_date, datetime.min.time())
    end_of_day = start_of_day + timedelta(days=1)
    
    all_data = storage.query(
        NoiseData, 
        point_id=point_id,
        timestamp=lambda t: start_of_day <= t < end_of_day
    )
    
    daytime_data = [d for d in all_data if d.time_period == TimePeriod.DAYTIME and d.is_valid]
    nighttime_data = [d for d in all_data if d.time_period == TimePeriod.NIGHTTIME and d.is_valid]
    
    total_expected_slots = 24 * 6  # 每10分钟一个数据点，共144个
    actual_valid_slots = len([d for d in all_data if d.is_valid])
    missing_slots = total_expected_slots - actual_valid_slots
    data_missing_hours = (missing_slots * 10) / 60
    
    is_valid_for_compliance = data_missing_hours <= DATA_MISSING_THRESHOLD_HOURS
    
    def calc_avg(data_list):
        if not data_list:
            return None
        return sum(d.value for d in data_list) / len(data_list)
    
    def calc_compliance_rate(data_list):
        if not data_list:
            return None
        valid = len([d for d in data_list if not d.is_exceeded])
        return valid / len(data_list) * 100
    
    daytime_avg = calc_avg(daytime_data)
    nighttime_avg = calc_avg(nighttime_data)
    overall_avg = calc_avg([d for d in all_data if d.is_valid])
    
    daytime_compliance = calc_compliance_rate(daytime_data) if is_valid_for_compliance else None
    nighttime_compliance = calc_compliance_rate(nighttime_data) if is_valid_for_compliance else None
    overall_compliance = calc_compliance_rate([d for d in all_data if d.is_valid]) if is_valid_for_compliance else None
    
    stats = DailyStatistics(
        id=0,
        point_id=point_id,
        date=target_date,
        daytime_avg=round(daytime_avg, 1) if daytime_avg is not None else None,
        nighttime_avg=round(nighttime_avg, 1) if nighttime_avg is not None else None,
        overall_avg=round(overall_avg, 1) if overall_avg is not None else None,
        daytime_compliance_rate=round(daytime_compliance, 1) if daytime_compliance is not None else None,
        nighttime_compliance_rate=round(nighttime_compliance, 1) if nighttime_compliance is not None else None,
        overall_compliance_rate=round(overall_compliance, 1) if overall_compliance is not None else None,
        data_missing_hours=round(data_missing_hours, 1),
        is_valid_for_compliance=is_valid_for_compliance,
    )
    
    existing = storage.query(DailyStatistics, point_id=point_id, date=target_date)
    if existing:
        stats.id = existing[0].id
        return storage.update(stats)
    
    return storage.create(stats)


def export_to_csv(data_list: List[NoiseData]) -> str:
    output = io.StringIO()
    writer = csv.writer(output)
    
    writer.writerow([
        'ID', '点位ID', '时间', '噪声值(dB)', '是否超标', '时段', '是否有效'
    ])
    
    for data in data_list:
        writer.writerow([
            data.id,
            data.point_id,
            data.timestamp.strftime('%Y-%m-%d %H:%M:%S'),
            round(data.value, 1),
            '是' if data.is_exceeded else '否',
            '昼间' if data.time_period == TimePeriod.DAYTIME else '夜间',
            '是' if data.is_valid else '否',
        ])
    
    return output.getvalue()


def export_daily_stats_to_csv(stats_list: List[DailyStatistics]) -> str:
    output = io.StringIO()
    writer = csv.writer(output)
    
    writer.writerow([
        'ID', '点位ID', '日期', '昼间均值', '夜间均值', '总体均值',
        '昼间达标率(%)', '夜间达标率(%)', '总体达标率(%)',
        '数据缺失小时数', '是否参与达标计算'
    ])
    
    for stats in stats_list:
        writer.writerow([
            stats.id,
            stats.point_id,
            stats.date.strftime('%Y-%m-%d'),
            round(stats.daytime_avg, 1) if stats.daytime_avg is not None else '',
            round(stats.nighttime_avg, 1) if stats.nighttime_avg is not None else '',
            round(stats.overall_avg, 1) if stats.overall_avg is not None else '',
            round(stats.daytime_compliance_rate, 1) if stats.daytime_compliance_rate is not None else '',
            round(stats.nighttime_compliance_rate, 1) if stats.nighttime_compliance_rate is not None else '',
            round(stats.overall_compliance_rate, 1) if stats.overall_compliance_rate is not None else '',
            round(stats.data_missing_hours, 1),
            '是' if stats.is_valid_for_compliance else '否',
        ])
    
    return output.getvalue()
