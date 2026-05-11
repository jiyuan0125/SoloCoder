from __future__ import annotations

import json
from datetime import date, datetime
from typing import Optional

import click

from .api_client import APIClient


@click.group()
@click.pass_context
def cli(ctx: click.Context) -> None:
    ctx.obj = APIClient()


@cli.group('well')
def well_group() -> None:
    pass


@well_group.command('create')
@click.option('--name', required=True, help='井名称')
@click.option('--block', required=True, help='区块')
@click.option('--planned-days', type=int, required=True, help='计划总天数')
@click.pass_obj
def create_well(client: APIClient, name: str, block: str, planned_days: int) -> None:
    result = client.create_well(name, block, planned_days)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@well_group.command('list')
@click.pass_obj
def list_wells(client: APIClient) -> None:
    result = client.list_wells()
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@well_group.command('get')
@click.option('--id', 'well_id', required=True, help='井ID')
@click.pass_obj
def get_well(client: APIClient, well_id: str) -> None:
    result = client.get_well(well_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@well_group.command('phases')
@click.option('--id', 'well_id', required=True, help='井ID')
@click.pass_obj
def list_phases(client: APIClient, well_id: str) -> None:
    result = client.list_phases(well_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@well_group.command('start-phase')
@click.option('--well-id', required=True, help='井ID')
@click.option('--phase', required=True, help='阶段名称 (PHASE_1 到 PHASE_8)')
@click.option('--start-date', required=True, help='开始日期 (YYYY-MM-DD)')
@click.option('--planned-days', type=int, required=True, help='计划天数')
@click.pass_obj
def start_phase(
    client: APIClient,
    well_id: str,
    phase: str,
    start_date: str,
    planned_days: int,
) -> None:
    result = client.start_phase(well_id, phase, start_date, planned_days)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@cli.command('complete-phase')
@click.option('--phase-id', required=True, help='阶段ID')
@click.option('--end-date', required=True, help='结束日期 (YYYY-MM-DD)')
@click.pass_obj
def complete_phase(client: APIClient, phase_id: str, end_date: str) -> None:
    result = client.complete_phase(phase_id, end_date)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@cli.group('todo')
def todo_group() -> None:
    pass


@todo_group.command('list')
@click.option('--well-id', help='井ID (可选)')
@click.pass_obj
def list_todos(client: APIClient, well_id: Optional[str]) -> None:
    result = client.list_todos(well_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@todo_group.command('accept')
@click.option('--id', 'todo_id', required=True, help='验收待办ID')
@click.pass_obj
def accept_todo(client: APIClient, todo_id: str) -> None:
    result = client.accept_todo(todo_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@cli.group('accident')
def accident_group() -> None:
    pass


@accident_group.command('record')
@click.option('--well-id', required=True, help='井ID')
@click.option('--type', 'accident_type', required=True,
              type=click.Choice(['well_kick', 'well_loss', 'stuck_pipe', 'well_collapse']),
              help='事故类型')
@click.option('--start-time', required=True, help='开始时间 (YYYY-MM-DDTHH:MM:SS)')
@click.option('--end-time', help='结束时间 (YYYY-MM-DDTHH:MM:SS)')
@click.option('--has-loss', is_flag=True, help='是否有损失')
@click.option('--responsible', help='负责人')
@click.pass_obj
def record_accident(
    client: APIClient,
    well_id: str,
    accident_type: str,
    start_time: str,
    end_time: Optional[str],
    has_loss: bool,
    responsible: Optional[str],
) -> None:
    result = client.create_accident(
        well_id=well_id,
        accident_type=accident_type,
        start_time=start_time,
        end_time=end_time,
        has_loss=has_loss,
        responsible_person=responsible,
    )
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@accident_group.command('update')
@click.option('--id', 'accident_id', required=True, help='事故ID')
@click.option('--cause', help='原因分析')
@click.option('--measures', help='防范措施')
@click.option('--end-time', help='结束时间 (YYYY-MM-DDTHH:MM:SS)')
@click.option('--has-loss', type=bool, help='是否有损失')
@click.option('--responsible', help='负责人')
@click.pass_obj
def update_accident(
    client: APIClient,
    accident_id: str,
    cause: Optional[str],
    measures: Optional[str],
    end_time: Optional[str],
    has_loss: Optional[bool],
    responsible: Optional[str],
) -> None:
    result = client.update_accident(
        accident_id=accident_id,
        cause_analysis=cause,
        preventive_measures=measures,
        end_time=end_time,
        has_loss=has_loss,
        responsible_person=responsible,
    )
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@accident_group.command('list')
@click.option('--well-id', help='井ID (可选)')
@click.pass_obj
def list_accidents(client: APIClient, well_id: Optional[str]) -> None:
    result = client.list_accidents(well_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@accident_group.command('get')
@click.option('--id', 'accident_id', required=True, help='事故ID')
@click.pass_obj
def get_accident(client: APIClient, accident_id: str) -> None:
    result = client.get_accident(accident_id)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@cli.group('stats')
def stats_group() -> None:
    pass


@stats_group.command('blocks')
@click.pass_obj
def list_blocks(client: APIClient) -> None:
    result = client.list_blocks()
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


@stats_group.command('block')
@click.option('--name', required=True, help='区块名称')
@click.pass_obj
def get_block_stats(client: APIClient, name: str) -> None:
    result = client.get_block_stats(name)
    click.echo(json.dumps(result, indent=2, ensure_ascii=False))


def main() -> None:
    cli()


if __name__ == '__main__':
    main()
