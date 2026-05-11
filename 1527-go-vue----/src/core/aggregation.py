from datetime import datetime
from typing import List, Optional, Dict, Tuple
from collections import defaultdict

from .models import (
    SensorReading,
    ShiftStats,
    DailyReport,
    ShiftType,
    ThresholdConfig,
    MetricType,
)
from .shifts import get_shift_info
from .alarms import is_over_threshold, get_metric_value


def calculate_shift_stats(
    readings: List[SensorReading],
    zone_id: int,
    shift_date: str,
    shift_type: ShiftType,
    thresholds: List[ThresholdConfig],
) -> ShiftStats:
    stats = ShiftStats(
        zone_id=zone_id,
        shift_date=shift_date,
        shift_type=shift_type,
    )
    
    if not readings:
        return stats
    
    methane_values = [r.methane for r in readings]
    co_values = [r.co for r in readings]
    wind_speed_values = [r.wind_speed for r in readings]
    temperature_values = [r.temperature for r in readings]
    dust_values = [r.dust for r in readings]
    
    stats.reading_count = len(readings)
    stats.methane_avg = round(sum(methane_values) / len(methane_values), 2)
    stats.methane_max = round(max(methane_values), 2)
    stats.co_avg = round(sum(co_values) / len(co_values), 2)
    stats.co_max = round(max(co_values), 2)
    stats.wind_speed_avg = round(sum(wind_speed_values) / len(wind_speed_values), 2)
    stats.wind_speed_max = round(max(wind_speed_values), 2)
    stats.temperature_avg = round(sum(temperature_values) / len(temperature_values), 2)
    stats.temperature_max = round(max(temperature_values), 2)
    stats.dust_avg = round(sum(dust_values) / len(dust_values), 2)
    stats.dust_max = round(max(dust_values), 2)
    
    threshold_map = {t.metric_type: t for t in thresholds}
    
    for reading in readings:
        for metric_type in MetricType:
            threshold = threshold_map.get(metric_type)
            if threshold is None:
                continue
            value = get_metric_value(reading, metric_type)
            if is_over_threshold(value, threshold):
                if metric_type == MetricType.METHANE:
                    stats.methane_over_count += 1
                elif metric_type == MetricType.CO:
                    stats.co_over_count += 1
                elif metric_type == MetricType.WIND_SPEED:
                    stats.wind_speed_over_count += 1
                elif metric_type == MetricType.TEMPERATURE:
                    stats.temperature_over_count += 1
                elif metric_type == MetricType.DUST:
                    stats.dust_over_count += 1
    
    return stats


def aggregate_readings_by_shift(
    readings: List[SensorReading],
    thresholds: List[ThresholdConfig],
) -> List[ShiftStats]:
    grouped: Dict[Tuple[int, str, ShiftType], List[SensorReading]] = defaultdict(list)
    
    for reading in readings:
        zone_id = reading.zone_id
        shift_type, shift_date = get_shift_info(reading.timestamp)
        key = (zone_id, shift_date, shift_type)
        grouped[key].append(reading)
    
    all_stats: List[ShiftStats] = []
    
    for (zone_id, shift_date, shift_type), readings_list in grouped.items():
        stats = calculate_shift_stats(
            readings_list, zone_id, shift_date, shift_type, thresholds
        )
        all_stats.append(stats)
    
    return all_stats


def calculate_daily_report(
    shift_stats_list: List[ShiftStats],
    zone_id: int,
    report_date: str,
    alarm_count: int = 0,
) -> DailyReport:
    report = DailyReport(
        zone_id=zone_id,
        report_date=report_date,
    )
    
    if not shift_stats_list:
        return report
    
    reading_counts = [s.reading_count for s in shift_stats_list]
    total_readings = sum(reading_counts)
    
    def weighted_avg(values: List[Optional[float]], weights: List[int]) -> Optional[float]:
        weighted_sum = 0.0
        total_weight = 0
        for v, w in zip(values, weights):
            if v is not None:
                weighted_sum += v * w
                total_weight += w
        if total_weight == 0:
            return None
        return round(weighted_sum / total_weight, 2)
    
    def max_or_none(values: List[Optional[float]]) -> Optional[float]:
        valid = [v for v in values if v is not None]
        if not valid:
            return None
        return round(max(valid), 2)
    
    methane_avgs = [s.methane_avg for s in shift_stats_list]
    co_avgs = [s.co_avg for s in shift_stats_list]
    wind_speed_avgs = [s.wind_speed_avg for s in shift_stats_list]
    temperature_avgs = [s.temperature_avg for s in shift_stats_list]
    dust_avgs = [s.dust_avg for s in shift_stats_list]
    
    methane_maxes = [s.methane_max for s in shift_stats_list]
    co_maxes = [s.co_max for s in shift_stats_list]
    wind_speed_maxes = [s.wind_speed_max for s in shift_stats_list]
    temperature_maxes = [s.temperature_max for s in shift_stats_list]
    dust_maxes = [s.dust_max for s in shift_stats_list]
    
    report.methane_avg = weighted_avg(methane_avgs, reading_counts)
    report.methane_max = max_or_none(methane_maxes)
    report.methane_over_count = sum(s.methane_over_count for s in shift_stats_list)
    
    report.co_avg = weighted_avg(co_avgs, reading_counts)
    report.co_max = max_or_none(co_maxes)
    report.co_over_count = sum(s.co_over_count for s in shift_stats_list)
    
    report.wind_speed_avg = weighted_avg(wind_speed_avgs, reading_counts)
    report.wind_speed_max = max_or_none(wind_speed_maxes)
    report.wind_speed_over_count = sum(s.wind_speed_over_count for s in shift_stats_list)
    
    report.temperature_avg = weighted_avg(temperature_avgs, reading_counts)
    report.temperature_max = max_or_none(temperature_maxes)
    report.temperature_over_count = sum(s.temperature_over_count for s in shift_stats_list)
    
    report.dust_avg = weighted_avg(dust_avgs, reading_counts)
    report.dust_max = max_or_none(dust_maxes)
    report.dust_over_count = sum(s.dust_over_count for s in shift_stats_list)
    
    report.reading_count = total_readings
    report.alarm_count = alarm_count
    
    return report


def aggregate_shift_stats_by_day(
    shift_stats_list: List[ShiftStats],
    alarm_counts_by_zone_date: Dict[Tuple[int, str], int],
) -> List[DailyReport]:
    grouped: Dict[Tuple[int, str], List[ShiftStats]] = defaultdict(list)
    
    for stats in shift_stats_list:
        key = (stats.zone_id, stats.shift_date)
        grouped[key].append(stats)
    
    all_reports: List[DailyReport] = []
    
    for (zone_id, report_date), stats_list in grouped.items():
        alarm_count = alarm_counts_by_zone_date.get((zone_id, report_date), 0)
        report = calculate_daily_report(stats_list, zone_id, report_date, alarm_count)
        all_reports.append(report)
    
    return all_reports
