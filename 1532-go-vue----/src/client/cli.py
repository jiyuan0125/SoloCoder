import json
import click
from typing import Optional, List
from src.client.api_client import APIClient


def print_json(data):
    click.echo(json.dumps(data, ensure_ascii=False, indent=2))


def get_client() -> APIClient:
    return APIClient()


@click.group()
@click.option("--url", envvar="REACTOR_API_URL", default="http://localhost:8000",
              help="Server API URL")
@click.pass_context
def cli(ctx, url):
    ctx.ensure_object(dict)
    ctx.obj["client"] = APIClient(base_url=url)


@cli.group()
def reactor():
    pass


@reactor.command("list")
@click.pass_context
def list_reactors(ctx):
    client = ctx.obj["client"]
    print_json(client.list_reactors())


@reactor.command("create")
@click.option("--name", required=True)
@click.option("--temp-max", type=float, required=True)
@click.option("--press-max", type=float, required=True)
@click.option("--description", default="")
@click.pass_context
def create_reactor(ctx, name, temp_max, press_max, description):
    client = ctx.obj["client"]
    result = client.create_reactor(name, temp_max, press_max, description)
    print_json(result)


@reactor.command("get")
@click.argument("reactor_id")
@click.pass_context
def get_reactor(ctx, reactor_id):
    client = ctx.obj["client"]
    print_json(client.get_reactor(reactor_id))


@reactor.command("delete")
@click.argument("reactor_id")
@click.pass_context
def delete_reactor(ctx, reactor_id):
    client = ctx.obj["client"]
    client.delete_reactor(reactor_id)
    click.echo(f"Reactor {reactor_id} deleted.")


@cli.group()
def recipe():
    pass


@recipe.command("list")
@click.pass_context
def list_recipes(ctx):
    client = ctx.obj["client"]
    print_json(client.list_recipes())


@recipe.command("create")
@click.option("--name", required=True)
@click.option("--time-window", type=int, required=True, help="Time window in seconds")
@click.option("--material", multiple=True, 
              help="Material in format: name,amount,unit,order (e.g. 'A,100,kg,1')")
@click.option("--description", default="")
@click.pass_context
def create_recipe(ctx, name, time_window, material, description):
    materials = []
    for m in material:
        parts = m.split(",")
        if len(parts) != 4:
            raise click.ClickException(
                "Material format: name,amount,unit,order"
            )
        materials.append({
            "name": parts[0].strip(),
            "required_amount": float(parts[1].strip()),
            "unit": parts[2].strip(),
            "order": int(parts[3].strip())
        })
    client = ctx.obj["client"]
    result = client.create_recipe(name, materials, time_window, description)
    print_json(result)


@recipe.command("get")
@click.argument("recipe_id")
@click.pass_context
def get_recipe(ctx, recipe_id):
    client = ctx.obj["client"]
    print_json(client.get_recipe(recipe_id))


@cli.group()
def sensor():
    pass


@sensor.command("report")
@click.option("--reactor-id", required=True)
@click.option("--temperature", type=float, required=True)
@click.option("--pressure", type=float, required=True)
@click.pass_context
def report_reading(ctx, reactor_id, temperature, pressure):
    client = ctx.obj["client"]
    result = client.report_sensor_reading(reactor_id, temperature, pressure)
    print_json(result)


@sensor.command("readings")
@click.option("--reactor-id", default=None)
@click.pass_context
def list_readings(ctx, reactor_id):
    client = ctx.obj["client"]
    print_json(client.list_readings(reactor_id))


@sensor.command("alarms")
@click.option("--reactor-id", default=None)
@click.option("--high-risk", is_flag=True, default=None)
@click.pass_context
def list_alarms(ctx, reactor_id, high_risk):
    client = ctx.obj["client"]
    is_high_risk = True if high_risk else None
    print_json(client.list_alarms(reactor_id, is_high_risk))


@cli.group()
def batch():
    pass


@batch.command("list")
@click.option("--reactor-id", default=None)
@click.option("--status", default=None, 
              type=click.Choice(["running", "completed", "aborted", "abnormal"]))
@click.pass_context
def list_batches(ctx, reactor_id, status):
    client = ctx.obj["client"]
    print_json(client.list_batches(reactor_id, status))


@batch.command("create")
@click.option("--reactor-id", required=True)
@click.option("--recipe-id", required=True)
@click.pass_context
def create_batch(ctx, reactor_id, recipe_id):
    client = ctx.obj["client"]
    result = client.create_batch(reactor_id, recipe_id)
    print_json(result)


@batch.command("get")
@click.argument("batch_id")
@click.pass_context
def get_batch(ctx, batch_id):
    client = ctx.obj["client"]
    print_json(client.get_batch(batch_id))


@batch.command("feed")
@click.option("--batch-id", required=True)
@click.option("--material", required=True)
@click.option("--req-amount", type=float, required=True)
@click.option("--act-amount", type=float, required=True)
@click.option("--unit", required=True)
@click.option("--order", type=int, required=True)
@click.pass_context
def record_feeding(ctx, batch_id, material, req_amount, act_amount, unit, order):
    client = ctx.obj["client"]
    result = client.record_feeding(
        batch_id, material, req_amount, act_amount, unit, order
    )
    print_json(result)


@batch.command("complete")
@click.argument("batch_id")
@click.pass_context
def complete_batch(ctx, batch_id):
    client = ctx.obj["client"]
    result = client.complete_batch(batch_id)
    print_json(result)


@batch.command("feedings")
@click.argument("batch_id")
@click.pass_context
def list_feedings(ctx, batch_id):
    client = ctx.obj["client"]
    print_json(client.list_feedings(batch_id))


@cli.group()
def stats():
    pass


@stats.command("daily")
@click.option("--start", default=None, help="Start date (YYYY-MM-DD)")
@click.option("--end", default=None, help="End date (YYYY-MM-DD)")
@click.pass_context
def daily_stats(ctx, start, end):
    client = ctx.obj["client"]
    print_json(client.get_daily_stats(start, end))


@stats.command("calculate")
@click.argument("date_str")
@click.pass_context
def calculate_daily(ctx, date_str):
    client = ctx.obj["client"]
    result = client.calculate_daily_stats(date_str)
    print_json(result)


@cli.group()
def scheduler():
    pass


@scheduler.command("status")
@click.pass_context
def scheduler_status(ctx):
    client = ctx.obj["client"]
    print_json(client.get_scheduler_status())


@scheduler.command("start")
@click.pass_context
def scheduler_start(ctx):
    client = ctx.obj["client"]
    print_json(client.start_scheduler())


@scheduler.command("stop")
@click.pass_context
def scheduler_stop(ctx):
    client = ctx.obj["client"]
    print_json(client.stop_scheduler())


def main():
    cli()


if __name__ == "__main__":
    main()
