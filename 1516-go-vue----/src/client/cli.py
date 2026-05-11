import typer
from typing import Optional, List
from datetime import date
from rich.console import Console
from rich.table import Table
from rich.prompt import Prompt, IntPrompt, Confirm
from src.client.api_client import APIClient

app = typer.Typer(help="宠物美容店预约管理系统客户端")
console = Console()
api = APIClient()


@app.command("list-projects")
def list_projects():
    projects = api.list_projects()
    table = Table(title="服务项目")
    table.add_column("ID", style="cyan")
    table.add_column("名称", style="green")
    table.add_column("时长(分钟)", style="yellow")
    table.add_column("价格(元)", style="magenta")
    
    for p in projects:
        table.add_row(str(p["id"]), p["name"], str(p["duration_minutes"]), f"{p['price']:.2f}")
    
    console.print(table)


@app.command("create-project")
def create_project():
    name = Prompt.ask("请输入项目名称")
    duration = IntPrompt.ask("请输入项目时长(分钟)", default=60)
    price = float(Prompt.ask("请输入项目价格(元)", default="100.0"))
    
    result = api.create_project(name, duration, price)
    console.print(f"[green]项目创建成功! ID: {result['id']}[/green]")


@app.command("delete-project")
def delete_project():
    list_projects()
    project_id = IntPrompt.ask("请输入要删除的项目ID")
    if Confirm.ask(f"确定要删除项目 {project_id} 吗?", default=False):
        api.delete_project(project_id)
        console.print("[green]项目已删除[/green]")


@app.command("list-stylists")
def list_stylists():
    stylists = api.list_stylists()
    table = Table(title="美容师列表")
    table.add_column("ID", style="cyan")
    table.add_column("姓名", style="green")
    table.add_column("状态", style="yellow")
    table.add_column("擅长项目", style="magenta")
    
    status_map = {
        "available": "可用",
        "busy": "忙碌",
        "off_duty": "休息"
    }
    
    for s in stylists:
        table.add_row(
            str(s["id"]),
            s["name"],
            status_map.get(s["status"], s["status"]),
            ",".join(map(str, s["skilled_projects"]))
        )
    
    console.print(table)


@app.command("create-stylist")
def create_stylist():
    name = Prompt.ask("请输入美容师姓名")
    status = Prompt.ask(
        "请输入状态",
        choices=["available", "busy", "off_duty"],
        default="available"
    )
    skilled_str = Prompt.ask("请输入擅长项目ID(用逗号分隔，可选)", default="")
    skilled_projects = [int(x.strip()) for x in skilled_str.split(",") if x.strip()]
    
    result = api.create_stylist(name, status, skilled_projects)
    console.print(f"[green]美容师创建成功! ID: {result['id']}[/green]")


@app.command("list-clients")
def list_clients():
    clients = api.list_clients()
    table = Table(title="客户列表")
    table.add_column("ID", style="cyan")
    table.add_column("姓名", style="green")
    table.add_column("电话", style="yellow")
    table.add_column("宠物", style="magenta")
    
    for c in clients:
        table.add_row(
            str(c["id"]),
            c["name"],
            c["phone"],
            ",".join(c["pets"])
        )
    
    console.print(table)


@app.command("create-client")
def create_client():
    name = Prompt.ask("请输入客户姓名")
    phone = Prompt.ask("请输入电话")
    pets_str = Prompt.ask("请输入宠物名称(用逗号分隔，可选)", default="")
    pets = [x.strip() for x in pets_str.split(",") if x.strip()]
    
    result = api.create_client(name, phone, pets)
    console.print(f"[green]客户创建成功! ID: {result['id']}[/green]")


@app.command("get-slots")
def get_slots():
    list_stylists()
    stylist_id = IntPrompt.ask("请选择美容师ID")
    list_projects()
    project_id = IntPrompt.ask("请选择项目ID")
    date_str = Prompt.ask("请输入预约日期(YYYY-MM-DD)", default=date.today().isoformat())
    
    appointment_date = date.fromisoformat(date_str)
    slots = api.get_available_slots(stylist_id, project_id, appointment_date)
    
    if slots:
        console.print(f"\n[green]可用时段({len(slots)}个):[/green]")
        for slot in slots:
            console.print(f"  - {slot}")
    else:
        console.print("[yellow]该美容师当天没有可用时段[/yellow]")


@app.command("create-appointment")
def create_appointment():
    list_clients()
    client_id = IntPrompt.ask("请选择客户ID")
    pet_name = Prompt.ask("请输入宠物名称")
    
    list_projects()
    project_id = IntPrompt.ask("请选择项目ID")
    
    list_stylists()
    stylist_id = IntPrompt.ask("请选择美容师ID")
    
    date_str = Prompt.ask("请输入预约日期(YYYY-MM-DD)", default=date.today().isoformat())
    appointment_date = date.fromisoformat(date_str)
    
    slots = api.get_available_slots(stylist_id, project_id, appointment_date)
    if slots:
        console.print(f"\n可用时段: {', '.join(slots)}")
        start_time = Prompt.ask("请选择时段", choices=slots)
    else:
        console.print("[yellow]该美容师当天没有可用时段[/yellow]")
        start_time = Prompt.ask("请输入希望的时段(HH:MM)，将进入候补")
    
    result = api.create_appointment(
        client_id, pet_name, project_id, stylist_id,
        appointment_date, start_time
    )
    
    status_map = {
        "confirmed": "已确认",
        "waitlist": "候补中",
        "pending": "待确认",
        "completed": "已完成",
        "cancelled": "已取消"
    }
    
    console.print(f"[green]预约创建成功! ID: {result['id']}[/green]")
    console.print(f"  时段: {result['start_time']} - {result['end_time']}")
    console.print(f"  状态: {status_map.get(result['status'], result['status'])}")


