import os
import json
import click
import httpx


BASE_URL = os.getenv("API_BASE_URL", "http://localhost:8000")


def get_client():
    return httpx.Client(base_url=BASE_URL, timeout=10.0)


def print_response(response):
    try:
        data = response.json()
        click.echo(json.dumps(data, indent=2, ensure_ascii=False, default=str))
    except Exception:
        click.echo(response.text)


@click.group()
def cli():
    """铁路货运调度系统命令行客户端"""
    pass


@cli.group()
def station():
    """车站管理"""
    pass


@station.command("create")
@click.option("--code", required=True, help="车站编码")
@click.option("--name", required=True, help="车站名称")
@click.option("--traction", required=True, type=float, help="线路牵引定数(吨)")
def station_create(code, name, traction):
    """创建车站"""
    with get_client() as client:
        response = client.post("/stations", json={"code": code, "name": name, "line_traction_tons": traction})
        print_response(response)


@station.command("list")
def station_list():
    """列出所有车站"""
    with get_client() as client:
        response = client.get("/stations")
        print_response(response)


@cli.group()
def wagon():
    """车皮管理"""
    pass


@wagon.command("create")
@click.option("--number", required=True, help="车皮编号")
@click.option("--type", "car_type", required=True, type=click.Choice(["敞车", "棚车", "罐车", "平车", "冷藏车"]), help="车种类型")
@click.option("--capacity", required=True, type=float, help="载重(吨)")
@click.option("--station", required=True, help="所在车站编码")
def wagon_create(number, car_type, capacity, station):
    """创建车皮"""
    with get_client() as client:
        response = client.post("/wagons", json={
            "wagon_number": number,
            "car_type": car_type,
            "capacity_tons": capacity,
            "current_station": station
        })
        print_response(response)


@wagon.command("list")
@click.option("--station", help="按车站筛选")
def wagon_list(station):
    """列出车皮"""
    params = {}
    if station:
        params["station"] = station
    with get_client() as client:
        response = client.get("/wagons", params=params)
        print_response(response)


@wagon.command("available")
@click.option("--station", required=True, help="车站编码")
@click.option("--type", "car_type", required=True, type=click.Choice(["敞车", "棚车", "罐车", "平车", "冷藏车"]), help="车种类型")
def wagon_available(station, car_type):
    """列出可用空车"""
    with get_client() as client:
        response = client.get("/wagons/available", params={"station": station, "car_type": car_type})
        print_response(response)


@cli.group()
def order():
    """订单管理"""
    pass


@order.command("record")
@click.option("--no", "order_no", required=True, help="订单号")
@click.option("--cargo-type", required=True, type=click.Choice(["散装货物", "箱装货物", "液体货物", "集装箱", "易腐货物"]), help="货物类型")
@click.option("--cargo-name", required=True, help="货物名称")
@click.option("--weight", required=True, type=float, help="重量(吨)")
@click.option("--origin", required=True, help="发站编码")
@click.option("--dest", required=True, help="到站编码")
def order_record(order_no, cargo_type, cargo_name, weight, origin, dest):
    """录单"""
    with get_client() as client:
        response = client.post(f"/orders/{order_no}/record", json={
            "cargo_type": cargo_type,
            "cargo_name": cargo_name,
            "weight_tons": weight,
            "origin_station": origin,
            "dest_station": dest
        })
        print_response(response)


@order.command("get")
@click.option("--no", "order_no", required=True, help="订单号")
def order_get(order_no):
    """查看订单详情"""
    with get_client() as client:
        response = client.get(f"/orders/{order_no}")
        print_response(response)


@order.command("list")
@click.option("--status", help="按状态筛选（CREATED/ACCEPTED/ASSIGNED/LOADED/DEPARTED/ARRIVED/DELIVERED/CANCELLED）")
def order_list(status):
    """列出订单"""
    params = {}
    if status:
        params["status"] = status
    with get_client() as client:
        response = client.get("/orders", params=params)
        print_response(response)


