import json
from datetime import datetime
import click

from .client import HydrologyClient


def get_client():
    return HydrologyClient()


def output_json(data):
    click.echo(json.dumps(data, ensure_ascii=False, indent=2, default=str))


@click.group()
def cli():
    """水文数据采集管理系统命令行客户端"""
    pass


@cli.group()
def station():
    """监测站管理"""
    pass


@station.command("list")
def station_list():
    """列出所有监测站"""
    client = get_client()
    stations = client.list_stations()
    for s in stations:
        click.echo(f"ID: {s['id']}, 名称: {s['name']}, 类型: {s['station_type']}, 设备: {s.get('device_model', '未设置')}")


@station.command("create")
@click.option("--name", required=True, help="监测站名称")
@click.option("--type", "station_type", required=True, type=click.Choice(['water_level', 'rainfall', 'flow']), help="监测站类型")
@click.option("--device-model", help="设备型号")
def station_create(name, station_type, device_model):
    """创建监测站"""
    client = get_client()
    result = client.create_station(name, station_type, device_model)
    output_json(result)


@station.command("get")
@click.argument("station_id", type=int)
def station_get(station_id):
    """获取单个监测站信息"""
    client = get_client()
    result = client.get_station(station_id)
    output_json(result)


@station.command("update")
@click.argument("station_id", type=int)
@click.option("--name", help="新名称")
@click.option("--type", "station_type", type=click.Choice(['water_level', 'rainfall', 'flow']), help="新类型")
@click.option("--device-model", help="新设备型号")
def station_update(station_id, name, station_type, device_model):
    """更新监测站信息"""
    client = get_client()
    kwargs = {}
    if name:
        kwargs['name'] = name
    if station_type:
        kwargs['station_type'] = station_type
    if device_model:
        kwargs['device_model'] = device_model

    result = client.update_station(station_id, **kwargs)
    output_json(result)


@station.command("delete")
@click.argument("station_id", type=int)
def station_delete(station_id):
    """删除监测站"""
    client = get_client()
    client.delete_station(station_id)
    click.echo(f"监测站 {station_id} 已删除")


@cli.group()
def data():
    """原始数据管理"""
    pass


@data.command("list")
@click.argument("station_id", type=int)
@click.option("--limit", default=50, help="获取数量")
def data_list(station_id, limit):
    """列出原始数据"""
    client = get_client()
    result = client.get_raw_data(station_id, limit=limit)
    for d in result:
        click.echo(f"[{d['timestamp']}] 类型: {d['data_type']}, 值: {d['value']}, 质量: {d['quality_status']}")


@data.command("add")
@click.argument("station_id", type=int)
@click.option("--type", "data_type", required=True, type=click.Choice(['water_level', 'rainfall', 'flow']), help="数据类型")
@click.option("--value", required=True, type=float, help="数据值")
@click.option("--timestamp", help="时间戳 (ISO格式，如 2024-01-01T10:00:00)")
def data_add(station_id, data_type, value, timestamp):
    """添加原始数据"""
    client = get_client()
    ts = None
    if timestamp:
        ts = datetime.fromisoformat(timestamp)
    result = client.create_raw_data(station_id, data_type, value, ts)
    output_json(result)


@cli.group()
def aggregation():
    """聚合数据管理"""
    pass


@aggregation.command("list")
@click.argument("station_id", type=int)
@click.option("--granularity", type=click.Choice(['hourly', 'daily']), help="聚合粒度")
@click.option("--limit", default=50, help="获取数量")
def aggregation_list(station_id, granularity, limit):
    """列出聚合数据"""
    client = get_client()
    result = client.get_aggregated_data(station_id, granularity=granularity, limit=limit)
    for agg in result:
        click.echo(
            f"[{agg['granularity']}] {agg['period_start']} -> {agg['period_end']}: "
            f"均值={agg['average_value']:.2f}, 最大={agg['max_value']:.2f}, "
            f"最小={agg['min_value']}, 状态={agg['missing_status']}, 样本={agg['sample_count']}"
        )


