import click
import httpx
import json
from tabulate import tabulate
from datetime import datetime

BASE_URL = "http://localhost:8000"


class APIClient:
    def __init__(self, base_url: str = BASE_URL):
        self.base_url = base_url
        self.client = httpx.Client(timeout=30.0)
    
    def get(self, endpoint: str):
        return self.client.get(f"{self.base_url}{endpoint}")
    
    def post(self, endpoint: str, data: dict):
        return self.client.post(f"{self.base_url}{endpoint}", json=data)
    
    def put(self, endpoint: str, data: dict):
        return self.client.put(f"{self.base_url}{endpoint}", json=data)
    
    def delete(self, endpoint: str):
        return self.client.delete(f"{self.base_url}{endpoint}")


def handle_response(response, expected_status: int = 200):
    if response.status_code != expected_status:
        click.echo(f"错误: {response.status_code}")
        try:
            click.echo(json.dumps(response.json(), indent=2, ensure_ascii=False))
        except:
            click.echo(response.text)
        return None
    return response.json()


@click.group()
@click.option("--url", default=BASE_URL, help="API服务地址")
@click.pass_context
def cli(ctx, url):
    ctx.ensure_object(dict)
    ctx.obj['client'] = APIClient(url)


@cli.group()
def waypoint():
    pass


@waypoint.command("list")
@click.pass_context
def list_waypoints(ctx):
    client = ctx.obj['client']
    response = client.get("/waypoints/")
    data = handle_response(response)
    if data:
        headers = ["ID", "代码", "名称", "纬度", "经度"]
        rows = [[w["id"], w["code"], w.get("name", "-"), w["latitude"], w["longitude"]] for w in data]
        click.echo(tabulate(rows, headers=headers))


@waypoint.command("create")
@click.option("--code", required=True)
@click.option("--name", default="")
@click.option("--lat", required=True, type=float)
@click.option("--lon", required=True, type=float)
@click.pass_context
def create_waypoint(ctx, code, name, lat, lon):
    client = ctx.obj['client']
    data = {"code": code, "name": name, "latitude": lat, "longitude": lon}
    response = client.post("/waypoints/", data)
    result = handle_response(response, 200)
    if result:
        click.echo(f"航路点创建成功: {result['code']} (ID: {result['id']})")


@cli.group()
def route():
    pass


@route.command("list")
@click.pass_context
def list_routes(ctx):
    client = ctx.obj['client']
    response = client.get("/routes/")
    data = handle_response(response)
    if data:
        headers = ["ID", "代码", "名称", "容量", "状态"]
        rows = [[r["id"], r["code"], r.get("name", "-"), r["capacity"], r["status"]] for r in data]
        click.echo(tabulate(rows, headers=headers))


@route.command("status")
@click.argument("route_id", type=int)
@click.pass_context
def route_status(ctx, route_id):
    client = ctx.obj['client']
    response = client.get(f"/routes/{route_id}/status")
    data = handle_response(response)
    if data:
        click.echo(f"\n航路状态: {data['route_code']}")
        click.echo(f"当前状态: {data['current_status']}")
        click.echo(f"活跃航班: {data['active_flights']}/{data['capacity']}")
        click.echo(f"等待航班: {data['waiting_flights']}")
        click.echo(f"利用率: {data['utilization']*100:.1f}%")
        click.echo(f"未解决冲突: {data['unresolved_conflicts']}")


@route.command("flow")
@click.argument("route_id", type=int)
@click.option("--limit", default=20, type=int)
@click.pass_context
def route_flow(ctx, route_id, limit):
    client = ctx.obj['client']
    response = client.get(f"/routes/{route_id}/flow?limit={limit}")
    data = handle_response(response)
    if data:
        headers = ["时间", "活跃", "等待", "容量", "利用率", "状态"]
        rows = [[
            f["timestamp"][:19],
            f["active_flights"],
            f["waiting_flights"],
            f["capacity"],
            f"{f['utilization']*100:.1f}%",
            f["status"]
        ] for f in data]
        click.echo(tabulate(rows, headers=headers))


@route.command("conflicts")
@click.argument("route_id", type=int)
@click.option("--resolved", is_flag=True, help="显示已解决的冲突")
@click.pass_context
def route_conflicts(ctx, route_id, resolved):
    client = ctx.obj['client']
    resolved_param = "true" if resolved else "false" if click.get_current_context().params.get('resolved') else ""
    endpoint = f"/routes/{route_id}/conflicts"
    if resolved_param:
        endpoint += f"?resolved={resolved_param}"
    response = client.get(endpoint)
    data = handle_response(response)
    if data:
        if not data:
            click.echo("无冲突记录")
            return
        headers = ["ID", "类型", "航班1", "航班2", "检测时间", "状态"]
        rows = [[
            c["id"],
            c["conflict_type"],
            c["flight1_id"],
            c["flight2_id"],
            c["detected_at"][:19],
            "已解决" if c["resolved"] else "未解决"
        ] for c in data]
        click.echo(tabulate(rows, headers=headers))


@route.command("reevaluate")
@click.argument("route_id", type=int)
@click.pass_context
def reevaluate_conflicts(ctx, route_id):
    client = ctx.obj['client']
    response = client.post(f"/routes/{route_id}/reevaluate", {})
    data = handle_response(response)
    if data:
        if data:
            click.echo(f"重新评估发现 {len(data)} 个冲突:")
            for c in data:
                click.echo(f"  - [{c['conflict_type']}] {c['description']}")
        else:
            click.echo("重新评估后无冲突")


@cli.group()
def flight():
    pass


