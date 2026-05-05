import typer

from cli.commands import (
    supplier_app,
    purchase_app,
    quote_app,
    order_app,
    report_app,
)

app = typer.Typer(
    name="quote-cli",
    help="供应商报价管理系统命令行客户端",
    add_completion=False,
)

app.add_typer(supplier_app, name="supplier", help="供应商管理")
app.add_typer(purchase_app, name="purchase", help="采购需求管理")
app.add_typer(quote_app, name="quote", help="报价管理")
app.add_typer(order_app, name="order", help="订单管理")
app.add_typer(report_app, name="report", help="报表管理")


@app.command(name="health")
def health_check() -> None:
    """检查服务端健康状态"""
    import httpx

    try:
        with httpx.Client(timeout=5.0) as client:
            response = client.get("http://localhost:8000/health")
            response.raise_for_status()
            data = response.json()
            if data.get("is_success") or data.get("success"):
                typer.secho(f"✓ 服务正常运行: {data.get('data')}", fg=typer.colors.GREEN)
            else:
                typer.secho(f"✗ 服务异常: {data.get('message')}", fg=typer.colors.RED)
    except Exception as e:
        typer.secho(f"✗ 无法连接到服务端: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


if __name__ == "__main__":
    app()
