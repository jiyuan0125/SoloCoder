import json
import sys
from datetime import date, datetime
from typing import Optional

import click

from .api_client import APIClient


def get_client() -> APIClient:
    return APIClient()


def print_json(data):
    if data is None:
        click.echo("操作成功")
        return
    click.echo(json.dumps(data, ensure_ascii=False, indent=2))


def parse_date(s: str) -> date:
    return date.fromisoformat(s)


def parse_datetime(s: str) -> datetime:
    return datetime.fromisoformat(s)


@click.group()
def cli():
    """矿山企业综合管理系统命令行客户端"""
    pass


@cli.group()
def mine():
    """矿区管理"""
    pass


@mine.command("list")
def mine_list():
    """列出所有矿区"""
    client = get_client()
    mines = client.list_mines()
    print_json(mines)


@mine.command("create")
@click.option("--name", required=True, help="矿区名称")
@click.option("--mineral-type", required=True, help="矿种")
@click.option("--annual-capacity", type=float, required=True, help="年产能（吨）")
def mine_create(name, mineral_type, annual_capacity):
    """创建新矿区"""
    client = get_client()
    data = {
        "name": name,
        "mineral_type": mineral_type,
        "annual_capacity": annual_capacity,
    }
    result = client.create_mine(data)
    print_json(result)


@mine.command("get")
@click.argument("mine_id", type=int)
def mine_get(mine_id):
    """获取矿区详情"""
    client = get_client()
    result = client.get_mine(mine_id)
    print_json(result)


@mine.command("update")
@click.argument("mine_id", type=int)
@click.option("--name", help="矿区名称")
@click.option("--mineral-type", help="矿种")
@click.option("--annual-capacity", type=float, help="年产能（吨）")
def mine_update(mine_id, name, mineral_type, annual_capacity):
    """更新矿区信息"""
    client = get_client()
    data = client.get_mine(mine_id)
    if name:
        data["name"] = name
    if mineral_type:
        data["mineral_type"] = mineral_type
    if annual_capacity:
        data["annual_capacity"] = annual_capacity
    result = client.update_mine(mine_id, data)
    print_json(result)


@mine.command("delete")
@click.argument("mine_id", type=int)
def mine_delete(mine_id):
    """删除矿区"""
    client = get_client()
    client.delete_mine(mine_id)
    click.echo("删除成功")


@cli.group()
def mining():
    """采矿作业管理"""
    pass


@mining.command("list")
@click.option("--mine-id", type=int, help="矿区ID")
def mining_list(mine_id):
    """列出采矿作业记录"""
    client = get_client()
    result = client.list_mining_operations(mine_id)
    print_json(result)


@mining.command("create")
@click.option("--mine-id", type=int, required=True, help="矿区ID")
@click.option("--operation-date", required=True, help="作业日期 (YYYY-MM-DD)")
@click.option("--planned-output", type=float, required=True, help="计划产量（吨）")
@click.option("--actual-output", type=float, required=True, help="实际产量（吨）")
def mining_create(mine_id, operation_date, planned_output, actual_output):
    """创建采矿作业记录"""
    client = get_client()
    data = {
        "mine_id": mine_id,
        "operation_date": parse_date(operation_date).isoformat(),
        "planned_output": planned_output,
        "actual_output": actual_output,
    }
    result = client.create_mining_operation(data)
    print_json(result)


@mining.command("get")
@click.argument("op_id", type=int)
def mining_get(op_id):
    """获取采矿作业详情"""
    client = get_client()
    result = client.get_mining_operation(op_id)
    print_json(result)


@mining.command("delete")
@click.argument("op_id", type=int)
def mining_delete(op_id):
    """删除采矿作业记录"""
    client = get_client()
    client.delete_mining_operation(op_id)
    click.echo("删除成功")


@cli.group()
def transport():
    """运输管理"""
    pass


@transport.command("list")
@click.option("--mine-id", type=int, help="矿区ID")
def transport_list(mine_id):
    """列出运输记录"""
    client = get_client()
    result = client.list_transports(mine_id)
    print_json(result)


@transport.command("create")
@click.option("--mine-id", type=int, required=True, help="矿区ID")
@click.option("--vehicle-number", required=True, help="车牌号")
@click.option("--transport-date", required=True, help="运输日期 (YYYY-MM-DD)")
@click.option("--departure-time", required=True, help="出发时间 (YYYY-MM-DDTHH:MM:SS)")
@click.option("--arrival-time", required=True, help="到达时间 (YYYY-MM-DDTHH:MM:SS)")
@click.option("--transport-volume", type=float, required=True, help="运输量（吨）")
def transport_create(mine_id, vehicle_number, transport_date, departure_time, arrival_time, transport_volume):
    """创建运输记录"""
    client = get_client()
    data = {
        "mine_id": mine_id,
        "vehicle_number": vehicle_number,
        "transport_date": parse_date(transport_date).isoformat(),
        "departure_time": parse_datetime(departure_time).isoformat(),
        "arrival_time": parse_datetime(arrival_time).isoformat(),
        "transport_volume": transport_volume,
    }
    result = client.create_transport(data)
    print_json(result)


