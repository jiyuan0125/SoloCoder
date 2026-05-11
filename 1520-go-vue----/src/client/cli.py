import typer
from typing import Optional
from datetime import datetime, date
from rich.console import Console
from rich.table import Table
from rich import print as rprint

from src.client.api_client import APIClient


app = typer.Typer(help="水产种苗场管理系统客户端")
console = Console()


def get_client():
    return APIClient()


@app.command()
def health():
    """检查服务端健康状态"""
    client = get_client()
    try:
        result = client.health_check()
        rprint(f"[green]✓ 服务运行正常: {result}[/green]")
    except Exception as e:
        rprint(f"[red]✗ 服务连接失败: {e}[/red]")
    finally:
        client.close()


@app.group()
def species():
    """品种管理"""
    pass


@species.command("list")
def list_species():
    """列出所有品种"""
    client = get_client()
    try:
        items = client.list_species()
        if not items:
            rprint("[yellow]暂无品种数据[/yellow]")
            return
        
        table = Table(title="品种列表")
        table.add_column("ID")
        table.add_column("名称")
        table.add_column("最低温度")
        table.add_column("最高温度")
        table.add_column("描述")
        
        for item in items:
            table.add_row(
                item['id'],
                item['name'],
                f"{item['min_temperature']}°C",
                f"{item['max_temperature']}°C",
                item.get('description', '-')
            )
        
        console.print(table)
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@species.command("create")
def create_species(
    name: str = typer.Option(..., help="品种名称"),
    min_temp: float = typer.Option(..., help="最低适宜温度"),
    max_temp: float = typer.Option(..., help="最高适宜温度"),
    description: Optional[str] = typer.Option(None, help="描述")
):
    """创建新品种"""
    client = get_client()
    try:
        data = {
            "name": name,
            "min_temperature": min_temp,
            "max_temperature": max_temp,
            "description": description
        }
        result = client.create_species(data)
        rprint(f"[green]✓ 品种创建成功[/green]")
        rprint(f"  ID: {result['id']}")
        rprint(f"  名称: {result['name']}")
    except Exception as e:
        rprint(f"[red]✗ 创建失败: {e}[/red]")
    finally:
        client.close()


@app.group()
def pond():
    """繁育池管理"""
    pass


@pond.command("list")
def list_ponds():
    """列出所有繁育池"""
    client = get_client()
    try:
        items = client.list_ponds()
        if not items:
            rprint("[yellow]暂无繁育池数据[/yellow]")
            return
        
        table = Table(title="繁育池列表")
        table.add_column("ID")
        table.add_column("名称")
        table.add_column("水体类型")
        table.add_column("容量")
        table.add_column("当前温度")
        table.add_column("品种ID")
        
        for item in items:
            table.add_row(
                item['id'],
                item['name'],
                item['water_type'],
                f"{item['capacity']} m³",
                f"{item['current_temperature']}°C" if item['current_temperature'] else '-',
                item.get('species_id', '-')
            )
        
        console.print(table)
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@pond.command("create")
def create_pond(
    name: str = typer.Option(..., help="繁育池名称"),
    water_type: str = typer.Option("fresh", help="水体类型: fresh/salt/brackish"),
    capacity: float = typer.Option(..., help="容量(立方米)"),
    species_id: Optional[str] = typer.Option(None, help="绑定的品种ID")
):
    """创建新繁育池"""
    client = get_client()
    try:
        data = {
            "name": name,
            "water_type": water_type,
            "capacity": capacity,
            "species_id": species_id
        }
        result = client.create_pond(data)
        rprint(f"[green]✓ 繁育池创建成功[/green]")
        rprint(f"  ID: {result['id']}")
        rprint(f"  名称: {result['name']}")
    except Exception as e:
        rprint(f"[red]✗ 创建失败: {e}[/red]")
    finally:
        client.close()


