import json
from typing import Any, Optional

import click
from rich.console import Console
from rich.table import Table

from shared.models import (
    AlertRecord,
    InventoryHealthReport,
    InventoryTransaction,
    InventoryTurnoverAnalysis,
    Product,
    ReplenishmentOrderDraft,
)


console = Console()


def print_success(message: str) -> None:
    console.print(f"[green]✓ {message}[/green]")


def print_error(message: str) -> None:
    console.print(f"[red]✗ {message}[/red]")


def print_warning(message: str) -> None:
    console.print(f"[yellow]⚠ {message}[/yellow]")


def print_info(message: str) -> None:
    console.print(f"[blue]ℹ {message}[/blue]")


def format_datetime(dt: Any) -> str:
    if hasattr(dt, "isoformat"):
        return str(dt.isoformat())
    return str(dt)


def print_product_table(products: list[Product]) -> None:
    if not products:
        print_info("No products found.")
        return

    table = Table(title="Products")
    table.add_column("SKU", style="cyan")
    table.add_column("Name", style="white")
    table.add_column("Category", style="magenta")
    table.add_column("Stock", style="yellow", justify="right")
    table.add_column("Min", style="red", justify="right")
    table.add_column("Max", style="green", justify="right")
    table.add_column("Seasonal", style="blue")

    for p in products:
        seasonal = "Yes" if p.safety_stock.is_seasonal else "No"
        table.add_row(
            p.sku,
            p.name,
            p.category,
            str(p.current_stock),
            str(p.safety_stock.min_stock),
            str(p.safety_stock.max_stock),
            seasonal,
        )

    console.print(table)


def print_product_detail(product: Product) -> None:
    table = Table(title=f"Product: {product.sku}")
    table.add_column("Field", style="cyan")
    table.add_column("Value", style="white")

    table.add_row("SKU", product.sku)
    table.add_row("Name", product.name)
    table.add_row("Category", product.category)
    table.add_row("Current Stock", str(product.current_stock))
    table.add_row("Is Seasonal", "Yes" if product.safety_stock.is_seasonal else "No")
    table.add_row("Default Min Stock", str(product.safety_stock.min_stock))
    table.add_row("Default Max Stock", str(product.safety_stock.max_stock))
    table.add_row("Created At", format_datetime(product.created_at))
    table.add_row("Updated At", format_datetime(product.updated_at))

    if product.safety_stock.is_seasonal and product.safety_stock.seasonal_config:
        seasonal_table = Table(title="Seasonal Safety Stock")
        seasonal_table.add_column("Month", style="cyan")
        seasonal_table.add_column("Min", style="red")
        seasonal_table.add_column("Max", style="green")

        for month, config in sorted(product.safety_stock.seasonal_config.items()):
            seasonal_table.add_row(month.name, str(config.min_stock), str(config.max_stock))

        console.print(table)
        console.print(seasonal_table)
    else:
        console.print(table)


def print_alert_table(alerts: list[AlertRecord]) -> None:
    if not alerts:
        print_info("No alerts found.")
        return

    table = Table(title="Alerts")
    table.add_column("ID", style="cyan")
    table.add_column("SKU", style="white")
    table.add_column("Type", style="magenta")
    table.add_column("Level", style="yellow")
    table.add_column("Stock", style="blue", justify="right")
    table.add_column("Suggestion", style="green", justify="right")
    table.add_column("Resolved", style="cyan")
    table.add_column("Triggered At", style="white")

    for a in alerts:
        resolved = "Yes" if a.resolved else "No"
        table.add_row(
            a.id[:8] + "...",
            a.sku,
            a.alert_type.value,
            a.alert_level.value,
            str(a.current_stock),
            str(a.suggestion_quantity),
            resolved,
            format_datetime(a.triggered_at),
        )

    console.print(table)


def print_alert_detail(alert: AlertRecord) -> None:
    table = Table(title=f"Alert: {alert.id}")
    table.add_column("Field", style="cyan")
    table.add_column("Value", style="white")

    table.add_row("ID", alert.id)
    table.add_row("SKU", alert.sku)
    table.add_row("Type", alert.alert_type.value)
    table.add_row("Level", alert.alert_level.value)
    table.add_row("Current Stock", str(alert.current_stock))
    table.add_row("Min Stock", str(alert.min_stock))
    table.add_row("Max Stock", str(alert.max_stock))
    table.add_row("Suggestion Quantity", str(alert.suggestion_quantity))
    table.add_row("Triggered At", format_datetime(alert.triggered_at))
    table.add_row("Resolved", "Yes" if alert.resolved else "No")
    if alert.resolved_at:
        table.add_row("Resolved At", format_datetime(alert.resolved_at))

    console.print(table)


