import typer
import json
import urllib.request
import urllib.parse
import urllib.error
from datetime import date, time, datetime
from typing import Optional
from rich.console import Console
from rich.table import Table
from rich.prompt import Prompt
from rich.panel import Panel


console = Console()
app = typer.Typer(help="图书馆管理系统命令行客户端")

BASE_URL = "http://localhost:8000/api/v1"


def http_get(path: str, params: dict = None):
    url = BASE_URL + path
    if params:
        url += "?" + urllib.parse.urlencode(params)
    try:
        with urllib.request.urlopen(url, timeout=10) as response:
            return json.loads(response.read().decode())
    except urllib.error.HTTPError as e:
        error_body = e.read().decode()
        try:
            error_data = json.loads(error_body)
            console.print(f"[red]错误: {error_data.get('detail', str(e))}[/red]")
        except:
            console.print(f"[red]HTTP错误 {e.code}: {error_body}[/red]")
        return None
    except urllib.error.URLError as e:
        console.print(f"[red]连接错误: {e.reason}[/red]")
        console.print("[yellow]请确保后端服务已启动: uvicorn app.main:app --reload[/yellow]")
        return None


def http_post(path: str, data: dict = None):
    url = BASE_URL + path
    req = urllib.request.Request(
        url,
        data=json.dumps(data).encode() if data else None,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            return json.loads(response.read().decode())
    except urllib.error.HTTPError as e:
        error_body = e.read().decode()
        try:
            error_data = json.loads(error_body)
            console.print(f"[red]错误: {error_data.get('detail', str(e))}[/red]")
        except:
            console.print(f"[red]HTTP错误 {e.code}: {error_body}[/red]")
        return None
    except urllib.error.URLError as e:
        console.print(f"[red]连接错误: {e.reason}[/red]")
        console.print("[yellow]请确保后端服务已启动: uvicorn app.main:app --reload[/yellow]")
        return None


def http_put(path: str, data: dict = None):
    url = BASE_URL + path
    req = urllib.request.Request(
        url,
        data=json.dumps(data).encode() if data else None,
        headers={"Content-Type": "application/json"},
        method="PUT",
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            return json.loads(response.read().decode())
    except urllib.error.HTTPError as e:
        error_body = e.read().decode()
        try:
            error_data = json.loads(error_body)
            console.print(f"[red]错误: {error_data.get('detail', str(e))}[/red]")
        except:
            console.print(f"[red]HTTP错误 {e.code}: {error_body}[/red]")
        return None
    except urllib.error.URLError as e:
        console.print(f"[red]连接错误: {e.reason}[/red]")
        console.print("[yellow]请确保后端服务已启动: uvicorn app.main:app --reload[/yellow]")
        return None


@app.command()
def readers(
    search: Optional[str] = typer.Option(None, help="搜索关键词"),
    limit: int = typer.Option(50, help="显示数量"),
):
    """查看读者列表"""
    params = {"limit": limit}
    if search:
        params["search"] = search
    
    result = http_get("/readers", params)
    if not result:
        return
    
    if not result:
        console.print("[yellow]没有找到读者[/yellow]")
        return
    
    table = Table(title="读者列表")
    table.add_column("ID", style="cyan")
    table.add_column("证号", style="green")
    table.add_column("姓名")
    table.add_column("类型", style="magenta")
    table.add_column("电话")
    table.add_column("状态")
    
    reader_type_map = {
        "normal": "普通",
        "student": "学生",
        "teacher": "教师",
    }
    
    for reader in result:
        status = "[green]活跃[/green]" if reader["is_active"] else "[red]停用[/red]"
        table.add_row(
            str(reader["id"]),
            reader["card_number"],
            reader["name"],
            reader_type_map.get(reader["reader_type"], reader["reader_type"]),
            reader.get("phone", "-"),
            status,
        )
    
    console.print(table)


@app.command()
def reader_borrowings(
    reader_id: int = typer.Argument(..., help="读者ID"),
    active: bool = typer.Option(False, help="仅显示未归还的借阅"),
):
    """查看读者的借阅记录"""
    if active:
        result = http_get(f"/readers/{reader_id}/active-borrowings")
    else:
        result = http_get(f"/readers/{reader_id}/borrowings")
    
    if result is None:
        return
    
    if not result:
        console.print("[yellow]没有借阅记录[/yellow]")
        return
    
    title = f"读者 {reader_id} 的活跃借阅" if active else f"读者 {reader_id} 的借阅记录"
    table = Table(title=title)
    table.add_column("借阅ID", style="cyan")
    table.add_column("副本ID")
    table.add_column("借阅日期")
    table.add_column("应还日期")
    table.add_column("归还日期")
    table.add_column("续借次数")
    table.add_column("滞纳金", style="yellow")
    table.add_column("状态")
    
    for borrowing in result:
        is_returned = borrowing["is_returned"]
        status = "[green]已归还[/green]" if is_returned else "[red]借阅中[/red]"
        due_date = borrowing["due_date"]
        if not is_returned:
            today = date.today().isoformat()
            if due_date < today:
                status = "[bold red]逾期[/bold red]"
        
        table.add_row(
            str(borrowing["id"]),
            str(borrowing["copy_id"]),
            borrowing["borrow_date"],
            due_date,
            borrowing.get("return_date", "-"),
            str(borrowing["renew_count"]),
            f"{borrowing['late_fee']}元",
            status,
        )
    
    console.print(table)


@app.command()
def seats(
    floor: Optional[str] = typer.Option(None, help="楼层筛选"),
    active: bool = typer.Option(True, help="仅显示可用座位"),
):
    """查看座位列表"""
    params = {}
    if floor:
        params["floor"] = floor
    if active:
        params["is_active"] = "true"
    
    result = http_get("/seats", params)
    if result is None:
        return
    
    if not result:
        console.print("[yellow]没有找到座位[/yellow]")
        return
    
    table = Table(title="座位列表")
    table.add_column("ID", style="cyan")
    table.add_column("座位号", style="green")
    table.add_column("楼层")
    table.add_column("区域")
    table.add_column("电源", style="yellow")
    table.add_column("状态")
    
    for seat in result:
        power = "[green]有[/green]" if seat["has_power"] else "[dim]无[/dim]"
        status = "[green]可用[/green]" if seat["is_active"] else "[red]停用[/red]"
        table.add_row(
            str(seat["id"]),
            seat["seat_number"],
            seat.get("floor", "-"),
            seat.get("area", "-"),
            power,
            status,
        )
    
    console.print(table)


@app.command()
def seat_availability(
    seat_id: int = typer.Argument(..., help="座位ID"),
    check_date: str = typer.Option(None, help="查询日期 (YYYY-MM-DD)，默认今天"),
):
    """查看座位可用时段"""
    if not check_date:
        check_date = date.today().isoformat()
    
    result = http_get(f"/seats/{seat_id}/availability", {"check_date": check_date})
    if result is None:
        return
    
    table = Table(title=f"座位 {seat_id} 在 {result['date']} 的可用时段")
    table.add_column("时段")
    table.add_column("状态")
    
    for slot in result["availability"]:
        time_range = f"{slot['start']}-{slot['end']}"
        status = "[green]可用[/green]" if slot["available"] else "[red]已预约[/red]"
        table.add_row(time_range, status)
    
    console.print(table)


@app.command()
def seat_reservations(
    seat_id: Optional[int] = typer.Option(None, help="座位ID"),
    reader_id: Optional[int] = typer.Option(None, help="读者ID"),
    reservation_date: Optional[str] = typer.Option(None, help="预约日期 (YYYY-MM-DD)"),
):
    """查看座位预约记录"""
    params = {}
    if seat_id:
        params["seat_id"] = seat_id
    if reader_id:
        params["reader_id"] = reader_id
    if reservation_date:
        params["reservation_date"] = reservation_date
    
    result = http_get("/seats/reservations/", params)
    if result is None:
        return
    
    if not result:
        console.print("[yellow]没有预约记录[/yellow]")
        return
    
    table = Table(title="座位预约记录")
    table.add_column("ID", style="cyan")
    table.add_column("读者ID")
    table.add_column("座位ID")
    table.add_column("日期")
    table.add_column("时段")
    table.add_column("状态")
    
    status_map = {
        "reserved": "[yellow]已预约[/yellow]",
        "checked_in": "[green]已签到[/green]",
        "cancelled": "[dim]已取消[/dim]",
        "expired": "[red]已过期[/red]",
    }
    
    for res in result:
        status = status_map.get(res["status"], res["status"])
        table.add_row(
            str(res["id"]),
            str(res["reader_id"]),
            str(res["seat_id"]),
            res["reservation_date"],
            f"{res['start_time']}-{res['end_time']}",
            status,
        )
    
    console.print(table)


@app.command()
def time_slots():
    """查看可用的时段"""
    result = http_get("/seats/time-slots")
    if result:
        console.print(Panel.fit(
            "\n".join([f"  [cyan]{s['start']}[/cyan] - [cyan]{s['end']}[/cyan]" for s in result["slots"]]),
            title=f"可用时段 ({result['note']})",
        ))


@app.command()
def check_server():
    """检查后端服务是否运行"""
    try:
        result = http_get("")
        if result:
            console.print(f"[green]✓ 服务运行正常: {result.get('name', '')}[/green]")
            console.print(f"  API版本: {result.get('api_version', '')}")
            console.print(f"  文档地址: /docs")
    except:
        console.print("[red]✗ 无法连接到后端服务[/red]")
        console.print("[yellow]请运行: uvicorn app.main:app --reload[/yellow]")


def show_borrowing_rules():
    """显示借阅规则"""
    rules = [
        {"类型": "普通读者", "最大数量": "5本", "借期": "30天"},
        {"类型": "学生", "最大数量": "10本", "借期": "60天"},
        {"类型": "教师", "最大数量": "15本", "借期": "90天"},
    ]
    
    table = Table(title="借阅规则")
    table.add_column("类型")
    table.add_column("最大数量", style="cyan")
    table.add_column("借期", style="green")
    for r in rules:
        table.add_row(r["类型"], r["最大数量"], r["借期"])
    
    console.print(table)
    console.print("\n[yellow]续借规则:[/yellow]")
    console.print("  • 续借一次期限为原借期一半")
    console.print("  • 到期前才能续借，逾期不能续借")
    console.print("\n[yellow]滞纳金规则:[/yellow]")
    console.print("  • 超期每天每本 0.10 元")
    console.print("  • 累计超书价 50% 按 50% 封顶")
    console.print("  • 还书当天不计逾期天数")
    console.print("\n[yellow]座位预约规则:[/yellow]")
    console.print("  • 2小时一档，8:00-21:00")
    console.print("  • 15分钟内未签到自动取消")
    console.print("  • 同一读者不能同时预约两个时段")


@app.command()
def rules():
    """显示借阅和预约规则"""
    show_borrowing_rules()


if __name__ == "__main__":
    app()
