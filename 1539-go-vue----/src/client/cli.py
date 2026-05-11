import click
import json
from datetime import datetime, date
from typing import Optional
from .api_client import APIClient

client = APIClient()


def print_json(data):
    click.echo(json.dumps(data, ensure_ascii=False, indent=2))


@click.group()
@click.option('--api-url', envvar='API_BASE_URL', default='http://localhost:8000', help='API 基础 URL')
def cli(api_url):
    """噪声监测点位管理系统命令行客户端"""
    global client
    client = APIClient(base_url=api_url)


# Zone limits commands
@cli.group()
def zone_limits():
    """区域限值管理"""
    pass


@zone_limits.command('create')
@click.argument('zone_type', type=click.Choice(['residential', 'commercial', 'industrial', 'roadside', 'mixed']))
@click.argument('daytime_limit', type=float)
@click.argument('nighttime_limit', type=float)
def create_zone_limit_cmd(zone_type, daytime_limit, nighttime_limit):
    """创建区域限值"""
    result = client.create_zone_limit(zone_type, daytime_limit, nighttime_limit)
    print_json(result)


@zone_limits.command('list')
def list_zone_limits():
    """列出所有区域限值"""
    result = client.get_zone_limits()
    print_json(result)


@zone_limits.command('get')
@click.argument('zone_limit_id', type=int)
def get_zone_limit_cmd(zone_limit_id):
    """获取指定区域限值"""
    result = client.get_zone_limit(zone_limit_id)
    print_json(result)


@zone_limits.command('update')
@click.argument('zone_limit_id', type=int)
@click.option('--daytime', type=float, help='昼间限值')
@click.option('--nighttime', type=float, help='夜间限值')
def update_zone_limit_cmd(zone_limit_id, daytime, nighttime):
    """更新区域限值"""
    kwargs = {}
    if daytime is not None:
        kwargs['daytime_limit'] = daytime
    if nighttime is not None:
        kwargs['nighttime_limit'] = nighttime
    
    if not kwargs:
        click.echo("请至少提供一个要更新的参数")
        return
    
    result = client.update_zone_limit(zone_limit_id, **kwargs)
    print_json(result)


@zone_limits.command('delete')
@click.argument('zone_limit_id', type=int)
def delete_zone_limit_cmd(zone_limit_id):
    """删除区域限值"""
    client.delete_zone_limit(zone_limit_id)
    click.echo(f"已删除区域限值 {zone_limit_id}")


# Monitoring points commands
@cli.group()
def points():
    """监测点位管理"""
    pass


@points.command('create')
@click.argument('name')
@click.argument('code')
@click.argument('zone_type', type=click.Choice(['residential', 'commercial', 'industrial', 'roadside', 'mixed']))
@click.argument('last_calibration', type=str)
@click.argument('next_calibration', type=str)
@click.option('--address', help='地址')
@click.option('--lat', type=float, help='纬度')
@click.option('--lng', type=float, help='经度')
@click.option('--active/--inactive', default=True, help='是否启用')
def create_point_cmd(name, code, zone_type, last_calibration, next_calibration, address, lat, lng, active):
    """创建监测点位"""
    result = client.create_point(
        name=name,
        code=code,
        zone_type=zone_type,
        last_calibration_date=last_calibration,
        next_calibration_date=next_calibration,
        address=address,
        latitude=lat,
        longitude=lng,
        is_active=active
    )
    print_json(result)


@points.command('list')
def list_points():
    """列出所有监测点位"""
    result = client.get_points()
    print_json(result)


@points.command('get')
@click.argument('point_id', type=int)
def get_point_cmd(point_id):
    """获取指定监测点位"""
    result = client.get_point(point_id)
    print_json(result)


@points.command('delete')
@click.argument('point_id', type=int)
def delete_point_cmd(point_id):
    """删除监测点位"""
    client.delete_point(point_id)
    click.echo(f"已删除监测点位 {point_id}")


# Noise data commands
@cli.group()
def noise():
    """噪声数据管理"""
    pass


@noise.command('create')
@click.argument('point_id', type=int)
@click.argument('value', type=float)
@click.option('--timestamp', help='时间戳 (ISO 格式)')
def create_noise_data_cmd(point_id, value, timestamp):
    """创建噪声数据"""
    ts = datetime.fromisoformat(timestamp) if timestamp else None
    result = client.create_noise_data(point_id, value, ts)
    print_json(result)


