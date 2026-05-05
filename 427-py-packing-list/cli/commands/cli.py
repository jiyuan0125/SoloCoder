from typing import Optional
import json

import click
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich import print as rprint

from cli.client import APIClient

console = Console()


@click.group()
@click.option("--server", "-s", default=None, help="服务端地址 (默认: http://localhost:8000)")
@click.pass_context
def main(ctx: click.Context, server: Optional[str]) -> None:
    """装箱管理系统命令行工具"""
    ctx.ensure_object(dict)
    ctx.obj["client"] = APIClient(base_url=server)


@main.group()
def order() -> None:
    """订单管理"""
    pass


@order.command("list")
@click.pass_context
def order_list(ctx: click.Context) -> None:
    """列出所有订单"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.list_orders()
        if result.get("success"):
            orders = result.get("data", {}).get("orders", [])
            if not orders:
                rprint("[yellow]暂无订单[/yellow]")
                return

            table = Table(title="订单列表")
            table.add_column("订单号", style="cyan")
            table.add_column("状态", style="green")
            table.add_column("商品数", justify="right")
            table.add_column("创建时间")

            for order_data in orders:
                items_count = len(order_data.get("items", []))
                table.add_row(
                    order_data.get("order_id"),
                    order_data.get("status"),
                    str(items_count),
                    order_data.get("created_at", "-"),
                )

            console.print(table)
        else:
            rprint(f"[red]错误: {result.get('message', '未知错误')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@order.command("create")
@click.argument("order_id")
@click.argument("items_json")
@click.pass_context
def order_create(ctx: click.Context, order_id: str, items_json: str) -> None:
    """创建订单

    ORDER_ID: 订单号

    ITEMS_JSON: 商品列表JSON, 格式: [{"product": {"product_id": "P001", "name": "商品1", "weight_kg": 5, "length_cm": 20, "width_cm": 15, "height_cm": 10, "product_type": "normal"}, "quantity": 2}]
    """
    client: APIClient = ctx.obj["client"]
    try:
        items = json.loads(items_json)
        result = client.create_order(order_id, items)
        if result.get("success"):
            rprint(f"[green]✓ 订单创建成功: {order_id}[/green]")
            order_data = result.get("data", {}).get("order", {})
            rprint(f"  状态: {order_data.get('status')}")
            rprint(f"  商品数: {len(order_data.get('items', []))}")
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
            rprint(f"  错误码: {result.get('error_code')}")
    except json.JSONDecodeError:
        rprint("[red]错误: JSON格式无效[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@order.command("get")
@click.argument("order_id")
@click.pass_context
def order_get(ctx: click.Context, order_id: str) -> None:
    """查看订单详情"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.get_order(order_id)
        if result.get("success"):
            order_data = result.get("data", {}).get("order", {})
            panel = Panel(
                f"[cyan]订单号:[/cyan] {order_data.get('order_id')}\n"
                f"[cyan]状态:[/cyan] {order_data.get('status')}\n"
                f"[cyan]创建时间:[/cyan] {order_data.get('created_at')}",
                title="订单详情",
            )
            console.print(panel)

            items = order_data.get("items", [])
            if items:
                table = Table(title="商品列表")
                table.add_column("商品ID", style="cyan")
                table.add_column("名称", style="green")
                table.add_column("类型")
                table.add_column("重量(kg)", justify="right")
                table.add_column("数量", justify="right")

                for item in items:
                    prod = item.get("product", {})
                    table.add_row(
                        prod.get("product_id"),
                        prod.get("name"),
                        prod.get("product_type"),
                        f"{prod.get('weight_kg'):.2f}",
                        str(item.get("quantity")),
                    )
                console.print(table)
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@main.group()
def package() -> None:
    """包裹管理"""
    pass


