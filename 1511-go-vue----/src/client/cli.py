import json
from datetime import date, datetime, timedelta
from typing import Optional

import click

from .api_client import APIClient


def format_json(data, pretty: bool = True):
    if pretty:
        return json.dumps(data, ensure_ascii=False, indent=2, default=str)
    return json.dumps(data, ensure_ascii=False, default=str)


def get_client() -> APIClient:
    return APIClient()


@click.group()
@click.version_option("0.1.0")
def cli():
    """茶叶加工厂管理系统命令行客户端"""
    pass


@cli.group()
def plot():
    """茶园地块管理"""
    pass


@plot.command("list")
def plot_list():
    """列出所有茶园地块"""
    client = get_client()
    plots = client.list_plots()
    if not plots:
        click.echo("暂无地块数据")
        return
    for p in plots:
        click.echo(f"[{p['id']}] {p['name']} - {p['location']} (面积: {p['area']}亩)")


@plot.command("create")
@click.option("--name", required=True, help="地块名称")
@click.option("--location", required=True, help="地块位置")
@click.option("--area", type=float, required=True, help="地块面积(亩)")
def plot_create(name, location, area):
    """创建茶园地块"""
    client = get_client()
    result = client.create_plot(name, location, area)
    click.echo(format_json(result))


@plot.command("get")
@click.argument("plot_id")
def plot_get(plot_id):
    """查看地块详情"""
    client = get_client()
    result = client.get_plot(plot_id)
    click.echo(format_json(result))


@plot.command("update")
@click.argument("plot_id")
@click.option("--name", help="地块名称")
@click.option("--location", help="地块位置")
@click.option("--area", type=float, help="地块面积(亩)")
def plot_update(plot_id, name, location, area):
    """更新地块信息"""
    client = get_client()
    data = {}
    if name is not None:
        data["name"] = name
    if location is not None:
        data["location"] = location
    if area is not None:
        data["area"] = area
    if not data:
        click.echo("请指定至少一个要更新的字段")
        return
    result = client.update_plot(plot_id, data)
    click.echo(format_json(result))


@plot.command("delete")
@click.argument("plot_id")
def plot_delete(plot_id):
    """删除地块"""
    client = get_client()
    result = client.delete_plot(plot_id)
    click.echo(format_json(result))


@cli.group()
def harvest():
    """采摘管理"""
    pass


@harvest.group()
def plan():
    """采摘计划管理"""
    pass


@plan.command("list")
def plan_list():
    """列出所有采摘计划"""
    client = get_client()
    plans = client.list_harvest_plans()
    if not plans:
        click.echo("暂无采摘计划")
        return
    for p in plans:
        click.echo(f"[{p['id']}] 地块: {p['plot_id']} 日期: {p['plan_date']} 预期: {p['expected_quantity']}kg")


@plan.command("create")
@click.option("--plot-id", required=True, help="地块ID")
@click.option("--date", "plan_date", help="采摘日期(YYYY-MM-DD)，默认明天")
@click.option("--qty", type=float, required=True, help="预期采摘量(kg)")
def plan_create(plot_id, plan_date, qty):
    """创建采摘计划"""
    client = get_client()
    if plan_date:
        d = datetime.strptime(plan_date, "%Y-%m-%d").date()
    else:
        d = date.today() + timedelta(days=1)
    result = client.create_harvest_plan(plot_id, d, qty)
    click.echo(format_json(result))


@plan.command("get")
@click.argument("plan_id")
def plan_get(plan_id):
    """查看采摘计划详情"""
    client = get_client()
    result = client.get_harvest_plan(plan_id)
    click.echo(format_json(result))


@plan.command("update")
@click.argument("plan_id")
@click.option("--date", "plan_date", help="采摘日期(YYYY-MM-DD)")
@click.option("--qty", type=float, help="预期采摘量(kg)")
def plan_update(plan_id, plan_date, qty):
    """更新采摘计划"""
    client = get_client()
    data = {}
    if plan_date:
        data["plan_date"] = datetime.strptime(plan_date, "%Y-%m-%d").date()
    if qty is not None:
        data["expected_quantity"] = qty
    if not data:
        click.echo("请指定至少一个要更新的字段")
        return
    result = client.update_harvest_plan(plan_id, data)
    click.echo(format_json(result))


