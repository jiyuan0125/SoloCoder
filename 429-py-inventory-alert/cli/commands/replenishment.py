from typing import Optional

import click
import httpx

from cli.client import APIClient
from cli.output import (
    console,
    print_error,
    print_json,
    print_success,
    print_replenishment_table,
)
from shared.models import ReplenishmentOrderDraft


@click.group(name="replenishment")
def replenishment_group() -> None:
    """Replenishment order draft commands"""
    pass


@replenishment_group.command(name="list")
@click.option("--sku", default=None, help="Filter by SKU")
@click.option("--confirmed", default=None, type=bool, help="Filter by confirmed status")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def list_drafts(
    sku: Optional[str],
    confirmed: Optional[bool],
    output_json: bool,
) -> None:
    """List replenishment order drafts"""
    with APIClient() as client:
        params: dict[str, str] = {}
        if sku:
            params["sku"] = sku
        if confirmed is not None:
            params["confirmed"] = str(confirmed).lower()

        try:
            response = client.get("/replenishment", params=params)
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                drafts = [ReplenishmentOrderDraft(**d) for d in result["data"]]
                print_replenishment_table(drafts)
            else:
                print_error(result.get("error", "Failed to list drafts"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@replenishment_group.command(name="get")
@click.argument("draft_id")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def get_draft(draft_id: str, output_json: bool) -> None:
    """Get replenishment draft details by ID"""
    with APIClient() as client:
        try:
            response = client.get(f"/replenishment/{draft_id}")
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                draft = ReplenishmentOrderDraft(**result["data"])
                console.print(f"[cyan]Replenishment Draft: {draft.id}[/cyan]")
                console.print(f"  SKU: {draft.sku}")
                console.print(f"  Alert ID: {draft.alert_id}")
                console.print(f"  Suggested Quantity: [yellow]{draft.suggested_quantity}[/yellow]")
                console.print(f"  Confirmed: {'Yes' if draft.confirmed else 'No'}")
                console.print(f"  Created At: {draft.created_at}")
                if draft.confirmed_at:
                    console.print(f"  Confirmed At: {draft.confirmed_at}")
            else:
                print_error(result.get("error", "Draft not found"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@replenishment_group.command(name="confirm")
@click.argument("draft_id")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def confirm_draft(draft_id: str, output_json: bool) -> None:
    """Confirm a replenishment order draft"""
    with APIClient() as client:
        try:
            response = client.post(f"/replenishment/{draft_id}/confirm")
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                draft = ReplenishmentOrderDraft(**result["data"])
                print_success(f"Replenishment draft confirmed: {draft_id}")
                console.print(f"  SKU: {draft.sku}")
                console.print(f"  Confirmed Quantity: [yellow]{draft.suggested_quantity}[/yellow]")
            else:
                print_error(result.get("error", "Failed to confirm draft"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")
