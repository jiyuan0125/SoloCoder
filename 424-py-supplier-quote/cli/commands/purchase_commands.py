from datetime import datetime
from decimal import Decimal
from typing import Optional
from uuid import UUID

import typer

from shared.models.enums import PurchaseStatus
from shared.models.purchase import PurchaseItem
from shared.protocols.purchase import (
    PurchaseCreateRequest,
    PurchaseListResponse,
    PurchaseResponse,
)
from cli.client.http_client import get_default_client

app = typer.Typer(help="采购需求管理命令")


@app.command(name="create")
def create_purchase(
    title: str = typer.Option(..., "--title", "-t", help="采购需求标题"),
    description: str = typer.Option("", "--desc", "-d", help="需求描述"),
    deadline: str = typer.Option(
        ..., "--deadline", "-l", help="报价截止时间 (ISO格式, 如 2026-06-01T18:00:00)"
    ),
    products: str = typer.Option(
        ..., "--products", "-p",
        help="商品列表 (JSON格式: [{\"name\":\"商品1\",\"quantity\":100,\"unit\":\"件\"},...])"
    ),
) -> None:
    """创建采购需求"""
    import json

    client = get_default_client()
    try:
        deadline_dt = datetime.fromisoformat(deadline)
        products_data = json.loads(products)

        items: list[PurchaseItem] = []
        for p in products_data:
            estimated_price = p.get("estimated_price")
            item = PurchaseItem(
                product_name=p["name"],
                product_code=p.get("product_code"),
                quantity=int(p["quantity"]),
                unit=p["unit"],
                estimated_price=Decimal(str(estimated_price)) if estimated_price else None,
                description=p.get("description"),
            )
            items.append(item)

        request = PurchaseCreateRequest(
            title=title,
            description=description if description else None,
            items=items,
            quote_deadline=deadline_dt,
        )

        response = client.post("/purchases", request.model_dump(mode="json"))
        purchase = client.parse_response(response, PurchaseResponse)
        typer.secho("✓ 采购需求创建成功", fg=typer.colors.GREEN)
        _print_purchase(purchase)
    except Exception as e:
        typer.secho(f"✗ 创建失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="get")
def get_purchase(
    purchase_id: UUID = typer.Argument(..., help="采购需求 ID"),
) -> None:
    """获取采购需求详情"""
    client = get_default_client()
    try:
        response = client.get(f"/purchases/{purchase_id}")
        purchase = client.parse_response(response, PurchaseResponse)
        _print_purchase(purchase)
    except Exception as e:
        typer.secho(f"✗ 获取失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="list")
def list_purchases(
    status: Optional[PurchaseStatus] = typer.Option(
        None, "--status", "-s", help="按状态筛选"
    ),
    published_only: bool = typer.Option(False, "--published", "-p", help="仅显示已发布的"),
) -> None:
    """列出所有采购需求"""
    client = get_default_client()
    try:
        params: dict[str, str | bool] = {}
        if status is not None:
            params["status"] = status.value
        if published_only:
            params["published_only"] = True

        response = client.get("/purchases", params)
        result = client.parse_response(response, PurchaseListResponse)

        if result.total == 0:
            typer.secho("暂无采购需求", fg=typer.colors.YELLOW)
            return

        typer.secho(f"共 {result.total} 个采购需求:\n", fg=typer.colors.BLUE)
        for purchase in result.purchases:
            _print_purchase_short(purchase)
    except Exception as e:
        typer.secho(f"✗ 获取列表失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="publish")
def publish_purchase(
    purchase_id: UUID = typer.Argument(..., help="采购需求 ID"),
) -> None:
    """发布采购需求"""
    client = get_default_client()
    try:
        response = client.post(f"/purchases/{purchase_id}/publish")
        purchase = client.parse_response(response, PurchaseResponse)
        typer.secho("✓ 采购需求已发布", fg=typer.colors.GREEN)
        _print_purchase(purchase)
    except Exception as e:
        typer.secho(f"✗ 发布失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="close")
def close_purchase(
    purchase_id: UUID = typer.Argument(..., help="采购需求 ID"),
) -> None:
    """关闭采购需求"""
    client = get_default_client()
    try:
        response = client.post(f"/purchases/{purchase_id}/close")
        purchase = client.parse_response(response, PurchaseResponse)
        typer.secho("✓ 采购需求已关闭", fg=typer.colors.GREEN)
        _print_purchase(purchase)
    except Exception as e:
        typer.secho(f"✗ 关闭失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


def _print_purchase(purchase: PurchaseResponse) -> None:
    status_color = (
        typer.colors.GREEN
        if purchase.status in (PurchaseStatus.PUBLISHED, PurchaseStatus.QUOTING)
        else (
            typer.colors.YELLOW
            if purchase.status == PurchaseStatus.DRAFT
            else (
                typer.colors.MAGENTA
                if purchase.status == PurchaseStatus.AWARDED
                else typer.colors.WHITE
            )
        )
    )
    typer.echo(f"  ID: {purchase.id}")
    typer.echo(f"  标题: {purchase.title}")
    if purchase.description:
        typer.echo(f"  描述: {purchase.description}")
    typer.secho(f"  状态: {purchase.status.value}", fg=status_color)
    typer.echo(f"  报价截止: {purchase.quote_deadline}")
    if purchase.published_at:
        typer.echo(f"  发布时间: {purchase.published_at}")
    if purchase.awarded_at:
        typer.echo(f"  定标时间: {purchase.awarded_at}")
        typer.echo(f"  中标供应商ID: {purchase.awarded_supplier_id}")

    typer.echo("\n  商品列表:")
    for i, item in enumerate(purchase.items, 1):
        typer.echo(f"    {i}. {item.product_name} (数量: {item.quantity} {item.unit})")
        if item.estimated_price:
            typer.echo(f"       预估单价: {item.estimated_price}")


def _print_purchase_short(purchase: PurchaseResponse) -> None:
    status_color = (
        typer.colors.GREEN
        if purchase.status in (PurchaseStatus.PUBLISHED, PurchaseStatus.QUOTING)
        else (
            typer.colors.YELLOW
            if purchase.status == PurchaseStatus.DRAFT
            else (
                typer.colors.MAGENTA
                if purchase.status == PurchaseStatus.AWARDED
                else typer.colors.WHITE
            )
        )
    )
    typer.echo(f"  [{purchase.id}] {purchase.title}")
    typer.secho(f"    状态: {purchase.status.value}, 截止: {purchase.quote_deadline}", fg=status_color)
