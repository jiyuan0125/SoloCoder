import json
from datetime import datetime
from typing import Optional
import click
from .api_client import APIClient


client = APIClient()


def print_json(data):
    click.echo(json.dumps(data, ensure_ascii=False, indent=2))


@click.group()
def cli():
    """温室大棚管理系统命令行客户端"""
    pass


@cli.group()
def greenhouse():
    """大棚管理"""
    pass


@greenhouse.command("create")
@click.option("--code", required=True, help="大棚编号")
@click.option("--area", type=float, required=True, help="面积")
@click.option("--crop", required=True, help="种植作物")
@click.option("--type", "greenhouse_type", type=click.Choice(["glass", "plastic"]), required=True, help="大棚类型")
def create_greenhouse(code, area, crop, greenhouse_type):
    """创建大棚"""
    result = client.create_greenhouse(code, area, crop, greenhouse_type)
    print_json(result)


@greenhouse.command("list")
def list_greenhouses():
    """列出所有大棚"""
    result = client.list_greenhouses()
    print_json(result)


@greenhouse.command("get")
@click.argument("greenhouse_id")
def get_greenhouse(greenhouse_id):
    """获取大棚详情"""
    result = client.get_greenhouse(greenhouse_id)
    print_json(result)


@cli.command("report-env")
@click.argument("greenhouse_id")
@click.option("--temp", type=float, required=True, help="温度")
@click.option("--hum", type=float, required=True, help="湿度")
@click.option("--soil", type=float, required=True, help="土壤湿度")
@click.option("--light", type=float, required=True, help="光照")
def report_env(greenhouse_id, temp, hum, soil, light):
    """上报环境数据"""
    result = client.report_environment(greenhouse_id, temp, hum, soil, light)
    print_json(result)


@cli.command("aggregate-env")
@click.argument("greenhouse_id")
@click.option(
    "--metric",
    type=click.Choice(["temperature", "humidity", "soil_moisture", "light"]),
    required=True,
    help="聚合指标",
)
@click.option("--start", required=True, help="开始时间 (ISO格式)")
@click.option("--end", required=True, help="结束时间 (ISO格式)")
def aggregate_env(greenhouse_id, metric, start, end):
    """聚合查询环境数据"""
    start_time = datetime.fromisoformat(start)
    end_time = datetime.fromisoformat(end)
    result = client.aggregate_environment(greenhouse_id, metric, start_time, end_time)
    print_json(result)


@cli.group()
def irrigation():
    """灌溉计划管理"""
    pass


@irrigation.command("create")
@click.argument("greenhouse_id")
@click.option("--time", required=True, help="执行时间 (ISO格式，过去时间表示立即执行)")
@click.option("--water", type=float, required=True, help="用水量")
def create_irrigation(greenhouse_id, time, water):
    """创建灌溉计划"""
    execution_time = datetime.fromisoformat(time)
    result = client.create_irrigation_plan(greenhouse_id, execution_time, water)
    print_json(result)


@irrigation.command("list")
@click.option("--greenhouse-id", help="大棚ID过滤")
@click.option("--status", type=click.Choice(["pending", "executed"]), help="状态过滤")
def list_irrigation(greenhouse_id, status):
    """列出灌溉计划"""
    result = client.list_irrigation_plans(greenhouse_id, status)
    print_json(result)


@cli.group()
def fertilizer():
    """施肥计划管理"""
    pass


@fertilizer.command("create")
@click.argument("greenhouse_id")
@click.option("--time", required=True, help="执行时间 (ISO格式，过去时间表示立即执行)")
@click.option("--amount", type=float, required=True, help="施肥量 (不能为零)")
def create_fertilizer(greenhouse_id, time, amount):
    """创建施肥计划"""
    execution_time = datetime.fromisoformat(time)
    result = client.create_fertilizer_plan(greenhouse_id, execution_time, amount)
    print_json(result)


@fertilizer.command("list")
@click.option("--greenhouse-id", help="大棚ID过滤")
@click.option("--status", type=click.Choice(["pending", "executed"]), help="状态过滤")
def list_fertilizer(greenhouse_id, status):
    """列出施肥计划"""
    result = client.list_fertilizer_plans(greenhouse_id, status)
    print_json(result)


@cli.command("run-scheduler")
def run_scheduler():
    """手动触发调度器执行"""
    result = client.run_scheduler()
    print_json(result)


@cli.command("health")
def health():
    """检查服务健康状态"""
    result = client.health_check()
    print_json(result)


if __name__ == "__main__":
    cli()
