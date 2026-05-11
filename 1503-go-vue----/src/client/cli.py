import click
import json
from datetime import date, timedelta
from typing import List

from src.client.api_client import APIClient


def get_client() -> APIClient:
    return APIClient()


def print_json(data):
    click.echo(json.dumps(data, ensure_ascii=False, indent=2))


@click.group()
@click.version_option(version="1.0.0")
def cli():
    pass


@cli.group()
def user():
    pass


@user.command("create")
@click.option("--name", required=True, help="用户姓名")
@click.option("--phone", required=True, help="电话号码")
@click.option("--role", required=True, type=click.Choice(["admin", "ranger", "area_manager"]), help="角色")
def user_create(name, phone, role):
    client = get_client()
    try:
        result = client.create_user(name, phone, role)
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@user.command("list")
def user_list():
    client = get_client()
    try:
        users = client.list_users()
        for u in users:
            click.echo(f"ID: {u['id']}, 姓名: {u['name']}, 电话: {u['phone']}, 角色: {u['role']}")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@user.command("get")
@click.argument("user_id", type=int)
def user_get(user_id):
    client = get_client()
    try:
        result = client.get_user(user_id)
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@cli.group()
def area():
    pass


@area.command("create")
@click.option("--name", required=True, help="区域名称")
@click.option("--area", "area_km2", required=True, type=float, help="面积(平方公里)")
@click.option("--species", required=True, help="主要树种")
@click.option("--ranger-id", type=int, default=None, help="护林员ID")
@click.option("--manager-id", type=int, default=None, help="区域管理员ID")
def area_create(name, area_km2, species, ranger_id, manager_id):
    client = get_client()
    try:
        result = client.create_area(name, area_km2, species, ranger_id, manager_id)
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@area.command("list")
def area_list():
    client = get_client()
    try:
        areas = client.list_areas()
        for a in areas:
            click.echo(
                f"ID: {a['id']}, 名称: {a['name']}, 面积: {a['area_km2']}km², 树种: {a['main_tree_species']}"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@area.command("get")
@click.argument("area_id", type=int)
def area_get(area_id):
    client = get_client()
    try:
        result = client.get_area(area_id)
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@cli.group()
def task():
    pass


@task.command("create")
@click.option("--area-id", required=True, type=int, help="区域ID")
@click.option("--ranger-id", required=True, type=int, help="护林员ID")
@click.option("--date", "patrol_date", required=True, help="巡护日期 (YYYY-MM-DD)")
@click.option("--waypoints", required=True, help='路线点 JSON, 如: [{"lat":30.0,"lng":120.0},...]')
def task_create(area_id, ranger_id, patrol_date, waypoints):
    client = get_client()
    try:
        route_waypoints = json.loads(waypoints)
        result = client.create_task(area_id, ranger_id, patrol_date, route_waypoints)
        print_json(result)
    except json.JSONDecodeError:
        click.echo("错误: waypoints 格式不正确，请使用正确的 JSON 数组", err=True)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@task.command("list")
def task_list():
    client = get_client()
    try:
        tasks = client.list_tasks()
        for t in tasks:
            area_name = t.get("area", {}).get("name", "未知") if t.get("area") else "未知"
            ranger_name = t.get("ranger", {}).get("name", "未知") if t.get("ranger") else "未知"
            click.echo(
                f"ID: {t['id']}, 区域: {area_name}, 护林员: {ranger_name}, 日期: {t['patrol_date']}"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@task.command("get")
@click.argument("task_id", type=int)
def task_get(task_id):
    client = get_client()
    try:
        result = client.get_task(task_id)
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@cli.group()
def report():
    pass


@report.command("submit")
@click.option("--task-id", required=True, type=int, help="任务ID")
@click.option("--route", required=True, help='实际路线点 JSON, 如: [{"lat":30.0,"lng":120.0},...]')
@click.option("--anomalies", default="[]", help='异常列表 JSON')
def report_submit(task_id, route, anomalies):
    client = get_client()
    try:
        actual_route = json.loads(route)
        anomaly_list = json.loads(anomalies)
        result = client.submit_report(task_id, actual_route, anomaly_list)
        qual = "合格" if result["is_qualified"] else "不合格"
        click.echo(f"报告提交成功! ID: {result['id']}, 状态: {qual}")
    except json.JSONDecodeError:
        click.echo("错误: 路线点或异常列表格式不正确", err=True)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@report.command("list")
def report_list():
    client = get_client()
    try:
        reports = client.list_reports()
        for r in reports:
            qual = "合格" if r["is_qualified"] else "不合格"
            click.echo(
                f"ID: {r['id']}, 任务ID: {r['task_id']}, 状态: {qual}, 提交时间: {r['submitted_at']}"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@report.command("get")
@click.argument("report_id", type=int)
def report_get(report_id):
    client = get_client()
    try:
        result = client.get_report(report_id)
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@cli.group()
def todo():
    pass


@todo.command("list")
@click.option("--overdue-only", is_flag=True, help="只显示逾期的")
def todo_list(overdue_only):
    client = get_client()
    try:
        todos = client.list_todos(overdue_only=overdue_only)
        for t in todos:
            status_map = {"pending": "待处理", "in_progress": "处理中", "done": "已完成", "overdue": "逾期"}
            priority_map = {1: "普通", 2: "高"}
            status = status_map.get(t["status"], t["status"])
            priority = priority_map.get(t["priority"], t["priority"])
            anom_type = t.get("anomaly", {}).get("anomaly_type", "") if t.get("anomaly") else ""
            duplicate = "(重复)" if t.get("anomaly", {}).get("is_duplicate") else ""
            click.echo(
                f"ID: {t['id']}, 优先级: {priority}, 状态: {status}, "
                f"类型: {anom_type}{duplicate}, 截止: {t['due_date']}"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@todo.command("get")
@click.argument("todo_id", type=int)
def todo_get(todo_id):
    client = get_client()
    try:
        result = client.get_todo(todo_id)
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@todo.command("update")
@click.argument("todo_id", type=int)
@click.option("--status", type=click.Choice(["pending", "in_progress", "done", "overdue"]), help="状态")
@click.option("--notes", help="备注")
def todo_update(todo_id, status, notes):
    if status is None and notes is None:
        click.echo("错误: 必须提供 --status 或 --notes", err=True)
        return
    client = get_client()
    try:
        result = client.update_todo(todo_id, status, notes)
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@cli.group()
def stats():
    pass


@stats.command("pest-monthly")
def stats_pest_monthly():
    client = get_client()
    try:
        result = client.get_monthly_pest_stats()
        stats = result.get("stats", [])
        if not stats:
            click.echo("暂无数据")
            return
        click.echo("=== 病虫害月度统计 ===")
        for s in stats:
            ratio_pct = round(s["severe_ratio"] * 100, 1)
            click.echo(
                f"{s['year']}年{s['month']}月: 新发现 {s['total_anomalies']} 起, "
                f"涉及树木 {s['trees_affected']} 棵, 重度占比 {ratio_pct}%"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@stats.command("health")
def stats_health():
    client = get_client()
    try:
        result = client.get_area_health_scores()
        scores = result.get("scores", [])
        if not scores:
            click.echo("暂无数据")
            return
        click.echo("=== 区域健康度看板 ===")
        for s in scores:
            click.echo(
                f"区域: {s['area_name']}, 健康评分: {s['score']}, "
                f"异常总数: {s['total_anomalies']}, 重度: {s['severe_anomalies']}, "
                f"火险: {s['fire_risk_count']}, 病虫害: {s['pest_count']}"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@cli.command("demo")
def demo():
    today = date.today()
    tomorrow = today + timedelta(days=1)
    day_after = tomorrow + timedelta(days=1)

    client = get_client()
    try:
        click.echo("=== 正在创建演示数据 ===")

        click.echo("\n1. 创建用户...")
        admin = client.create_user("张管理员", "13800000001", "admin")
        manager = client.create_user("李经理", "13800000002", "area_manager")
        ranger = client.create_user("王护林员", "13800000003", "ranger")
        click.echo(f"   管理员: {admin['name']}(ID:{admin['id']})")
        click.echo(f"   区域经理: {manager['name']}(ID:{manager['id']})")
        click.echo(f"   护林员: {ranger['name']}(ID:{ranger['id']})")

        click.echo("\n2. 创建区域...")
        area = client.create_area(
            "青松岭林区",
            50.5,
            "松树、柏树",
            ranger_id=ranger['id'],
            manager_id=manager['id']
        )
        click.echo(f"   区域: {area['name']}(ID:{area['id']})")

        click.echo("\n3. 派发巡护任务...")
        waypoints1 = [
            {"lat": 30.123, "lng": 120.456},
            {"lat": 30.124, "lng": 120.457},
            {"lat": 30.125, "lng": 120.458},
            {"lat": 30.126, "lng": 120.459}
        ]
        task1 = client.create_task(
            area['id'],
            ranger['id'],
            tomorrow.strftime("%Y-%m-%d"),
            waypoints1
        )
        click.echo(f"   任务ID: {task1['id']}, 日期: {tomorrow}")

        waypoints2 = [
            {"lat": 30.200, "lng": 120.500},
            {"lat": 30.201, "lng": 120.501}
        ]
        task2 = client.create_task(
            area['id'],
            ranger['id'],
            day_after.strftime("%Y-%m-%d"),
            waypoints2
        )
        click.echo(f"   任务ID: {task2['id']}, 日期: {day_after}")

        click.echo("\n4. 提交巡护报告(含异常)...")
        actual_route1 = [
            {"lat": 30.123, "lng": 120.456},
            {"lat": 30.124, "lng": 120.457},
            {"lat": 30.125, "lng": 120.458}
        ]
        anomalies1 = [
            {
                "location_lat": 30.124,
                "location_lng": 120.457,
                "description": "发现松树枯萎",
                "pest_data": {
                    "pest_type": "disease",
                    "severity": "severe",
                    "trees_affected": 15,
                    "notes": "疑似松材线虫病"
                }
            },
            {
                "location_lat": 30.125,
                "location_lng": 120.458,
                "description": "发现干枯植被堆积",
                "fire_risk_data": {
                    "risk_type": "dry_vegetation",
                    "status": "pending"
                }
            }
        ]
        report1 = client.submit_report(task1['id'], actual_route1, anomalies1)
        qual1 = "合格" if report1["is_qualified"] else "不合格"
        click.echo(f"   报告ID: {report1['id']}, 状态: {qual1}")

        click.echo("\n5. 提交不合格报告(路线点不足)...")
        actual_route2 = [{"lat": 30.200, "lng": 120.500}]
        report2 = client.submit_report(task2['id'], actual_route2, [])
        qual2 = "合格" if report2["is_qualified"] else "不合格"
        click.echo(f"   报告ID: {report2['id']}, 状态: {qual2}")

        click.echo("\n6. 查看待办事项...")
        todos = client.list_todos()
        if todos:
            for t in todos:
                priority_map = {1: "普通", 2: "高(火险优先)"}
                click.echo(
                    f"   待办ID: {t['id']}, 优先级: {priority_map.get(t['priority'], t['priority'])}, "
                    f"类型: {t.get('anomaly', {}).get('anomaly_type', '')}"
                )
        else:
            click.echo("   暂无待办")

        click.echo("\n7. 查看月度病虫害统计...")
        pest_stats = client.get_monthly_pest_stats()
        if pest_stats.get("stats"):
            for s in pest_stats["stats"]:
                ratio = round(s["severe_ratio"] * 100, 1)
                click.echo(
                    f"   {s['year']}年{s['month']}月: {s['total_anomalies']}起, "
                    f"{s['trees_affected']}棵树, 重度占比{ratio}%"
                )
        else:
            click.echo("   暂无统计数据")

        click.echo("\n8. 查看区域健康度...")
        health = client.get_area_health_scores()
        for h in health.get("scores", []):
            click.echo(
                f"   {h['area_name']}: 评分 {h['score']}, 异常 {h['total_anomalies']} 起"
            )

        click.echo("\n=== 演示数据创建完成 ===")
        click.echo(f"服务地址: {client.base_url}")
        click.echo("使用以下命令查看详情:")
        click.echo("  python -m client task list")
        click.echo("  python -m client report list")
        click.echo("  python -m client todo list")
        click.echo("  python -m client stats pest-monthly")
        click.echo("  python -m client stats health")

    except Exception as e:
        click.echo(f"\n错误: {e}", err=True)
    finally:
        client.close()


if __name__ == "__main__":
    cli()
