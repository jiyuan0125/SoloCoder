import json
import click
from .http_client import APIClient


def get_client():
    return APIClient()


def echo_json(data):
    click.echo(json.dumps(data, ensure_ascii=False, indent=2))


@click.group()
def cli():
    pass


@cli.group()
def enterprise():
    pass


@enterprise.command("create")
@click.option("--name", required=True, help="企业名称")
def enterprise_create(name):
    client = get_client()
    result = client.create_enterprise(name)
    echo_json(result)


@enterprise.command("list")
def enterprise_list():
    client = get_client()
    result = client.list_enterprises()
    echo_json(result)


@enterprise.command("get")
@click.argument("enterprise_id", type=int)
def enterprise_get(enterprise_id):
    client = get_client()
    result = client.get_enterprise(enterprise_id)
    echo_json(result)


@cli.group()
def facility():
    pass


@facility.command("create")
@click.option("--enterprise-id", type=int, required=True, help="企业ID")
@click.option("--name", required=True, help="设施名称")
@click.option("--type", "facility_type", required=True, help="设施类型")
@click.option("--installed-at", required=True, help="安装日期 (YYYY-MM-DD)")
@click.option("--running/--not-running", default=True, help="是否运行中")
def facility_create(enterprise_id, name, facility_type, installed_at, running):
    client = get_client()
    result = client.create_facility(enterprise_id, name, facility_type, installed_at, running)
    echo_json(result)


@facility.command("list")
@click.option("--enterprise-id", type=int, help="按企业ID筛选")
def facility_list(enterprise_id):
    client = get_client()
    result = client.list_facilities(enterprise_id)
    echo_json(result)


@facility.command("get")
@click.argument("facility_id", type=int)
def facility_get(facility_id):
    client = get_client()
    result = client.get_facility(facility_id)
    echo_json(result)


@cli.group()
def maintenance():
    pass


@maintenance.command("create")
@click.option("--facility-id", type=int, required=True, help="设施ID")
@click.option("--date", "maintenance_date", required=True, help="维保日期 (YYYY-MM-DD)")
@click.option("--description", help="维保描述")
@click.option("--part", multiple=True, help="配件: 名称,数量,单价 (可重复)")
def maintenance_create(facility_id, maintenance_date, description, part):
    parts = []
    for p in part:
        tokens = [t.strip() for t in p.split(",")]
        if len(tokens) >= 3:
            parts.append({
                "part_name": tokens[0],
                "quantity": int(tokens[1]),
                "unit_cost": float(tokens[2])
            })
        elif len(tokens) >= 1:
            parts.append({"part_name": tokens[0], "quantity": 1, "unit_cost": 0.0})
    client = get_client()
    result = client.create_maintenance(facility_id, maintenance_date, description, parts)
    echo_json(result)


@maintenance.command("list")
@click.option("--facility-id", type=int, help="按设施ID筛选")
@click.option("--start-date", help="开始日期 (YYYY-MM-DD)")
@click.option("--end-date", help="结束日期 (YYYY-MM-DD)")
def maintenance_list(facility_id, start_date, end_date):
    client = get_client()
    result = client.list_maintenances(facility_id, start_date, end_date)
    echo_json(result)


@maintenance.command("monthly-cost")
@click.option("--year", type=int, required=True)
@click.option("--month", type=int, required=True)
@click.option("--enterprise-id", type=int, help="按企业ID筛选")
def maintenance_monthly_cost(year, month, enterprise_id):
    client = get_client()
    result = client.monthly_maintenance_cost(year, month, enterprise_id)
    echo_json(result)


@cli.group()
def emission():
    pass


@emission.command("create")
@click.option("--facility-id", type=int, required=True, help="设施ID")
@click.option("--recorded-at", required=True, help="记录时间 (ISO 格式)")
@click.option("--pollutant", required=True, help="污染物名称")
@click.option("--value", type=float, required=True, help="排放值")
@click.option("--limit", "limit_value", type=float, required=True, help="限值")
def emission_create(facility_id, recorded_at, pollutant, value, limit_value):
    client = get_client()
    result = client.create_emission(facility_id, recorded_at, pollutant, value, limit_value)
    echo_json(result)


@emission.command("list")
@click.option("--facility-id", type=int, help="按设施ID筛选")
def emission_list(facility_id):
    client = get_client()
    result = client.list_emissions(facility_id)
    echo_json(result)


@cli.group()
def todo():
    pass


@todo.command("generate-maintenance")
def todo_generate_maintenance():
    client = get_client()
    result = client.generate_maintenance_todos()
    echo_json(result)


@todo.command("generate-compliance")
def todo_generate_compliance():
    client = get_client()
    result = client.generate_compliance_todos()
    echo_json(result)


@todo.command("list")
@click.option("--facility-id", type=int, help="按设施ID筛选")
@click.option("--status", help="按状态筛选")
def todo_list(facility_id, status):
    client = get_client()
    result = client.list_todos(facility_id, status)
    echo_json(result)


@todo.command("update-status")
@click.argument("todo_id", type=int)
@click.option("--status", required=True, help="新状态")
def todo_update_status(todo_id, status):
    client = get_client()
    result = client.update_todo_status(todo_id, status)
    echo_json(result)


@cli.group()
def alert():
    pass


@alert.command("list")
@click.option("--facility-id", type=int, help="按设施ID筛选")
@click.option("--resolved/--unresolved", default=None, help="按是否解决筛选")
def alert_list(facility_id, resolved):
    client = get_client()
    result = client.list_alerts(facility_id, resolved)
    echo_json(result)


@alert.command("resolve")
@click.argument("alert_id", type=int)
def alert_resolve(alert_id):
    client = get_client()
    result = client.resolve_alert(alert_id)
    echo_json(result)


@cli.group()
def stats():
    pass


@stats.command("monthly")
@click.option("--year", type=int, required=True)
@click.option("--month", type=int, required=True)
def stats_monthly(year, month):
    client = get_client()
    result = client.monthly_stats(year, month)
    echo_json(result)


@stats.command("monthly-facility")
@click.option("--year", type=int, required=True)
@click.option("--month", type=int, required=True)
@click.option("--facility-id", type=int, required=True)
def stats_monthly_facility(year, month, facility_id):
    client = get_client()
    result = client.facility_monthly_stats(year, month, facility_id)
    echo_json(result)


def main():
    cli()
