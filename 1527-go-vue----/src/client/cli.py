import os
import json
from datetime import datetime
from typing import Optional

import click

from client.api_client import MonitorAPIClient


def get_client(ctx: click.Context) -> MonitorAPIClient:
    return ctx.obj["client"]


@click.group()
@click.option(
    "--server",
    default=None,
    envvar="MONITOR_API_URL",
    help="API 服务器地址 (默认: http://localhost:8000)",
)
@click.option(
    "--timeout",
    type=int,
    default=30,
    help="请求超时时间 (秒)",
)
@click.pass_context
def cli(ctx: click.Context, server: Optional[str], timeout: int):
    """煤矿通风瓦斯监测管理系统命令行客户端"""
    ctx.ensure_object(dict)
    ctx.obj["client"] = MonitorAPIClient(base_url=server, timeout=timeout)


@cli.group()
def mine():
    """矿井管理"""
    pass


@mine.command("list")
@click.pass_context
def list_mines(ctx: click.Context):
    """列出所有矿井"""
    client = get_client(ctx)
    mines = client.list_mines()
    click.echo(json.dumps(mines, indent=2, ensure_ascii=False))


@mine.command("create")
@click.option("--name", required=True, help="矿井名称")
@click.option("--description", help="矿井描述")
@click.pass_context
def create_mine(ctx: click.Context, name: str, description: Optional[str]):
    """创建矿井"""
    client = get_client(ctx)
    result = client.create_mine(name=name, description=description)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@mine.command("delete")
