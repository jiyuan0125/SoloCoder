import json
from typing import Optional

import typer
from rich.console import Console
from rich.table import Table
from rich.panel import Panel

from client.api_client import APIClient

app = typer.Typer(help="危险品运输管理系统命令行客户端")
console = Console()
api = APIClient()


def _print_json(data):
    console.print(json.dumps(data, ensure_ascii=False, indent=2))


def _print_error(message: str):
    console.print(f"[red]错误:[/red] {message}")


@app.command()
def health():
    """检查服务健康状态"""
    try:
        result = api._request("GET", "/health")
        console.print(f"[green]服务状态:[/green] {result.get('status', 'unknown')}")
    except Exception as e:
        _print_error(str(e))


@app.group()
def ship():
    """船舶管理"""
    pass


@ship.command("create")
def ship_create(
    name: str = typer.Option(..., "--name", "-n", help="船舶名称"),
    imo_number: str = typer.Option(..., "--imo", help="IMO编号"),
    flag: str = typer.Option(..., "--flag", "-f", help="船旗国"),
    has_hazardous: bool = typer.Option(False, "--hazardous", help="是否有危险品运输资质"),
    actor: str = typer.Option(..., "--actor", "-a", help="操作人"),
):
    """创建船舶"""
    try:
        result = api.create_ship(name, imo_number, flag, has_hazardous, actor)
        console.print(Panel(f"[green]船舶创建成功[/green]\nID: {result['id']}\n名称: {result['name']}"))
    except Exception as e:
        _print_error(str(e))


@ship.command("list")
def ship_list():
    """列出所有船舶"""
    try:
        ships = api.list_ships()
        if not ships:
            console.print("暂无船舶数据")
            return
        table = Table(title="船舶列表")
        table.add_column("ID", style="cyan")
        table.add_column("名称", style="green")
        table.add_column("IMO编号", style="yellow")
        table.add_column("船旗国", style="magenta")
        table.add_column("危险品资质", style="blue")
        for ship in ships:
            table.add_row(
                str(ship["id"]),
                ship["name"],
                ship["imo_number"],
                ship["flag"],
                "[green]是[/green]" if ship["has_hazardous_qualification"] else "[red]否[/red]",
            )
        console.print(table)
    except Exception as e:
        _print_error(str(e))


@ship.command("get")
def ship_get(ship_id: int = typer.Argument(..., help="船舶ID")):
    """获取船舶详情"""
    try:
        result = api.get_ship(ship_id)
        _print_json(result)
    except Exception as e:
        _print_error(str(e))


@app.group()
def declaration():
    """申报管理"""
    pass


@declaration.command("create")
def declaration_create(
    number: str = typer.Option(..., "--number", help="申报编号"),
    ship_id: int = typer.Option(..., "--ship-id", help="船舶ID"),
    voyage: str = typer.Option(..., "--voyage", help="航次号"),
    category: int = typer.Option(..., "--category", "-c", help="危险品类别(1-9)"),
    cargo: str = typer.Option(..., "--cargo", help="货物名称"),
    quantity: float = typer.Option(..., "--quantity", "-q", help="货物数量"),
    packaging: bool = typer.Option(True, "--packaging", help="包装是否合规"),
    submitted_by: str = typer.Option(..., "--submitter", help="提交人"),
    actor: str = typer.Option(..., "--actor", "-a", help="操作人"),
):
    """创建申报"""
    try:
        result = api.create_declaration(
            number, ship_id, voyage, category, cargo, quantity, packaging, submitted_by, actor
        )
        console.print(
            Panel(
                f"[green]申报创建成功[/green]\n"
                f"ID: {result['id']}\n"
                f"申报编号: {result['declaration_number']}\n"
                f"状态: {result['status_description']}"
            )
        )
    except Exception as e:
        _print_error(str(e))


@declaration.command("list")
def declaration_list():
    """列出所有申报"""
    try:
        declarations = api.list_declarations()
        if not declarations:
            console.print("暂无申报数据")
            return
        table = Table(title="申报列表")
        table.add_column("ID", style="cyan")
        table.add_column("申报编号", style="green")
        table.add_column("船舶", style="yellow")
        table.add_column("航次", style="magenta")
        table.add_column("危险品类别", style="blue")
        table.add_column("货物", style="white")
        table.add_column("状态", style="bold")
        for d in declarations:
            table.add_row(
                str(d["id"]),
                d["declaration_number"],
                d["ship_name"],
                d["voyage_number"],
                d["hazardous_category_name"],
                d["cargo_name"],
                d["status_description"],
            )
        console.print(table)
    except Exception as e:
        _print_error(str(e))


@declaration.command("get")
def declaration_get(declaration_id: int = typer.Argument(..., help="申报ID")):
    """获取申报详情"""
    try:
        result = api.get_declaration(declaration_id)
        _print_json(result)
    except Exception as e:
        _print_error(str(e))


@declaration.group()
def review():
    """审核管理"""
    pass


@review.command("initial")
def review_initial(
    declaration_id: int = typer.Argument(..., help="申报ID"),
    reviewer: str = typer.Option(..., "--reviewer", "-r", help="审核人"),
    approved: bool = typer.Option(..., "--approved/--rejected", help="是否通过"),
    comments: Optional[str] = typer.Option(None, "--comments", "-c", help="审核意见"),
    actor: str = typer.Option(..., "--actor", "-a", help="操作人"),
):
    """初审"""
    try:
        result = api.initial_review(declaration_id, reviewer, approved, comments, actor)
        status = "[green]通过[/green]" if approved else "[red]驳回[/red]"
        console.print(Panel(f"初审{status}\n当前状态: {result['status_description']}"))
    except Exception as e:
        _print_error(str(e))