@noise.command('list')
@click.option('--point-id', type=int, help='点位 ID')
@click.option('--start-time', help='开始时间 (ISO 格式)')
@click.option('--end-time', help='结束时间 (ISO 格式)')
def list_noise_data(point_id, start_time, end_time):
    """列出噪声数据"""
    start = datetime.fromisoformat(start_time) if start_time else None
    end = datetime.fromisoformat(end_time) if end_time else None
    result = client.get_noise_data(point_id, start, end)
    print_json(result)


@noise.command('export-csv')
@click.argument('output_file')
@click.option('--point-id', type=int, help='点位 ID')
@click.option('--start-time', help='开始时间 (ISO 格式)')
@click.option('--end-time', help='结束时间 (ISO 格式)')
def export_noise_csv(output_file, point_id, start_time, end_time):
    """导出噪声数据到 CSV"""
    start = datetime.fromisoformat(start_time) if start_time else None
    end = datetime.fromisoformat(end_time) if end_time else None
    csv_data = client.export_noise_data_csv(point_id, start, end)
    
    with open(output_file, 'wb') as f:
        f.write(csv_data)
    
    click.echo(f"数据已导出到 {output_file}")


# Alerts commands
@cli.group()
def alerts():
    """预警事件管理"""
    pass


@alerts.command('check-calibration')
def check_calibration_cmd():
    """检查设备检定状态"""
    result = client.check_calibration()
    print_json(result)


@alerts.command('list')
@click.option('--point-id', type=int, help='点位 ID')
@click.option('--type', 'alert_type', type=click.Choice(['exceed', 'calibration', 'device_fault']), help='预警类型')
@click.option('--resolved/--unresolved', default=None, help='是否已解决')
def list_alerts(point_id, alert_type, resolved):
    """列出预警事件"""
    result = client.get_alerts(point_id, alert_type, resolved)
    print_json(result)


@alerts.command('resolve')
@click.argument('alert_id', type=int)
def resolve_alert_cmd(alert_id):
    """解决预警事件"""
    result = client.resolve_alert(alert_id)
    print_json(result)


# Construction permits commands
@cli.group()
def permits():
    """施工许可管理"""
    pass


@permits.command('create')
@click.argument('permit_number')
@click.argument('start_date')
@click.argument('end_date')
@click.option('--point-ids', required=True, help='点位 ID 列表，逗号分隔')
@click.option('--description', help='描述')
@click.option('--active/--inactive', default=True, help='是否启用')
def create_permit_cmd(permit_number, start_date, end_date, point_ids, description, active):
    """创建施工许可"""
    ids = [int(x.strip()) for x in point_ids.split(',')]
    result = client.create_permit(
        point_ids=ids,
        permit_number=permit_number,
        start_date=start_date,
        end_date=end_date,
        description=description,
        is_active=active
    )
    print_json(result)


@permits.command('list')
def list_permits():
    """列出所有施工许可"""
    result = client.get_permits()
    print_json(result)


# Statistics commands
@cli.group()
def stats():
    """统计数据管理"""
    pass


@stats.command('compute-daily')
@click.argument('point_id', type=int)
@click.argument('target_date')
def compute_daily_stats(point_id, target_date):
    """计算日均统计"""
    d = date.fromisoformat(target_date)
    result = client.compute_daily_statistics(point_id, d)
    print_json(result)


@stats.command('list-daily')
@click.option('--point-id', type=int, help='点位 ID')
@click.option('--start-date', help='开始日期')
@click.option('--end-date', help='结束日期')
def list_daily_stats(point_id, start_date, end_date):
    """列出日均统计"""
    start = date.fromisoformat(start_date) if start_date else None
    end = date.fromisoformat(end_date) if end_date else None
    result = client.get_daily_statistics(point_id, start, end)
    print_json(result)


@stats.command('export-csv')
@click.argument('output_file')
@click.option('--point-id', type=int, help='点位 ID')
@click.option('--start-date', help='开始日期')
@click.option('--end-date', help='结束日期')
def export_stats_csv(output_file, point_id, start_date, end_date):
    """导出统计数据到 CSV"""
    start = date.fromisoformat(start_date) if start_date else None
    end = date.fromisoformat(end_date) if end_date else None
    csv_data = client.export_statistics_csv(point_id, start, end)
    
    with open(output_file, 'wb') as f:
        f.write(csv_data)
    
    click.echo(f"数据已导出到 {output_file}")


if __name__ == '__main__':
    cli()
