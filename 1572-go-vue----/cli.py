#!/usr/bin/env python3
import click
from datetime import datetime, date, timedelta
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich import box
from sqlalchemy.orm import Session

from app.database import SessionLocal, init_db
from app.models import TrainStatus, RouteStatus, CrewType, AlertStatus
from app.services.train_service import TrainService, RouteService
from app.services.crew_service import CrewMemberService, CrewGroupService, DutyRulesService
from app.services.optimization_service import WorkHoursService, OptimizationService
from app.services.delay_service import DelayService

console = Console()
init_db()


def get_session() -> Session:
    return SessionLocal()


@click.group()
def cli():
    """铁路客运调度管理系统 - 命令行客户端"""
    pass


@cli.group()
def train():
    """列车管理"""
    pass


@train.command("list")
@click.option("--status", type=click.Choice(['idle', 'in_service', 'maintenance']), help="按状态筛选")
def train_list(status):
    """列出所有列车"""
    db = get_session()
    try:
        status_enum = TrainStatus(status) if status else None
        trains = TrainService.list_trains(db, status_enum)
        
        table = Table(title="列车列表", box=box.ROUNDED)
        table.add_column("ID", style="cyan")
        table.add_column("车次", style="green")
        table.add_column("车型", style="yellow")
        table.add_column("定员", justify="right")
        table.add_column("乘务定额", justify="right")
        table.add_column("状态", style="magenta")
        
        for t in trains:
            status_style = "green" if t.status == TrainStatus.IDLE else "yellow" if t.status == TrainStatus.IN_SERVICE else "red"
            table.add_row(
                str(t.id),
                t.train_number,
                t.train_type,
                str(t.seat_capacity),
                str(t.crew_quota),
                f"[{status_style}]{t.status.value}[/{status_style}]"
            )
        
        console.print(table)
    finally:
        db.close()


@cli.group()
def route():
    """交路管理"""
    pass


@route.command("list")
@click.option("--status", type=click.Choice(['scheduled', 'running', 'delayed', 'completed', 'cancelled']), help="按状态筛选")
@click.option("--train-id", type=int, help="按列车ID筛选")
@click.option("--today", is_flag=True, help="仅显示今日交路")
def route_list(status, train_id, today):
    """列出交路信息"""
    db = get_session()
    try:
        status_enum = RouteStatus(status) if status else None
        start_date = datetime.combine(date.today(), datetime.min.time()) if today else None
        end_date = datetime.combine(date.today(), datetime.max.time()) if today else None
        
        routes = RouteService.list_routes(db, status_enum, train_id, start_date, end_date)
        
        table = Table(title="交路列表", box=box.ROUNDED)
        table.add_column("ID", style="cyan")
        table.add_column("交路号", style="green")
        table.add_column("区间", style="yellow")
        table.add_column("计划出发")
        table.add_column("计划到达")
        table.add_column("晚点(分)", justify="right")
        table.add_column("状态", style="magenta")
        
        for r in routes:
            status_style = "green" if r.status in [RouteStatus.COMPLETED] else \
                          "yellow" if r.status in [RouteStatus.RUNNING, RouteStatus.SCHEDULED] else \
                          "red" if r.status == RouteStatus.DELAYED else "dim"
            table.add_row(
                str(r.id),
                r.route_code,
                f"{r.departure_station}→{r.arrival_station}",
                r.scheduled_departure.strftime("%m-%d %H:%M"),
                r.scheduled_arrival.strftime("%m-%d %H:%M"),
                f"[red]{r.delay_minutes}[/red]" if r.delay_minutes > 0 else "0",
                f"[{status_style}]{r.status.value}[/{status_style}]"
            )
        
        console.print(table)
    finally:
        db.close()


@route.command("detail")
@click.argument("route_id", type=int)
def route_detail(route_id):
    """查看交路详情"""
    db = get_session()
    try:
        r = RouteService.get_route(db, route_id)
        if not r:
            console.print("[red]交路不存在[/red]")
            return
        
        panel = Panel.fit(
            f"交路号: [green]{r.route_code}[/green]\n"
            f"区间: [yellow]{r.departure_station} → {r.arrival_station}[/yellow]\n\n"
            f"计划出发: {r.scheduled_departure.strftime('%Y-%m-%d %H:%M:%S')}\n"
            f"计划到达: {r.scheduled_arrival.strftime('%Y-%m-%d %H:%M:%S')}\n"
            f"实际出发: {(r.actual_departure.strftime('%Y-%m-%d %H:%M:%S') if r.actual_departure else '未出发')}\n"
            f"实际到达: {(r.actual_arrival.strftime('%Y-%m-%d %H:%M:%S') if r.actual_arrival else '未到达')}\n\n"
            f"状态: [magenta]{r.status.value}[/magenta]\n"
            f"晚点: [red]{r.delay_minutes}分钟[/red]" if r.delay_minutes > 0 else f"晚点: 0分钟",
            title=f"交路详情 - {r.route_code}",
            box=box.ROUNDED
        )
        console.print(panel)
    finally:
        db.close()


