from decimal import Decimal
from typing import Any

from rich.console import Console
from rich.table import Table
from rich.box import ROUNDED
from rich.panel import Panel

console = Console()


def format_decimal(value: Decimal | str | None) -> str:
    if value is None:
        return "-"
    if isinstance(value, str):
        value = Decimal(value)
    return f"{value:,.2f}"


def format_quantity(value: int | None) -> str:
    if value is None:
        return "-"
    return f"{value:,}"


def print_warehouses(warehouses: list[dict[str, Any]]) -> None:
    if not warehouses:
        console.print("[yellow]没有找到仓库数据[/yellow]")
        return

    table = Table(
        title="仓库列表",
        box=ROUNDED,
        header_style="bold blue",
    )
    table.add_column("ID", style="dim")
    table.add_column("名称")
    table.add_column("位置")
    table.add_column("公司ID", style="dim")

    for wh in warehouses:
        table.add_row(
            str(wh["warehouse_id"])[:8] + "...",
            wh["name"],
            wh["location"],
            str(wh["company_id"])[:8] + "...",
        )

    console.print(table)


def print_products(products: list[dict[str, Any]]) -> None:
    if not products:
        console.print("[yellow]没有找到商品数据[/yellow]")
        return

    table = Table(
        title="商品列表",
        box=ROUNDED,
        header_style="bold blue",
    )
    table.add_column("ID", style="dim")
    table.add_column("SKU")
    table.add_column("名称")
    table.add_column("单价", justify="right")
    table.add_column("单位")

    for p in products:
        table.add_row(
            str(p["product_id"])[:8] + "...",
            p["sku"],
            p["name"],
            format_decimal(p["unit_price"]),
            p["unit"],
        )

    console.print(table)


def print_inventory_overview(response: dict[str, Any]) -> None:
    inventories = response.get("inventories", [])
    if not inventories:
        console.print("[yellow]没有找到库存数据[/yellow]")
        return

    table = Table(
        title=f"库存总览 (共 {response.get('total', 0)} 条)",
        box=ROUNDED,
        header_style="bold blue",
    )
    table.add_column("仓库")
    table.add_column("商品SKU")
    table.add_column("商品名称")
    table.add_column("可用数量", justify="right")
    table.add_column("冻结数量", justify="right")
    table.add_column("总数量", justify="right")
    table.add_column("可用金额", justify="right")
    table.add_column("冻结金额", justify="right")

    for inv in inventories:
        table.add_row(
            inv.get("warehouse_name", "-"),
            inv.get("product_sku", "-"),
            inv.get("product_name", "-"),
            format_quantity(inv.get("available_quantity")),
            f"[yellow]{format_quantity(inv.get('frozen_quantity'))}[/yellow]",
            format_quantity(inv.get("total_quantity")),
            format_decimal(inv.get("available_amount")),
            f"[yellow]{format_decimal(inv.get('frozen_amount'))}[/yellow]",
        )

    console.print(table)