@click.argument("mine_id", type=int)
@click.pass_context
def delete_mine(ctx: click.Context, mine_id: int):
    """删除矿井"""
    client = get_client(ctx)
    result = client.delete_mine(mine_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@cli.group()
def zone():
    """监测区域管理"""
    pass


@zone.command("list")
@click.option("--mine-id", type=int, help="按矿井ID筛选")
@click.pass_context
def list_zones(ctx: click.Context, mine_id: Optional[int]):
    """列出监测区域"""
    client = get_client(ctx)
    zones = client.list_zones(mine_id=mine_id)
    click.echo(json.dumps(zones, indent=2, ensure_ascii=False))


@zone.command("create")
@click.option("--mine-id", type=int, required=True, help="所属矿井ID")
@click.option("--name", required=True, help="区域名称")
@click.option("--description", help="区域描述")
@click.pass_context
def create_zone(ctx: click.Context, mine_id: int, name: str, description: Optional[str]):
    """创建监测区域"""
    client = get_client(ctx)
    result = client.create_zone(mine_id=mine_id, name=name, description=description)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@zone.command("delete")
@click.argument("zone_id", type=int)
@click.pass_context
def delete_zone(ctx: click.Context, zone_id: int):
    """删除监测区域"""
    client = get_client(ctx)
    result = client.delete_zone(zone_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@cli.group()
def threshold():
    """报警阈值管理"""
    pass


@threshold.command("list")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.pass_context
def list_thresholds(ctx: click.Context, zone_id: Optional[int]):
    """列出报警阈值配置"""
    client = get_client(ctx)
    thresholds = client.list_thresholds(zone_id=zone_id)
    click.echo(json.dumps(thresholds, indent=2, ensure_ascii=False))


@threshold.command("create")
@click.option("--zone-id", type=int, required=True, help="区域ID")
@click.option(
    "--metric",
    required=True,
    type=click.Choice(["methane", "co", "wind_speed", "temperature", "dust"]),
    help="指标类型: methane(瓦斯), co(一氧化碳), wind_speed(风速), temperature(温度), dust(粉尘)",
)
@click.option("--level1", type=float, required=True, help="一级报警阈值")
@click.option("--level2", type=float, required=True, help="二级报警阈值")
@click.option("--level3", type=float, required=True, help="三级报警阈值")
@click.pass_context
def create_threshold(
    ctx: click.Context,
    zone_id: int,
    metric: str,
    level1: float,
    level2: float,
    level3: float,
):
    """创建报警阈值配置"""
    client = get_client(ctx)
    result = client.create_threshold(
        zone_id=zone_id,
        metric_type=metric,
        level1=level1,
        level2=level2,
        level3=level3,
    )
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@cli.group()
def reading():
    """传感器数据管理"""
    pass


@reading.command("submit")
@click.option("--zone-id", type=int, required=True, help="区域ID")
@click.option("--methane", type=float, required=True, help="瓦斯浓度 (%)")
@click.option("--co", type=float, required=True, help="一氧化碳 (ppm)")
@click.option("--wind-speed", type=float, required=True, help="风速 (m/s)")
@click.option("--temperature", type=float, required=True, help="温度 (°C)")
@click.option("--dust", type=float, required=True, help="粉尘 (mg/m³)")
@click.option(
    "--timestamp",
    default=None,
    help="时间戳 (格式: YYYY-MM-DDTHH:MM:SS, 默认当前时间)",
)
@click.pass_context
def submit_reading(
    ctx: click.Context,
    zone_id: int,
    methane: float,
    co: float,
    wind_speed: float,
    temperature: float,
    dust: float,
    timestamp: Optional[str],
):
    """上报传感器数据"""
    client = get_client(ctx)
    
    ts = None
    if timestamp:
        ts = datetime.fromisoformat(timestamp)
    
    result = client.submit_reading(
        zone_id=zone_id,
        methane=methane,
        co=co,
        wind_speed=wind_speed,
        temperature=temperature,
        dust=dust,
        timestamp=ts,
    )
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@reading.command("list")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--start-time", help="开始时间 (格式: YYYY-MM-DDTHH:MM:SS)")
@click.option("--end-time", help="结束时间 (格式: YYYY-MM-DDTHH:MM:SS)")
@click.option("--limit", type=int, default=100, help="返回条数限制")
@click.pass_context
def list_readings(
    ctx: click.Context,
    zone_id: Optional[int],
    start_time: Optional[str],
    end_time: Optional[str],
    limit: int,
):
    """查询传感器数据"""
    client = get_client(ctx)
    
    start = datetime.fromisoformat(start_time) if start_time else None
    end = datetime.fromisoformat(end_time) if end_time else None
    
    readings = client.list_readings(
        zone_id=zone_id,
        start_time=start,
        end_time=end,
        limit=limit,
    )
    click.echo(json.dumps(readings, indent=2, ensure_ascii=False))


@reading.command("export")
@click.option("--output", "-o", required=True, help="输出CSV文件路径")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--start-time", help="开始时间 (格式: YYYY-MM-DDTHH:MM:SS)")
@click.option("--end-time", help="结束时间 (格式: YYYY-MM-DDTHH:MM:SS)")
@click.pass_context
def export_readings(
    ctx: click.Context,
    output: str,
    zone_id: Optional[int],
    start_time: Optional[str],
    end_time: Optional[str],
):
    """导出传感器数据到CSV"""
    client = get_client(ctx)
    
    start = datetime.fromisoformat(start_time) if start_time else None
    end = datetime.fromisoformat(end_time) if end_time else None
    
    client.export_readings(
        output_path=output,
        zone_id=zone_id,
        start_time=start,
        end_time=end,
    )
    click.echo(f"数据已导出到: {output}")


@cli.group()
def alarm():
    """报警事件管理"""
    pass


@alarm.command("list")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--start-time", help="开始时间")
@click.option("--end-time", help="结束时间")
@click.option("--acknowledged/--unacknowledged", default=None, help="是否已确认")
@click.option("--limit", type=int, default=100, help="返回条数限制")
@click.pass_context
def list_alarms(
    ctx: click.Context,
    zone_id: Optional[int],
    start_time: Optional[str],
    end_time: Optional[str],
    acknowledged: Optional[bool],
    limit: int,
):
    """查询报警事件"""
    client = get_client(ctx)
    
    start = datetime.fromisoformat(start_time) if start_time else None
    end = datetime.fromisoformat(end_time) if end_time else None
    
    alarms = client.list_alarms(
        zone_id=zone_id,
        start_time=start,
        end_time=end,
        acknowledged=acknowledged,
        limit=limit,
    )
    click.echo(json.dumps(alarms, indent=2, ensure_ascii=False))


@alarm.command("acknowledge")
@click.argument("alarm_id", type=int)
@click.option("--by", help="确认人")
@click.pass_context
def acknowledge_alarm(ctx: click.Context, alarm_id: int, by: Optional[str]):
    """确认报警"""
    client = get_client(ctx)
    result = client.acknowledge_alarm(alarm_id, acknowledged_by=by)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@alarm.command("export")
@click.option("--output", "-o", required=True, help="输出CSV文件路径")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--start-time", help="开始时间")
@click.option("--end-time", help="结束时间")
@click.pass_context
def export_alarms(
    ctx: click.Context,
    output: str,
    zone_id: Optional[int],
    start_time: Optional[str],
    end_time: Optional[str],
):
    """导出报警事件到CSV"""
    client = get_client(ctx)
    
    start = datetime.fromisoformat(start_time) if start_time else None
    end = datetime.fromisoformat(end_time) if end_time else None
    
    client.export_alarms(
        output_path=output,
        zone_id=zone_id,
        start_time=start,
        end_time=end,
    )
    click.echo(f"报警数据已导出到: {output}")


@cli.group()
def stats():
    """统计数据管理"""
    pass


@stats.command("generate-shift")
@click.option("--zone-id", type=int, help="指定区域ID (可选)")
@click.pass_context
def generate_shift(ctx: click.Context, zone_id: Optional[int]):
    """生成班次统计"""
    client = get_client(ctx)
    result = client.generate_shift_stats(zone_id=zone_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@stats.command("generate-daily")
@click.option("--zone-id", type=int, help="指定区域ID (可选)")
@click.pass_context
def generate_daily(ctx: click.Context, zone_id: Optional[int]):
    """生成日报"""
    client = get_client(ctx)
    result = client.generate_daily_reports(zone_id=zone_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@stats.command("list-shift")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--shift-date", help="班次日期 (格式: YYYY-MM-DD)")
@click.option("--limit", type=int, default=100, help="返回条数限制")
@click.pass_context
def list_shift_stats(
    ctx: click.Context,
    zone_id: Optional[int],
    shift_date: Optional[str],
    limit: int,
):
    """查询班次统计"""
    client = get_client(ctx)
    stats = client.list_shift_stats(
        zone_id=zone_id,
        shift_date=shift_date,
        limit=limit,
    )
    click.echo(json.dumps(stats, indent=2, ensure_ascii=False))


@stats.command("export-shift")
@click.option("--output", "-o", required=True, help="输出CSV文件路径")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--start-date", help="开始日期 (格式: YYYY-MM-DD)")
@click.option("--end-date", help="结束日期 (格式: YYYY-MM-DD)")
@click.pass_context
def export_shift_stats(
    ctx: click.Context,
    output: str,
    zone_id: Optional[int],
    start_date: Optional[str],
    end_date: Optional[str],
):
    """导出版次统计到CSV"""
    client = get_client(ctx)
    client.export_shift_stats(
        output_path=output,
        zone_id=zone_id,
        start_date=start_date,
        end_date=end_date,
    )
    click.echo(f"班次统计已导出到: {output}")


@stats.command("list-daily")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--report-date", help="报表日期 (格式: YYYY-MM-DD)")
@click.option("--start-date", help="开始日期")
@click.option("--end-date", help="结束日期")
@click.option("--limit", type=int, default=100, help="返回条数限制")
@click.pass_context
def list_daily_reports(
    ctx: click.Context,
    zone_id: Optional[int],
    report_date: Optional[str],
    start_date: Optional[str],
    end_date: Optional[str],
    limit: int,
):
    """查询日报"""
    client = get_client(ctx)
    reports = client.list_daily_reports(
        zone_id=zone_id,
        report_date=report_date,
        start_date=start_date,
        end_date=end_date,
        limit=limit,
    )
    click.echo(json.dumps(reports, indent=2, ensure_ascii=False))


@stats.command("export-daily")
@click.option("--output", "-o", required=True, help="输出CSV文件路径")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--start-date", help="开始日期 (格式: YYYY-MM-DD)")
@click.option("--end-date", help="结束日期 (格式: YYYY-MM-DD)")
@click.pass_context
def export_daily_reports(
    ctx: click.Context,
    output: str,
    zone_id: Optional[int],
    start_date: Optional[str],
    end_date: Optional[str],
):
    """导出日报到CSV"""
    client = get_client(ctx)
    client.export_daily_reports(
        output_path=output,
        zone_id=zone_id,
        start_date=start_date,
        end_date=end_date,
    )
    click.echo(f"日报已导出到: {output}")


def main():
    cli()


if __name__ == "__main__":
    main()