@pond.command("temp")
def record_temperature(
    pond_id: str = typer.Option(..., help="繁育池ID"),
    temperature: float = typer.Option(..., help="当前水温")
):
    """记录水温（自动检测是否超范围）"""
    client = get_client()
    try:
        result = client.record_temperature(pond_id, temperature)
        rprint(f"[green]✓ 温度记录成功[/green]")
        rprint(f"  时间: {result['recorded_at']}")
        rprint(f"  温度: {result['temperature']}°C")
        rprint(f"\n  提示: 请查看待办事项确认是否有温度异常")
    except Exception as e:
        rprint(f"[red]✗ 记录失败: {e}[/red]")
    finally:
        client.close()


@app.group()
def breeding():
    """繁育管理"""
    pass


@breeding.command("list")
def list_breeding():
    """列出所有繁育周期"""
    client = get_client()
    try:
        items = client.list_breeding_cycles()
        if not items:
            rprint("[yellow]暂无繁育周期数据[/yellow]")
            return
        
        table = Table(title="繁育周期列表")
        table.add_column("ID")
        table.add_column("繁育池")
        table.add_column("品种")
        table.add_column("产卵日期")
        table.add_column("出苗日期")
        table.add_column("实际数量")
        
        for item in items:
            table.add_row(
                item['id'],
                item['pond_id'][:8] + '...',
                item['species_id'][:8] + '...',
                item['spawning_date'],
                item.get('emergence_date', '-'),
                str(item.get('actual_quantity', '-'))
            )
        
        console.print(table)
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@breeding.command("create")
def create_breeding(
    pond_id: str = typer.Option(..., help="繁育池ID"),
    species_id: str = typer.Option(..., help="品种ID"),
    spawning_date: str = typer.Option(..., help="产卵日期 (YYYY-MM-DD)"),
    expected_qty: Optional[int] = typer.Option(None, help="预期数量")
):
    """创建新繁育周期"""
    client = get_client()
    try:
        data = {
            "pond_id": pond_id,
            "species_id": species_id,
            "spawning_date": spawning_date,
            "expected_quantity": expected_qty
        }
        result = client.create_breeding_cycle(data)
        rprint(f"[green]✓ 繁育周期创建成功[/green]")
        rprint(f"  ID: {result['id']}")
        rprint(f"  产卵日期: {result['spawning_date']}")
    except Exception as e:
        rprint(f"[red]✗ 创建失败: {e}[/red]")
    finally:
        client.close()


@breeding.command("emerge")
def record_emergence(
    cycle_id: str = typer.Option(..., help="繁育周期ID"),
    emergence_date: str = typer.Option(..., help="出苗日期 (YYYY-MM-DD)"),
    actual_qty: int = typer.Option(..., help="实际出苗数量")
):
    """记录出苗信息（自动更新库存）"""
    client = get_client()
    try:
        data = {
            "emergence_date": emergence_date,
            "actual_quantity": actual_qty
        }
        result = client.update_breeding_cycle(cycle_id, data)
        rprint(f"[green]✓ 出苗信息记录成功[/green]")
        rprint(f"  出苗日期: {result['emergence_date']}")
        rprint(f"  实际数量: {result['actual_quantity']} 尾")
        rprint(f"\n  提示: 库存已自动更新")
    except Exception as e:
        rprint(f"[red]✗ 记录失败: {e}[/red]")
    finally:
        client.close()


@app.command()
def inventory():
    """查看库存"""
    client = get_client()
    try:
        items = client.list_inventory()
        if not items:
            rprint("[yellow]当前库存为空[/yellow]")
            return
        
        table = Table(title="库存列表")
        table.add_column("ID")
        table.add_column("品种ID")
        table.add_column("数量")
        table.add_column("繁育池")
        table.add_column("最后更新")
        
        for item in items:
            table.add_row(
                item['id'],
                item['species_id'][:8] + '...',
                f"{item['quantity']} 尾",
                (item.get('pond_id', '')[:8] + '...') if item.get('pond_id') else '-',
                item['last_updated'][:19]
            )
        
        console.print(table)
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@app.group()
def sales():
    """销售管理"""
    pass


