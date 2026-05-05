from typing import Optional

import click
import httpx

from cli.client import APIClient
from cli.output import (
    console,
    print_error,
    print_json,
    print_success,
    print_alert_detail,
    print_alert_table,
)
from shared.models import AlertRecord


@click.group(name="alert")
def alert_group() -> None:
    """Alert management commands"""
    pass


@alert_group.command(name="list")
@click.option("--sku", default=None, help="Filter by SKU")
@click.option("--category", default=None, help="Filter by category")
@click.option("--type", "alert_type", default=None, type=click.Choice(["low_stock", "overstock"]), help="Filter by alert type")
@click.option("--level", "alert_level", default=None, type=click.Choice(["warning", "critical"]), help="Filter by alert level")
@click.option("--resolved", default=None, type=bool, help="Filter by resolved status")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def list_alerts(
    sku: Optional[str],
    category: Optional[str],
    alert_type: Optional[str],
    alert_level: Optional[str],
    resolved: Optional[bool],
    output_json: bool,
) -> None:
    """List alerts with optional filters"""
    with APIClient() as client:
        params: dict[str, str] = {}
        if sku:
            params["sku"] = sku
        if category:
            params["category"] = category
        if alert_type:
            params["alert_type"] = alert_type
        if alert_level:
            params["alert_level"] = alert_level
        if resolved is not None:
            params["resolved"] = str(resolved).lower()

        try:
            response = client.get("/alerts", params=params)
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                alerts = [AlertRecord(**a) for a in result["data"]]
                print_alert_table(alerts)
            else:
                print_error(result.get("error", "Failed to list alerts"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@alert_group.command(name="get")
@click.argument("alert_id")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def get_alert(alert_id: str, output_json: bool) -> None:
    """Get alert details by ID"""
    with APIClient() as client:
        try:
            response = client.get(f"/alerts/{alert_id}")
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                alert = AlertRecord(**result["data"])
                print_alert_detail(alert)
            else:
                print_error(result.get("error", "Alert not found"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@alert_group.command(name="resolve")
@click.argument("alert_id")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def resolve_alert(alert_id: str, output_json: bool) -> None:
    """Manually resolve an alert"""
    with APIClient() as client:
        try:
            response = client.post(f"/alerts/{alert_id}/resolve")
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                alert = AlertRecord(**result["data"])
                print_success(f"Alert resolved: {alert_id}")
                print_alert_detail(alert)
            else:
                print_error(result.get("error", "Failed to resolve alert"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")
