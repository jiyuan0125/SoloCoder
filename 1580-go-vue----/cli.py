import typer
import requests
import os
from typing import Optional
from datetime import datetime

app = typer.Typer()

BASE_URL = os.getenv("API_BASE_URL", "http://localhost:8000")

@app.command()
def pending(inspector_id: int):
    """查询巡检员的待办任务"""
    try:
        response = requests.get(f"{BASE_URL}/todos/{inspector_id}/pending")
        response.raise_for_status()
        tasks = response.json()
        
        if not tasks:
            typer.echo("暂无待办任务")
            return
        
        typer.echo(f"\n=== 巡检员 {inspector_id} 的待办任务 ===")
        for task in tasks:
            typer.echo(f"\n任务编号: {task['inspection_code']}")
            typer.echo(f"  类型: {task['inspection_type']}")
            typer.echo(f"  计划开始: {task['planned_start_time']}")
            typer.echo(f"  计划结束: {task['planned_end_time']}")
            typer.echo(f"  状态: {task['status']}")
    except requests.exceptions.RequestException as e:
        typer.echo(f"请求失败: {e}", err=True)

@app.command()
def overdue(inspector_id: int):
    """查询巡检员的逾期任务"""
    try:
        response = requests.get(f"{BASE_URL}/todos/{inspector_id}/overdue")
        response.raise_for_status()
        tasks = response.json()
        
        if not tasks:
            typer.echo("暂无逾期任务")
            return
        
        typer.echo(f"\n=== 巡检员 {inspector_id} 的逾期任务 ===")
        for task in tasks:
            typer.echo(f"\n任务编号: {task['inspection_code']}")
            typer.echo(f"  类型: {task['inspection_type']}")
            typer.echo(f"  计划结束: {task['planned_end_time']}")
            typer.echo(f"  状态: {task['status']}")
    except requests.exceptions.RequestException as e:
        typer.echo(f"请求失败: {e}", err=True)

@app.command()
def statistics():
    """查询巡检统计信息"""
    try:
        response = requests.get(f"{BASE_URL}/statistics/")
        response.raise_for_status()
        stats = response.json()
        
        typer.echo("\n=== 本月巡检统计 ===")
        typer.echo(f"\n巡检完成率: {stats['completion_rate']}%")
        
        typer.echo(f"\n问题数量按等级分布:")
        for level, count in stats['problems_by_level'].items():
            typer.echo(f"  {level}: {count}")
        
        typer.echo(f"\n设备状态分布:")
        for status, count in stats['device_status_distribution'].items():
            typer.echo(f"  {status}: {count}")
    except requests.exceptions.RequestException as e:
        typer.echo(f"请求失败: {e}", err=True)

if __name__ == "__main__":
    app()
