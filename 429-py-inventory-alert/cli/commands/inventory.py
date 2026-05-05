from typing import Optional

import click
import httpx

from cli.client import APIClient
from cli.output import (
    console,
    print_error,
    print_json,
    print_success,
    print_transaction_table,
)
from shared.models import InventoryTransaction, StockAdjustment


@click.group(name="inventory")
def inventory_group() -> None:
    """Inventory management commands"""
    pass


@inventory_group.command(name="inbound")
@click.argument("sku")
@click.option("--quantity", "-q", required=True, type=int, help="Quantity to add")
@click.option("--reference", "-r", default=None, help="Reference ID")
@click.option("--notes", "-n", default=None, help="Notes")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def stock_inbound(
    sku: str,
    quantity: int,
    reference: Optional[str],
    notes: Optional[str],
    output_json: bool,
) -> None:
    """Add stock to inventory (inbound)"""
    with APIClient() as client:
        adjustment = StockAdjustment(
            quantity=quantity,
            reference_id=reference,
            notes=notes,
        )

        try:
            response = client.post(
                f"/inventory/{sku}/inbound",
                json=adjustment.model_dump(mode="json"),
            )
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                transaction = InventoryTransaction(**result["data"])
                print_success(f"Inbound transaction created: +{transaction.quantity} for {sku}")
                print_transaction_table([transaction])
            else:
                print_error(result.get("error", "Failed to create inbound transaction"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@inventory_group.command(name="outbound")
@click.argument("sku")
@click.option("--quantity", "-q", required=True, type=int, help="Quantity to remove")
@click.option("--reference", "-r", default=None, help="Reference ID")
@click.option("--notes", "-n", default=None, help="Notes")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def stock_outbound(
    sku: str,
    quantity: int,
    reference: Optional[str],
    notes: Optional[str],
    output_json: bool,
) -> None:
    """Remove stock from inventory (outbound)"""
    with APIClient() as client:
        adjustment = StockAdjustment(
            quantity=quantity,
            reference_id=reference,
            notes=notes,
        )

        try:
            response = client.post(
                f"/inventory/{sku}/outbound",
                json=adjustment.model_dump(mode="json"),
            )
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                transaction = InventoryTransaction(**result["data"])
                print_success(f"Outbound transaction created: -{transaction.quantity} for {sku}")
                print_transaction_table([transaction])
            else:
                print_error(result.get("error", "Failed to create outbound transaction"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@inventory_group.command(name="transfer")
@click.argument("sku")
@click.option("--quantity", "-q", required=True, type=int, help="Quantity to transfer out")
@click.option("--reference", "-r", default=None, help="Reference ID")
@click.option("--notes", "-n", default=None, help="Notes")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def stock_transfer(
    sku: str,
    quantity: int,
    reference: Optional[str],
    notes: Optional[str],
    output_json: bool,
) -> None:
    """Transfer stock out (for warehouse transfer)"""
    with APIClient() as client:
        adjustment = StockAdjustment(
            quantity=quantity,
            reference_id=reference,
            notes=notes,
        )

        try:
            response = client.post(
                f"/inventory/{sku}/transfer",
                json=adjustment.model_dump(mode="json"),
            )
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                transaction = InventoryTransaction(**result["data"])
                print_success(f"Transfer transaction created: -{transaction.quantity} for {sku}")
                print_transaction_table([transaction])
            else:
                print_error(result.get("error", "Failed to create transfer transaction"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@inventory_group.command(name="transactions")
@click.option("--sku", default=None, help="Filter by SKU")
@click.option("--type", "transaction_type", default=None, type=click.Choice(["inbound", "outbound", "transfer"]), help="Filter by transaction type")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def list_transactions(
    sku: Optional[str],
    transaction_type: Optional[str],
    output_json: bool,
) -> None:
    """List inventory transactions"""
    from shared.models import TransactionType

    with APIClient() as client:
        params: dict[str, str] = {}
        if sku:
            params["sku"] = sku
        if transaction_type:
            params["transaction_type"] = transaction_type

        try:
            response = client.get("/inventory/transactions", params=params)
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                transactions = [InventoryTransaction(**t) for t in result["data"]]
                print_transaction_table(transactions)
            else:
                print_error(result.get("error", "Failed to list transactions"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@inventory_group.command(name="status")
@click.argument("sku")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def inventory_status(sku: str, output_json: bool) -> None:
    """Get inventory status for a product"""
    with APIClient() as client:
        try:
            response = client.get(f"/inventory/{sku}/status")
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                data = result["data"]
                console.print(f"[cyan]Inventory Status for {data['sku']} ({data['name']})[/cyan]")
                console.print(f"  Current Stock: [yellow]{data['current_stock']}[/yellow]")
                console.print(f"  Min Stock: [red]{data['min_stock']}[/red]")
                console.print(f"  Max Stock: [green]{data['max_stock']}[/green]")

                status = data["status"]
                status_style = "green"
                if status == "low_stock":
                    status_style = "red"
                elif status == "overstock":
                    status_style = "yellow"
                console.print(f"  Status: [{status_style}]{status}[/{status_style}]")
                console.print(f"  Seasonal: {'Yes' if data['is_seasonal'] else 'No'}")
            else:
                print_error(result.get("error", "Failed to get inventory status"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")
