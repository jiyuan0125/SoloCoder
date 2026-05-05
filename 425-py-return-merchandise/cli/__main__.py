from datetime import datetime, timedelta
from typing import Optional

import typer
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.text import Text

from shared.models import (
    OutboundOrderCreate,
    OutboundOrderItem,
    ReturnApplyRequest,
    ReturnOrderItem,
    WarehouseReceiveRequest,
    InspectionRequest,
    StockInRequest,
    DefectiveToNormalRequest,
    InspectionResult,
    DisposalReason,
)
from cli.client import APIClient

app = typer.Typer(
    name="rma",
    help="退货入库管理系统命令行工具",
    add_completion=False,
)
console = Console()


def get_client() -> APIClient:
    return APIClient()


def format_datetime(dt: Optional[datetime]) -> str:
    if dt is None:
        return "-"
    return dt.strftime("%Y-%m-%d %H:%M:%S")


def format_status(status: str) -> str:
    status_colors = {
        "pending": "yellow",
        "shipped": "blue",
        "delivered": "cyan",
        "completed": "green",
        "applied": "yellow",
        "warehouse_received": "blue",
        "inspected": "cyan",
        "stocked_in": "green",
        "returned_to_customer": "red",
        "expired": "red",
        "cancelled": "magenta",
    }
    color = status_colors.get(status, "white")
    return f"[{color}]{status}[/{color}]"


@app.command()
def health() -> None:
    """检查服务器健康状态"""
    try:
        with get_client() as client:
            response = client._make_request("GET", "/health")
            data = response.json()
            console.print(f"[green]服务器状态: {data['status']}[/green]")
    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


outbound_app = typer.Typer(name="outbound", help="出库单管理")
app.add_typer(outbound_app)


