import io
from datetime import datetime
from typing import List, Iterator, Optional

from .models import SensorReading, MetricType
from .shifts import get_shift_display_name
from .alarms import get_metric_display_name, get_metric_unit


CSV_BATCH_SIZE = 100000
TIMESTAMP_FORMAT = "%Y-%m-%d %H:%M:%S"


def format_timestamp(dt: datetime) -> str:
    return dt.strftime(TIMESTAMP_FORMAT)


def format_value(value: Optional[float]) -> str:
    if value is None:
        return ""
    return f"{value:.2f}"


def get_reading_csv_header() -> List[str]:
    return [
        "区域ID",
        "时间",
        "班次",
        "瓦斯浓度(%)",
        "一氧化碳(ppm)",
        "风速(m/s)",
        "温度(°C)",
        "粉尘(mg/m³)",
    ]


def reading_to_row(reading: SensorReading) -> List[str]:
    shift_name = get_shift_display_name(reading.shift_type) if reading.shift_type else ""
    return [
        str(reading.zone_id),
        format_timestamp(reading.timestamp),
        shift_name,
        format_value(reading.methane),
        format_value(reading.co),
        format_value(reading.wind_speed),
        format_value(reading.temperature),
        format_value(reading.dust),
    ]


def batch_readings(readings: List[SensorReading], batch_size: int = CSV_BATCH_SIZE) -> Iterator[List[SensorReading]]:
    for i in range(0, len(readings), batch_size):
        yield readings[i:i + batch_size]


def _write_csv_row(writer, row: List[str]) -> None:
    writer.writerow(row)


def export_readings_to_csv(readings: List[SensorReading]) -> bytes:
    import csv
    
    output = io.StringIO()
    writer = csv.writer(output)
    writer.writerow(get_reading_csv_header())
    
    for batch in batch_readings(readings):
        for reading in batch:
            writer.writerow(reading_to_row(reading))
    
    content = output.getvalue()
    bom = '\ufeff'
    return (bom + content).encode('utf-8')


def export_readings_to_csv_stream(
    readings: List[SensorReading],
    batch_size: int = CSV_BATCH_SIZE
) -> Iterator[bytes]:
    import csv
    
    bom = b'\xef\xbb\xbf'
    yield bom
    
    header_row = ','.join(get_reading_csv_header()) + '\r\n'
    yield header_row.encode('utf-8')
    
    for batch in batch_readings(readings, batch_size):
        output = io.StringIO()
        writer = csv.writer(output)
        for reading in batch:
            writer.writerow(reading_to_row(reading))
        yield output.getvalue().encode('utf-8')


def get_stats_csv_header() -> List[str]:
    return [
        "区域ID",
        "统计日期",
        "班次",
        "瓦斯浓度平均(%)",
        "瓦斯浓度最大(%)",
        "瓦斯超限次数",
        "一氧化碳平均(ppm)",
        "一氧化碳最大(ppm)",
        "一氧化碳超限次数",
        "风速平均(m/s)",
        "风速最大(m/s)",
        "风速超限次数",
        "温度平均(°C)",
        "温度最大(°C)",
        "温度超限次数",
        "粉尘平均(mg/m³)",
        "粉尘最大(mg/m³)",
        "粉尘超限次数",
        "数据条数",
    ]


def shift_stats_to_row(stats) -> List[str]:
    from .models import ShiftStats
    return [
        str(stats.zone_id),
        stats.shift_date,
        get_shift_display_name(stats.shift_type),
        format_value(stats.methane_avg),
        format_value(stats.methane_max),
        str(stats.methane_over_count),
        format_value(stats.co_avg),
        format_value(stats.co_max),
        str(stats.co_over_count),
        format_value(stats.wind_speed_avg),
        format_value(stats.wind_speed_max),
        str(stats.wind_speed_over_count),
        format_value(stats.temperature_avg),
        format_value(stats.temperature_max),
        str(stats.temperature_over_count),
        format_value(stats.dust_avg),
        format_value(stats.dust_max),
        str(stats.dust_over_count),
        str(stats.reading_count),
    ]


def export_shift_stats_to_csv(stats_list) -> bytes:
    import csv
    from .models import ShiftStats
    
    output = io.StringIO()
    writer = csv.writer(output)
    writer.writerow(get_stats_csv_header())
    
    for stats in stats_list:
        writer.writerow(shift_stats_to_row(stats))
    
    content = output.getvalue()
    bom = '\ufeff'
    return (bom + content).encode('utf-8')


def get_daily_report_csv_header() -> List[str]:
    return [
        "区域ID",
        "报表日期",
        "瓦斯浓度平均(%)",
        "瓦斯浓度最大(%)",
        "瓦斯超限次数",
        "一氧化碳平均(ppm)",
        "一氧化碳最大(ppm)",
        "一氧化碳超限次数",
        "风速平均(m/s)",
        "风速最大(m/s)",
        "风速超限次数",
        "温度平均(°C)",
        "温度最大(°C)",
        "温度超限次数",
        "粉尘平均(mg/m³)",
        "粉尘最大(mg/m³)",
        "粉尘超限次数",
        "报警次数",
        "数据条数",
    ]


def daily_report_to_row(report) -> List[str]:
    from .models import DailyReport
    return [
        str(report.zone_id),
        report.report_date,
        format_value(report.methane_avg),
        format_value(report.methane_max),
        str(report.methane_over_count),
        format_value(report.co_avg),
        format_value(report.co_max),
        str(report.co_over_count),
        format_value(report.wind_speed_avg),
        format_value(report.wind_speed_max),
        str(report.wind_speed_over_count),
        format_value(report.temperature_avg),
        format_value(report.temperature_max),
        str(report.temperature_over_count),
        format_value(report.dust_avg),
        format_value(report.dust_max),
        str(report.dust_over_count),
        str(report.alarm_count),
        str(report.reading_count),
    ]


def export_daily_reports_to_csv(reports) -> bytes:
    import csv
    from .models import DailyReport
    
    output = io.StringIO()
    writer = csv.writer(output)
    writer.writerow(get_daily_report_csv_header())
    
    for report in reports:
        writer.writerow(daily_report_to_row(report))
    
    content = output.getvalue()
    bom = '\ufeff'
    return (bom + content).encode('utf-8')


def get_alarm_csv_header() -> List[str]:
    return [
        "区域ID",
        "指标类型",
        "报警级别",
        "报警值",
        "阈值",
        "开始时间",
        "结束时间",
        "是否已确认",
        "确认时间",
        "确认人",
    ]


def alarm_event_to_row(alarm) -> List[str]:
    from .models import AlarmEvent
    from .alarms import get_alarm_level_display
    
    return [
        str(alarm.zone_id),
        get_metric_display_name(alarm.metric_type),
        get_alarm_level_display(alarm.alarm_level),
        format_value(alarm.value),
        format_value(alarm.threshold),
        format_timestamp(alarm.start_time),
        format_timestamp(alarm.end_time) if alarm.end_time else "",
        "是" if alarm.acknowledged else "否",
        format_timestamp(alarm.acknowledged_at) if alarm.acknowledged_at else "",
        alarm.acknowledged_by or "",
    ]


def export_alarms_to_csv(alarms) -> bytes:
    import csv
    from .models import AlarmEvent
    
    output = io.StringIO()
    writer = csv.writer(output)
    writer.writerow(get_alarm_csv_header())
    
    for alarm in alarms:
        writer.writerow(alarm_event_to_row(alarm))
    
    content = output.getvalue()
    bom = '\ufeff'
    return (bom + content).encode('utf-8')
