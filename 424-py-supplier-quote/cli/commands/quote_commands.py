from decimal import Decimal
from uuid import UUID

import typer

from shared.models.enums import QuoteStatus
from shared.protocols.quote import (
    AwardResultMasked,
    QuoteHistoryResponse,
    QuoteListResponse,
    QuoteResponse,
    QuoteSubmitRequest,
)
from cli.client.http_client import get_default_client

app = typer.Typer(help="报价管理命令")


@app.command(name="submit")
def submit_quote(
    purchase_id: UUID = typer.Option(..., "--purchase", "-p", help="采购需求 ID"),
    supplier_id: UUID = typer.Option(..., "--supplier", "-s", help="供应商 ID"),
    price: str = typer.Option(..., "--price", help="单价"),
    delivery_days: int = typer.Option(..., "--delivery", "-d", help="承诺交货天数"),
    remarks: str = typer.Option("", "--remarks", "-r", help="备注"),
) -> None:
    """提交报价（截止前可多次修改）"""
    client = get_default_client()
    try:
        request = QuoteSubmitRequest(
            unit_price=Decimal(price),
            delivery_days=delivery_days,
            remarks=remarks,
        )

        params = {"purchase_id": str(purchase_id), "supplier_id": str(supplier_id)}
        response = client.post("/quotes", request.model_dump(mode="json"), params=params)
        quote = client.parse_response(response, QuoteResponse)
        typer.secho("✓ 报价提交成功", fg=typer.colors.GREEN)
        _print_quote(quote)
    except Exception as e:
        typer.secho(f"✗ 提交失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="get")
def get_quote(
    quote_id: UUID = typer.Argument(..., help="报价 ID"),
) -> None:
    """获取报价详情"""
    client = get_default_client()
    try:
        response = client.get(f"/quotes/{quote_id}")
        quote = client.parse_response(response, QuoteResponse)
        _print_quote(quote)
    except Exception as e:
        typer.secho(f"✗ 获取失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="list-by-purchase")
def list_quotes_by_purchase(
    purchase_id: UUID = typer.Argument(..., help="采购需求 ID"),
) -> None:
    """列出某采购需求的所有报价"""
    client = get_default_client()
    try:
        response = client.get(f"/quotes/purchase/{purchase_id}")
        result = client.parse_response(response, QuoteListResponse)

        if result.total == 0:
            typer.secho("暂无报价", fg=typer.colors.YELLOW)
            return

        typer.secho(f"共 {result.total} 个报价:\n", fg=typer.colors.BLUE)
        for quote in result.quotes:
            _print_quote_short(quote)
    except Exception as e:
        typer.secho(f"✗ 获取列表失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="list-by-supplier")
def list_quotes_by_supplier(
    supplier_id: UUID = typer.Argument(..., help="供应商 ID"),
) -> None:
    """列出某供应商的所有报价"""
    client = get_default_client()
    try:
        response = client.get(f"/quotes/supplier/{supplier_id}")
        result = client.parse_response(response, QuoteListResponse)

        if result.total == 0:
            typer.secho("暂无报价", fg=typer.colors.YELLOW)
            return

        typer.secho(f"共 {result.total} 个报价:\n", fg=typer.colors.BLUE)
        for quote in result.quotes:
            _print_quote_short(quote)
    except Exception as e:
        typer.secho(f"✗ 获取列表失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="history")
def get_quote_history(
    quote_id: UUID = typer.Argument(..., help="报价 ID"),
) -> None:
    """查看报价修改历史"""
    client = get_default_client()
    try:
        response = client.get(f"/quotes/{quote_id}/history")
        history = client.parse_response(response, QuoteHistoryResponse)

        typer.secho(f"报价 ID: {history.quote_id}", fg=typer.colors.BLUE)
        typer.echo(f"\n版本历史 (共 {len(history.versions)} 个版本):")

        for version in history.versions:
            late_mark = " [迟到]" if version.is_late else ""
            typer.echo(f"\n  版本 {version.version} ({version.submitted_at}){late_mark}")
            typer.echo(f"    单价: {version.unit_price}")
            typer.echo(f"    交货天数: {version.delivery_days}")
            if version.remarks:
                typer.echo(f"    备注: {version.remarks}")

        if history.history:
            typer.echo(f"\n变更记录:")
            for entry in history.history:
                typer.echo(f"\n  版本 {entry.version} ({entry.changed_at})")
                for field, change in entry.changes.items():
                    typer.echo(f"    {field}: {change}")
    except Exception as e:
        typer.secho(f"✗ 获取历史失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="award-result")
def get_award_result(
    purchase_id: UUID = typer.Option(..., "--purchase", "-p", help="采购需求 ID"),
    supplier_id: UUID = typer.Option(..., "--supplier", "-s", help="供应商 ID"),
) -> None:
    """查看脱敏中标结果（供应商视角）"""
    client = get_default_client()
    try:
        response = client.get(f"/quotes/award-result/{purchase_id}/{supplier_id}")
        result = client.parse_response(response, AwardResultMasked)

        typer.echo(f"采购需求: {result.purchase_title}")
        if result.is_awarded:
            typer.secho("✓ 您已中标！", fg=typer.colors.GREEN)
        else:
            typer.secho("✗ 您未中标", fg=typer.colors.RED)

        typer.echo(f"中标供应商: {result.awarded_supplier_name}")
        comparison_text = {
            "lower": "比中标价格低",
            "higher": "比中标价格高",
            "same": "与中标价格相同",
        }.get(result.price_comparison, result.price_comparison)
        typer.echo(f"您的报价: {comparison_text}")
    except Exception as e:
        typer.secho(f"✗ 获取结果失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


def _print_quote(quote: QuoteResponse) -> None:
    status_color = (
        typer.colors.GREEN
        if quote.status == QuoteStatus.AWARDED
        else (
            typer.colors.RED
            if quote.status in (QuoteStatus.LOST, QuoteStatus.EXPIRED)
            else (
                typer.colors.YELLOW
                if quote.status == QuoteStatus.LATE
                else typer.colors.WHITE
            )
        )
    )
    typer.echo(f"  报价 ID: {quote.id}")
    typer.echo(f"  采购需求 ID: {quote.purchase_id}")
    typer.echo(f"  供应商: {quote.supplier_name} (等级: {quote.qualification_level.value})")
    typer.secho(f"  状态: {quote.status.value}", fg=status_color)
    typer.echo(f"  当前版本: {quote.current_version}")
    typer.echo(f"  有效期至: {quote.valid_until}")

    if quote.latest_version:
        typer.echo("\n  当前报价内容:")
        typer.echo(f"    单价: {quote.latest_version.unit_price}")
        typer.echo(f"    交货天数: {quote.latest_version.delivery_days}")
        if quote.latest_version.is_late:
            typer.secho("    这是一个迟到报价", fg=typer.colors.YELLOW)
        if quote.latest_version.remarks:
            typer.echo(f"    备注: {quote.latest_version.remarks}")


def _print_quote_short(quote: QuoteResponse) -> None:
    status_color = (
        typer.colors.GREEN
        if quote.status == QuoteStatus.AWARDED
        else (
            typer.colors.RED
            if quote.status in (QuoteStatus.LOST, QuoteStatus.EXPIRED)
            else typer.colors.WHITE
        )
    )
    typer.echo(f"  [{quote.id}] {quote.supplier_name} (等级: {quote.qualification_level.value})")
    if quote.latest_version:
        typer.echo(f"    单价: {quote.latest_version.unit_price}, 交货: {quote.latest_version.delivery_days}天")
    typer.secho(f"    状态: {quote.status.value}", fg=status_color)