@sales.command("list")
def list_orders():
    """列出所有销售订单"""
    client = get_client()
    try:
        items = client.list_sales_orders()
        if not items:
            rprint("[yellow]暂无销售订单[/yellow]")
            return
        
        table = Table(title="销售订单列表")
        table.add_column("ID")
        table.add_column("客户")
        table.add_column("品种")
        table.add_column("请求数量")
        table.add_column("批准数量")
        table.add_column("总金额")
        
        for item in items:
            table.add_row(
                item['id'][:12] + '...',
                item['customer_name'],
                item['species_id'][:8] + '...',
                str(item['requested_quantity']),
                str(item.get('approved_quantity', '-')),
                f"¥{item.get('total_amount', '-')}"
            )
        
        console.print(table)
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@sales.command("create")
def create_order(
    customer: str = typer.Option(..., help="客户名称"),
    species_id: str = typer.Option(..., help="品种ID"),
    quantity: int = typer.Option(..., help="请求数量"),
    unit_price: float = typer.Option(..., help="单价(元/尾)"),
    phone: Optional[str] = typer.Option(None, help="联系电话")
):
    """创建销售订单（自动检查库存）"""
    client = get_client()
    try:
        data = {
            "customer_name": customer,
            "customer_phone": phone,
            "species_id": species_id,
            "requested_quantity": quantity,
            "unit_price": unit_price
        }
        result = client.create_sales_order(data)
        
        order = result['order']
        rprint(f"[green]✓ 销售订单创建成功[/green]")
        rprint(f"  订单ID: {order['id']}")
        rprint(f"  客户: {order['customer_name']}")
        rprint(f"  请求数量: {order['requested_quantity']} 尾")
        rprint(f"  批准数量: {order.get('approved_quantity', 0)} 尾")
        rprint(f"  总金额: ¥{order.get('total_amount', 0)}")
        
        if not result['can_fullfill']:
            rprint(f"\n[yellow]⚠ 库存不足: {result['message']}[/yellow]")
    except Exception as e:
        rprint(f"[red]✗ 创建失败: {e}[/red]")
    finally:
        client.close()


@app.group()
def vehicle():
    """配送车辆管理"""
    pass


@vehicle.command("list")
def list_vehicles():
    """列出所有配送车辆"""
    client = get_client()
    try:
        items = client.list_vehicles()
        if not items:
            rprint("[yellow]暂无配送车辆[/yellow]")
            return
        
        table = Table(title="配送车辆列表")
        table.add_column("ID")
        table.add_column("车牌号")
        table.add_column("名称")
        table.add_column("容量")
        
        for item in items:
            table.add_row(
                item['id'],
                item['license_plate'],
                item['name'],
                f"{item.get('capacity', '-')} 尾"
            )
        
        console.print(table)
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@vehicle.command("create")
def create_vehicle(
    plate: str = typer.Option(..., help="车牌号"),
    name: str = typer.Option(..., help="车辆名称"),
    capacity: Optional[int] = typer.Option(None, help="容量(尾)")
):
    """创建新配送车辆"""
    client = get_client()
    try:
        data = {
            "license_plate": plate,
            "name": name,
            "capacity": capacity
        }
        result = client.create_vehicle(data)
        rprint(f"[green]✓ 配送车辆创建成功[/green]")
        rprint(f"  ID: {result['id']}")
        rprint(f"  车牌号: {result['license_plate']}")
    except Exception as e:
        rprint(f"[red]✗ 创建失败: {e}[/red]")
    finally:
        client.close()


@app.group()
def delivery():
    """配送任务管理"""
    pass