@app.command("list-appointments")
def list_appointments():
    appointments = api.list_appointments()
    table = Table(title="预约列表")
    table.add_column("ID", style="cyan")
    table.add_column("日期", style="green")
    table.add_column("时段", style="yellow")
    table.add_column("客户ID", style="magenta")
    table.add_column("宠物", style="blue")
    table.add_column("项目ID", style="red")
    table.add_column("美容师ID", style="cyan")
    table.add_column("状态", style="green")
    table.add_column("超时", style="red")
    
    status_map = {
        "confirmed": "已确认",
        "waitlist": "候补中",
        "pending": "待确认",
        "completed": "已完成",
        "cancelled": "已取消"
    }
    
    for a in appointments:
        table.add_row(
            str(a["id"]),
            a["appointment_date"],
            f"{a['start_time']}-{a['end_time']}",
            str(a["client_id"]),
            a["pet_name"],
            str(a["project_id"]),
            str(a["stylist_id"]),
            status_map.get(a["status"], a["status"]),
            "是" if a.get("is_overtime") else "否"
        )
    
    console.print(table)


@app.command("cancel-appointment")
def cancel_appointment():
    list_appointments()
    appointment_id = IntPrompt.ask("请输入要取消的预约ID")
    if Confirm.ask(f"确定要取消预约 {appointment_id} 吗?", default=False):
        result = api.cancel_appointment(appointment_id)
        console.print(f"[green]预约已取消[/green]")


@app.command("complete-appointment")
def complete_appointment():
    list_appointments()
    appointment_id = IntPrompt.ask("请输入要完成的预约ID")
    actual_duration = IntPrompt.ask("请输入实际用时(分钟)")
    
    result = api.complete_appointment(appointment_id, actual_duration)
    console.print(f"[green]预约已完成[/green]")
    if result.get("is_overtime"):
        console.print("[red]警告: 超时50%以上[/red]")


@app.command("create-review")
def create_review():
    list_appointments()
    appointment_id = IntPrompt.ask("请输入预约ID")
    client_id = IntPrompt.ask("请输入客户ID")
    rating = IntPrompt.ask("请输入评分(1-5)", default=5)
    comment = Prompt.ask("请输入评价内容(可选)", default="")
    
    result = api.create_review(appointment_id, client_id, rating, comment or None)
    console.print(f"[green]评价创建成功! ID: {result['id']}[/green]")


@app.command("list-reviews")
def list_reviews():
    reviews = api.list_reviews()
    table = Table(title="评价列表")
    table.add_column("ID", style="cyan")
    table.add_column("预约ID", style="green")
    table.add_column("客户ID", style="yellow")
    table.add_column("评分", style="magenta")
    table.add_column("评价", style="blue")
    
    for r in reviews:
        table.add_row(
            str(r["id"]),
            str(r["appointment_id"]),
            str(r["client_id"]),
            "★" * r["rating"],
            r.get("comment", "") or ""
        )
    
    console.print(table)


@app.command("list-supplies")
def list_supplies():
    supplies = api.list_supplies()
    table = Table(title="耗材库存")
    table.add_column("ID", style="cyan")
    table.add_column("名称", style="green")
    table.add_column("当前库存", style="yellow")
    table.add_column("最低库存", style="magenta")
    table.add_column("单位", style="blue")
    table.add_column("状态", style="red")
    
    for s in supplies:
        status = "[red]库存不足[/red]" if s["current_stock"] < s["min_stock"] else "[green]充足[/green]"
        table.add_row(
            str(s["id"]),
            s["name"],
            str(s["current_stock"]),
            str(s["min_stock"]),
            s["unit"],
            status
        )
    
    console.print(table)


@app.command("create-supply")
def create_supply():
    name = Prompt.ask("请输入耗材名称")
    current_stock = IntPrompt.ask("请输入当前库存", default=10)
    min_stock = IntPrompt.ask("请输入最低库存", default=5)
    unit = Prompt.ask("请输入单位", default="个")
    
    result = api.create_supply(name, current_stock, min_stock, unit)
    console.print(f"[green]耗材创建成功! ID: {result['id']}[/green]")


@app.command("list-purchase-todos")
def list_purchase_todos():
    todos = api.list_purchase_todos()
    table = Table(title="采购待办")
    table.add_column("ID", style="cyan")
    table.add_column("耗材ID", style="green")
    table.add_column("需要数量", style="yellow")
    table.add_column("状态", style="magenta")
    
    for t in todos:
        status = "[green]已完成[/green]" if t["is_completed"] else "[red]待采购[/red]"
        table.add_row(
            str(t["id"]),
            str(t["supply_id"]),
            str(t["quantity_needed"]),
            status
        )
    
    console.print(table)


@app.command("complete-purchase")
def complete_purchase():
    list_purchase_todos()
    todo_id = IntPrompt.ask("请输入采购待办ID")
    result = api.complete_purchase_todo(todo_id)
    console.print(f"[green]采购已完成，库存已补充[/green]")


def run():
    app()


if __name__ == "__main__":
    run()
