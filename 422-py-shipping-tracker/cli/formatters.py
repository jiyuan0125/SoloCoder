from datetime import datetime
from typing import Any

from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.tree import Tree

from shared.models import (
    AlertResponse,
    BatchQueryResponse,
    CustomStatus,
    Location,
    NodeType,
    SignedBy,
    TrackingNode,
    Waybill,
    WaybillQueryResponse,
    WaybillSplitResponse,
    WaybillStatus,
)


console = Console()


def format_datetime(dt: datetime | None) -> str:
    if dt is None:
        return "N/A"
    return dt.strftime("%Y-%m-%d %H:%M:%S UTC")


def format_location(location: Location) -> str:
    parts: list[str] = [location.address]
    if location.latitude is not None and location.longitude is not None:
        parts.append(f"({location.latitude:.6f}, {location.longitude:.6f})")
    return " ".join(parts)


def format_signed_by(signed_by: SignedBy | None) -> str:
    if signed_by is None:
        return "N/A"
    if signed_by.is_authorized:
        relation = signed_by.authorized_relation or "未知"
        return f"{signed_by.name} (代签收, {relation})"
    return signed_by.name


def format_waybill_status(status: WaybillStatus, is_abnormal: bool) -> str:
    if is_abnormal:
        return f"[bold red]{str(status.value)} (异常)[/bold red]"
    if status == WaybillStatus.SIGNED:
        return f"[bold green]{str(status.value)}[/bold green]"
    if status == WaybillStatus.IN_TRANSIT:
        return f"[bold blue]{str(status.value)}[/bold blue]"
    return str(status.value)


def format_custom_status(status: CustomStatus | None) -> str:
    if status is None:
        return "N/A (国内运单)"
    status_colors: dict[CustomStatus, str] = {
        CustomStatus.PENDING_DECLARATION: "yellow",
        CustomStatus.DECLARING: "blue",
        CustomStatus.DECLARED: "green",
        CustomStatus.CUSTOMS_CLEARED: "bold green",
        CustomStatus.CUSTOMS_HELD: "bold red",
    }
    color = status_colors.get(status, "white")
    return f"[{color}]{status.value}[/{color}]"


def print_waybill(waybill: Waybill) -> None:
    table = Table(show_header=False, box=None)
    table.add_column("Field", style="bold cyan")
    table.add_column("Value")
    table.add_row("运单号", waybill.waybill_number)
    if waybill.parent_waybill_number:
        table.add_row("父运单号", waybill.parent_waybill_number)
    table.add_row("发件人", waybill.sender)
    table.add_row("收件人", waybill.receiver)
    table.add_row("收件地址", waybill.receiver_address)
    table.add_row(
        "当前状态",
        format_waybill_status(waybill.status, waybill.is_abnormal),
    )
    table.add_row("是否国际运单", "是" if waybill.is_international else "否")
    if waybill.is_international:
        table.add_row("海关状态", format_custom_status(waybill.custom_status))
    table.add_row("创建时间", format_datetime(waybill.created_at))
    table.add_row("更新时间", format_datetime(waybill.updated_at))
    if waybill.last_node_time:
        table.add_row("最后节点时间", format_datetime(waybill.last_node_time))
    table.add_row("节点数量", str(len(waybill.nodes)))
    console.print(Panel(table, title="运单信息", border_style="cyan"))


def print_tracking_nodes(nodes: list[TrackingNode]) -> None:
    if not nodes:
        console.print("[yellow]暂无物流轨迹[/yellow]")
        return
    tree = Tree("[bold green]物流轨迹 (按时间正序)[/bold green]")
    for i, node in enumerate(nodes, 1):
        node_branch = tree.add(
            f"[bold cyan]{i}. {node.node_type.value}[/bold cyan] "
            f"[dim]- {format_datetime(node.timestamp)}[/dim]"
        )
        node_branch.add(f"地点: {format_location(node.location)}")
        node_branch.add(f"操作人: {node.operator.operator_name or node.operator.operator_id}")
        if node.signed_by:
            node_branch.add(f"签收人: {format_signed_by(node.signed_by)}")
    console.print(tree)


def print_waybill_query(response: WaybillQueryResponse) -> None:
    print_waybill(response.waybill)
    console.print()
    print_tracking_nodes(response.tracking_history)


def print_batch_query(response: BatchQueryResponse) -> None:
    if response.results:
        console.print(f"[bold green]成功查询到 {len(response.results)} 个运单[/bold green]")
        for waybill_number, query_response in response.results.items():
            console.print(f"\n{'='*60}")
            print_waybill_query(query_response)
    if response.not_found:
        console.print(f"\n[bold yellow]未找到 {len(response.not_found)} 个运单:[/bold yellow]")
        for number in response.not_found:
            console.print(f"  - {number}")


def print_split_response(response: WaybillSplitResponse) -> None:
    console.print(
        f"[bold green]成功拆分运单 {response.parent_waybill_number}[/bold green]"
    )
    console.print(f"生成 {len(response.sub_waybill_numbers)} 个子运单:")
    for i, number in enumerate(response.sub_waybill_numbers, 1):
        console.print(f"  {i}. [bold cyan]{number}[/bold cyan]")


def print_alerts(alerts: list[AlertResponse]) -> None:
    if not alerts:
        console.print("[bold green]暂无异常告警[/bold green]")
        return
    console.print(f"[bold red]发现 {len(alerts)} 个异常运单[/bold red]")
    table = Table(show_header=True, box=None)
    table.add_column("运单号", style="bold cyan")
    table.add_column("当前状态", style="bold red")
    table.add_column("最后节点时间")
    table.add_column("超时(小时)", justify="right")
    table.add_column("告警时间")
    for alert in alerts:
        table.add_row(
            alert.waybill_number,
            alert.current_status.value,
            format_datetime(alert.last_node_time),
            f"{alert.hours_since_last_node:.1f}",
            format_datetime(alert.alert_time),
        )
    console.print(table)


def print_created_waybill(waybill: Waybill) -> None:
    console.print("[bold green]运单创建成功![/bold green]")
    console.print()
    print_waybill(waybill)


def print_added_node(node: TrackingNode) -> None:
    console.print("[bold green]节点添加成功![/bold green]")
    console.print()
    table = Table(show_header=False, box=None)
    table.add_column("Field", style="bold cyan")
    table.add_column("Value")
    table.add_row("节点类型", node.node_type.value)
    table.add_row("时间", format_datetime(node.timestamp))
    table.add_row("地点", format_location(node.location))
    table.add_row("操作人", node.operator.operator_name or node.operator.operator_id)
    if node.signed_by:
        table.add_row("签收人", format_signed_by(node.signed_by))
    console.print(Panel(table, title="节点信息", border_style="green"))


def print_error(message: str, error_code: int | None = None) -> None:
    if error_code is not None:
        console.print(f"[bold red]错误 [{error_code}]: {message}[/bold red]")
    else:
        console.print(f"[bold red]错误: {message}[/bold red]")


def print_success(message: str) -> None:
    console.print(f"[bold green]{message}[/bold green]")
