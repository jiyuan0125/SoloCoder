import click
import requests
import json
from datetime import datetime

API_BASE = "http://localhost:8600/api/v1"

@click.group()
def cli():
    """高铁车站客运管理系统 - 命令行客户端"""
    pass

@cli.group()
def station():
    """车站管理"""
    pass

@station.command("list")
def list_stations():
    """列出所有车站"""
    try:
        response = requests.get(f"{API_BASE}/stations")
        response.raise_for_status()
        stations = response.json()
        
        click.echo("\n=== 车站列表 ===")
        for s in stations:
            click.echo(f"代码: {s['code']:10} 名称: {s['name']:20} 最大容量: {s['max_capacity']}")
    except Exception as e:
        click.echo(f"错误: {e}")

@station.command("create")
@click.option("--code", prompt="车站代码")
@click.option("--name", prompt="车站名称")
@click.option("--capacity", prompt="最大容量", default=5000, type=int)
def create_station(code, name, capacity):
    """创建新车站"""
    try:
        data = {"code": code, "name": name, "max_capacity": capacity}
        response = requests.post(f"{API_BASE}/stations", json=data)
        response.raise_for_status()
        click.echo(f"成功创建车站: {code}")
    except Exception as e:
        click.echo(f"错误: {e}")

@cli.group()
def zone():
    """区域管理"""
    pass

@zone.command("list")
@click.option("--station", prompt="车站代码")
def list_zones(station):
    """列出车站的所有区域"""
    try:
        response = requests.get(f"{API_BASE}/stations/{station}/zones")
        response.raise_for_status()
        zones = response.json()
        
        click.echo(f"\n=== 车站 {station} 的区域列表 ===")
        for z in zones:
            click.echo(f"ID: {z['id']:5} 名称: {z['name']:15} 类型: {z['zone_type']:10} 当前人数: {z['current_count']}/{z['max_capacity']}")
    except Exception as e:
        click.echo(f"错误: {e}")

@zone.command("create")
@click.option("--station", prompt="车站代码")
@click.option("--name", prompt="区域名称")
@click.option("--type", prompt="区域类型")
@click.option("--capacity", prompt="最大容量", default=1000, type=int)
def create_zone(station, name, type, capacity):
    """创建新区域"""
    try:
        data = {"name": name, "zone_type": type, "max_capacity": capacity}
        response = requests.post(f"{API_BASE}/stations/{station}/zones", json=data)
        response.raise_for_status()
        click.echo(f"成功创建区域: {name}")
    except Exception as e:
        click.echo(f"错误: {e}")

@cli.group()
def security():
    """安检通道管理"""
    pass

@security.command("status")
@click.option("--station", prompt="车站代码")
def security_status(station):
    """查看安检通道状态"""
    try:
        response = requests.get(f"{API_BASE}/stations/{station}/security/status")
        response.raise_for_status()
        gates = response.json()
        
        click.echo(f"\n=== 车站 {station} 安检通道状态 ===")
        for g in gates:
            status = "开放" if g['is_open'] else "关闭"
            fault = "故障" if g['is_faulty'] else "正常"
            click.echo(f"通道: {g['gate_number']:10} 状态: {status:8} 故障: {fault:8} 排队: {g['queue_length']}")
    except Exception as e:
        click.echo(f"错误: {e}")

@security.command("open")
@click.option("--station", prompt="车站代码")
@click.option("--gate", prompt="通道号")
def open_gate(station, gate):
    """开启安检通道"""
    try:
        response = requests.post(f"{API_BASE}/stations/{station}/security/{gate}/open")
        response.raise_for_status()
        click.echo(f"安检通道 {gate} 已开启")
    except Exception as e:
        click.echo(f"错误: {e}")

@security.command("close")
@click.option("--station", prompt="车站代码")
@click.option("--gate", prompt="通道号")
def close_gate(station, gate):
    """关闭安检通道"""
    try:
        response = requests.post(f"{API_BASE}/stations/{station}/security/{gate}/close")
        response.raise_for_status()
        click.echo(f"安检通道 {gate} 已关闭")
    except Exception as e:
        click.echo(f"错误: {e}")

@security.command("suggestions")
@click.option("--station", prompt="车站代码")
def security_suggestions(station):
    """获取安检通道开关建议"""
    try:
        response = requests.get(f"{API_BASE}/stations/{station}/security/suggestions")
        response.raise_for_status()
        data = response.json()
        
        click.echo(f"\n=== 车站 {station} 安检通道建议 ===")
        for s in data['suggestions']:
            action_map = {
                "OPEN_MORE": "建议增加开放",
                "CLOSE_SOME": "建议减少开放",
                "MAINTAIN": "维持现状"
            }
            click.echo(f"\n区域: {s['zone_name']}")
            click.echo(f"  开放/总数: {s['open_gates']}/{s['total_gates']}  故障: {s['faulty_gates']}")
            click.echo(f"  平均排队: {s['avg_queue_per_gate']}  建议: {action_map.get(s['action'], s['action'])}")
    except Exception as e:
        click.echo(f"错误: {e}")

@cli.group()
def flow():
    """客流管理"""
    pass