@plan.command("delete")
@click.argument("plan_id")
def plan_delete(plan_id):
    """删除采摘计划"""
    client = get_client()
    result = client.delete_harvest_plan(plan_id)
    click.echo(format_json(result))


@harvest.group()
def record():
    """采摘记录管理"""
    pass


@record.command("list")
def record_list():
    """列出所有采摘记录"""
    client = get_client()
    records = client.list_harvest_records()
    if not records:
        click.echo("暂无采摘记录")
        return
    for r in records:
        click.echo(f"[{r['id']}] 计划: {r['plan_id']} 数量: {r['actual_quantity']}kg 等级: {r['fresh_leaf_grade']}")


@record.command("create")
@click.option("--plan-id", required=True, help="采摘计划ID")
@click.option("--qty", type=float, required=True, help="实际采摘量(kg)")
@click.option("--grade", type=click.Choice(["A", "B", "C"]), required=True, help="鲜叶等级(A/B/C)")
@click.option("--time", "harvest_time", help="采摘时间(YYYY-MM-DD HH:MM:SS)，默认当前时间")
def record_create(plan_id, qty, grade, harvest_time):
    """创建采摘记录"""
    client = get_client()
    t = None
    if harvest_time:
        t = datetime.strptime(harvest_time, "%Y-%m-%d %H:%M:%S")
    result = client.create_harvest_record(plan_id, qty, grade, t)
    click.echo(format_json(result))


@record.command("get")
@click.argument("record_id")
def record_get(record_id):
    """查看采摘记录详情"""
    client = get_client()
    result = client.get_harvest_record(record_id)
    click.echo(format_json(result))


@record.command("update")
@click.argument("record_id")
@click.option("--qty", type=float, help="实际采摘量(kg)")
@click.option("--grade", type=click.Choice(["A", "B", "C"]), help="鲜叶等级(A/B/C)")
@click.option("--time", "harvest_time", help="采摘时间(YYYY-MM-DD HH:MM:SS)")
def record_update(record_id, qty, grade, harvest_time):
    """更新采摘记录"""
    client = get_client()
    data = {}
    if qty is not None:
        data["actual_quantity"] = qty
    if grade:
        data["fresh_leaf_grade"] = grade
    if harvest_time:
        data["harvest_time"] = datetime.strptime(harvest_time, "%Y-%m-%d %H:%M:%S")
    if not data:
        click.echo("请指定至少一个要更新的字段")
        return
    result = client.update_harvest_record(record_id, data)
    click.echo(format_json(result))


@record.command("delete")
@click.argument("record_id")
def record_delete(record_id):
    """删除采摘记录"""
    client = get_client()
    result = client.delete_harvest_record(record_id)
    click.echo(format_json(result))


@cli.group()
def processing():
    """加工管理"""
    pass


@processing.group()
def batch():
    """炒制批次管理"""
    pass


@batch.command("list")
def batch_list():
    """列出所有炒制批次"""
    client = get_client()
    batches = client.list_processing_batches()
    if not batches:
        click.echo("暂无炒制批次")
        return
    for b in batches:
        output = b.get('output_quantity')
        status = "已完成" if output is not None else "进行中"
        click.echo(f"[{b['id']}] 采摘记录: {b['harvest_record_id']} 投叶: {b['input_quantity']}kg 状态: {status}")


@batch.command("create")
@click.option("--record-id", required=True, help="采摘记录ID")
@click.option("--qty", type=float, required=True, help="投叶量(kg)")
@click.option("--time", "start_time", help="开始时间(YYYY-MM-DD HH:MM:SS)，默认当前时间")
def batch_create(record_id, qty, start_time):
    """创建炒制批次"""
    client = get_client()
    t = None
    if start_time:
        t = datetime.strptime(start_time, "%Y-%m-%d %H:%M:%S")
    result = client.create_processing_batch(record_id, qty, t)
    click.echo(format_json(result))


@batch.command("get")
@click.argument("batch_id")
def batch_get(batch_id):
    """查看炒制批次详情"""
    client = get_client()
    result = client.get_processing_batch(batch_id)
    click.echo(format_json(result))


@batch.command("complete")
@click.argument("batch_id")
@click.option("--qty", type=float, required=True, help="成品量(kg)")
def batch_complete(batch_id, qty):
    """完成炒制批次"""
    client = get_client()
    result = client.complete_processing_batch(batch_id, qty)
    click.echo(format_json(result))


