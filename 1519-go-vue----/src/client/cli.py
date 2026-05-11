import click
import json
from decimal import Decimal
from datetime import datetime, date
from typing import List, Tuple

from src.client.api_client import client


def parse_materials(ctx, param, value) -> List[dict]:
    if not value:
        return []
    materials = []
    try:
        for item in value:
            name, pct_str = item.rsplit(":", 1)
            pct = Decimal(pct_str)
            materials.append({"name": name, "percentage": str(pct)})
        return materials
    except Exception as e:
        raise click.BadParameter(f"原材料格式错误，请使用 '名称:占比' 格式: {e}")


@click.group()
def cli():
    """饲料加工厂管理系统 - 命令行客户端"""
    pass


@cli.group()
def recipe():
    """配方管理"""
    pass


@recipe.command("create")
@click.option("--name", required=True, help="配方名称")
@click.option("--description", help="配方描述")
@click.option("--material", "materials", multiple=True, callback=parse_materials,
              help="原材料，格式为 '名称:占比'，可多次使用（至少2种）")
def create_recipe(name, description, materials):
    """创建新配方"""
    if len(materials) < 2:
        click.echo("错误: 配方至少需要2种原材料", err=True)
        return

    data = {
        "name": name,
        "description": description,
        "raw_materials": materials
    }
    try:
        result = client.post("/recipes/", json=data)
        click.echo(f"配方创建成功: ID={result['id']}")
        _print_recipe(result)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@recipe.command("list")
def list_recipes():
    """列出所有配方"""
    try:
        recipes = client.get("/recipes/")
        if not recipes:
            click.echo("暂无配方")
            return
        for r in recipes:
            _print_recipe(r)
            click.echo("-" * 40)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@recipe.command("show")
@click.argument("recipe_id", type=int)
def show_recipe(recipe_id):
    """查看配方详情"""
    try:
        result = client.get(f"/recipes/{recipe_id}")
        _print_recipe(result)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@recipe.command("deactivate")
@click.argument("recipe_id", type=int)
def deactivate_recipe_cmd(recipe_id):
    """停用配方"""
    try:
        result = client.post(f"/recipes/{recipe_id}/deactivate")
        click.echo(result.get("message", "操作成功"))
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@recipe.command("activate")
@click.argument("recipe_id", type=int)
def activate_recipe_cmd(recipe_id):
    """启用配方"""
    try:
        result = client.post(f"/recipes/{recipe_id}/activate")
        click.echo(result.get("message", "操作成功"))
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


def _print_recipe(recipe: dict):
    click.echo(f"配方ID: {recipe['id']}")
    click.echo(f"名称: {recipe['name']}")
    click.echo(f"状态: {recipe['status']}")
    if recipe.get('description'):
        click.echo(f"描述: {recipe['description']}")
    click.echo("原材料:")
    for mat in recipe.get("raw_materials", []):
        click.echo(f"  - {mat['name']}: {mat['percentage']}%")


@cli.group()
def batch():
    """生产批次管理"""
    pass


@batch.command("create")
@click.option("--recipe-id", required=True, type=int, help="关联的配方ID")
@click.option("--planned", required=True, help="计划产量")
@click.option("--actual", help="实际产量（可选）")
def create_batch(recipe_id, planned, actual):
    """创建生产批次"""
    data = {
        "recipe_id": recipe_id,
        "planned_quantity": str(Decimal(planned))
    }
    if actual:
        data["actual_quantity"] = str(Decimal(actual))

    try:
        result = client.post("/batches/", json=data)
        click.echo(f"批次创建成功: ID={result['id']}")
        _print_batch(result)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@batch.command("list")
def list_batches():
    """列出所有生产批次"""
    try:
        batches = client.get("/batches/")
        if not batches:
            click.echo("暂无生产批次")
            return
        for b in batches:
            _print_batch(b)
            click.echo("-" * 40)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@batch.command("show")
@click.argument("batch_id", type=int)
def show_batch(batch_id):
    """查看批次详情"""
    try:
        result = client.get(f"/batches/{batch_id}")
        _print_batch(result)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@batch.command("update")
@click.argument("batch_id", type=int)
@click.option("--actual", help="实际产量")
@click.option("--status", type=click.Choice(["pending", "in_progress", "completed"]), help="批次状态")
def update_batch_cmd(batch_id, actual, status):
    """更新生产批次"""
    data = {}
    if actual:
        data["actual_quantity"] = str(Decimal(actual))
    if status:
        data["status"] = status

    if not data:
        click.echo("错误: 至少需要指定一个更新项", err=True)
        return

    try:
        result = client.put(f"/batches/{batch_id}", json=data)
        click.echo("批次更新成功")
        _print_batch(result)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