@flow.command("realtime")
@click.option("--station", prompt="车站代码")
def realtime_flow(station):
    """查看实时客流"""
    try:
        response = requests.get(f"{API_BASE}/stations/{station}/passenger-flow/realtime")
        response.raise_for_status()
        data = response.json()
        
        click.echo(f"\n=== 车站 {station} 实时客流 ===")
        click.echo(f"车站: {data['station_name']}")
        click.echo(f"总人数: {data['total_count']} / {data['total_capacity']}")
        click.echo(f"整体拥挤度: {data['overall_occupancy_rate']}%")
        
        click.echo("\n各区域详情:")
        for z in data['zones']:
            click.echo(f"  {z['zone_name']:15} {z['current_count']:6}/{z['max_capacity']:6}  ({z['occupancy_rate']}%)")
    except Exception as e:
        click.echo(f"错误: {e}")

@flow.command("hourly")
@click.option("--station", prompt="车站代码")
@click.option("--date", default=None, help="日期 (YYYY-MM-DD)")
def hourly_flow(station, date):
    """查看小时统计"""
    try:
        params = {"date": date} if date else {}
        response = requests.get(f"{API_BASE}/stations/{station}/passenger-flow/hourly", params=params)
        response.raise_for_status()
        data = response.json()
        
        click.echo(f"\n=== 车站 {station} {data['date']} 小时客流统计 ===")
        click.echo(f"{'小时':<10}{'平均人数':<12}{'最大人数':<12}{'最小人数':<12}")
        click.echo("-" * 46)
        for h in data['hourly_stats']:
            click.echo(f"{h['hour']:02d}:00     {h['average_count']:<12}{h['max_count']:<12}{h['min_count']:<12}")
    except Exception as e:
        click.echo(f"错误: {e}")

@flow.command("daily")
@click.option("--station", prompt="车站代码")
@click.option("--start", default=None, help="开始日期 (YYYY-MM-DD)")
@click.option("--end", default=None, help="结束日期 (YYYY-MM-DD)")
def daily_flow(station, start, end):
    """查看日统计"""
    try:
        params = {}
        if start:
            params["start_date"] = start
        if end:
            params["end_date"] = end
        response = requests.get(f"{API_BASE}/stations/{station}/passenger-flow/daily", params=params)
        response.raise_for_status()
        data = response.json()
        
        click.echo(f"\n=== 车站 {station} 日客流统计 ===")
        click.echo(f"{'日期':<15}{'平均人数':<12}{'最大人数':<12}{'总人数':<12}")
        click.echo("-" * 51)
        for d in data['daily_stats']:
            click.echo(f"{d['date']:<15}{d['average_count']:<12}{d['max_count']:<12}{d['total_count']:<12}")
    except Exception as e:
        click.echo(f"错误: {e}")

@flow.command("alerts")
@click.option("--station", prompt="车站代码")
@click.option("--all", is_flag=True, help="显示所有预警（包括已解决的）")
def flow_alerts(station, all):
    """查看客流预警"""
    try:
        params = {"active_only": not all}
        response = requests.get(f"{API_BASE}/stations/{station}/alerts", params=params)
        response.raise_for_status()
        alerts = response.json()
        
        click.echo(f"\n=== 车站 {station} 预警列表 ===")
        if not alerts:
            click.echo("暂无预警")
            return
        
        for a in alerts:
            status = "活跃" if a['is_active'] else "已解决"
            click.echo(f"\n类型: {a['alert_type']}  严重程度: {a['severity']}  状态: {status}")
            click.echo(f"消息: {a['message']}")
            click.echo(f"时间: {a['created_at']}")
    except Exception as e:
        click.echo(f"错误: {e}")

@cli.group()
def checkin():
    """检票管理"""
    pass

@checkin.command("list")
@click.option("--platform", default=None, help="站台号")
def list_trains(platform):
    """列出车次"""
    try:
        params = {"platform": platform} if platform else {}
        response = requests.get(f"{API_BASE}/trains", params=params)
        response.raise_for_status()
        trains = response.json()
        
        click.echo("\n=== 车次列表 ===")
        for t in trains:
            checkin = "检票中" if t['is_checkin_active'] else "未检票"
            click.echo(f"车次: {t['train_number']:10} 站台: {t['platform']:8} 发车: {t['departure_time']}  {checkin}")
    except Exception as e:
        click.echo(f"错误: {e}")

@checkin.command("start")
@click.option("--train", prompt="车次号")
def start_checkin(train):
    """开始检票"""
    try:
        response = requests.post(f"{API_BASE}/trains/{train}/checkin/start")
        response.raise_for_status()
        click.echo(f"车次 {train} 检票已开始")
    except Exception as e:
        click.echo(f"错误: {e}")

@checkin.command("stop")
@click.option("--train", prompt="车次号")
def stop_checkin(train):
    """停止检票"""
    try:
        response = requests.post(f"{API_BASE}/trains/{train}/checkin/stop")
        response.raise_for_status()
        click.echo(f"车次 {train} 检票已停止")
    except Exception as e:
        click.echo(f"错误: {e}")

if __name__ == "__main__":
    cli()