@transport.command("get")
@click.argument("transport_id", type=int)
def transport_get(transport_id):
    """获取运输记录详情"""
    client = get_client()
    result = client.get_transport(transport_id)
    print_json(result)


@transport.command("delete")
@click.argument("transport_id", type=int)
def transport_delete(transport_id):
    """删除运输记录"""
    client = get_client()
    client.delete_transport(transport_id)
    click.echo("删除成功")


@cli.group()
def safety():
    """安全检查管理"""
    pass


@safety.command("list")
@click.option("--mine-id", type=int, help="矿区ID")
def safety_list(mine_id):
    """列出安全检查记录"""
    client = get_client()
    result = client.list_safety_checks(mine_id)
    print_json(result)


@safety.command("create")
@click.option("--mine-id", type=int, required=True, help="矿区ID")
@click.option("--check-date", required=True, help="检查日期 (YYYY-MM-DD)")
@click.option("--items", required=True, help="检查项JSON，格式: [{'name': '项名', 'is_abnormal': true/false}]")
def safety_create(mine_id, check_date, items):
    """创建安全检查记录"""
    client = get_client()
    items_data = json.loads(items)
    data = {
        "mine_id": mine_id,
        "check_date": parse_date(check_date).isoformat(),
        "items": items_data,
    }
    result = client.create_safety_check(data)
    print_json(result)


@safety.command("get")
@click.argument("check_id", type=int)
def safety_get(check_id):
    """获取安全检查详情"""
    client = get_client()
    result = client.get_safety_check(check_id)
    print_json(result)


@safety.command("delete")
@click.argument("check_id", type=int)
def safety_delete(check_id):
    """删除安全检查记录"""
    client = get_client()
    client.delete_safety_check(check_id)
    click.echo("删除成功")


@cli.group()
def todo():
    """待办整改管理"""
    pass


@todo.command("list")
@click.option("--mine-id", type=int, help="矿区ID")
@click.option("--status", type=click.Choice(["pending", "in_progress", "completed"]), help="状态")
def todo_list(mine_id, status):
    """列出待办整改记录"""
    client = get_client()
    result = client.list_todos(mine_id, status)
    print_json(result)


@todo.command("create")
@click.option("--mine-id", type=int, required=True, help="矿区ID")
@click.option("--safety-check-id", type=int, help="关联的安全检查ID")
@click.option("--description", required=True, help="整改描述")
@click.option("--responsible-person", required=True, help="责任人")
@click.option("--deadline", required=True, help="整改期限 (YYYY-MM-DD)")
@click.option("--priority", type=click.Choice(["low", "medium", "high", "urgent"]), default="medium", help="优先级")
def todo_create(mine_id, safety_check_id, description, responsible_person, deadline, priority):
    """创建待办整改记录"""
    client = get_client()
    data = {
        "mine_id": mine_id,
        "safety_check_id": safety_check_id,
        "description": description,
        "responsible_person": responsible_person,
        "deadline": parse_date(deadline).isoformat(),
        "priority": priority,
    }
    result = client.create_todo(data)
    print_json(result)


@todo.command("get")
@click.argument("todo_id", type=int)
def todo_get(todo_id):
    """获取待办整改详情"""
    client = get_client()
    result = client.get_todo(todo_id)
    print_json(result)


@todo.command("update-status")
@click.argument("todo_id", type=int)
@click.option("--status", type=click.Choice(["pending", "in_progress", "completed"]), required=True, help="新状态")
def todo_update_status(todo_id, status):
    """更新待办整改状态"""
    client = get_client()
    data = client.get_todo(todo_id)
    data["status"] = status
    result = client.update_todo(todo_id, data)
    print_json(result)


@todo.command("delete")
@click.argument("todo_id", type=int)
def todo_delete(todo_id):
    """删除待办整改记录"""
    client = get_client()
    client.delete_todo(todo_id)
    click.echo("删除成功")


@todo.command("upgrade-overdue")
def todo_upgrade_overdue():
    """升级超期待办整改的优先级"""
    client = get_client()
    count = client.upgrade_overdue_todos()
    click.echo(f"已升级 {count} 个超期待办的优先级")


@cli.command()
@click.option("--year", type=int, help="年份")
@click.option("--month", type=int, help="月份")
def statistics(year, month):
    """获取月度统计指标"""
    client = get_client()
    result = client.get_monthly_statistics(year, month)
    print_json(result)


def main():
    try:
        cli()
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)
        sys.exit(1)
    except Exception as e:
        click.echo(f"意外错误: {e}", err=True)
        sys.exit(1)


if __name__ == "__main__":
    main()