@outbound_app.command("create")
def create_outbound(
    order_id: str = typer.Option(..., "--order-id", "-o", help="出库单ID"),
    customer_id: str = typer.Option(..., "--customer-id", "-c", help="客户ID"),
    items: str = typer.Option(
        ...,
        "--items",
        "-i",
        help="商品列表，格式: sku:name:qty:price:category,sku2:name2:qty2:price2:...",
    ),
) -> None:
    """创建出库单"""
    try:
        item_list: list[OutboundOrderItem] = []
        for item_str in items.split(","):
            parts = item_str.split(":")
            if len(parts) < 4:
                raise ValueError(f"商品格式错误: {item_str}，需要 sku:name:qty:price:category")
            sku = parts[0].strip()
            name = parts[1].strip()
            qty = int(parts[2].strip())
            price = float(parts[3].strip())
            category = parts[4].strip() if len(parts) > 4 else ""
            item_list.append(
                OutboundOrderItem(
                    sku=sku,
                    product_name=name,
                    quantity=qty,
                    unit_price=price,
                    category=category,
                )
            )

        request = OutboundOrderCreate(
            order_id=order_id,
            customer_id=customer_id,
            items=item_list,
        )

        with get_client() as client:
            result = client.create_outbound_order(request)

        table = Table(title="出库单创建成功")
        table.add_column("字段", style="cyan")
        table.add_column("值", style="green")
        table.add_row("订单ID", result.order_id)
        table.add_row("客户ID", result.customer_id)
        table.add_row("状态", format_status(result.status.value))
        table.add_row("总金额", f"¥{result.total_amount:.2f}")
        table.add_row("创建时间", format_datetime(result.created_at))
        console.print(table)

        items_table = Table(title="商品明细")
        items_table.add_column("SKU")
        items_table.add_column("商品名称")
        items_table.add_column("数量")
        items_table.add_column("单价")
        items_table.add_column("类别")
        for item in result.items:
            items_table.add_row(
                item.sku,
                item.product_name,
                str(item.quantity),
                f"¥{item.unit_price:.2f}",
                item.category,
            )
        console.print(items_table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@outbound_app.command("list")
def list_outbound() -> None:
    """列出所有出库单"""
    try:
        with get_client() as client:
            orders = client.list_outbound_orders()

        if not orders:
            console.print("[yellow]暂无出库单[/yellow]")
            return

        table = Table(title="出库单列表")
        table.add_column("订单ID", style="cyan")
        table.add_column("客户ID")
        table.add_column("状态")
        table.add_column("总金额")
        table.add_column("创建时间")
        for order in orders:
            table.add_row(
                order.order_id,
                order.customer_id,
                format_status(order.status.value),
                f"¥{order.total_amount:.2f}",
                format_datetime(order.created_at),
            )
        console.print(table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@outbound_app.command("get")
def get_outbound(order_id: str = typer.Argument(..., help="出库单ID")) -> None:
    """获取出库单详情"""
    try:
        with get_client() as client:
            order = client.get_outbound_order(order_id)

        table = Table(title=f"出库单详情: {order.order_id}")
        table.add_column("字段", style="cyan")
        table.add_column("值", style="green")
        table.add_row("订单ID", order.order_id)
        table.add_row("客户ID", order.customer_id)
        table.add_row("状态", format_status(order.status.value))
        table.add_row("总金额", f"¥{order.total_amount:.2f}")
        table.add_row("创建时间", format_datetime(order.created_at))
        table.add_row("更新时间", format_datetime(order.updated_at))
        table.add_row("发货时间", format_datetime(order.shipped_at))
        table.add_row("签收时间", format_datetime(order.delivered_at))
        console.print(table)

        items_table = Table(title="商品明细")
        items_table.add_column("SKU")
        items_table.add_column("商品名称")
        items_table.add_column("数量")
        items_table.add_column("单价")
        items_table.add_column("类别")
        for item in order.items:
            items_table.add_row(
                item.sku,
                item.product_name,
                str(item.quantity),
                f"¥{item.unit_price:.2f}",
                item.category,
            )
        console.print(items_table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@outbound_app.command("ship")
def ship_outbound(order_id: str = typer.Argument(..., help="出库单ID")) -> None:
    """标记出库单已发货"""
    try:
        with get_client() as client:
            order = client.ship_outbound_order(order_id)
        console.print(f"[green]出库单 {order_id} 已标记为已发货[/green]")
        console.print(f"当前状态: {format_status(order.status.value)}")
    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@outbound_app.command("deliver")
def deliver_outbound(order_id: str = typer.Argument(..., help="出库单ID")) -> None:
    """标记出库单已签收"""
    try:
        with get_client() as client:
            order = client.deliver_outbound_order(order_id)
        console.print(f"[green]出库单 {order_id} 已标记为已签收[/green]")
        console.print(f"当前状态: {format_status(order.status.value)}")
    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@outbound_app.command("complete")
def complete_outbound(order_id: str = typer.Argument(..., help="出库单ID")) -> None:
    """标记出库单已完成"""
    try:
        with get_client() as client:
            order = client.complete_outbound_order(order_id)
        console.print(f"[green]出库单 {order_id} 已标记为已完成[/green]")
        console.print(f"当前状态: {format_status(order.status.value)}")
    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


return_app = typer.Typer(name="return", help="退货管理")
app.add_typer(return_app)


@return_app.command("apply")
def apply_return(
    return_id: str = typer.Option(..., "--return-id", "-r", help="退货单ID"),
    outbound_id: str = typer.Option(..., "--outbound-id", "-o", help="原出库单ID"),
    reason: str = typer.Option(..., "--reason", "-n", help="退货原因: quality_issue/size_mismatch/dislike/wrong_item/other"),
    items: str = typer.Option(
        ...,
        "--items",
        "-i",
        help="退货商品列表，格式: sku:qty,sku2:qty2:...",
    ),
    note: str = typer.Option("", "--note", "-N", help="客户备注"),
) -> None:
    """申请退货"""
    try:
        reason_enum = DisposalReason(reason)

        item_list: list[ReturnOrderItem] = []
        for item_str in items.split(","):
            parts = item_str.split(":")
            if len(parts) != 2:
                raise ValueError(f"商品格式错误: {item_str}，需要 sku:qty")
            sku = parts[0].strip()
            qty = int(parts[1].strip())
            item_list.append(
                ReturnOrderItem(
                    sku=sku,
                    product_name="",
                    quantity=qty,
                    original_unit_price=0.0,
                )
            )

        request = ReturnApplyRequest(
            return_order_id=return_id,
            outbound_order_id=outbound_id,
            items=item_list,
            reason=reason_enum,
            customer_note=note,
        )

        with get_client() as client:
            result = client.apply_return(request)

        table = Table(title="退货申请提交成功")
        table.add_column("字段", style="cyan")
        table.add_column("值", style="green")
        table.add_row("退货单ID", result.return_order_id)
        table.add_row("原出库单ID", result.outbound_order_id)
        table.add_row("状态", format_status(result.status.value))
        table.add_row("创建时间", format_datetime(result.created_at))
        table.add_row("有效期至", format_datetime(result.expiry_date))
        console.print(table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@return_app.command("receive")
def receive_return(
    return_id: str = typer.Argument(..., help="退货单ID"),
    receiver: str = typer.Option(..., "--receiver", "-r", help="收货人员"),
) -> None:
    """仓库收货"""
    try:
        request = WarehouseReceiveRequest(
            return_order_id=return_id,
            received_by=receiver,
        )

        with get_client() as client:
            result = client.warehouse_receive(request)

        console.print(f"[green]退货单 {return_id} 已确认收货[/green]")
        console.print(f"当前状态: {format_status(result.status.value)}")

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@return_app.command("inspect")
def inspect_return(
    return_id: str = typer.Argument(..., help="退货单ID"),
    inspector: str = typer.Option(..., "--inspector", "-i", help="质检人员"),
    results: str = typer.Option(
        ...,
        "--results",
        "-r",
        help="质检结果，格式: sku:good/sku:minor_defect/sku:severe_damage",
    ),
) -> None:
    """质检"""
    try:
        result_map: dict[str, InspectionResult] = {}
        for item_str in results.split(","):
            parts = item_str.split(":")
            if len(parts) != 2:
                raise ValueError(f"质检结果格式错误: {item_str}，需要 sku:result")
            sku = parts[0].strip()
            result = InspectionResult(parts[1].strip())
            result_map[sku] = result

        request = InspectionRequest(
            return_order_id=return_id,
            inspector=inspector,
            item_results=result_map,
        )

        with get_client() as client:
            result = client.inspect_return(request)

        console.print(f"[green]退货单 {return_id} 质检完成[/green]")
        console.print(f"当前状态: {format_status(result.status.value)}")

        for item in result.items:
            result_text = (
                "[green]完好[/green]" if item.inspection_result == InspectionResult.GOOD
                else "[yellow]轻微瑕疵[/yellow]" if item.inspection_result == InspectionResult.MINOR_DEFECT
                else "[red]严重损坏[/red]"
            )
            price_text = f" (处理价: ¥{item.disposal_price:.2f})" if item.disposal_price else ""
            console.print(f"  {item.sku}: {result_text}{price_text}")

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@return_app.command("stock-in")
def stock_in_return(
    return_id: str = typer.Argument(..., help="退货单ID"),
    operator: str = typer.Option(..., "--operator", "-o", help="入库操作人员"),
) -> None:
    """入库"""
    try:
        request = StockInRequest(
            return_order_id=return_id,
            stock_in_by=operator,
        )

        with get_client() as client:
            result = client.stock_in_return(request)

        console.print(f"[green]退货单 {return_id} 入库完成[/green]")
        console.print(f"当前状态: {format_status(result.status.value)}")

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@return_app.command("list")
def list_returns() -> None:
    """列出所有退货单"""
    try:
        with get_client() as client:
            returns = client.list_return_orders()

        if not returns:
            console.print("[yellow]暂无退货单[/yellow]")
            return

        table = Table(title="退货单列表")
        table.add_column("退货单ID", style="cyan")
        table.add_column("原出库单ID")
        table.add_column("状态")
        table.add_column("创建时间")
        table.add_column("有效期至")
        for ret in returns:
            table.add_row(
                ret.return_order_id,
                ret.outbound_order_id,
                format_status(ret.status.value),
                format_datetime(ret.created_at),
                format_datetime(ret.expiry_date),
            )
        console.print(table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@return_app.command("get")
def get_return(return_id: str = typer.Argument(..., help="退货单ID")) -> None:
    """获取退货单详情"""
    try:
        with get_client() as client:
            ret = client.get_return_order(return_id)

        table = Table(title=f"退货单详情: {ret.return_order_id}")
        table.add_column("字段", style="cyan")
        table.add_column("值", style="green")
        table.add_row("退货单ID", ret.return_order_id)
        table.add_row("原出库单ID", ret.outbound_order_id)
        table.add_row("状态", format_status(ret.status.value))
        table.add_row("退货原因", ret.reason.value)
        table.add_row("客户备注", ret.customer_note or "-")
        table.add_row("创建时间", format_datetime(ret.created_at))
        table.add_row("有效期至", format_datetime(ret.expiry_date))
        table.add_row("收货时间", format_datetime(ret.warehouse_received_at))
        table.add_row("收货人", ret.warehouse_received_by or "-")
        table.add_row("质检时间", format_datetime(ret.inspected_at))
        table.add_row("质检员", ret.inspected_by or "-")
        table.add_row("入库时间", format_datetime(ret.stocked_in_at))
        table.add_row("入库人", ret.stocked_in_by or "-")
        console.print(table)

        items_table = Table(title="商品明细")
        items_table.add_column("SKU")
        items_table.add_column("商品名称")
        items_table.add_column("数量")
        items_table.add_column("原单价")
        items_table.add_column("质检结果")
        items_table.add_column("处理价")
        for item in ret.items:
            result_text = item.inspection_result.value if item.inspection_result else "-"
            price_text = f"¥{item.disposal_price:.2f}" if item.disposal_price else "-"
            items_table.add_row(
                item.sku,
                item.product_name,
                str(item.quantity),
                f"¥{item.original_unit_price:.2f}",
                result_text,
                price_text,
            )
        console.print(items_table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


defective_app = typer.Typer(name="defective", help="瑕疵品管理")
app.add_typer(defective_app)


@defective_app.command("list")
def list_defective() -> None:
    """列出所有瑕疵品"""
    try:
        with get_client() as client:
            items = client.list_defective_items()

        if not items:
            console.print("[yellow]暂无瑕疵品[/yellow]")
            return

        table = Table(title="瑕疵品列表")
        table.add_column("瑕疵品ID", style="cyan")
        table.add_column("SKU")
        table.add_column("商品名称")
        table.add_column("数量")
        table.add_column("原单价")
        table.add_column("瑕疵价")
        table.add_column("状态")
        for item in items:
            status_text = "[green]已转正[/green]" if item.converted_to_normal else "[yellow]瑕疵品[/yellow]"
            table.add_row(
                item.defective_id,
                item.sku,
                item.product_name,
                str(item.quantity),
                f"¥{item.unit_price:.2f}",
                f"¥{item.defective_price:.2f}",
                status_text,
            )
        console.print(table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@defective_app.command("get")
def get_defective(defective_id: str = typer.Argument(..., help="瑕疵品ID")) -> None:
    """获取瑕疵品详情"""
    try:
        with get_client() as client:
            item = client.get_defective_item(defective_id)

        table = Table(title=f"瑕疵品详情: {item.defective_id}")
        table.add_column("字段", style="cyan")
        table.add_column("值", style="green")
        table.add_row("瑕疵品ID", item.defective_id)
        table.add_row("SKU", item.sku)
        table.add_row("商品名称", item.product_name)
        table.add_row("数量", str(item.quantity))
        table.add_row("原单价", f"¥{item.unit_price:.2f}")
        table.add_row("瑕疵价", f"¥{item.defective_price:.2f}")
        table.add_row("类别", item.category)
        table.add_row("原退货单", item.return_order_id)
        table.add_row("创建时间", format_datetime(item.created_at))
        table.add_row("是否已转正", "是" if item.converted_to_normal else "否")
        table.add_row("转正时间", format_datetime(item.converted_at))
        table.add_row("转正质检人", item.converted_by or "-")
        console.print(table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@defective_app.command("convert")
def convert_defective(
    defective_id: str = typer.Argument(..., help="瑕疵品ID"),
    inspector: str = typer.Option(..., "--inspector", "-i", help="二次质检人员"),
) -> None:
    """瑕疵品转正品（二次质检合格）"""
    try:
        request = DefectiveToNormalRequest(
            defective_id=defective_id,
            inspector=inspector,
        )

        with get_client() as client:
            result = client.convert_defective_to_normal(request)

        console.print(f"[green]瑕疵品 {defective_id} 已转为正品[/green]")
        console.print(f"质检人: {inspector}")

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


stats_app = typer.Typer(name="stats", help="统计分析")
app.add_typer(stats_app)


@stats_app.command("returns")
def stats_returns(
    days: int = typer.Option(30, "--days", "-d", help="统计天数"),
    category: Optional[str] = typer.Option(None, "--category", "-c", help="商品类别（可选）"),
) -> None:
    """退货统计（完好率、报废率）"""
    try:
        end_date = datetime.now()
        start_date = end_date - timedelta(days=days)

        with get_client() as client:
            result = client.get_return_statistics(start_date, end_date, category)

        table = Table(title=f"退货统计（最近{days}天）")
        table.add_column("指标", style="cyan")
        table.add_column("数值", style="green")
        table.add_row("退货单总数", str(result.total_returns))
        table.add_row("退货商品总数", str(result.total_quantity))
        table.add_row("完好数量", str(result.good_quantity))
        table.add_row("瑕疵品数量", str(result.defective_quantity))
        table.add_row("报废数量", str(result.scrap_quantity))
        table.add_row("完好率", f"{result.good_rate:.2%}")
        table.add_row("报废率", f"{result.scrap_rate:.2%}")
        console.print(table)

        if result.category:
            console.print(f"\n[cyan]筛选类别: {result.category}[/cyan]")

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@stats_app.command("scrap-ledger")
def stats_scrap_ledger(
    year: int = typer.Option(None, "--year", "-y", help="年份（默认当前年）"),
    month: int = typer.Option(None, "--month", "-m", help="月份（默认当前月）"),
    reason: Optional[str] = typer.Option(None, "--reason", "-r", help="退货原因过滤（可选）"),
) -> None:
    """报废台账（按月汇总）"""
    try:
        now = datetime.now()
        if year is None:
            year = now.year
        if month is None:
            month = now.month

        with get_client() as client:
            result = client.get_scrap_ledger(year, month, reason)

        table = Table(title=f"报废台账 ({year}年{month}月)")
        table.add_column("指标", style="cyan")
        table.add_column("数值", style="green")
        table.add_row("报废记录数", str(result.total_records))
        table.add_row("报废数量", str(result.total_quantity))
        table.add_row("报废总价值", f"¥{result.total_value:.2f}")
        console.print(table)

        if result.breakdown:
            breakdown_table = Table(title="按原因分类")
            breakdown_table.add_column("原因", style="cyan")
            breakdown_table.add_column("数量")
            breakdown_table.add_column("价值")
            for item in result.breakdown:
                breakdown_table.add_row(
                    item.reason.value,
                    str(item.quantity),
                    f"¥{item.total_value:.2f}",
                )
            console.print(breakdown_table)

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


@stats_app.command("quality-alert")
def stats_quality_alert(
    supplier_id: str = typer.Argument(..., help="供应商ID"),
    days: int = typer.Option(30, "--days", "-d", help="统计周期天数"),
    threshold: Optional[float] = typer.Option(None, "--threshold", "-t", help="阈值（默认10%）"),
) -> None:
    """检查供应商质量问题退货率"""
    try:
        with get_client() as client:
            result = client.check_quality_alert(supplier_id, days, threshold)

        alert_color = "red" if result.alert_triggered else "green"
        alert_text = "[bold red]⚠️ 已触发质量整改通知[/bold red]" if result.alert_triggered else "[green]正常[/green]"

        table = Table(title=f"供应商质量检查: {supplier_id}")
        table.add_column("指标", style="cyan")
        table.add_column("数值", style=alert_color)
        table.add_row("统计周期", f"最近{result.period_days}天")
        table.add_row("质量问题退货率", f"{result.actual_rate:.2%}")
        table.add_row("阈值", f"{result.threshold:.2%}")
        table.add_row("状态", alert_text)
        console.print(table)

        console.print(f"\n{result.message}")

    except ValueError as e:
        console.print(f"[red]错误: {e}[/red]")
        raise typer.Exit(1)


if __name__ == "__main__":
    app()
