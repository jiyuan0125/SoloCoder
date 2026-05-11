import os
import sys
import json
from datetime import date, datetime

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..')))

import click

from .api_client import get_client


def _print_json(data):
    click.echo(json.dumps(data, ensure_ascii=False, indent=2))


def _parse_date(s: str):
    if not s:
        return None
    return date.fromisoformat(s)


@click.group()
@click.option('--host', envvar='API_BASE_URL', default='http://localhost:8000', help='服务端地址')
@click.pass_context
def cli(ctx, host):
    ctx.obj = {'client': get_client(base_url=host)}


@cli.group()
def storehouse():
    pass


@storehouse.command('list')
@click.pass_obj
def storehouse_list(ctx):
    client = ctx['client']
    result = client.get('/storehouses')
    _print_json(result)


@storehouse.command('create')
@click.option('--name', required=True, help='仓库名称')
@click.option('--location', required=True, help='仓库位置')
@click.pass_obj
def storehouse_create(ctx, name, location):
    client = ctx['client']
    result = client.post('/storehouses', json={'name': name, 'location': location})
    _print_json(result)


@storehouse.command('get')
@click.argument('storehouse_id')
@click.pass_obj
def storehouse_get(ctx, storehouse_id):
    client = ctx['client']
    result = client.get(f'/storehouses/{storehouse_id}')
    _print_json(result)


@cli.group()
def area():
    pass


@area.command('list')
@click.option('--storehouse-id', default=None, help='仓库ID')
@click.pass_obj
def area_list(ctx, storehouse_id):
    client = ctx['client']
    params = {}
    if storehouse_id:
        params['storehouse_id'] = storehouse_id
    result = client.get('/areas', params=params)
    _print_json(result)


@area.command('create')
@click.option('--storehouse-id', required=True, help='仓库ID')
@click.option('--name', required=True, help='库区名称')
@click.option('--area-type', required=True, type=click.Choice(['grain', 'oil', 'mixed']), help='库区类型')
@click.option('--capacity', required=True, type=float, help='容量上限(kg)')
@click.option('--min-temp', default=10.0, type=float, help='最低温度')
@click.option('--max-temp', default=25.0, type=float, help='最高温度')
@click.pass_obj
def area_create(ctx, storehouse_id, name, area_type, capacity, min_temp, max_temp):
    client = ctx['client']
    result = client.post('/areas', json={
        'storehouse_id': storehouse_id,
        'name': name,
        'area_type': area_type,
        'capacity_kg': capacity,
        'min_temp': min_temp,
        'max_temp': max_temp,
    })
    _print_json(result)


@area.command('get')
@click.argument('area_id')
@click.pass_obj
def area_get(ctx, area_id):
    client = ctx['client']
    result = client.get(f'/areas/{area_id}')
    _print_json(result)


@cli.group()
def stock():
    pass


@stock.group()
def in_():
    pass


@in_.command('list')
@click.option('--area-id', default=None, help='库区ID')
@click.pass_obj
def stock_in_list(ctx, area_id):
    client = ctx['client']
    params = {}
    if area_id:
        params['area_id'] = area_id
    result = client.get('/stock-in', params=params)
    _print_json(result)


@in_.command('create')
@click.option('--area-id', required=True, help='库区ID')
@click.option('--variety', required=True, type=click.Choice([
    'wheat', 'rice', 'corn', 'soybean', 'rapeseed_oil', 'soybean_oil', 'other'
]), help='品种')
@click.option('--quantity', required=True, type=float, help='数量(kg)')
@click.option('--source', required=True, help='来源')
@click.option('--in-date', default=None, help='入库日期 (YYYY-MM-DD)')
@click.pass_obj
def stock_in_create(ctx, area_id, variety, quantity, source, in_date):
    client = ctx['client']
    payload = {
        'area_id': area_id,
        'variety': variety,
        'quantity_kg': quantity,
        'source': source,
    }
    if in_date:
        payload['in_date'] = in_date
    result = client.post('/stock-in', json=payload)
    _print_json(result)


@in_.command('fifo')
@click.option('--area-id', required=True, help='库区ID')
@click.option('--variety', required=True, help='品种')
@click.option('--quantity', required=True, type=float, help='出库数量(kg)')
@click.pass_obj
def stock_in_fifo(ctx, area_id, variety, quantity):
    client = ctx['client']
    result = client.get('/stock-in/fifo-recommendation', params={
        'area_id': area_id,
        'variety': variety,
        'quantity_kg': quantity,
    })
    _print_json(result)


