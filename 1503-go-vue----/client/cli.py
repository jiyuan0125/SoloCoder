import click
from tabulate import tabulate
from datetime import date, datetime, timedelta
from typing import List, Dict
from .api_client import ApiClient


def get_client():
    return ApiClient()


@click.group()
def cli():
    """林火巡护管理系统命令行客户端"""
    pass


@cli.group()
def user():
    """用户管理"""
    pass


@user.command("list")
def list_users():
    """列出所有用户"""
    client = get_client()
    users = client.get_users()
    
    if not users:
        click.echo("没有用户")
        return
    
    table = []
    for u in users:
        table.append([u["id"], u["name"], u["role"], u.get("phone", "-")])
    
    click.echo(tabulate(table, headers=["ID", "姓名", "角色", "电话"], tablefmt="grid"))


@user.command("create")
@click.option("--name", required=True, help="用户姓名")
@click.option("--role", type=click.Choice(["admin", "area_manager", "ranger"]), required=True, help="用户角色")
@click.option("--phone", default=None, help="电话")
def create_user_cmd(name, role, phone):
    """创建新用户"""
    client = get_client()
    result = client.create_user(name, role, phone)
    click.echo(f"用户创建成功: ID={result['id']}, 姓名={result['name']}")


@cli.group()
def area():
    """林区管理"""
    pass


@area.command("list")
def list_areas():
    """列出所有林区"""
    client = get_client()
    areas = client.get_areas()
    
    if not areas:
        click.echo("没有林区")
        return
    
    table = []
    for a in areas:
        table.append([
            a["id"], a["name"], f"{a['area_km2']} km²",
            a["main_tree_species"], a.get("ranger_id", "-")
        ])
    
    click.echo(tabulate(table, headers=["ID", "名称", "面积", "主要树种", "护林员ID"], tablefmt="grid"))


@area.command("create")
@click.option("--name", required=True, help="林区名称")
@click.option("--area", type=float, required=True, help="面积(平方公里)")
@click.option("--trees", required=True, help="主要树种")
@click.option("--ranger", type=int, default=None, help="护林员ID")
def create_area_cmd(name, area, trees, ranger):
    """创建新林区"""
    client = get_client()
    result = client.create_area(name, area, trees, ranger)
    click.echo(f"林区创建成功: ID={result['id']}, 名称={result['name']}")


@cli.group()
def task():
    """巡护任务管理"""
    pass


@task.command("list")
@click.option("--pending", is_flag=True, help="只显示未提交报告的任务")
def list_tasks(pending):
    """列出巡护任务"""
    client = get_client()
    
    if pending:
        tasks = client.get_pending_tasks()
        click.echo("=== 未提交报告的任务 ===")
    else:
        tasks = client.get_tasks()
    
    if not tasks:
        click.echo("没有任务")
        return
    
    table = []
    for t in tasks:
        keypoints = len(t["route_keypoints"])
        table.append([
            t["id"], t["area_id"], t["ranger_id"],
            t["patrol_date"], f"{keypoints} 个关键点"
        ])
    
    click.echo(tabulate(table, headers=["ID", "林区ID", "护林员ID", "巡护日期", "关键点"], tablefmt="grid"))


@task.command("create")
@click.option("--area-id", type=int, required=True, help="林区ID")
@click.option("--ranger-id", type=int, required=True, help="护林员ID")
@click.option("--date", "patrol_date", required=True, help="巡护日期 (YYYY-MM-DD)")
@click.option("--keypoints", required=True, help="路线关键点，格式: lat1,lng1;lat2,lng2;...")
def create_task_cmd(area_id, ranger_id, patrol_date, keypoints):
    """创建新巡护任务"""
    try:
        route_points = []
        for kp in keypoints.split(";"):
            if kp.strip():
                parts = kp.split(",")
                if len(parts) == 2:
                    route_points.append({
                        "latitude": float(parts[0].strip()),
                        "longitude": float(parts[1].strip())
                    })
        
        if not route_points:
            raise ValueError("没有有效的路线关键点")
        
        client = get_client()
        result = client.create_task(area_id, ranger_id, patrol_date, route_points)
        click.echo(f"任务创建成功: ID={result['id']}, 日期={result['patrol_date']}")
    except Exception as e:
        click.echo(f"错误: {str(e)}")
        click.echo("关键点格式示例: 34.123,116.456;34.234,116.567")


@task.command("show")
@click.option("--id", type=int, required=True, help="任务ID")
def show_task(id):
    """显示任务详情"""
    client = get_client()
    task = client.get_task(id)
    
    click.echo(f"任务ID: {task['id']}")
    click.echo(f"林区ID: {task['area_id']}")
    click.echo(f"护林员ID: {task['ranger_id']}")
    click.echo(f"巡护日期: {task['patrol_date']}")
    click.echo(f"创建时间: {task.get('created_at', '-')}")
    click.echo(f"\n路线关键点 ({len(task['route_keypoints'])} 个):")
    for i, kp in enumerate(task['route_keypoints'], 1):
        click.echo(f"  {i}. ({kp['latitude']:.4f}, {kp['longitude']:.4f})")


@cli.group()
def report():
    """巡护报告管理"""
    pass


@report.command("list")
def list_reports():
    """列出所有报告"""
    client = get_client()
    reports = client.get_reports()
    
    if not reports:
        click.echo("没有报告")
        return
    
    table = []
    for r in reports:
        status_color = "green" if r["status"] == "qualified" else "red"
        table.append([
            r["id"], r["task_id"],
            click.style(r["status"], fg=status_color),
            len(r["actual_route"]),
            len(r["anomalies"]),
            r.get("submitted_at", "-")
        ])
    
    click.echo(tabulate(table, headers=["ID", "任务ID", "状态", "路线点", "异常数", "提交时间"], tablefmt="grid"))