def print_transfer_detail(transfer: dict[str, Any]) -> None:
    status_colors = {
        "pending_confirm": "yellow",
        "confirmed": "blue",
        "in_transit": "cyan",
        "arrived": "magenta",
        "stocked": "green",
        "cancelled": "red",
    }
    status = transfer.get("status", "unknown")
    status_color = status_colors.get(status, "white")

    status_display = {
        "pending_confirm": "待确认",
        "confirmed": "已确认",
        "in_transit": "运输中",
        "arrived": "已到达",
        "stocked": "已入库",
        "cancelled": "已取消",
    }

    info_lines = [
        f"调拨单号: [bold]{transfer.get('transfer_no')}[/bold]",
        f"状态: [{status_color}]{status_display.get(status, status)}[/{status_color}]",
        f"源仓库: {transfer.get('source_warehouse_name')}",
        f"目标仓库: {transfer.get('target_warehouse_name')}",
        f"调拨类型: {'公司内调拨' if transfer.get('transfer_type') == 'intra_company' else '跨公司调拨'}",
        f"需要审批: {transfer.get('required_approval_level', 'none')}",
        f"已审批: {'是' if transfer.get('is_approved') else '否'}",
        "",
        f"申请金额: {format_decimal(transfer.get('total_requested_amount'))}",
        f"发出金额: {format_decimal(transfer.get('total_shipped_amount'))}",
        f"到达金额: {format_decimal(transfer.get('total_arrived_amount'))}",
        f"损耗金额: [red]{format_decimal(transfer.get('total_loss_amount'))}[/red]",
        "",
        f"创建时间: {transfer.get('created_at')}",
    ]

    if transfer.get("confirmed_at"):
        info_lines.append(f"确认时间: {transfer.get('confirmed_at')}")
    if transfer.get("shipped_at"):
        info_lines.append(f"发货时间: {transfer.get('shipped_at')}")
    if transfer.get("arrived_at"):
        info_lines.append(f"到达时间: {transfer.get('arrived_at')}")
    if transfer.get("stocked_at"):
        info_lines.append(f"入库时间: {transfer.get('stocked_at')}")

    console.print(Panel("\n".join(info_lines), title="调拨单详情", border_style="blue"))

    items = transfer.get("items", [])
    if items:
        item_table = Table(box=ROUNDED, header_style="bold blue")
        item_table.add_column("商品ID", style="dim")
        item_table.add_column("申请数量", justify="right")
        item_table.add_column("发出数量", justify="right")
        item_table.add_column("到达数量", justify="right")
        item_table.add_column("损耗数量", justify="right")
        item_table.add_column("单价", justify="right")
        item_table.add_column("申请金额", justify="right")

        for item in items:
            loss_qty = item.get("loss_quantity", 0)
            loss_str = f"[red]{format_quantity(loss_qty)}[/red]" if loss_qty > 0 else "-"

            item_table.add_row(
                str(item.get("product_id"))[:8] + "...",
                format_quantity(item.get("requested_quantity")),
                format_quantity(item.get("shipped_quantity")),
                format_quantity(item.get("arrived_quantity")),
                loss_str,
                format_decimal(item.get("unit_price")),
                format_decimal(item.get("requested_amount")),
            )

        console.print(item_table)


def print_transfer_list(response: dict[str, Any]) -> None:
    transfers = response.get("transfers", [])
    if not transfers:
        console.print("[yellow]没有找到调拨单[/yellow]")
        return

    status_display = {
        "pending_confirm": "待确认",
        "confirmed": "已确认",
        "in_transit": "运输中",
        "arrived": "已到达",
        "stocked": "已入库",
        "cancelled": "已取消",
    }

    status_colors = {
        "pending_confirm": "yellow",
        "confirmed": "blue",
        "in_transit": "cyan",
        "arrived": "magenta",
        "stocked": "green",
        "cancelled": "red",
    }

    table = Table(
        title=f"调拨单列表 (共 {response.get('total', 0)} 条, 第 {response.get('page', 1)} 页)",
        box=ROUNDED,
        header_style="bold blue",
    )
    table.add_column("调拨单号")
    table.add_column("状态")
    table.add_column("源仓库")
    table.add_column("目标仓库")
    table.add_column("申请金额", justify="right")
    table.add_column("损耗金额", justify="right")
    table.add_column("创建时间")

    for t in transfers:
        status = t.get("status", "unknown")
        color = status_colors.get(status, "white")
        display = status_display.get(status, status)

        table.add_row(
            t.get("transfer_no"),
            f"[{color}]{display}[/{color}]",
            t.get("source_warehouse_name", "-"),
            t.get("target_warehouse_name", "-"),
            format_decimal(t.get("total_requested_amount")),
            f"[red]{format_decimal(t.get('total_loss_amount'))}[/red]"
            if t.get("total_loss_amount") and Decimal(str(t.get("total_loss_amount"))) > 0
            else "-",
            t.get("created_at", "-"),
        )

    console.print(table)


def print_loss_list(response: dict[str, Any]) -> None:
    losses = response.get("losses", [])
    if not losses:
        console.print("[yellow]没有找到损耗记录[/yellow]")
        return

    table = Table(
        title=f"损耗记录 (共 {response.get('total', 0)} 条)",
        box=ROUNDED,
        header_style="bold red",
    )
    table.add_column("调拨单号")
    table.add_column("仓库")
    table.add_column("商品")
    table.add_column("损耗数量", justify="right")
    table.add_column("损耗金额", justify="right")
    table.add_column("损耗日期")

    for loss in losses:
        table.add_row(
            loss.get("transfer_no", "-"),
            loss.get("warehouse_name", "-"),
            loss.get("product_name", "-"),
            f"[red]{format_quantity(loss.get('loss_quantity'))}[/red]",
            f"[red]{format_decimal(loss.get('loss_amount'))}[/red]",
            loss.get("loss_date", "-"),
        )

    console.print(table)


