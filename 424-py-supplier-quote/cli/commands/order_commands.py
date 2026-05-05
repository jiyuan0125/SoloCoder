from typing import Optional
from uuid import UUID

import typer

from shared.models.enums import OrderStatus
from shared.protocols.order import (
    OrderConfirmRequest,
    OrderListResponse,
    OrderResponse,
)
from cli.client.http_client import get_default_client

app = typer.Typer(help="订单管理命令")


@app.command(name="create")
def create_order(
    quote_id: UUID = typer.Argument(..., help="中标报价 ID"),
) -> None:
    """从报价创建订单（选择中标供应商）"""
    client = get_default_client()
    try:
        response = client.post(f"/orders/from-quote/{quote_id}")
        order = client.parse_response(response, OrderResponse)
        typer.secho("✓ 订单创建成功", fg=typer.colors.GREEN)
        _print_order(order)
    except Exception as e:
        typer.secho(f"✗ 创建失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="get")
def get_order(
    order_id: UUID = typer.Argument(..., help="订单 ID"),
) -> None:
    """获取订单详情"""
    client = get_default_client()
    try:
        response = client.get(f"/orders/{order_id}")
        order = client.parse_response(response, OrderResponse)
        _print_order(order)
    except Exception as e:
        typer.secho(f"✗ 获取失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="get-by-number")
def get_order_by_number(
    order_number: str = typer.Argument(..., help="订单编号"),
) -> None:
    """按订单编号获取订单"""
    client = get_default_client()
    try:
        response = client.get(f"/orders/number/{order_number}")
        order = client.parse_response(response, OrderResponse)
        _print_order(order)
    except Exception as e:
        typer.secho(f"✗ 获取失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="list")
def list_orders(
    status: Optional[OrderStatus] = typer.Option(
        None, "--status", "-s", help="按状态筛选"
    ),
    supplier_id: Optional[UUID] = typer.Option(
        None, "--supplier", "-p", help="按供应商筛选"
    ),
) -> None:
    """列出所有订单"""
    client = get_default_client()
    try:
        params: dict[str, str] = {}
        if status is not None:
            params["status"] = status.value
        if supplier_id is not None:
            params["supplier_id"] = str(supplier_id)

        response = client.get("/orders", params)
        result = client.parse_response(response, OrderListResponse)

        if result.total == 0:
            typer.secho("暂无订单", fg=typer.colors.YELLOW)
            return

        typer.secho(f"共 {result.total} 个订单:\n", fg=typer.colors.BLUE)
        for order in result.orders:
            _print_order_short(order)
    except Exception as e:
        typer.secho(f"✗ 获取列表失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="confirm")
def confirm_order(
    order_id: UUID = typer.Argument(..., help="订单 ID"),
    remarks: str = typer.Option("", "--remarks", "-r", help="备注"),
) -> None:
    """确认订单"""
    client = get_default_client()
    try:
        request = OrderConfirmRequest(remarks=remarks)
        response = client.post(f"/orders/{order_id}/confirm", request.model_dump(mode="json"))
        order = client.parse_response(response, OrderResponse)
        typer.secho("✓ 订单已确认", fg=typer.colors.GREEN)
        _print_order(order)
    except Exception as e:
        typer.secho(f"✗ 确认失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="start")
def start_order(
    order_id: UUID = typer.Argument(..., help="订单 ID"),
) -> None:
    """开始执行订单"""
    client = get_default_client()
    try:
        response = client.post(f"/orders/{order_id}/start")
        order = client.parse_response(response, OrderResponse)
        typer.secho("✓ 订单已开始执行", fg=typer.colors.GREEN)
        _print_order(order)
    except Exception as e:
        typer.secho(f"✗ 操作失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="complete")
def complete_order(
    order_id: UUID = typer.Argument(..., help="订单 ID"),
) -> None:
    """完成订单"""
    client = get_default_client()
    try:
        response = client.post(f"/orders/{order_id}/complete")
        order = client.parse_response(response, OrderResponse)
        typer.secho("✓ 订单已完成", fg=typer.colors.GREEN)
        _print_order(order)
    except Exception as e:
        typer.secho(f"✗ 操作失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="cancel")
def cancel_order(
    order_id: UUID = typer.Argument(..., help="订单 ID"),
) -> None:
    """取消订单"""
    client = get_default_client()
    try:
        response = client.post(f"/orders/{order_id}/cancel")
        order = client.parse_response(response, OrderResponse)
        typer.secho("✓ 订单已取消", fg=typer.colors.GREEN)
        _print_order(order)
    except Exception as e:
        typer.secho(f"✗ 操作失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


def _print_order(order: OrderResponse) -> None:
    status_color = (
        typer.colors.GREEN
        if order.status == OrderStatus.COMPLETED
        else (
            typer.colors.RED
            if order.status == OrderStatus.CANCELLED
            else (
                typer.colors.MAGENTA
                if order.status == OrderStatus.IN_PROGRESS
                else typer.colors.WHITE
            )
        )
    )
    typer.echo(f"  订单编号: {order.order_number}")
    typer.echo(f"  订单 ID: {order.id}")
    typer.echo(f"  采购需求 ID: {order.purchase_id}")
    typer.echo(f"  报价 ID: {order.quote_id}")
    typer.echo(f"  供应商: {order.supplier_name}")
    typer.secho(f"  状态: {order.status.value}", fg=status_color)
    typer.echo(f"  总金额: {order.total_amount}")
    typer.echo(f"  承诺交货天数: {order.delivery_days} 天")
    if order.expected_delivery_date:
        typer.echo(f"  预计交货日期: {order.expected_delivery_date}")
    if order.confirmed_at:
        typer.echo(f"  确认时间: {order.confirmed_at}")
    if order.completed_at:
        typer.echo(f"  完成时间: {order.completed_at}")
    if order.remarks:
        typer.echo(f"  备注: {order.remarks}")

    typer.echo("\n  商品列表:")
    for i, item in enumerate(order.items, 1):
        typer.echo(
            f"    {i}. {item.product_name} "
            f"(数量: {item.quantity} {item.unit}, "
            f"单价: {item.unit_price}, 小计: {item.total_amount})"
        )


def _print_order_short(order: OrderResponse) -> None:
    status_color = (
        typer.colors.GREEN
        if order.status == OrderStatus.COMPLETED
        else (
            typer.colors.RED
            if order.status == OrderStatus.CANCELLED
            else (
                typer.colors.MAGENTA
                if order.status == OrderStatus.IN_PROGRESS
                else typer.colors.WHITE
            )
        )
    )
    typer.echo(f"  [{order.order_number}] 供应商: {order.supplier_name}")
    typer.echo(f"    金额: {order.total_amount}")
    typer.secho(f"    状态: {order.status.value}", fg=status_color)
