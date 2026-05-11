import json
from datetime import date, datetime
from typing import List, Optional

import typer
from rich.console import Console
from rich.table import Table

from client.api_client import VetAPIClient

app = typer.Typer()
console = Console()
err_console = Console(stderr=True)


def print_table(title: str, columns: List[str], rows: List[List[str]]):
    table = Table(title=title, show_lines=True)
    for col in columns:
        table.add_column(col)
    for row in rows:
        table.add_row(*[str(item) if item is not None else "" for item in row])
    console.print(table)


@app.command()
def health():
    with VetAPIClient() as client:
        result = client.health_check()
        console.print(f"[green]Health: {result}[/green]")


@app.command()
def farmers():
    with VetAPIClient() as client:
        farmers = client.list_farmers()
        if not farmers:
            console.print("[yellow]No farmers found[/yellow]")
            return
        rows = [
            [str(f["id"]), f["name"], f["address"], f["livestock_type"], str(f["scale"]), str(f.get("last_visit_date") or "")]
            for f in farmers
        ]
        print_table("养殖户列表", ["ID", "姓名", "地址", "养殖类型", "规模", "上次巡诊日期"], rows)


@app.command()
def add_farmer(
    name: str,
    address: str,
    livestock_type: str,
    scale: int
):
    with VetAPIClient() as client:
        farmer = client.create_farmer(name, address, livestock_type, scale)
        console.print(f"[green]Created farmer: {json.dumps(farmer, ensure_ascii=False)}[/green]")


@app.command()
def medicines():
    with VetAPIClient() as client:
        medicines = client.list_medicines()
        if not medicines:
            console.print("[yellow]No medicines found[/yellow]")
            return
        rows = [
            [str(m["id"]), m["name"], m["specification"], str(m["stock"]), m["unit"]]
            for m in medicines
        ]
        print_table("药品列表", ["ID", "名称", "规格", "库存", "单位"], rows)


@app.command()
def add_medicine(
    name: str,
    specification: str,
    unit: str,
    stock: float
):
    with VetAPIClient() as client:
        medicine = client.create_medicine(name, specification, unit, stock)
        console.print(f"[green]Created medicine: {json.dumps(medicine, ensure_ascii=False)}[/green]")


@app.command()
def add_stock(
    medicine_id: int,
    amount: float
):
    with VetAPIClient() as client:
        medicine = client.add_medicine_stock(medicine_id, amount)
        console.print(f"[green]Updated medicine: {json.dumps(medicine, ensure_ascii=False)}[/green]")


@app.command()
def generate_route(route_date: Optional[str] = None):
    target_date = date.fromisoformat(route_date) if route_date else date.today()
    with VetAPIClient() as client:
        route = client.generate_route(target_date)
        items = route.get("items", [])
        if not items:
            console.print("[yellow]No farmers in route[/yellow]")
            return
        rows = [
            [str(i["order_index"]), str(i["farmer_id"]), i["farmer"]["name"] if i.get("farmer") else "",
             i["farmer"]["address"] if i.get("farmer") else "", "已完成" if i["is_completed"] else "待完成"]
            for i in items
        ]
        print_table(f"巡诊路线 - {route['route_date']}", ["顺序", "养殖户ID", "姓名", "地址", "状态"], rows)


@app.command()
def route(route_date: Optional[str] = None):
    target_date = date.fromisoformat(route_date) if route_date else date.today()
    with VetAPIClient() as client:
        try:
            route = client.get_route(target_date)
        except Exception:
            console.print("[yellow]Route not found, try generate-route first[/yellow]")
            return
        items = route.get("items", [])
        if not items:
            console.print("[yellow]No farmers in route[/yellow]")
            return
        rows = [
            [str(i["order_index"]), str(i["farmer_id"]), i["farmer"]["name"] if i.get("farmer") else "",
             i["farmer"]["address"] if i.get("farmer") else "", "已完成" if i["is_completed"] else "待完成", str(i["id"])]
            for i in items
        ]
        print_table(f"巡诊路线 - {route['route_date']}", ["顺序", "养殖户ID", "姓名", "地址", "状态", "Item ID"], rows)


@app.command()
def record_visit(
    farmer_id: int,
    visit_date: str,
    has_abnormality: bool = False,
    notes: Optional[str] = None
):
    v_date = date.fromisoformat(visit_date)
    with VetAPIClient() as client:
        visit = client.create_visit(farmer_id, v_date, has_abnormality, notes)
        console.print(f"[green]Created visit: {json.dumps(visit, ensure_ascii=False)}[/green]")


@app.command()
def todos(pending: bool = False, todo_date: Optional[str] = None):
    target_date = date.fromisoformat(todo_date) if todo_date else None
    with VetAPIClient() as client:
        todos_list = client.list_todos(todo_date=target_date, pending=pending)
        if not todos_list:
            console.print("[yellow]No todos found[/yellow]")
            return
        rows = []
        for t in todos_list:
            presc = t.get("prescription", {})
            med = presc.get("medicine", {})
            rows.append([
                str(t["id"]),
                t["todo_date"],
                str(presc.get("case_id") or ""),
                med.get("name", ""),
                str(presc.get("dosage_per_day") or ""),
                med.get("unit", ""),
                "已完成" if t["is_completed"] else "待完成"
            ])
        print_table("待办事项", ["ID", "日期", "病例ID", "药品", "日用量", "单位", "状态"], rows)


@app.command()
def complete_todo(todo_id: int, notes: Optional[str] = None):
    with VetAPIClient() as client:
        todo = client.update_todo(todo_id, is_completed=True, notes=notes)
        console.print(f"[green]Updated todo: {json.dumps(todo, ensure_ascii=False)}[/green]")


@app.command()
def dashboard(target_date: Optional[str] = None):
    t_date = date.fromisoformat(target_date) if target_date else None
    with VetAPIClient() as client:
        stats = client.get_dashboard_stats(t_date)
        console.print(f"\n[bold cyan]===== 指标看板 - {stats['month']} =====[/bold cyan]\n")
        console.print(f"巡诊次数: {stats['visit_count']}")
        console.print(f"覆盖户数: {stats['covered_farmers']}")
        console.print(f"新增病例: {stats['new_cases']}")
        if stats.get("consumption_ranking"):
            rows = [
                [str(i + 1), cr["medicine_name"], str(cr["total_dosage"]), cr["unit"]]
                for i, cr in enumerate(stats["consumption_ranking"])
            ]
            print_table("药品消耗排名", ["排名", "药品名称", "消耗量", "单位"], rows)


def main():
    app()


if __name__ == "__main__":
    main()