@cli.group()
def crew():
    """乘务组管理"""
    pass


@crew.command("list")
@click.option("--type", "crew_type", type=click.Choice(['driver', 'conductor', 'attendant']), help="按类型筛选")
def crew_list(crew_type):
    """列出乘务人员"""
    db = get_session()
    try:
        type_enum = CrewType(crew_type) if crew_type else None
        members = CrewMemberService.list_crew_members(db, type_enum)
        
        table = Table(title="乘务人员列表", box=box.ROUNDED)
        table.add_column("ID", style="cyan")
        table.add_column("工号", style="green")
        table.add_column("姓名", style="yellow")
        table.add_column("类型", style="magenta")
        table.add_column("电话")
        
        for m in members:
            type_style = "red" if m.crew_type == CrewType.DRIVER else "cyan" if m.crew_type == CrewType.CONDUCTOR else "white"
            table.add_row(
                str(m.id),
                m.employee_id,
                m.name,
                f"[{type_style}]{m.crew_type.value}[/{type_style}]",
                m.phone or "-"
            )
        
        console.print(table)
    finally:
        db.close()


@crew.command("groups")
def crew_groups():
    """列出乘务组"""
    db = get_session()
    try:
        groups = CrewGroupService.list_crew_groups(db)
        
        table = Table(title="乘务组列表", box=box.ROUNDED)
        table.add_column("ID", style="cyan")
        table.add_column("组号", style="green")
        table.add_column("名称", style="yellow")
        table.add_column("成员数", justify="right")
        
        for g in groups:
            members = CrewGroupService.get_group_members(db, g.id)
            table.add_row(str(g.id), g.group_code, g.name, str(len(members)))
        
        console.print(table)
    finally:
        db.close()


@crew.command("schedule")
@click.option("--year", type=int, default=lambda: date.today().year, help="年份")
@click.option("--month", type=int, default=lambda: date.today().month, help="月份")
@click.option("--type", "crew_type", type=click.Choice(['driver', 'conductor', 'attendant']), help="按类型筛选")
def crew_schedule(year, month, crew_type):
    """查看乘务排班工时统计"""
    db = get_session()
    try:
        type_enum = CrewType(crew_type) if crew_type else None
        report = WorkHoursService.get_monthly_report(db, year, month, type_enum)
        
        table = Table(title=f"乘务排班工时统计 - {year}年{month}月", box=box.ROUNDED)
        table.add_column("工号", style="green")
        table.add_column("姓名", style="yellow")
        table.add_column("类型", style="magenta")
        table.add_column("工时(小时)", justify="right")
        table.add_column("上限(小时)", justify="right")
        table.add_column("剩余(小时)", justify="right")
        table.add_column("状态")
        
        for item in report:
            remaining = item["remaining_hours"]
            status = "[red]超限[/red]" if item["is_over_limit"] else \
                    "[yellow]接近上限[/yellow]" if remaining < 10 else "[green]正常[/green]"
            
            type_style = "red" if item["crew_type"] == "driver" else "cyan" if item["crew_type"] == "conductor" else "white"
            hours_style = "red" if item["is_over_limit"] else "white"
            
            table.add_row(
                item["employee_id"],
                item["name"],
                f"[{type_style}]{item['crew_type']}[/{type_style}]",
                f"[{hours_style}]{item['total_hours']:.1f}[/{hours_style}]",
                str(item["max_limit_hours"]),
                f"{remaining:.1f}",
                status
            )
        
        console.print(table)
    finally:
        db.close()


@cli.group()
def alert():
    """预警管理"""
    pass