@in_.command('expiring')
@click.option('--days', default=30, type=int, help='未来多少天内到期')
@click.pass_obj
def stock_in_expiring(ctx, days):
    client = ctx['client']
    result = client.get('/stock-in/expiring-soon', params={'days': days})
    _print_json(result)


@stock.group()
def out():
    pass


@out.command('list')
@click.option('--area-id', default=None, help='库区ID')
@click.pass_obj
def stock_out_list(ctx, area_id):
    client = ctx['client']
    params = {}
    if area_id:
        params['area_id'] = area_id
    result = client.get('/stock-out', params=params)
    _print_json(result)


@out.command('create')
@click.option('--area-id', required=True, help='库区ID')
@click.option('--variety', required=True, type=click.Choice([
    'wheat', 'rice', 'corn', 'soybean', 'rapeseed_oil', 'soybean_oil', 'other'
]), help='品种')
@click.option('--quantity', required=True, type=float, help='数量(kg)')
@click.option('--destination', required=True, help='去向')
@click.option('--purpose', required=True, help='用途')
@click.option('--out-date', default=None, help='出库日期 (YYYY-MM-DD)')
@click.pass_obj
def stock_out_create(ctx, area_id, variety, quantity, destination, purpose, out_date):
    client = ctx['client']
    payload = {
        'area_id': area_id,
        'variety': variety,
        'quantity_kg': quantity,
        'destination': destination,
        'purpose': purpose,
    }
    if out_date:
        payload['out_date'] = out_date
    result = client.post('/stock-out', json=payload)
    _print_json(result)


@cli.group()
def temperature():
    pass


@temperature.command('list')
@click.option('--area-id', default=None, help='库区ID')
@click.pass_obj
def temperature_list(ctx, area_id):
    client = ctx['client']
    params = {}
    if area_id:
        params['area_id'] = area_id
    result = client.get('/temperatures', params=params)
    _print_json(result)


@temperature.command('report')
@click.option('--area-id', required=True, help='库区ID')
@click.option('--value', required=True, type=float, help='温度值')
@click.option('--time', default=None, help='上报时间 (ISO格式)')
@click.pass_obj
def temperature_report(ctx, area_id, value, time):
    client = ctx['client']
    payload = {
        'area_id': area_id,
        'temperature': value,
    }
    if time:
        payload['record_time'] = time
    result = client.post('/temperatures', json=payload)
    _print_json(result)


@temperature.command('alarms')
@click.pass_obj
def temperature_alarms(ctx):
    client = ctx['client']
    result = client.get('/temperatures/alarms')
    _print_json(result)


@cli.group()
def pest():
    pass


@pest.command('list')
@click.option('--area-id', default=None, help='库区ID')
@click.pass_obj
def pest_list(ctx, area_id):
    client = ctx['client']
    params = {}
    if area_id:
        params['area_id'] = area_id
    result = client.get('/pest-inspections', params=params)
    _print_json(result)


@pest.command('record')
@click.option('--area-id', required=True, help='库区ID')
@click.option('--count', required=True, type=float, help='每公斤虫数')
@click.option('--variety', default=None, help='品种')
@click.option('--date', default=None, help='检查日期')
@click.option('--notes', default=None, help='备注')
@click.pass_obj
def pest_record(ctx, area_id, count, variety, date, notes):
    client = ctx['client']
    payload = {
        'area_id': area_id,
        'pest_count_per_kg': count,
    }
    if variety:
        payload['variety'] = variety
    if date:
        payload['inspection_date'] = date
    if notes:
        payload['notes'] = notes
    result = client.post('/pest-inspections', json=payload)
    _print_json(result)


@cli.group()
def todo():
    pass


@todo.command('list')
@click.option('--area-id', default=None, help='库区ID')
@click.option('--pending', is_flag=True, help='仅显示未完成')
@click.pass_obj
def todo_list(ctx, area_id, pending):
    client = ctx['client']
    params = {}
    if area_id:
        params['area_id'] = area_id
    if pending:
        params['pending_only'] = True
    result = client.get('/todos', params=params)
    _print_json(result)


@todo.command('complete')
@click.argument('todo_id')
@click.pass_obj
def todo_complete(ctx, todo_id):
    client = ctx['client']
    result = client.post(f'/todos/{todo_id}/complete')
    _print_json(result)


@cli.command()
@click.pass_obj
def metrics(ctx):
    client = ctx['client']
    result = client.get('/metrics')
    _print_json(result)


if __name__ == '__main__':
    cli()