@report.command("submit")
@click.option("--task-id", type=int, required=True, help="任务ID")
@click.option("--route", required=True, help="实际路线点，格式: lat1,lng1;lat2,lng2;...")
@click.option("--anomaly", multiple=True, help="异常信息，可多次使用。格式: type;location;detail")
def submit_report(task_id, route, anomaly):
    """提交巡护报告"""
    try:
        route_points = []
        for p in route.split(";"):
            if p.strip():
                parts = p.split(",")
                if len(parts) == 2:
                    route_points.append({
                        "latitude": float(parts[0].strip()),
                        "longitude": float(parts[1].strip())
                    })
        
        if not route_points:
            raise ValueError("没有有效的路线点")
        
        anomalies = []
        for a in anomaly:
            parts = a.split(";", 3)
            if len(parts) >= 3:
                anomaly_type = parts[0].strip()
                loc_parts = parts[1].split(",")
                if len(loc_parts) == 2:
                    loc = {
                        "latitude": float(loc_parts[0].strip()),
                        "longitude": float(loc_parts[1].strip())
                    }
                    detail = parts[2].strip()
                    
                    if anomaly_type == "pest":
                        pest_parts = detail.split(";")
                        anomalies.append({
                            "report_id": 0,
                            "anomaly_type": "pest",
                            "location": loc,
                            "pest": {
                                "type": pest_parts[0] if len(pest_parts) > 0 else "unknown",
                                "severity": pest_parts[1] if len(pest_parts) > 1 else "mild",
                                "trees_affected": int(pest_parts[2]) if len(pest_parts) > 2 else 0,
                                "location": loc
                            }
                        })
                    elif anomaly_type == "fire_risk":
                        anomalies.append({
                            "report_id": 0,
                            "anomaly_type": "fire_risk",
                            "location": loc,
                            "fire_risk": {
                                "type": detail,
                                "status": "pending",
                                "location": loc
                            }
                        })
        
        client = get_client()
        result = client.create_report(task_id, route_points, anomalies)
        
        status = "合格" if result["status"] == "qualified" else "不合格"
        status_color = "green" if result["status"] == "qualified" else "red"
        
        click.echo(f"报告提交成功: ID={result['id']}")
        click.echo(f"状态: {click.style(status, fg=status_color)}")
        click.echo(f"异常数量: {len(result['anomalies'])}")
        
        for a in result["anomalies"]:
            dup_text = " (重复上报)" if a.get("is_duplicate") else ""
            click.echo(f"  - 异常: {a['anomaly_type']}{dup_text}")
        
    except Exception as e:
        click.echo(f"错误: {str(e)}")


@cli.group()
def todo():
    """待办事项管理"""
    pass


@todo.command("list")
def list_todos():
    """列出所有待办事项"""
    client = get_client()
    todos = client.get_todos()
    
    if not todos:
        click.echo("没有待办事项")
        return
    
    status_colors = {
        "pending": "yellow",
        "in_progress": "blue",
        "resolved": "green",
        "overdue": "red"
    }
    
    type_colors = {
        "fire_risk": "red",
        "pest": "yellow"
    }
    
    table = []
    for t in todos:
        status_color = status_colors.get(t["status"], "white")
        type_color = type_colors.get(t["anomaly_type"], "white")
        
        table.append([
            t["id"],
            click.style(t["anomaly_type"], fg=type_color, bold=True),
            click.style(t["status"], fg=status_color),
            t["area_id"],
            t["assigned_user_id"],
            t.get("created_at", "-")
        ])
    
    click.echo(tabulate(table, headers=["ID", "类型", "状态", "林区ID", "分配给", "创建时间"], tablefmt="grid"))
    click.echo("\n说明: fire_risk (火险隐患) 优先级高于 pest (病虫害)")


@todo.command("resolve")
@click.option("--id", type=int, required=True, help="待办事项ID")
def resolve_todo_cmd(id):
    """解决待办事项"""
    client = get_client()
    result = client.resolve_todo(id)
    click.echo(f"待办事项已解决: ID={result['id']}")


@cli.group()
def stats():
    """统计分析"""
    pass


@stats.command("pest")
@click.option("--year", type=int, default=lambda: datetime.now().year, help="年份")
@click.option("--month", type=int, default=lambda: datetime.now().month, help="月份")
def pest_stats(year, month):
    """病虫害月度统计"""
    client = get_client()
    stats = client.get_monthly_pest_stats(year, month)
    
    click.echo(f"\n=== {year}年{month}月 病虫害统计 ===")
    click.echo(f"新发现数量: {stats['new_discovery_count']}")
    click.echo(f"涉及树木总数: {stats['trees_affected_total']}")
    click.echo(f"重度占比: {stats['severe_percentage']:.2f}%")


@stats.command("health")
def health_scores():
    """区域健康度看板"""
    client = get_client()
    scores = client.get_area_health_scores()
    
    if not scores:
        click.echo("没有数据")
        return
    
    def get_health_color(score):
        if score >= 80:
            return "green"
        elif score >= 60:
            return "yellow"
        elif score >= 40:
            return "orange"
        else:
            return "red"
    
    table = []
    for s in scores:
        color = get_health_color(s["score"])
        table.append([
            s["area_id"],
            s["area_name"],
            click.style(f"{s['score']:.2f}", fg=color, bold=True),
            s["anomaly_count"],
            s["severe_anomaly_count"]
        ])
    
    click.echo(tabulate(table, headers=["ID", "名称", "健康度", "异常数", "严重异常数"], tablefmt="grid"))


if __name__ == "__main__":
    cli()