@order.command("accept")
@click.option("--no", "order_no", required=True, help="订单号")
def order_accept(order_no):
    """受理订单"""
    with get_client() as client:
        response = client.post(f"/orders/{order_no}/accept")
        print_response(response)


@order.command("assign")
@click.option("--no", "order_no", required=True, help="订单号")
@click.option("--wagon-id", required=True, type=int, help="车皮ID")
def order_assign(order_no, wagon_id):
    """配车"""
    with get_client() as client:
        response = client.post(f"/orders/{order_no}/assign", json={"wagon_id": wagon_id})
        print_response(response)


@order.command("load")
@click.option("--no", "order_no", required=True, help="订单号")
@click.option("--train-id", required=True, type=int, help="列车ID")
def order_load(order_no, train_id):
    """装车（编入列车）"""
    with get_client() as client:
        response = client.post(f"/orders/{order_no}/load", json={"train_id": train_id})
        print_response(response)


@order.command("depart")
@click.option("--no", "order_no", required=True, help="订单号")
def order_depart(order_no):
    """发运"""
    with get_client() as client:
        response = client.post(f"/orders/{order_no}/depart")
        print_response(response)


@order.command("arrive")
@click.option("--no", "order_no", required=True, help="订单号")
def order_arrive(order_no):
    """到达"""
    with get_client() as client:
        response = client.post(f"/orders/{order_no}/arrive")
        print_response(response)


@order.command("deliver")
@click.option("--no", "order_no", required=True, help="订单号")
def order_deliver(order_no):
    """交付"""
    with get_client() as client:
        response = client.post(f"/orders/{order_no}/deliver")
        print_response(response)


@order.command("cancel")
@click.option("--no", "order_no", required=True, help="订单号")
def order_cancel(order_no):
    """取消订单"""
    with get_client() as client:
        response = client.post(f"/orders/{order_no}/cancel")
        print_response(response)


@cli.group()
def train():
    """列车管理"""
    pass


@train.command("create")
@click.option("--no", "train_no", required=True, help="车次")
@click.option("--origin", required=True, help="发站编码")
@click.option("--dest", required=True, help="到站编码")
@click.option("--traction", required=True, type=float, help="线路牵引定数(吨)")
@click.option("--max-wagons", default=50, type=int, help="最大车皮数")
def train_create(train_no, origin, dest, traction, max_wagons):
    """创建列车编组"""
    with get_client() as client:
        response = client.post("/trains", json={
            "train_no": train_no,
            "origin_station": origin,
            "dest_station": dest,
            "line_traction_tons": traction,
            "max_wagons": max_wagons
        })
        print_response(response)


@train.command("get")
@click.option("--no", "train_no", required=True, help="车次")
def train_get(train_no):
    """查看列车详情"""
    with get_client() as client:
        response = client.get(f"/trains/{train_no}")
        print_response(response)


@train.command("list")
def train_list():
    """列出所有列车"""
    with get_client() as client:
        response = client.get("/trains")
        print_response(response)


@train.command("arrival")
@click.option("--no", "train_no", required=True, help="车次")
def train_arrival(train_no):
    """记录列车到达"""
    with get_client() as client:
        response = client.post(f"/trains/{train_no}/arrival")
        print_response(response)


@train.command("departure")
@click.option("--no", "train_no", required=True, help="车次")
def train_departure(train_no):
    """记录列车出发"""
    with get_client() as client:
        response = client.post(f"/trains/{train_no}/departure")
        print_response(response)


@cli.group()
def stats():
    """统计查询"""
    pass


@stats.command("summary")
@click.option("--station", "station_code", required=True, help="车站编码")
def stats_summary(station_code):
    """车站统计摘要"""
    with get_client() as client:
        response = client.get(f"/stats/stations/{station_code}/summary")
        print_response(response)


if __name__ == "__main__":
    cli()
