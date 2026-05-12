#!/usr/bin/env python3
"""地铁运营调度系统命令行客户端"""

import typer
import requests
from typing import Optional
from datetime import datetime
import json

app = typer.Typer(name="metro-cli", help="地铁运营调度系统命令行客户端")
trains_app = typer.Typer(name="trains", help="列车管理命令")
schedules_app = typer.Typer(name="schedules", help="运行计划管理命令")
signals_app = typer.Typer(name="signals", help="信号机管理命令")
power_app = typer.Typer(name="power", help="供电管理命令")
logs_app = typer.Typer(name="logs", help="调度日志命令")

app.add_typer(trains_app)
app.add_typer(schedules_app)
app.add_typer(signals_app)
app.add_typer(power_app)
app.add_typer(logs_app)

BASE_URL = "http://localhost:8701"


def make_request(method: str, endpoint: str, **kwargs):
    url = f"{BASE_URL}{endpoint}"
    try:
        response = requests.request(method, url, **kwargs)
        if response.status_code >= 400:
            typer.echo(f"Error: {response.status_code} - {response.text}")
            raise typer.Exit(code=1)
        return response.json()
    except requests.exceptions.ConnectionError:
        typer.echo(f"Error: 无法连接到服务器 {BASE_URL}")
        typer.echo("请先启动 FastAPI 服务器: uvicorn backend.main:app --reload")
        raise typer.Exit(code=1)


@trains_app.command("list")
def list_trains(skip: int = 0, limit: int = 100):
    """列出所有列车"""
    result = make_request("GET", "/trains/", params={"skip": skip, "limit": limit})
    if not result:
        typer.echo("没有找到列车")
        return

    typer.echo("\n" + "=" * 80)
    typer.echo(f"{'ID':<6} {'车号':<12} {'状态':<12} {'限速(km/h)':<12} {'容量':<8}")
    typer.echo("=" * 80)
    for train in result:
        typer.echo(
            f"{train['id']:<6} {train['train_number']:<12} {train['status']:<12} "
            f"{train['speed_limit']:<12} {train['capacity']:<8}"
        )
    typer.echo("=" * 80 + "\n")


@trains_app.command("status")
def train_status(train_id: int):
    """查看列车详情"""
    result = make_request("GET", f"/trains/{train_id}")
    typer.echo(json.dumps(result, indent=2, ensure_ascii=False, default=str))


@trains_app.command("create")
def create_train(number: str = typer.Option(..., help="列车编号"), capacity: int = typer.Option(1000, help="载客容量")):
    """创建新列车"""
    data = {"train_number": number, "capacity": capacity}
    result = make_request("POST", "/trains/", json=data)
    typer.echo(f"成功创建列车: {result['train_number']} (ID: {result['id']})")


@schedules_app.command("list")
def list_schedules(line_id: Optional[int] = None, active_only: bool = False):
    """列出运行计划"""
    params = {}
    if line_id:
        params["line_id"] = line_id
    if active_only:
        params["active_only"] = "true"
    result = make_request("GET", "/schedules/", params=params)
    if not result:
        typer.echo("没有找到运行计划")
        return

    typer.echo("\n" + "=" * 100)
    typer.echo(f"{'ID':<6} {'线路ID':<8} {'名称':<20} {'首班':<20} {'末班':<20} {'间隔(秒)':<10} {'状态':<10}")
    typer.echo("=" * 100)
    for s in result:
        typer.echo(
            f"{s['id']:<6} {s['line_id']:<8} {s['name']:<20} "
            f"{s['first_departure'][:19] if s['first_departure'] else '-':<20} "
            f"{s['last_departure'][:19] if s['last_departure'] else '-':<20} "
            f"{s['interval_seconds']:<10} {'活跃' if s['is_active'] else '停用':<10}"
        )
    typer.echo("=" * 100 + "\n")