@batch.command("delete")
@click.argument("batch_id")
def batch_delete(batch_id):
    """删除炒制批次"""
    client = get_client()
    result = client.delete_processing_batch(batch_id)
    click.echo(format_json(result))


@processing.group()
def rating():
    """等级评定管理"""
    pass


@rating.command("list")
def rating_list():
    """列出所有等级评定"""
    client = get_client()
    ratings = client.list_quality_ratings()
    if not ratings:
        click.echo("暂无等级评定")
        return
    for r in ratings:
        click.echo(f"[{r['id']}] 批次: {r['batch_id']} 等级: {r['grade']}")


@rating.command("create")
@click.option("--batch-id", required=True, help="炒制批次ID")
@click.option("--grade", type=click.Choice(["特级", "一级", "二级", "三级"]), required=True, help="成品等级")
@click.option("--desc", "sensory_description", required=True, help="感官描述")
def rating_create(batch_id, grade, sensory_description):
    """创建等级评定"""
    client = get_client()
    result = client.create_quality_rating(batch_id, grade, sensory_description)
    click.echo(format_json(result))


@rating.command("get")
@click.argument("rating_id")
def rating_get(rating_id):
    """查看等级评定详情"""
    client = get_client()
    result = client.get_quality_rating(rating_id)
    click.echo(format_json(result))


@rating.command("update")
@click.argument("rating_id")
@click.option("--grade", type=click.Choice(["特级", "一级", "二级", "三级"]), help="成品等级")
@click.option("--desc", "sensory_description", help="感官描述")
def rating_update(rating_id, grade, sensory_description):
    """更新等级评定"""
    client = get_client()
    data = {}
    if grade:
        data["grade"] = grade
    if sensory_description:
        data["sensory_description"] = sensory_description
    if not data:
        click.echo("请指定至少一个要更新的字段")
        return
    result = client.update_quality_rating(rating_id, data)
    click.echo(format_json(result))


@rating.command("delete")
@click.argument("rating_id")
def rating_delete(rating_id):
    """删除等级评定"""
    client = get_client()
    result = client.delete_quality_rating(rating_id)
    click.echo(format_json(result))


@cli.group()
def scheduler():
    """调度器管理"""
    pass


@scheduler.command("run")
def scheduler_run():
    """运行调度器（生成待办、检查告警）"""
    client = get_client()
    result = client.run_scheduler()
    click.echo(f"生成待办: {result['todos_created']}个")
    click.echo(f"生成告警: {result['alerts_created']}个")
    if result["todos"]:
        click.echo("\n待办事项:")
        for t in result["todos"]:
            click.echo(f"  - {t['title']}")
    if result["alerts"]:
        click.echo("\n紧急告警:")
        for a in result["alerts"]:
            click.echo(f"  - {a['message']}")


@cli.group()
def todo():
    """待办事项管理"""
    pass


@todo.command("list")
@click.option("--status", type=click.Choice(["pending", "completed"]), help="状态过滤")
def todo_list(status):
    """列出待办事项"""
    client = get_client()
    todos = client.list_todos(status)
    if not todos:
        click.echo("暂无待办事项")
        return
    for t in todos:
        status_str = "✓" if t["status"] == "completed" else "○"
        click.echo(f"{status_str} [{t['id']}] {t['title']} (截止: {t['due_date']})")


@todo.command("complete")
@click.argument("todo_id")
def todo_complete(todo_id):
    """完成待办事项"""
    client = get_client()
    result = client.complete_todo(todo_id)
    click.echo(format_json(result))


@cli.group()
def alert():
    """告警管理"""
    pass


@alert.command("list")
@click.option("--status", type=click.Choice(["active", "resolved"]), help="状态过滤")
def alert_list(status):
    """列出告警"""
    client = get_client()
    alerts = client.list_alerts(status)
    if not alerts:
        click.echo("暂无告警")
        return
    for a in alerts:
        status_str = "紧急" if a["status"] == "active" else "已处理"
        click.echo(f"[{status_str}] [{a['id']}] {a['message']}")


@alert.command("resolve")
@click.argument("alert_id")
def alert_resolve(alert_id):
    """标记告警已解决"""
    client = get_client()
    result = client.resolve_alert(alert_id)
    click.echo(format_json(result))


def main():
    cli()


if __name__ == "__main__":
    main()
