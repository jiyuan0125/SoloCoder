from datetime import datetime, timedelta
from typing import Optional, List, Tuple

from .models import (
    AlarmEvent,
    AlarmLevel,
    MetricType,
    ThresholdConfig,
    SensorReading,
)


ALARM_DEDUPLICATION_INTERVAL = timedelta(minutes=10)


def get_alarm_level(value: float, threshold: ThresholdConfig) -> Optional[AlarmLevel]:
    if value >= threshold.level3_threshold:
        return AlarmLevel.LEVEL3
    elif value >= threshold.level2_threshold:
        return AlarmLevel.LEVEL2
    elif value >= threshold.level1_threshold:
        return AlarmLevel.LEVEL1
    return None


def get_metric_value(reading: SensorReading, metric_type: MetricType) -> float:
    values = {
        MetricType.METHANE: reading.methane,
        MetricType.CO: reading.co,
        MetricType.WIND_SPEED: reading.wind_speed,
        MetricType.TEMPERATURE: reading.temperature,
        MetricType.DUST: reading.dust,
    }
    return values[metric_type]


def get_threshold_value(threshold: ThresholdConfig, level: AlarmLevel) -> float:
    thresholds = {
        AlarmLevel.LEVEL1: threshold.level1_threshold,
        AlarmLevel.LEVEL2: threshold.level2_threshold,
        AlarmLevel.LEVEL3: threshold.level3_threshold,
    }
    return thresholds[level]


def should_create_new_alarm(
    existing_alarm: Optional[AlarmEvent],
    new_alarm_time: datetime,
    new_zone_id: int,
    new_metric_type: MetricType,
    new_alarm_level: AlarmLevel,
) -> bool:
    if existing_alarm is None:
        return True
    
    if (existing_alarm.zone_id != new_zone_id or
        existing_alarm.metric_type != new_metric_type):
        return True
    
    if existing_alarm.end_time is not None:
        time_since_end = new_alarm_time - existing_alarm.end_time
        if time_since_end > ALARM_DEDUPLICATION_INTERVAL:
            return True
        if new_alarm_level.value > existing_alarm.alarm_level.value:
            return True
        return False
    
    time_since_start = new_alarm_time - existing_alarm.start_time
    if time_since_start > ALARM_DEDUPLICATION_INTERVAL:
        return True
    if new_alarm_level.value > existing_alarm.alarm_level.value:
        return True
    
    return False


def check_alarms(
    reading: SensorReading,
    thresholds: List[ThresholdConfig],
    existing_alarms: List[AlarmEvent],
) -> Tuple[List[AlarmEvent], List[AlarmEvent]]:
    new_alarms: List[AlarmEvent] = []
    updated_alarms: List[AlarmEvent] = []
    
    for metric_type in MetricType:
        threshold = next(
            (t for t in thresholds if t.metric_type == metric_type),
            None
        )
        if threshold is None:
            continue
        
        value = get_metric_value(reading, metric_type)
        alarm_level = get_alarm_level(value, threshold)
        
        existing_alarm = next(
            (a for a in existing_alarms 
             if a.zone_id == reading.zone_id and 
                a.metric_type == metric_type and
                a.end_time is None),
            None
        )
        
        if alarm_level is not None:
            threshold_value = get_threshold_value(threshold, alarm_level)
            
            if existing_alarm is None:
                new_alarms.append(AlarmEvent(
                    zone_id=reading.zone_id,
                    metric_type=metric_type,
                    alarm_level=alarm_level,
                    value=value,
                    threshold=threshold_value,
                    start_time=reading.timestamp,
                ))
            else:
                if should_create_new_alarm(
                    existing_alarm, reading.timestamp,
                    reading.zone_id, metric_type, alarm_level
                ):
                    existing_alarm.end_time = reading.timestamp
                    updated_alarms.append(existing_alarm)
                    new_alarms.append(AlarmEvent(
                        zone_id=reading.zone_id,
                        metric_type=metric_type,
                        alarm_level=alarm_level,
                        value=value,
                        threshold=threshold_value,
                        start_time=reading.timestamp,
                    ))
                else:
                    existing_alarm.value = max(existing_alarm.value, value)
                    existing_alarm.alarm_level = max(existing_alarm.alarm_level, alarm_level)
                    if alarm_level.value > existing_alarm.alarm_level:
                        existing_alarm.threshold = threshold_value
                    updated_alarms.append(existing_alarm)
        else:
            if existing_alarm is not None:
                existing_alarm.end_time = reading.timestamp
                updated_alarms.append(existing_alarm)
    
    return new_alarms, updated_alarms


def is_over_threshold(value: float, threshold: ThresholdConfig) -> bool:
    return value >= threshold.level1_threshold


def get_alarm_level_display(level: AlarmLevel) -> str:
    displays = {
        AlarmLevel.LEVEL1: "一级报警",
        AlarmLevel.LEVEL2: "二级报警",
        AlarmLevel.LEVEL3: "三级报警",
    }
    return displays.get(level, "未知级别")


def get_metric_display_name(metric_type: MetricType) -> str:
    displays = {
        MetricType.METHANE: "瓦斯浓度",
        MetricType.CO: "一氧化碳",
        MetricType.WIND_SPEED: "风速",
        MetricType.TEMPERATURE: "温度",
        MetricType.DUST: "粉尘",
    }
    return displays.get(metric_type, "未知指标")


def get_metric_unit(metric_type: MetricType) -> str:
    units = {
        MetricType.METHANE: "%",
        MetricType.CO: "ppm",
        MetricType.WIND_SPEED: "m/s",
        MetricType.TEMPERATURE: "°C",
        MetricType.DUST: "mg/m³",
    }
    return units.get(metric_type, "")