@alert.command("list")
@click.option("--status", type=click.Choice(['pending', 'acknowledged', 'resolved']), help="按状态筛选")
def alert_list(status):
    """列出预警信息"""
    db = get_session()
    try:
        status_enum = AlertStatus(status) if status else None
        alerts = WorkHoursService.get_alerts(db, status_enum)
        
        table = Table(title="预警信息列表", box=box.ROUNDED)
        table.add_column("ID", style="cyan")
        table.add_column("类型", style="magenta")
        table.add_column("标题", style="yellow")
        table.add_column("接收角色")
        table.add_column("状态")
        table.add_column("创建时间")
        
        for a in alerts:
            status_style = "red" if a.status == AlertStatus.PENDING else \
                          "yellow" if a.status == AlertStatus.ACKNOWLEDGED else "green"
            type_style = "red" if a.alert_type.value == "overtime" else "yellow"
            table.add_row(
                str(a.id),
                f"[{type_style}]{a.alert_type.value}[/{type_style}]",
                a.title,
                a.target_role,
                f"[{status_style}]{a.status.value}[/{status_style}]",
                a.created_at.strftime("%m-%d %H:%M")
            )
        
        console.print(table)
    finally:
        db.close()


@cli.group()
def optimization():
    """交路优化"""
    pass


@optimization.command("list")
@click.option("--implemented", is_flag=True, help="仅显示已实施的")
@click.option("--pending", is_flag=True, help="仅显示待实施的")
def optimization_list(implemented, pending):
    """列出优化建议"""
    db = get_session()
    try:
        is_implemented = None
        if implemented:
            is_implemented = True
        elif pending:
            is_implemented = False
            
        optimizations = OptimizationService.get_optimizations(db, None, is_implemented)
        
        table = Table(title="交路优化建议", box=box.ROUNDED)
        table.add_column("ID", style="cyan")
        table.add_column("类型", style="magenta")
        table.add_column("描述", style="yellow")
        table.add_column("日均运用(小时)", justify="right")
        table.add_column("平均折返(分钟)", justify="right")
        table.add_column("状态")
        
        for o in optimizations:
            status = "[green]已实施[/green]" if o.is_implemented else "[yellow]待实施[/yellow]"
            type_style = "red" if "利用率" in o.optimization_type else "cyan"
            table.add_row(
                str(o.id),
                f"[{type_style}]{o.optimization_type}[/{type_style}]",
                o.description[:40] + "..." if len(o.description) > 40 else o.description,
                f"{o.avg_utilization_hours:.1f}" if o.avg_utilization_hours else "-",
                f"{o.avg_turnaround_minutes:.0f}" if o.avg_turnaround_minutes else "-",
                status
            )
        
        console.print(table)
    finally:
        db.close()


@optimization.command("detail")
@click.argument("opt_id", type=int)
def optimization_detail(opt_id):
    """查看优化建议详情"""
    db = get_session()
    try:
        optimizations = OptimizationService.get_optimizations(db)
        o = next((x for x in optimizations if x.id == opt_id), None)
        
        if not o:
            console.print("[red]优化建议不存在[/red]")
            return
        
        panel = Panel(
            f"{o.suggestion}",
            title=f"优化建议 - {o.optimization_type}",
            box=box.ROUNDED
        )
        console.print(panel)
    finally:
        db.close()


@cli.group()
def delay():
    """晚点管理"""
    pass


@delay.command("list")
@click.option("--route-id", type=int, help="按交路ID筛选")
def delay_list(route_id):
    """列出晚点调整方案"""
    db = get_session()
    try:
        adjustments = DelayService.get_delay_adjustments(db, route_id)
        
        table = Table(title="晚点调整方案列表", box=box.ROUNDED)
        table.add_column("ID", style="cyan")
        table.add_column("交路ID", style="green")
        table.add_column("晚点(分钟)", justify="right")
        table.add_column("类型", style="magenta")
        table.add_column("通知对象")
        table.add_column("创建时间")
        
        for a in adjustments:
            table.add_row(
                str(a.id),
                str(a.route_id),
                f"[red]{a.delay_minutes}[/red]",
                a.adjustment_type,
                a.notified_to,
                a.created_at.strftime("%m-%d %H:%M")
            )
        
        console.print(table)
    finally:
        db.close()


@delay.command("detail")
@click.argument("adjustment_id", type=int)
def delay_detail(adjustment_id):
    """查看晚点调整方案详情"""
    db = get_session()
    try:
        adjustments = DelayService.get_delay_adjustments(db)
        a = next((x for x in adjustments if x.id == adjustment_id), None)
        
        if not a:
            console.print("[red]调整方案不存在[/red]")
            return
        
        panel = Panel(
            f"{a.adjustment_plan}",
            title=f"晚点调整方案 - 交路{a.route_id}",
            box=box.ROUNDED
        )
        console.print(panel)
    finally:
        db.close()


if __name__ == "__main__":
    cli()
