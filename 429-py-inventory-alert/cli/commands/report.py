from datetime import datetime, timedelta, timezone
from typing import Optional

import click
import httpx

from cli.client import APIClient
from cli.output import (
    console,
    print_error,
    print_json,
    print_success,
    print_health_report,
    print_turnover_analysis,
)
from shared.models import InventoryHealthReport, InventoryTurnoverAnalysis


@click.group(name="report")
def report_group() -> None:
    """Report and analysis commands"""
    pass


@report_group.command(name="health")
@click.option("--generate", "-g", is_flag=True, help="Generate a new health report")
@click.option("--latest", "-l", is_flag=True, help="Get the latest health report")
@click.option("--report-id", "-r", default=None, help="Get report by ID")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def health_report(
    generate: bool,
    latest: bool,
    report_id: Optional[str],
    output_json: bool,
) -> None:
    """Inventory health report commands"""
    with APIClient() as client:
        try:
            if generate:
                response = client.post("/reports/health")
            elif report_id:
                response = client.get(f"/reports/health/{report_id}")
            else:
                response = client.get("/reports/health/latest")

            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                report = InventoryHealthReport(**result["data"])
                if generate:
                    print_success("Health report generated")
                print_health_report(report)
            else:
                print_error(result.get("error", "Failed to get health report"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@report_group.command(name="turnover")
@click.option("--start-date", "-s", default=None, help="Start date (YYYY-MM-DD, UTC)")
@click.option("--end-date", "-e", default=None, help="End date (YYYY-MM-DD, UTC)")
@click.option("--days", "-d", default=30, type=int, help="Days to analyze (default: 30)")
@click.option("--analysis-id", "-a", default=None, help="Get analysis by ID")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def turnover_analysis(
    start_date: Optional[str],
    end_date: Optional[str],
    days: int,
    analysis_id: Optional[str],
    output_json: bool,
) -> None:
    """Inventory turnover analysis"""
    from datetime import date

    with APIClient() as client:
        try:
            if analysis_id:
                response = client.get(f"/reports/turnover/{analysis_id}")
            else:
                now = datetime.now(timezone.utc)
                if end_date:
                    end_dt = datetime.strptime(end_date, "%Y-%m-%d").replace(tzinfo=timezone.utc)
                else:
                    end_dt = now

                if start_date:
                    start_dt = datetime.strptime(start_date, "%Y-%m-%d").replace(tzinfo=timezone.utc)
                else:
                    start_dt = end_dt - timedelta(days=days)

                params = {
                    "start_date": start_dt.isoformat(),
                    "end_date": end_dt.isoformat(),
                }
                response = client.post("/reports/turnover", params=params)

            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                analysis = InventoryTurnoverAnalysis(**result["data"])
                print_success("Turnover analysis generated")
                print_turnover_analysis(analysis)
            else:
                print_error(result.get("error", "Failed to get turnover analysis"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")