@delivery.command("list")
def list_deliveries():
    """列出所有配送任务"""
    client = get_client()
    try:
        items = client.list_delivery_tasks()
        if not items:
            rprint("[yellow]暂无配送任务[/yellow]")
            return
        
        table = Table(title="配送任务列表")
        table.add_column("ID")
        table.add_column("订单")
        table.add_column("车辆")
        table.add_column("司机")
        table.add_column("计划开始")
        table.add_column("状态")
        
        for item in items:
            table.add_row(
                item['id'][:12] + '...',
                item['order_id'][:8] + '...',
                item['vehicle_id'][:8] + '...',
                item['driver_name'],
                item['scheduled_start'][:19],
                item['status']
            )
        
        console.print(table)
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@delivery.command("create")
def create_delivery(
    order_id: str = typer.Option(..., help="销售订单ID"),
    vehicle_id: str = typer.Option(..., help="车辆ID"),
    driver: str = typer.Option(..., help="司机姓名"),
    address: str = typer.Option(..., help="配送地址"),
    start_time: str = typer.Option(..., help="计划开始时间 (YYYY-MM-DD HH:MM)"),
    driver_phone: Optional[str] = typer.Option(None, help="司机电话"),
    end_time: Optional[str] = typer.Option(None, help="计划结束时间")
):
    """创建配送任务（自动检查车辆可用性）"""
    client = get_client()
    try:
        scheduled_start = datetime.strptime(start_time, "%Y-%m-%d %H:%M")
        scheduled_end = None
        if end_time:
            scheduled_end = datetime.strptime(end_time, "%Y-%m-%d %H:%M")
        
        data = {
            "order_id": order_id,
            "vehicle_id": vehicle_id,
            "driver_name": driver,
            "driver_phone": driver_phone,
            "delivery_address": address,
            "scheduled_start": scheduled_start.isoformat(),
            "scheduled_end": scheduled_end.isoformat() if scheduled_end else None
        }
        result = client.create_delivery_task(data)
        
        if result.get('success'):
            task = result['task']
            rprint(f"[green]✓ 配送任务创建成功[/green]")
            rprint(f"  任务ID: {task['id']}")
            rprint(f"  计划时间: {task['scheduled_start']}")
        else:
            rprint(f"[red]✗ 创建失败: {result['message']}[/red]")
    except Exception as e:
        rprint(f"[red]✗ 创建失败: {e}[/red]")
    finally:
        client.close()


@app.group()
def todo():
    """待办事项管理"""
    pass


@todo.command("list")
def list_todos():
    """列出所有待办事项"""
    client = get_client()
    try:
        items = client.list_todos()
        if not items:
            rprint("[yellow]暂无待办事项[/yellow]")
            return
        
        table = Table(title="待办事项列表")
        table.add_column("ID")
        table.add_column("标题")
        table.add_column("优先级")
        table.add_column("状态")
        table.add_column("连续违规")
        
        for item in items:
            priority_color = {
                "urgent": "red",
                "high": "yellow",
                "medium": "blue",
                "low": "green"
            }.get(item['priority'], "white")
            
            status_color = {
                "pending": "yellow",
                "in_progress": "blue",
                "resolved": "green"
            }.get(item['status'], "white")
            
            table.add_row(
                item['id'][:12] + '...',
                item['title'],
                f"[{priority_color}]{item['priority']}[/{priority_color}]",
                f"[{status_color}]{item['status']}[/{status_color}]",
                str(item.get('consecutive_violations', 0))
            )
        
        console.print(table)
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@todo.command("show")
def show_todo(
    todo_id: str = typer.Option(..., help="待办事项ID")
):
    """查看待办详情"""
    client = get_client()
    try:
        item = client.get_todo(todo_id)
        rprint(f"[bold]待办详情[/bold]")
        rprint(f"  ID: {item['id']}")
        rprint(f"  标题: {item['title']}")
        rprint(f"  描述: {item['description']}")
        rprint(f"  优先级: {item['priority']}")
        rprint(f"  状态: {item['status']}")
        rprint(f"  分类: {item.get('category', '-')}")
        if item.get('consecutive_violations', 0) > 0:
            rprint(f"  连续违规: {item['consecutive_violations']} 次")
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


@todo.command("resolve")
def resolve_todo(
    todo_id: str = typer.Option(..., help="待办事项ID")
):
    """标记待办为已解决"""
    client = get_client()
    try:
        result = client.update_todo(todo_id, {"status": "resolved"})
        rprint(f"[green]✓ 待办已标记为已解决[/green]")
    except Exception as e:
        rprint(f"[red]✗ 操作失败: {e}[/red]")
    finally:
        client.close()


if __name__ == "__main__":
    app()
