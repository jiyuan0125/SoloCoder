from typing import Optional, List
from datetime import datetime, timedelta, timezone
import json

from shared.models import InboundItem, CrossDockOrderResponse

import typer
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich import print as rprint

from shared.models import (
    CrossDockOrderCreate,
    OutboundItem,
    InboundItem,
    InboundScanRequest,
    OutboundScanRequest,
    BatchCrossDockCreate,
)
from cli.client import CrossDockClient


app = typer.Typer(name="cross-dock", help="越库管理命令行工具")
console = Console()
DEFAULT_SERVER_URL = "http://localhost:8000"


def get_client(server_url: str) -> CrossDockClient:
    return CrossDockClient(server_url)


@app.command()
def create(
    order_number: str = typer.Option(..., "--order-number", "-n", help="越库单号"),
    outbound_order_number: Optional[str] = typer.Option(
        None, "--outbound-order", help="出库单号（批量越库时使用）"
    ),
    items: str = typer.Option(
        ..., "--items", "-i", help="商品列表 JSON: [{'sku': 'SKU001', 'quantity': 10}]"
    ),
    warehouse_id: str = typer.Option(..., "--warehouse", "-w", help="仓库ID"),
    destination: str = typer.Option(..., "--dest", "-d", help="目的地"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """创建越库单"""
    try:
        item_list = json.loads(items)
        expected_items = [OutboundItem(**item) for item in item_list]
    except (json.JSONDecodeError, KeyError, TypeError) as e:
        console.print(f"[red]商品列表格式错误: {e}[/red]")
        raise typer.Exit(code=1)

    client = get_client(server)
    order_create = CrossDockOrderCreate(
        order_number=order_number,
        outbound_order_number=outbound_order_number,
        expected_items=expected_items,
        warehouse_id=warehouse_id,
        destination=destination,
    )

    try:
        response = client.create_order(order_create)
        console.print(Panel(f"越库单创建成功!", title="成功", style="green"))
        _display_order(response)
    except Exception as e:
        console.print(f"[red]创建失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="list")
def list_orders(
    status: Optional[str] = typer.Option(
        None, "--status", help="状态筛选: pending_inbound, inbound_completed, outbound_completed, timeout_alert"
    ),
    warehouse_id: Optional[str] = typer.Option(None, "--warehouse", "-w", help="仓库ID筛选"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """列出越库单"""
    client = get_client(server)
    try:
        orders = client.list_orders(status, warehouse_id)
        if not orders:
            console.print("[yellow]没有找到越库单[/yellow]")
            return

        table = Table(title="越库单列表")
        table.add_column("ID", style="cyan")
        table.add_column("单号", style="cyan")
        table.add_column("状态", style="magenta")
        table.add_column("仓库", style="green")
        table.add_column("目的地", style="blue")
        table.add_column("入库操作员", style="yellow")
        table.add_column("出库操作员", style="yellow")
        table.add_column("停留时长(分钟)", style="cyan")

        for order in orders:
            duration = f"{order.duration_minutes:.1f}" if order.duration_minutes else "-"
            status_style = _get_status_style(order.status.value)
            table.add_row(
                order.id[:8] + "...",
                order.order_number,
                f"[{status_style}]{order.status.value}[/{status_style}]",
                order.warehouse_id,
                order.destination,
                order.inbound_operator_id or "-",
                order.outbound_operator_id or "-",
                duration,
            )

        console.print(table)
    except Exception as e:
        console.print(f"[red]获取列表失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command()
def get(
    order_id: str = typer.Argument(..., help="越库单ID"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """获取越库单详情"""
    client = get_client(server)
    try:
        order = client.get_order(order_id)
        _display_order(order)
    except Exception as e:
        console.print(f"[red]获取失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="inbound-scan")
def inbound_scan(
    order_id: str = typer.Option(..., "--order-id", "-o", help="越库单ID"),
    operator_id: str = typer.Option(..., "--operator", "-p", help="操作员ID"),
    items: str = typer.Option(
        ..., "--items", "-i", help="商品列表 JSON: [{'sku': 'SKU001', 'quantity': 10}]"
    ),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """入库扫码"""
    try:
        item_list = json.loads(items)
        inbound_items = [InboundItem(**item) for item in item_list]
    except (json.JSONDecodeError, KeyError, TypeError) as e:
        console.print(f"[red]商品列表格式错误: {e}[/red]")
        raise typer.Exit(code=1)

    client = get_client(server)
    request = InboundScanRequest(
        order_id=order_id,
        operator_id=operator_id,
        items=inbound_items,
        scan_time=None,
    )

    try:
        response = client.inbound_scan(request)
        console.print(Panel(f"入库扫码成功!", title="成功", style="green"))
        _display_order(response)
    except Exception as e:
        console.print(f"[red]入库扫码失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="outbound-scan")
def outbound_scan(
    order_id: str = typer.Option(..., "--order-id", "-o", help="越库单ID"),
    operator_id: str = typer.Option(..., "--operator", "-p", help="操作员ID"),
    items: str = typer.Option(
        ..., "--items", "-i", help="商品列表 JSON: [{'sku': 'SKU001', 'quantity': 10}]"
    ),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """出库扫码"""
    try:
        item_list = json.loads(items)
        outbound_items = [OutboundItem(**item) for item in item_list]
    except (json.JSONDecodeError, KeyError, TypeError) as e:
        console.print(f"[red]商品列表格式错误: {e}[/red]")
        raise typer.Exit(code=1)

    client = get_client(server)
    request = OutboundScanRequest(
        order_id=order_id,
        operator_id=operator_id,
        items=outbound_items,
        scan_time=None,
    )

    try:
        response = client.outbound_scan(request)
        console.print(Panel(f"出库扫码成功!", title="成功", style="green"))
        _display_order(response)
    except Exception as e:
        console.print(f"[red]出库扫码失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="batch-create")
def batch_create(
    outbound_order_number: str = typer.Option(..., "--outbound-order", "-o", help="出库单号"),
    inbound_order_numbers: str = typer.Option(
        ..., "--inbound-orders", "-i", help="入库单号列表 JSON: ['IN001', 'IN002']"
    ),
    items: str = typer.Option(
        ..., "--items", help="商品列表 JSON: [{'sku': 'SKU001', 'quantity': 10}]"
    ),
    warehouse_id: str = typer.Option(..., "--warehouse", "-w", help="仓库ID"),
    destination: str = typer.Option(..., "--dest", "-d", help="目的地"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """创建批量越库"""
    try:
        inbound_orders = json.loads(inbound_order_numbers)
        item_list = json.loads(items)
        expected_items = [OutboundItem(**item) for item in item_list]
    except (json.JSONDecodeError, KeyError, TypeError) as e:
        console.print(f"[red]参数格式错误: {e}[/red]")
        raise typer.Exit(code=1)

    client = get_client(server)
    batch_create = BatchCrossDockCreate(
        outbound_order_number=outbound_order_number,
        inbound_order_numbers=inbound_orders,
        expected_items=expected_items,
        warehouse_id=warehouse_id,
        destination=destination,
    )

    try:
        response = client.create_batch_cross_dock(batch_create)
        console.print(Panel(f"批量越库创建成功!", title="成功", style="green"))
        table = Table(title="批量越库信息")
        table.add_column("出库单号", style="cyan")
        table.add_column("状态", style="magenta")
        table.add_column("完成/总数", style="green")
        table.add_row(
            response.outbound_order_number,
            response.status.value,
            f"{response.completed_count}/{response.total_count}",
        )
        console.print(table)
        console.print(f"\n入库单ID: {response.inbound_order_ids}")
    except Exception as e:
        console.print(f"[red]创建失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="batch-status")
def batch_status(
    outbound_order_number: str = typer.Argument(..., help="出库单号"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """查看批量越库状态"""
    client = get_client(server)
    try:
        response = client.get_batch_status(outbound_order_number)
        table = Table(title="批量越库状态")
        table.add_column("出库单号", style="cyan")
        table.add_column("状态", style="magenta")
        table.add_column("完成/总数", style="green")
        table.add_row(
            response.outbound_order_number,
            response.status.value,
            f"{response.completed_count}/{response.total_count}",
        )
        console.print(table)
        console.print(f"\n入库单ID: {response.inbound_order_ids}")
    except Exception as e:
        console.print(f"[red]获取状态失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="stats")
def get_stats(
    days: int = typer.Option(7, "--days", "-d", help="统计天数"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """获取越库效率统计"""
    end_time = datetime.now(timezone.utc)
    start_time = end_time - timedelta(days=days)

    client = get_client(server)
    try:
        stats = client.get_efficiency_statistics(start_time, end_time)
        table = Table(title=f"越库效率统计 (最近 {days} 天)")
        table.add_column("指标", style="cyan")
        table.add_column("数值", style="green")
        table.add_row("平均停留时长(分钟)", f"{stats.average_duration_minutes}")
        table.add_row("超时率(%)", f"{stats.timeout_rate}")
        table.add_row("日处理量", f"{stats.daily_volume}")
        console.print(table)
    except Exception as e:
        console.print(f"[red]获取统计失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="report")
def get_report(
    date: Optional[str] = typer.Option(None, "--date", help="报告日期 (格式: YYYY-MM-DD)"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """获取越库日报"""
    if date:
        try:
            report_date = datetime.strptime(date, "%Y-%m-%d")
        except ValueError as e:
            console.print(f"[red]日期格式错误: {e}[/red]")
            raise typer.Exit(code=1)
    else:
        report_date = datetime.now(timezone.utc)

    client = get_client(server)
    try:
        report = client.get_daily_report(report_date)
        console.print(Panel(f"越库作业日报 - {report.report_date.date()}", style="cyan"))
        table = Table(title="报告摘要")
        table.add_column("指标", style="cyan")
        table.add_column("数值", style="green")
        table.add_row("总越库单数", f"{report.total_orders}")
        table.add_row("已完成单数", f"{report.completed_orders}")
        table.add_row("待处理单数", f"{report.pending_orders}")
        table.add_row("超时单数", f"{report.timeout_orders}")
        table.add_row("跨日单数", f"{report.cross_day_orders}")
        console.print(table)
        console.print("\n[bold]效率统计:[/bold]")
        stats_table = Table()
        stats_table.add_column("平均停留时长(分钟)", style="cyan")
        stats_table.add_column("超时率(%)", style="magenta")
        stats_table.add_column("日处理量", style="green")
        stats_table.add_row(
            f"{report.efficiency.average_duration_minutes}",
            f"{report.efficiency.timeout_rate}",
            f"{report.efficiency.daily_volume}",
        )
        console.print(stats_table)
    except Exception as e:
        console.print(f"[red]获取报告失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="monitor")
def get_monitor(
    warehouse_id: str = typer.Argument(..., help="仓库ID"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """越库区域监控"""
    client = get_client(server)
    try:
        monitor = client.get_zone_monitor(warehouse_id)
        console.print(Panel(f"越库区域监控 - {monitor.zone_name}", style="cyan"))
        table = Table(title="当前状态")
        table.add_column("状态", style="cyan")
        table.add_column("数量", style="green")
        table.add_row("待入库", f"{monitor.pending_inbound_count}")
        table.add_row("已入库待出库", f"{monitor.inbound_completed_count}")
        table.add_row("超时", f"[red]{monitor.timeout_count}[/red]")
        console.print(table)
        if monitor.peak_hour_suggestion:
            console.print(f"\n[yellow]高峰期建议: {monitor.peak_hour_suggestion}[/yellow]")
    except Exception as e:
        console.print(f"[red]获取监控失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command(name="plan")
def get_plan(
    exception_type: str = typer.Argument(..., help="异常类型: timeout, item_mismatch, operator_insufficient"),
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """获取异常处理预案"""
    client = get_client(server)
    try:
        plan = client.get_exception_plan(exception_type)
        console.print(Panel(f"异常处理预案 - {plan.exception_type}", style="cyan"))
        console.print(f"[bold]严重程度:[/bold] {plan.severity}")
        console.print(f"[bold]负责角色:[/bold] {plan.responsible_role}")
        console.print("\n[bold]触发条件:[/bold]")
        for condition in plan.trigger_conditions:
            console.print(f"  • {condition}")
        console.print("\n[bold]响应步骤:[/bold]")
        for step in plan.response_steps:
            console.print(f"  {step}")
        console.print("\n[bold]升级路径:[/bold]")
        for role in plan.escalation_path:
            console.print(f"  → {role}")
    except Exception as e:
        console.print(f"[red]获取预案失败: {e}[/red]")
        raise typer.Exit(code=1)


@app.command()
def health(
    server: str = typer.Option(DEFAULT_SERVER_URL, "--server", "-s", help="服务端URL"),
) -> None:
    """检查服务健康状态"""
    client = get_client(server)
    try:
        result = client.health_check()
        console.print(f"[green]服务状态: {result['status']}[/green]")
        console.print(f"[cyan]时间戳: {result['timestamp']}[/cyan]")
    except Exception as e:
        console.print(f"[red]健康检查失败: {e}[/red]")
        raise typer.Exit(code=1)


def _display_order(order: CrossDockOrderResponse) -> None:
    status_style = _get_status_style(order.status.value)
    console.print(Panel(f"越库单详情", style=status_style))
    table = Table(show_header=False)
    table.add_column("字段", style="cyan")
    table.add_column("值", style="green")
    table.add_row("ID", order.id)
    table.add_row("单号", order.order_number)
    table.add_row("出库单号", order.outbound_order_number or "-")
    table.add_row("状态", f"[{status_style}]{order.status.value}[/{status_style}]")
    table.add_row("仓库", order.warehouse_id)
    table.add_row("目的地", order.destination)
    table.add_row("入库操作员", order.inbound_operator_id or "-")
    table.add_row("出库操作员", order.outbound_operator_id or "-")
    table.add_row("入库时间", str(order.inbound_time) if order.inbound_time else "-")
    table.add_row("出库时间", str(order.outbound_time) if order.outbound_time else "-")
    table.add_row("创建时间", str(order.created_at))
    table.add_row("是否跨日", f"[red]是[/red]" if order.is_cross_day else "否")
    table.add_row(
        "停留时长(分钟)",
        f"{order.duration_minutes:.1f}" if order.duration_minutes else "-",
    )
    console.print(table)

    if order.expected_items:
        console.print("\n[bold]期望商品:[/bold]")
        item_table = Table()
        item_table.add_column("SKU", style="cyan")
        item_table.add_column("数量", style="green")
        for item in order.expected_items:
            item_table.add_row(item.sku, str(item.quantity))
        console.print(item_table)

    if order.inbound_items:
        console.print("\n[bold]入库商品:[/bold]")
        inbound_table = Table()
        inbound_table.add_column("SKU", style="cyan")
        inbound_table.add_column("数量", style="green")
        inbound_table.add_column("批次号", style="yellow")
        for i in range(len(order.inbound_items)):
            item = order.inbound_items[i]  # type: ignore[assignment]
            inbound_table.add_row(
                item.sku, str(item.quantity), item.batch_number or "-"  # type: ignore[attr-defined]
            )
        console.print(inbound_table)

    if order.alerts:
        console.print("\n[red][bold]告警列表:[/bold][/red]")
        for alert in order.alerts:
            console.print(f"  • {alert}")


def _get_status_style(status: str) -> str:
    styles: dict[str, str] = {
        "pending_inbound": "yellow",
        "inbound_completed": "cyan",
        "outbound_completed": "green",
        "timeout_alert": "red",
    }
    return styles.get(status, "white")


if __name__ == "__main__":
    app()