@package.command("create")
@click.argument("order_id")
@click.argument("box_type_id")
@click.pass_context
def package_create(ctx: click.Context, order_id: str, box_type_id: str) -> None:
    """为订单创建新包裹"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.create_package(order_id, box_type_id)
        if result.get("success"):
            pkg = result.get("data", {}).get("package", {})
            rprint(f"[green]✓ 包裹创建成功[/green]")
            rprint(f"  箱号: {pkg.get('box_number')}")
            rprint(f"  包裹ID: {pkg.get('package_id')}")
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@package.command("list")
@click.argument("order_id")
@click.pass_context
def package_list(ctx: click.Context, order_id: str) -> None:
    """列出订单的所有包裹"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.list_packages(order_id)
        if result.get("success"):
            packages = result.get("data", {}).get("packages", [])
            if not packages:
                rprint("[yellow]该订单暂无包裹[/yellow]")
                return

            table = Table(title=f"订单 {order_id} 包裹列表")
            table.add_column("箱号", style="cyan")
            table.add_column("状态", style="green")
            table.add_column("重量(kg)", justify="right")
            table.add_column("体积(m³)", justify="right")
            table.add_column("商品数", justify="right")
            table.add_column("物流单号")

            for pkg in packages:
                items_count = len(pkg.get("items", []))
                tracking = pkg.get("tracking_number", "-")
                table.add_row(
                    pkg.get("box_number"),
                    pkg.get("status"),
                    f"{pkg.get('total_weight_kg', 0):.2f}",
                    f"{pkg.get('total_volume_m3', 0):.4f}",
                    str(items_count),
                    tracking,
                )

            console.print(table)
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@package.command("add-item")
@click.argument("package_id")
@click.argument("product_id")
@click.option("--quantity", "-q", type=int, default=1, help="数量 (默认: 1)")
@click.pass_context
def package_add_item(
    ctx: click.Context, package_id: str, product_id: str, quantity: int
) -> None:
    """向包裹添加商品"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.add_item_to_package(package_id, product_id, quantity)
        if result.get("success"):
            pkg = result.get("data", {}).get("package", {})
            rprint(f"[green]✓ 商品已添加到包裹[/green]")
            rprint(f"  箱号: {pkg.get('box_number')}")
            rprint(f"  当前重量: {pkg.get('total_weight_kg'):.2f} kg")
            rprint(f"  当前体积: {pkg.get('total_volume_m3'):.4f} m³")
            rprint(f"  填充物体积: {pkg.get('filler_volume_m3'):.4f} m³")
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
            rprint(f"  错误码: {result.get('error_code')}")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@package.command("ship")
@click.argument("package_id")
@click.argument("tracking_number")
@click.pass_context
def package_ship(
    ctx: click.Context, package_id: str, tracking_number: str
) -> None:
    """标记包裹已发出并添加物流单号"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.update_tracking(package_id, tracking_number)
        if result.get("success"):
            rprint(f"[green]✓ 包裹已标记为发出[/green]")
            rprint(f"  物流单号: {tracking_number}")
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@main.command("packing-list")
@click.argument("order_id")
@click.pass_context
def packing_list(ctx: click.Context, order_id: str) -> None:
    """生成格式化装箱单"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.get_packing_list_text(order_id)
        if result.get("success"):
            text = result.get("data", {}).get("packing_list_text", "")
            console.print(text)
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@main.command("suggest")
@click.argument("order_id")
@click.pass_context
def packing_suggest(ctx: click.Context, order_id: str) -> None:
    """获取装箱优化建议"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.get_packing_suggestions(order_id)
        if result.get("success"):
            suggestions = result.get("data", {}).get("suggestions", [])
            if not suggestions:
                rprint("[yellow]暂无装箱建议[/yellow]")
                return

            rprint(f"[cyan]订单 {order_id} 装箱建议:[/cyan]")
            rprint(f"共建议 {len(suggestions)} 个包裹\n")

            for i, sug in enumerate(suggestions, 1):
                box = sug.get("suggested_box_type", {})
                items = sug.get("items", [])
                confidence = sug.get("confidence_score", 0) * 100

                rprint(f"[bold]包裹 {i}:[/bold]")
                rprint(f"  推荐箱型: {box.get('name')} ({box.get('length_cm')}x{box.get('width_cm')}x{box.get('height_cm')}cm)")
                rprint(f"  预估重量: {sug.get('estimated_weight_kg'):.2f} kg")
                rprint(f"  预估体积: {sug.get('estimated_volume_m3'):.4f} m³")
                rprint(f"  置信度: {confidence:.0f}%")
                rprint(f"  商品:")
                for item in items:
                    prod = item.get("product", {})
                    rprint(f"    - {prod.get('name')} x{item.get('quantity')}")
                rprint("")
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@main.group()
def box() -> None:
    """箱型管理"""
    pass