@aggregation.command("recalculate")
@click.argument("station_id", type=int)
@click.option("--start", required=True, help="开始时间 (ISO格式)")
@click.option("--end", required=True, help="结束时间 (ISO格式)")
def aggregation_recalculate(station_id, start, end):
    """重新计算聚合数据"""
    client = get_client()
    start_dt = datetime.fromisoformat(start)
    end_dt = datetime.fromisoformat(end)
    result = client.recalculate_aggregations(station_id, start_dt, end_dt)
    click.echo(f"已计算 {len(result)} 个聚合结果")


@cli.group()
def quality():
    """质量记录管理"""
    pass


@quality.command("list")
@click.argument("station_id", type=int)
@click.option("--limit", default=50, help="获取数量")
def quality_list(station_id, limit):
    """列出质量记录"""
    client = get_client()
    result = client.get_quality_records(station_id, limit=limit)
    for q in result:
        status = "已审核" if q['reviewed'] else "待审核"
        click.echo(f"ID: {q['id']}, 类型: {q['issue_type']}, 状态: {status}")
        click.echo(f"  描述: {q['description']}")


@quality.command("review")
@click.argument("station_id", type=int)
@click.argument("record_id", type=int)
@click.option("--invalid/--valid", "invalid", required=True, help="标记为无效/恢复为有效")
@click.option("--reviewer", help="审核人")
@click.option("--notes", help="审核备注")
def quality_review(station_id, record_id, invalid, reviewer, notes):
    """审核质量记录"""
    client = get_client()
    result = client.review_quality_record(station_id, record_id, invalid, reviewer, notes)
    click.echo("审核完成")
    output_json(result)


@cli.group()
def calibration():
    """设备检定管理"""
    pass


@calibration.command("list")
@click.argument("station_id", type=int)
@click.option("--limit", default=50, help="获取数量")
def calibration_list(station_id, limit):
    """列出检定记录"""
    client = get_client()
    result = client.get_calibrations(station_id, limit=limit)
    for c in result:
        status = "已完成" if c['is_done'] else "待完成"
        click.echo(f"ID: {c['id']}, 上次: {c['last_calibration_date']}, 下次: {c['next_calibration_date']}, 状态: {status}")


@calibration.command("add")
@click.argument("station_id", type=int)
@click.option("--last-date", required=True, help="上次检定日期 (ISO格式)")
@click.option("--next-date", help="下次检定日期 (默认两年后)")
@click.option("--done/--pending", "is_done", default=False, help="是否已完成")
def calibration_add(station_id, last_date, next_date, is_done):
    """添加检定记录"""
    client = get_client()
    last_dt = datetime.fromisoformat(last_date)
    next_dt = None
    if next_date:
        next_dt = datetime.fromisoformat(next_date)
    result = client.create_calibration(station_id, last_dt, next_dt, is_done)
    output_json(result)


@calibration.command("todos")
def calibration_todos():
    """查看检定待办"""
    client = get_client()
    todos = client.get_calibration_todos()
    if not todos:
        click.echo("没有待办")
        return

    for todo in todos:
        click.echo(
            f"监测站 {todo['station_id']} ({todo['station_name']}): "
            f"到期 {todo['next_calibration_date']}, 剩余 {todo['days_remaining']} 天"
        )


@calibration.command("update")
@click.argument("station_id", type=int)
@click.argument("calibration_id", type=int)
@click.option("--is-done/--not-done", "is_done", help="标记为完成/未完成")
def calibration_update(station_id, calibration_id, is_done):
    """更新检定记录"""
    client = get_client()
    kwargs = {}
    if is_done is not None:
        kwargs['is_done'] = is_done

    result = client.update_calibration(station_id, calibration_id, **kwargs)
    output_json(result)


@cli.command()
def health():
    """检查服务健康状态"""
    try:
        client = get_client()
        result = client.health_check()
        click.echo(f"服务状态: {result['status']}")
    except Exception as e:
        click.echo(f"服务不可用: {e}")


if __name__ == "__main__":
    cli()