@flight.command("list")
@click.pass_context
def list_flights(ctx):
    client = ctx.obj['client']
    response = client.get("/flights/")
    data = handle_response(response)
    if data:
        headers = ["ID", "航班号", "航路ID", "方向", "高度", "类型", "预计进入", "状态"]
        rows = []
        for f in data:
            status = "完成" if f["is_completed"] else ("等待" if f["is_waiting"] else "活跃")
            rows.append([
                f["id"],
                f["flight_number"],
                f["route_id"],
                f["direction"],
                f["altitude"],
                f["flight_type"],
                f["estimated_entry_time"][:19],
                status
            ])
        click.echo(tabulate(rows, headers=headers))


@flight.command("create")
@click.option("--number", required=True)
@click.option("--route-id", required=True, type=int)
@click.option("--direction", required=True, help="east/west 或 东/西")
@click.option("--altitude", required=True, type=float)
@click.option("--type", "flight_type", default="domestic", type=click.Choice(['domestic', 'international']))
@click.option("--entry-time", required=True, help="格式: YYYY-MM-DDTHH:MM:SS")
@click.pass_context
def create_flight(ctx, number, route_id, direction, altitude, flight_type, entry_time):
    client = ctx.obj['client']
    data = {
        "flight_number": number,
        "route_id": route_id,
        "direction": direction,
        "altitude": altitude,
        "flight_type": flight_type,
        "estimated_entry_time": entry_time
    }
    response = client.post("/flights/", data)
    result = handle_response(response, 200)
    if result:
        status = "等待队列" if result["is_waiting"] else "已进入航路"
        click.echo(f"航班创建成功: {result['flight_number']} - {status}")


@flight.command("complete")
@click.argument("flight_id", type=int)
@click.pass_context
def complete_flight(ctx, flight_id):
    client = ctx.obj['client']
    response = client.post(f"/flights/{flight_id}/complete", {})
    result = handle_response(response, 200)
    if result:
        click.echo(f"航班 {result['flight_number']} 已标记为完成")


@cli.group()
def simulation():
    pass


@simulation.command("run")
@click.option("--file", required=True, type=click.File('r'), help="JSON文件包含航班计划")
@click.pass_context
def run_simulation(ctx, file):
    client = ctx.obj['client']
    try:
        data = json.load(file)
        response = client.post("/simulation/run", data)
        result = handle_response(response, 200)
        if result:
            click.echo(f"\n=== 模拟推演报告 ===")
            click.echo(f"航班数量: {result['total_flights']}")
            click.echo(f"冲突数量: {len(result['conflicts'])}")
            
            if result['conflicts']:
                click.echo(f"\n--- 冲突详情 ---")
                for c in result['conflicts']:
                    click.echo(f"\n[{c['conflict_type'].upper()}] {c['description']}")
                    click.echo(f"建议: {c['resolution_suggestion']}")
            
            click.echo(f"\n--- 建议 ---")
            for s in result['suggestions']:
                click.echo(f"  * {s}")
    except json.JSONDecodeError as e:
        click.echo(f"JSON解析错误: {e}")


@cli.command("demo")
@click.pass_context
def demo(ctx):
    click.echo("=== 空管调度辅助系统演示 ===\n")
    
    client = ctx.obj['client']
    
    click.echo("1. 创建航路点...")
    wps = [
        {"code": "PEK", "name": "北京首都", "latitude": 40.0799, "longitude": 116.5891},
        {"code": "PVG", "name": "上海浦东", "latitude": 31.1434, "longitude": 121.8058},
        {"code": "CAN", "name": "广州白云", "latitude": 23.3924, "longitude": 113.2988},
    ]
    for wp in wps:
        client.post("/waypoints/", wp)
    
    click.echo("2. 创建航路...")
    route_data = {
        "code": "BJSHA",
        "name": "北京-上海航路",
        "capacity": 5,
        "busy_threshold": 0.8,
        "saturated_threshold": 1.0,
        "min_interval_minutes": 10,
        "average_speed": 800.0,
        "waypoints": [
            {"waypoint_id": 1, "sequence": 1, "estimated_time_minutes": 0},
            {"waypoint_id": 2, "sequence": 2, "estimated_time_minutes": 120}
        ]
    }
    client.post("/routes/", route_data)
    
    click.echo("3. 添加航班...")
    now = datetime.utcnow()
    flights = [
        {
            "flight_number": "CA1001",
            "route_id": 1,
            "direction": "east",
            "altitude": 9000.0,
            "flight_type": "international",
            "estimated_entry_time": (now).isoformat()
        },
        {
            "flight_number": "MU2002",
            "route_id": 1,
            "direction": "east",
            "altitude": 9000.0,
            "flight_type": "domestic",
            "estimated_entry_time": (now).isoformat()
        },
        {
            "flight_number": "CZ3003",
            "route_id": 1,
            "direction": "west",
            "altitude": 9600.0,
            "flight_type": "domestic",
            "estimated_entry_time": (now).isoformat()
        },
    ]
    for f in flights:
        client.post("/flights/", f)
    
    click.echo("\n4. 查询航路状态...")
    response = client.get("/routes/1/status")
    data = response.json()
    click.echo(f"   航路: {data['route_code']}")
    click.echo(f"   状态: {data['current_status']}")
    click.echo(f"   活跃航班: {data['active_flights']}/{data['capacity']}")
    click.echo(f"   利用率: {data['utilization']*100:.1f}%")
    click.echo(f"   冲突数: {data['unresolved_conflicts']}")
    
    click.echo("\n5. 查看冲突...")
    response = client.get("/routes/1/conflicts?resolved=false")
    data = response.json()
    for c in data:
        click.echo(f"   [{c['conflict_type']}] {c['description']}")
    
    click.echo("\n=== 演示完成 ===")


if __name__ == "__main__":
    cli()