def _print_batch(batch: dict):
    click.echo(f"批次ID: {batch['id']}")
    click.echo(f"配方ID: {batch['recipe_id']}")
    click.echo(f"状态: {batch['status']}")
    click.echo(f"计划产量: {batch['planned_quantity']}")
    click.echo(f"实际产量: {batch.get('actual_quantity', '-')}")
    click.echo(f"创建时间: {batch['created_at']}")


@cli.group()
def inspection():
    """质检管理"""
    pass


@inspection.command("create")
@click.option("--batch-id", required=True, type=int, help="批次ID")
@click.option("--status", required=True, type=click.Choice(["pass", "fail"]), help="质检结果")
@click.option("--notes", help="备注")
def create_inspection_cmd(batch_id, status, notes):
    """创建质检记录"""
    data = {
        "batch_id": batch_id,
        "status": status
    }
    if notes:
        data["notes"] = notes

    try:
        result = client.post("/inspections/", json=data)
        click.echo(f"质检记录创建成功: ID={result['id']}")
        _print_inspection(result)
        if status == "fail":
            click.echo("注意: 质检不合格，已生成待办事项")
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@inspection.command("list")
@click.argument("batch_id", type=int)
def list_inspections(batch_id):
    """列出批次的质检记录"""
    try:
        inspections = client.get(f"/batches/{batch_id}/inspections")
        if not inspections:
            click.echo("该批次暂无质检记录")
            return
        for i in inspections:
            _print_inspection(i)
            click.echo("-" * 40)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


def _print_inspection(inspection: dict):
    click.echo(f"质检ID: {inspection['id']}")
    click.echo(f"批次ID: {inspection['batch_id']}")
    click.echo(f"结果: {inspection['status']}")
    if inspection.get('notes'):
        click.echo(f"备注: {inspection['notes']}")
    click.echo(f"时间: {inspection['created_at']}")


@cli.group()
def todo():
    """待办事项管理"""
    pass


@todo.command("list")
def list_todos():
    """列出所有待办事项"""
    try:
        todos = client.get("/todos/")
        if not todos:
            click.echo("暂无待办事项")
            return
        for t in todos:
            _print_todo(t)
            click.echo("-" * 40)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@todo.command("show")
@click.argument("todo_id", type=int)
def show_todo(todo_id):
    """查看待办详情"""
    try:
        result = client.get(f"/todos/{todo_id}")
        _print_todo(result)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@todo.command("update")
@click.argument("todo_id", type=int)
@click.option("--status", required=True,
              type=click.Choice(["in_progress", "resolved"]),
              help="状态")
@click.option("--resolution",
              type=click.Choice(["rework", "downgrade", "scrap"]),
              help="处理方式（仅resolved状态需要）")
@click.option("--notes", help="备注")
def update_todo_cmd(todo_id, status, resolution, notes):
    """更新待办事项"""
    data = {"status": status}
    if notes:
        data["notes"] = notes
    if resolution:
        data["resolution"] = resolution

    if status == "resolved" and not resolution:
        click.echo("错误: 标记为已处理时必须指定处理方式", err=True)
        return

    try:
        result = client.put(f"/todos/{todo_id}", json=data)
        click.echo("待办更新成功")
        _print_todo(result)
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


def _print_todo(todo: dict):
    overdue_marker = " [逾期]" if todo.get('is_overdue') and todo['status'] in ['pending', 'overdue'] else ""
    click.echo(f"待办ID: {todo['id']}{overdue_marker}")
    click.echo(f"批次ID: {todo['batch_id']}")
    click.echo(f"状态: {todo['status']}")
    if todo.get('resolution'):
        click.echo(f"处理方式: {todo['resolution']}")
    if todo.get('notes'):
        click.echo(f"备注: {todo['notes']}")
    click.echo(f"创建时间: {todo['created_at']}")
    if todo.get('resolved_at'):
        click.echo(f"解决时间: {todo['resolved_at']}")


@cli.command("export")
@click.option("--start", required=True, help="开始日期 (YYYY-MM-DD)")
@click.option("--end", required=True, help="结束日期 (YYYY-MM-DD)")
def export_batches_cmd(start, end):
    """导出生产批次汇总"""
    try:
        params = {"start": start, "end": end}
        result = client.get("/export/batches", params=params)
        if "text" in result:
            click.echo(result["text"])
        else:
            click.echo("无数据")
    except RuntimeError as e:
        click.echo(f"错误: {e}", err=True)


@cli.command("server")
@click.option("--port", default=8000, help="端口号")
def run_server(port):
    """启动服务端（仅用于便捷启动）"""
    import os
    os.environ["PORT"] = str(port)
    from src.server import __main__ as server_main


if __name__ == "__main__":
    cli()