def print_loss_summary(response: dict[str, Any]) -> None:
    summaries = response.get("summaries", [])
    if not summaries:
        console.print("[yellow]没有找到损耗统计数据[/yellow]")
        return

    table = Table(
        title="月度损耗统计",
        box=ROUNDED,
        header_style="bold magenta",
    )
    table.add_column("年月")
    table.add_column("仓库ID", style="dim")
    table.add_column("调拨总数量", justify="right")
    table.add_column("调拨总金额", justify="right")
    table.add_column("损耗数量", justify="right")
    table.add_column("损耗金额", justify="right")
    table.add_column("损耗率", justify="right")

    for s in summaries:
        year = s.get("year")
        month = s.get("month")
        loss_rate = Decimal(str(s.get("loss_rate", 0)))
        rate_color = "red" if loss_rate > Decimal("1") else "yellow" if loss_rate > Decimal("0.5") else "green"

        table.add_row(
            f"{year}-{month:02d}",
            str(s.get("warehouse_id"))[:8] + "...",
            format_quantity(s.get("total_transfer_quantity")),
            format_decimal(s.get("total_transfer_amount")),
            f"[red]{format_quantity(s.get('total_loss_quantity'))}[/red]",
            f"[red]{format_decimal(s.get('total_loss_amount'))}[/red]",
            f"[{rate_color}]{loss_rate:.2f}%[/{rate_color}]",
        )

    console.print(table)


def print_transaction_list(response: dict[str, Any]) -> None:
    transactions = response.get("transactions", [])
    if not transactions:
        console.print("[yellow]没有找到库存变动流水[/yellow]")
        return

    type_display = {
        "transfer_out": "调出",
        "transfer_in": "调入",
        "freeze": "冻结",
        "unfreeze": "解冻",
        "loss": "损耗",
    }

    type_colors = {
        "transfer_out": "red",
        "transfer_in": "green",
        "freeze": "yellow",
        "unfreeze": "blue",
        "loss": "red",
    }

    table = Table(
        title=f"库存变动流水 (共 {response.get('total', 0)} 条)",
        box=ROUNDED,
        header_style="bold blue",
    )
    table.add_column("仓库")
    table.add_column("商品")
    table.add_column("类型")
    table.add_column("数量", justify="right")
    table.add_column("金额", justify="right")
    table.add_column("时间")

    for tx in transactions:
        tx_type = tx.get("transaction_type", "unknown")
        display = type_display.get(tx_type, tx_type)
        color = type_colors.get(tx_type, "white")

        qty = tx.get("quantity", 0)
        if tx_type in ["transfer_out", "freeze", "loss"]:
            qty_str = f"[red]-{qty}[/red]"
        elif tx_type in ["transfer_in", "unfreeze"]:
            qty_str = f"[green]+{qty}[/green]"
        else:
            qty_str = str(qty)

        table.add_row(
            tx.get("warehouse_name", "-"),
            tx.get("product_name", "-"),
            f"[{color}]{display}[/{color}]",
            qty_str,
            format_decimal(tx.get("amount")),
            tx.get("transaction_time", "-"),
        )

    console.print(table)


def print_print_document(response: dict[str, Any]) -> None:
    console.print(f"\n[bold]文档类型: {response.get('document_type')}[/bold]")
    console.print(f"调拨单号: {response.get('transfer_no')}")
    console.print(f"生成时间: {response.get('generated_at')}\n")
    console.print(response.get("content", ""))


def print_error(message: str, details: dict[str, Any] | None = None) -> None:
    console.print(f"[bold red]错误:[/bold red] {message}")
    if details:
        for key, value in details.items():
            console.print(f"  [dim]{key}:[/dim] {value}")


def print_success(message: str) -> None:
    console.print(f"[bold green]成功:[/bold green] {message}")