def print_transaction_table(transactions: list[InventoryTransaction]) -> None:
    if not transactions:
        print_info("No transactions found.")
        return

    table = Table(title="Inventory Transactions")
    table.add_column("ID", style="cyan")
    table.add_column("SKU", style="white")
    table.add_column("Type", style="magenta")
    table.add_column("Qty", style="yellow", justify="right")
    table.add_column("Previous", style="blue", justify="right")
    table.add_column("New", style="green", justify="right")
    table.add_column("Time", style="white")

    for t in transactions:
        table.add_row(
            t.id[:8] + "...",
            t.sku,
            t.transaction_type.value,
            str(t.quantity),
            str(t.previous_stock),
            str(t.new_stock),
            format_datetime(t.created_at),
        )

    console.print(table)


def print_replenishment_table(drafts: list[ReplenishmentOrderDraft]) -> None:
    if not drafts:
        print_info("No replenishment drafts found.")
        return

    table = Table(title="Replenishment Order Drafts")
    table.add_column("ID", style="cyan")
    table.add_column("SKU", style="white")
    table.add_column("Suggested Qty", style="yellow", justify="right")
    table.add_column("Confirmed", style="magenta")
    table.add_column("Created At", style="white")

    for d in drafts:
        confirmed = "Yes" if d.confirmed else "No"
        table.add_row(
            d.id[:8] + "...",
            d.sku,
            str(d.suggested_quantity),
            confirmed,
            format_datetime(d.created_at),
        )

    console.print(table)


def print_health_report(report: InventoryHealthReport) -> None:
    table = Table(title=f"Inventory Health Report - {format_datetime(report.generated_at)}")
    table.add_column("Metric", style="cyan")
    table.add_column("Count", style="white", justify="right")
    table.add_column("Percentage", style="yellow")

    total = report.total_products or 1
    table.add_row("Total Products", str(report.total_products), "100%")
    table.add_row(
        "Normal",
        str(report.normal_count),
        f"{(report.normal_count / total) * 100:.1f}%",
    )
    table.add_row(
        "Low Stock",
        str(report.low_stock_count),
        f"{(report.low_stock_count / total) * 100:.1f}%",
    )
    table.add_row(
        "Overstock",
        str(report.overstock_count),
        f"{(report.overstock_count / total) * 100:.1f}%",
    )

    console.print(table)

    if report.items:
        items_table = Table(title="Product Status")
        items_table.add_column("SKU", style="cyan")
        items_table.add_column("Name", style="white")
        items_table.add_column("Category", style="magenta")
        items_table.add_column("Stock", style="yellow", justify="right")
        items_table.add_column("Status", style="green")

        for item in report.items:
            status_style = ""
            if item.status.value == "low_stock":
                status_style = "red"
            elif item.status.value == "overstock":
                status_style = "yellow"
            else:
                status_style = "green"
            items_table.add_row(
                item.sku,
                item.name,
                item.category,
                str(item.current_stock),
                f"[{status_style}]{item.status.value}[/{status_style}]",
            )

        console.print(items_table)


def print_turnover_analysis(analysis: InventoryTurnoverAnalysis) -> None:
    table = Table(
        title=f"Turnover Analysis ({format_datetime(analysis.start_date)} to {format_datetime(analysis.end_date)})"
    )
    table.add_column("SKU", style="cyan")
    table.add_column("Name", style="white")
    table.add_column("Category", style="magenta")
    table.add_column("Outbound", style="yellow", justify="right")
    table.add_column("Avg Stock", style="blue", justify="right")
    table.add_column("Turnover Rate", style="green", justify="right")
    table.add_column("Turnover Days", style="red", justify="right")

    for item in analysis.items:
        table.add_row(
            item.sku,
            item.name,
            item.category,
            str(item.total_outbound),
            f"{item.average_stock:.1f}",
            f"{item.turnover_rate:.2f}",
            f"{item.turnover_days:.1f}" if item.turnover_days != float("inf") else "N/A",
        )

    console.print(table)


def print_json(data: Any) -> None:
    console.print_json(json.dumps(data, indent=2, default=str))
