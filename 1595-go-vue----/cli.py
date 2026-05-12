import os
import click
import requests
import json
from datetime import datetime

BASE_URL = os.environ.get("API_BASE_URL", "http://localhost:8000")


def get_json(url):
    try:
        response = requests.get(url)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        click.echo(f"请求失败: {e}", err=True)
        return None


@click.group()
def cli():
    """文化馆综合管理系统命令行客户端"""
    pass


@cli.group()
def performances():
    """演出管理命令"""
    pass


@performances.command("list")
@click.option("--limit", default=20, help="显示数量")
@click.option("--skip", default=0, help="跳过数量")
def list_performances(limit, skip):
    """列出所有演出"""
    data = get_json(f"{BASE_URL}/api/performances?skip={skip}&limit={limit}")
    if data:
        click.echo("\n演出列表:")
        click.echo("-" * 80)
        for perf in data:
            start = datetime.fromisoformat(perf["start_time"].replace("Z", "+00:00"))
            end = datetime.fromisoformat(perf["end_time"].replace("Z", "+00:00"))
            click.echo(f"ID: {perf['id']}")
            click.echo(f"  名称: {perf['name']}")
            click.echo(f"  时间: {start.strftime('%Y-%m-%d %H:%M')} ~ {end.strftime('%H:%M')}")
            click.echo(f"  场地ID: {perf['venue_id']}")
            click.echo(f"  价格: {perf['price']}分")
            click.echo(f"  状态: {perf['status']}")
            click.echo()


@performances.command("show")
@click.argument("performance_id", type=int)
def show_performance(performance_id):
    """显示演出详情"""
    data = get_json(f"{BASE_URL}/api/performances/{performance_id}")
    if data:
        click.echo(json.dumps(data, indent=2, ensure_ascii=False))


@performances.command("tickets")
@click.argument("performance_id", type=int)
def list_tickets(performance_id):
    """列出演出的所有票"""
    data = get_json(f"{BASE_URL}/api/performances/{performance_id}/tickets")
    if data:
        click.echo(f"\n演出 {performance_id} 的票列表:")
        click.echo("-" * 80)
        for ticket in data:
            click.echo(f"票ID: {ticket['id']}")
            click.echo(f"  客户: {ticket['customer_name']}")
            click.echo(f"  电话: {ticket.get('customer_phone', 'N/A')}")
            click.echo(f"  座位: {ticket.get('seat_number', 'N/A')}")
            click.echo(f"  价格: {ticket['price']}分")
            click.echo(f"  状态: {ticket['status']}")
            click.echo()


@cli.group()
def trainings():
    """培训管理命令"""
    pass


@trainings.command("list")
@click.option("--limit", default=20, help="显示数量")
@click.option("--skip", default=0, help="跳过数量")
def list_trainings(limit, skip):
    """列出所有培训"""
    data = get_json(f"{BASE_URL}/api/trainings?skip={skip}&limit={limit}")
    if data:
        click.echo("\n培训列表:")
        click.echo("-" * 80)
        for training in data:
            start = datetime.fromisoformat(training["start_date"].replace("Z", "+00:00"))
            end = datetime.fromisoformat(training["end_date"].replace("Z", "+00:00"))
            click.echo(f"ID: {training['id']}")
            click.echo(f"  名称: {training['name']}")
            click.echo(f"  讲师: {training.get('instructor', 'N/A')}")
            click.echo(f"  时间: {start.strftime('%Y-%m-%d')} ~ {end.strftime('%Y-%m-%d')}")
            click.echo(f"  时间安排: {training.get('schedule', 'N/A')}")
            click.echo(f"  最大人数: {training['max_participants']}")
            click.echo(f"  场地ID: {training['venue_id']}")
            click.echo(f"  状态: {training['status']}")
            click.echo()


@trainings.command("show")
@click.argument("training_id", type=int)
def show_training(training_id):
    """显示培训详情"""
    data = get_json(f"{BASE_URL}/api/trainings/{training_id}")
    if data:
        click.echo(json.dumps(data, indent=2, ensure_ascii=False))


@trainings.command("registrations")
@click.argument("training_id", type=int)
def list_registrations(training_id):
    """列出培训的所有报名"""
    data = get_json(f"{BASE_URL}/api/trainings/{training_id}/registrations")
    if data:
        click.echo(f"\n培训 {training_id} 的报名列表:")
        click.echo("-" * 80)
        for reg in data:
            click.echo(f"报名ID: {reg['id']}")
            click.echo(f"  学生: {reg['student_name']}")
            click.echo(f"  电话: {reg.get('student_phone', 'N/A')}")
            click.echo(f"  邮箱: {reg.get('student_email', 'N/A')}")
            click.echo(f"  状态: {reg['status']}")
            if reg['status'] == 'waitlist':
                click.echo(f"  候补顺序: {reg['waitlist_order']}")
            click.echo(f"  连续缺勤: {reg['consecutive_absences']}次")
            click.echo()


@cli.command("info")
def info():
    """显示系统信息"""
    data = get_json(f"{BASE_URL}/")
    if data:
        click.echo(json.dumps(data, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    cli()