@schedules_app.command("trains")
def list_schedule_trains(schedule_id: int):
    """列出运行计划中的班次"""
    result = make_request("GET", f"/schedules/{schedule_id}/trains")
    if not result:
        typer.echo("没有找到班次")
        return

    typer.echo(f"\n运行计划 {schedule_id} 的班次:")
    typer.echo("=" * 100)
    typer.echo(
        f"{'序号':<6} {'计划发车':<25} {'实际发车':<25} {'计划到达':<25} {'状态':<10}"
    )
    typer.echo("=" * 100)
    for st in result:
        actual_dep = st.get("actual_departure") or "-"
        if actual_dep != "-":
            actual_dep = actual_dep[:19]
        typer.echo(
            f"{st['sequence']:<6} "
            f"{st['scheduled_departure'][:19] if st['scheduled_departure'] else '-':<25} "
            f"{actual_dep:<25} "
            f"{st['scheduled_arrival'][:19] if st['scheduled_arrival'] else '-':<25} "
            f"{'已完成' if st['is_completed'] else '待发车':<10}"
        )
    typer.echo("=" * 100 + "\n")


@schedules_app.command("optimal")
def optimal_interval(schedule_id: int):
    """计算最优发车间隔"""
    result = make_request("POST", f"/schedules/{schedule_id}/optimal-interval")
    typer.echo("\n" + "=" * 50)
    typer.echo(f"运行计划 ID: {result['schedule_id']}")
    typer.echo(f"当前间隔: {result['current_interval_seconds']} 秒")
    typer.echo(f"最优间隔: {result['optimal_interval_seconds']} 秒")
    typer.echo(f"最小安全间隔: {result['min_safe_interval_seconds']} 秒")
    typer.echo("=" * 50 + "\n")


@signals_app.command("list")
def list_signals(station_id: Optional[int] = None):
    """列出信号机"""
    params = {}
    if station_id:
        params["station_id"] = station_id
    result = make_request("GET", "/signals/", params=params)
    if not result:
        typer.echo("没有找到信号机")
        return

    typer.echo("\n" + "=" * 80)
    typer.echo(f"{'ID':<6} {'编号':<12} {'车站ID':<10} {'状态':<12} {'故障时间':<25}")
    typer.echo("=" * 80)
    for s in result:
        fault_time = s.get("fault_time")
        if fault_time:
            fault_time = fault_time[:19]
        else:
            fault_time = "-"
        typer.echo(
            f"{s['id']:<6} {s['signal_code']:<12} {s['station_id']:<10} "
            f"{s['status']:<12} {fault_time:<25}"
        )
    typer.echo("=" * 80 + "\n")


@signals_app.command("fault")
def report_signal_fault(
    signal_id: int = typer.Option(..., help="信号机ID"),
    description: str = typer.Option("", help="故障描述"),
):
    """报告信号机故障"""
    result = make_request("POST", f"/signals/{signal_id}/fault", params={"description": description})
    typer.echo(f"\n信号机故障已报告:")
    typer.echo(f"  编号: {result['signal_code']}")
    typer.echo(f"  故障时间: {result['fault_time']}")
    if result.get('affected_section_name'):
        typer.echo(f"  影响区间: {result['affected_section_name']}")
    typer.echo(f"  注意: 受影响区段列车已自动限速")


@signals_app.command("recover")
def confirm_recovery(
    signal_id: int = typer.Option(..., help="信号机ID"),
    operator: str = typer.Option(..., help="操作员姓名"),
):
    """确认信号机恢复"""
    result = make_request("POST", f"/signals/{signal_id}/confirm-recovery", params={"operator": operator})
    typer.echo(f"信号机 {result['signal_code']} 已确认恢复 (状态: {result['status']})")


@signals_app.command("restore")
def restore_normal(signal_id: int):
    """将信号机恢复到正常状态"""
    result = make_request("POST", f"/signals/{signal_id}/restore-normal")
    typer.echo(f"信号机 {result['signal_code']} 已恢复正常运行")