@box.command("list")
@click.pass_context
def box_list(ctx: click.Context) -> None:
    """列出所有可用箱型"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.list_box_types()
        if result.get("success"):
            boxes = result.get("data", {}).get("box_types", [])
            if not boxes:
                rprint("[yellow]暂无箱型[/yellow]")
                return

            table = Table(title="可用箱型")
            table.add_column("箱型ID", style="cyan")
            table.add_column("名称", style="green")
            table.add_column("尺寸(cm)")
            table.add_column("体积(m³)", justify="right")
            table.add_column("限重(kg)", justify="right")

            for box in boxes:
                size = f"{box.get('length_cm')}x{box.get('width_cm')}x{box.get('height_cm')}"
                table.add_row(
                    box.get("box_type_id"),
                    box.get("name"),
                    size,
                    f"{box.get('volume_m3', 0):.4f}",
                    f"{box.get('max_weight_kg'):.1f}",
                )

            console.print(table)
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@box.command("stats")
@click.pass_context
def box_stats(ctx: click.Context) -> None:
    """查看包装材料库存统计"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.get_material_stats()
        if result.get("success"):
            stats = result.get("data", {}).get("stats", [])
            if not stats:
                rprint("[yellow]暂无库存数据[/yellow]")
                return

            table = Table(title="包装材料库存统计")
            table.add_column("箱型", style="cyan")
            table.add_column("库存", justify="right", style="green")
            table.add_column("已使用", justify="right", style="yellow")

            for stat in stats:
                box = stat.get("box_type", {})
                table.add_row(
                    box.get("name"),
                    str(stat.get("stock_quantity", 0)),
                    str(stat.get("used_quantity", 0)),
                )

            console.print(table)
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@main.group()
def template() -> None:
    """装箱模板管理"""
    pass


@template.command("save")
@click.argument("order_id")
@click.argument("template_name")
@click.pass_context
def template_save(ctx: click.Context, order_id: str, template_name: str) -> None:
    """将装箱方案保存为模板"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.save_template(order_id, template_name)
        if result.get("success"):
            rprint(f"[green]✓ 模板保存成功: {template_name}[/green]")
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@template.command("list")
@click.pass_context
def template_list(ctx: click.Context) -> None:
    """列出所有装箱模板"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.list_templates()
        if result.get("success"):
            templates = result.get("data", {}).get("templates", [])
            if not templates:
                rprint("[yellow]暂无装箱模板[/yellow]")
                return

            table = Table(title="装箱模板列表")
            table.add_column("模板ID", style="cyan")
            table.add_column("名称", style="green")
            table.add_column("箱数", justify="right")
            table.add_column("创建时间")

            for tmpl in templates:
                table.add_row(
                    tmpl.get("template_id"),
                    tmpl.get("template_name"),
                    str(tmpl.get("total_boxes")),
                    tmpl.get("created_at", "-"),
                )

            console.print(table)
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@template.command("apply")
@click.argument("order_id")
@click.argument("template_id")
@click.pass_context
def template_apply(ctx: click.Context, order_id: str, template_id: str) -> None:
    """应用装箱模板到订单"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.apply_template(order_id, template_id)
        if result.get("success"):
            rprint(f"[green]✓ 模板应用成功[/green]")
            pk_list = result.get("data", {}).get("packing_list", {})
            rprint(f"  创建了 {pk_list.get('total_boxes')} 个包裹")
        else:
            rprint(f"[red]✗ 错误: {result.get('message')}[/red]")
    except Exception as e:
        rprint(f"[red]连接失败: {e}[/red]")


@main.command("health")
@click.pass_context
def health_check(ctx: click.Context) -> None:
    """检查服务端健康状态"""
    client: APIClient = ctx.obj["client"]
    try:
        result = client.health_check()
        if result.get("status") == "healthy":
            rprint("[green]✓ 服务端运行正常[/green]")
        else:
            rprint(f"[yellow]服务端状态: {result}[/yellow]")
    except Exception as e:
        rprint(f"[red]✗ 无法连接到服务端: {e}[/red]")


if __name__ == "__main__":
    main()