@review.command("final")
def review_final(
    declaration_id: int = typer.Argument(..., help="申报ID"),
    reviewer: str = typer.Option(..., "--reviewer", "-r", help="审核人"),
    approved: bool = typer.Option(..., "--approved/--rejected", help="是否通过"),
    comments: Optional[str] = typer.Option(None, "--comments", "-c", help="审核意见"),
    actor: str = typer.Option(..., "--actor", "-a", help="操作人"),
):
    """复审"""
    try:
        result = api.final_review(declaration_id, reviewer, approved, comments, actor)
        status = "[green]通过[/green]" if approved else "[red]驳回[/red]"
        console.print(Panel(f"复审{status}\n当前状态: {result['status_description']}"))
    except Exception as e:
        _print_error(str(e))


@review.command("list")
def review_list(declaration_id: int = typer.Argument(..., help="申报ID")):
    """查询审核记录"""
    try:
        reviews = api.get_reviews(declaration_id)
        if not reviews:
            console.print("暂无审核记录")
            return
        table = Table(title=f"申报 {declaration_id} 审核记录")
        table.add_column("ID", style="cyan")
        table.add_column("类型", style="green")
        table.add_column("审核人", style="yellow")
        table.add_column("结果", style="magenta")
        table.add_column("时间", style="blue")
        for r in reviews:
            table.add_row(
                str(r["id"]),
                "初审" if r["review_type"] == "initial" else "复审",
                r["reviewer"],
                "[green]通过[/green]" if r["approved"] else "[red]驳回[/red]",
                r["reviewed_at"],
            )
        console.print(table)
    except Exception as e:
        _print_error(str(e))


@declaration.group()
def loading():
    """装卸管理"""
    pass


@loading.command("start")
def loading_start(
    declaration_id: int = typer.Argument(..., help="申报ID"),
    operator: str = typer.Option(..., "--operator", "-o", help="操作员"),
    temperature: Optional[float] = typer.Option(None, "--temp", "-t", help="环境温度(爆炸品必填)"),
    radiation: Optional[float] = typer.Option(None, "--radiation", "-r", help="辐射剂量率(放射性物质必填)"),
    notes: Optional[str] = typer.Option(None, "--notes", "-n", help="备注"),
    actor: str = typer.Option(..., "--actor", "-a", help="操作人"),
):
    """开始装卸"""
    try:
        result = api.start_loading(declaration_id, operator, temperature, radiation, notes, actor)
        console.print(Panel(f"[green]装卸开始[/green]\n当前状态: {result['status_description']}"))
    except Exception as e:
        _print_error(str(e))


@loading.command("complete")
def loading_complete(
    declaration_id: int = typer.Argument(..., help="申报ID"),
    notes: Optional[str] = typer.Option(None, "--notes", "-n", help="备注"),
    actor: str = typer.Option(..., "--actor", "-a", help="操作人"),
):
    """完成装卸"""
    try:
        result = api.complete_loading(declaration_id, notes, actor)
        console.print(Panel(f"[green]装卸完成[/green]\n当前状态: {result['status_description']}"))
    except Exception as e:
        _print_error(str(e))


@loading.command("get")
def loading_get(declaration_id: int = typer.Argument(..., help="申报ID")):
    """查询装卸记录"""
    try:
        result = api.get_loading(declaration_id)
        if not result:
            console.print("暂无装卸记录")
            return
        _print_json(result)
    except Exception as e:
        _print_error(str(e))


@declaration.group()
def emergency():
    """应急处置"""
    pass


@emergency.command("create")
def emergency_create(
    declaration_id: int = typer.Argument(..., help="申报ID"),
    impact_range: str = typer.Option(..., "--impact", "-i", help="影响范围(单个舱室/多个舱室/全船/码头区域)"),
    triggered_by: Optional[str] = typer.Option(None, "--triggered", help="触发人"),
    actor: str = typer.Option(..., "--actor", "-a", help="操作人"),
):
    """创建应急事件"""
    try:
        result = api.create_emergency(declaration_id, impact_range, triggered_by, actor)
        console.print(
            Panel(
                f"[red]应急事件创建[/red]\n"
                f"事件等级: {result['level_description']}\n"
                f"\n[bold]处置方案:[/bold]\n{result['plan']}"
            )
        )
    except Exception as e:
        _print_error(str(e))


@emergency.command("list")
def emergency_list(declaration_id: int = typer.Argument(..., help="申报ID")):
    """查询应急记录"""
    try:
        emergencies = api.get_emergencies(declaration_id)
        if not emergencies:
            console.print("暂无应急记录")
            return
        for e in emergencies:
            console.print(
                Panel(
                    f"应急事件 #{e['id']}\n"
                    f"等级: {e['level_description']}\n"
                    f"影响范围: {e['impact_range']}\n"
                    f"\n[bold]处置方案:[/bold]\n{e['plan']}",
                    title=f"应急事件 {e['id']}",
                )
            )
    except Exception as e:
        _print_error(str(e))


@app.command()
def audit(declaration_id: Optional[int] = typer.Option(None, "--declaration", "-d", help="筛选申报ID")):
    """查询审计日志"""
    try:
        logs = api.list_audit_logs(declaration_id)
        if not logs:
            console.print("暂无审计日志")
            return
        table = Table(title="审计日志")
        table.add_column("ID", style="cyan")
        table.add_column("申报ID", style="green")
        table.add_column("操作", style="yellow")
        table.add_column("操作人", style="magenta")
        table.add_column("时间", style="blue")
        for log in logs:
            table.add_row(
                str(log["id"]),
                str(log["declaration_id"] or "-"),
                log["action"],
                log["actor"],
                log["timestamp"],
            )
        console.print(table)
    except Exception as e:
        _print_error(str(e))


if __name__ == "__main__":
    app()