@power_app.command("list")
def list_power_sections():
    """列出供电区间"""
    result = make_request("GET", "/power-sections/")
    if not result:
        typer.echo("没有找到供电区间")
        return

    typer.echo("\n" + "=" * 100)
    typer.echo(f"{'ID':<6} {'编号':<12} {'名称':<25} {'状态':<10} {'停电时间':<25}")
    typer.echo("=" * 100)
    for p in result:
        outage_time = p.get("outage_time") or "-"
        if outage_time != "-":
            outage_time = outage_time[:19]
        typer.echo(
            f"{p['id']:<6} {p['section_code']:<12} {p['name']:<25} "
            f"{'正常' if p['status'] == 'active' else '停电':<10} "
            f"{outage_time:<25}"
        )
    typer.echo("=" * 100 + "\n")


@power_app.command("outage")
def report_outage(section_id: int):
    """报告供电区间停电"""
    result = make_request("POST", f"/power-sections/{section_id}/outage")
    typer.echo(f"\n供电区间停电已报告:")
    typer.echo(f"  名称: {result['section_name']}")
    typer.echo(f"  停电时间: {result['outage_time']}")
    typer.echo(f"  注意: 该区间列车已停运")


@power_app.command("restore")
def restore_power_section(section_id: int):
    """恢复供电区间供电"""
    result = make_request("POST", f"/power-sections/{section_id}/restore")
    typer.echo(f"供电区间 {result['name']} 供电已恢复，需调度员确认运行计划")


@power_app.command("confirm")
def confirm_operations(
    section_id: int = typer.Option(..., help="供电区间ID"),
    operator: str = typer.Option(..., help="操作员姓名"),
):
    """确认供电恢复后运行计划"""
    result = make_request("POST", f"/power-sections/{section_id}/confirm-operations", params={"operator": operator})
    typer.echo(result["message"])


@logs_app.command("list")
def list_logs(
    entity_type: Optional[str] = None,
    level: Optional[str] = None,
    limit: int = 50,
):
    """查看调度日志"""
    params = {"limit": limit}
    if entity_type:
        params["entity_type"] = entity_type
    if level:
        params["level"] = level

    result = make_request("GET", "/logs/", params=params)
    if not result:
        typer.echo("没有找到日志")
        return

    typer.echo("\n" + "=" * 120)
    typer.echo(f"{'时间':<25} {'级别':<10} {'实体类型':<15} {'实体ID':<10} {'操作员':<15} {'消息'}")
    typer.echo("=" * 120)
    for log in result:
        typer.echo(
            f"{log['created_at'][:19] if log['created_at'] else '-':<25} "
            f"{log['level']:<10} "
            f"{log.get('entity_type') or '-':<15} "
            f"{log.get('entity_id') or '-':<10} "
            f"{log.get('operator') or '-':<15} "
            f"{log['message']}"
        )
    typer.echo("=" * 120 + "\n")


@app.command("init-sample")
def init_sample_data():
    """初始化示例数据"""
    result = make_request("POST", "/init-sample-data")
    if "message" in result and "already exists" in result["message"]:
        typer.echo("示例数据已存在")
    else:
        typer.echo(f"示例数据创建成功:")
        typer.echo(f"  线路: {result.get('line_id', '-')}")
        typer.echo(f"  车站: {result.get('stations', 0)} 个")
        typer.echo(f"  供电区间: {result.get('power_sections', 0)} 个")
        typer.echo(f"  信号机: {result.get('signals', 0)} 个")
        typer.echo(f"  列车: {result.get('trains', 0)} 列")
        typer.echo(f"  运行计划: {result.get('schedules', 0)} 个")


@app.command("health")
def health_check():
    """检查服务器健康状态"""
    result = make_request("GET", "/health")
    typer.echo(f"服务器状态: {result['status']}")
    typer.echo(f"检查时间: {result['timestamp']}")


if __name__ == "__main__":
    app()
