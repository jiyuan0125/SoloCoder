#!/usr/bin/env python3
import click
import requests
import json
from datetime import datetime

BASE_URL = "http://localhost:8000"


def api_get(endpoint):
    try:
        response = requests.get(f"{BASE_URL}{endpoint}")
        response.raise_for_status()
        return response.json()
    except requests.exceptions.ConnectionError:
        click.echo(click.style("错误: 无法连接到API服务器，请确保服务器已启动", fg="red"))
        return None
    except requests.exceptions.HTTPError as e:
        click.echo(click.style(f"HTTP错误: {e}", fg="red"))
        return None


@click.group()
def cli():
    """公园综合管理系统 - 命令行客户端"""
    pass


@cli.group()
def maintenance():
    """养护任务相关命令"""
    pass


@maintenance.command("list")
@click.option("--park", "-p", help="公园编号")
@click.option("--zone", "-z", help="区域编号")
@click.option("--status", "-s", type=click.Choice(['pending', 'completed', 'overdue', 'in_progress']), help="任务状态")
def list_maintenance(park, zone, status):
    """列出养护任务"""
    if park and zone:
        endpoint = f"/parks/{park}/green-zones/{zone}/maintenance/tasks"
        if status:
            endpoint += f"?status={status}"
        tasks = api_get(endpoint)
    else:
        endpoint = "/scheduler/schedule?limit=100"
        tasks = api_get(endpoint)
    
    if tasks is None:
        return
    
    if not tasks:
        click.echo(click.style("暂无养护任务", fg="yellow"))
        return
    
    click.echo(click.style("\n=== 养护任务列表 ===", fg="cyan", bold=True))
    click.echo(f"共 {len(tasks)} 个任务\n")
    
    for task in tasks:
        status_color = {
            'pending': 'yellow',
            'completed': 'green',
            'overdue': 'red',
            'in_progress': 'blue'
        }.get(task.get('status', ''), 'white')
        
        scheduled_date = task.get('scheduled_date', '')
        if scheduled_date:
            try:
                dt = datetime.fromisoformat(scheduled_date.replace('Z', '+00:00'))
                scheduled_date = dt.strftime('%Y-%m-%d %H:%M')
            except:
                pass
        
        click.echo(f"任务ID: {click.style(str(task.get('task_id', '-')), fg='white')}")
        click.echo(f"任务编号: {click.style(task.get('task_code', '-'), fg='white')}")
        click.echo(f"区域: {click.style(task.get('green_zone_name', '-'), fg='white')}")
        click.echo(f"养护等级: {click.style(task.get('maintenance_grade', '-'), fg='white')}")
        click.echo(f"计划日期: {click.style(scheduled_date, fg='white')}")
        click.echo(f"状态: {click.style(task.get('status', '-'), fg=status_color)}")
        if 'tasks_in_zone' in task:
            click.echo(f"同区域任务数: {click.style(str(task['tasks_in_zone']), fg='white')}")
        click.echo(click.style("-" * 40, fg="bright_black"))


@maintenance.command("overdue")
def list_overdue():
    """列出逾期养护任务"""
    tasks = api_get("/scheduler/overdue")
    
    if tasks is None:
        return
    
    if not tasks:
        click.echo(click.style("暂无逾期养护任务", fg="green"))
        return
    
    click.echo(click.style("\n=== 逾期养护任务 ===", fg="red", bold=True))
    click.echo(f"共 {len(tasks)} 个逾期任务\n")
    
    for task in tasks:
        scheduled_date = task.get('scheduled_date', '')
        if scheduled_date:
            try:
                dt = datetime.fromisoformat(scheduled_date.replace('Z', '+00:00'))
                scheduled_date = dt.strftime('%Y-%m-%d %H:%M')
            except:
                pass
        
        click.echo(f"任务ID: {click.style(str(task['task_id']), fg='white')}")
        click.echo(f"区域: {click.style(task['green_zone_name'], fg='white')}")
        click.echo(f"计划日期: {click.style(scheduled_date, fg='white')}")
        click.echo(f"逾期天数: {click.style(str(task.get('days_overdue', 0)), fg='red', bold=True)}")
        click.echo(click.style("-" * 40, fg="bright_black"))


@maintenance.command("summary")
def maintenance_summary():
    """按区域统计养护任务"""
    summary = api_get("/scheduler/zone-summary")
    
    if summary is None:
        return
    
    click.echo(click.style("\n=== 各区域养护任务统计 ===", fg="cyan", bold=True))
    click.echo(f"{'区域名称':<20} {'等级':<10} {'待处理':<8} {'逾期':<6} {'已完成':<8}")
    click.echo(click.style("-" * 60, fg="bright_black"))
    
    for zone in summary:
        overdue_str = click.style(str(zone['overdue_tasks']), fg='red') if zone['overdue_tasks'] > 0 else str(zone['overdue_tasks'])
        click.echo(f"{zone['name']:<20} {zone['current_grade']:<10} {zone['pending_tasks']:<8} {overdue_str:<6} {zone['completed_tasks']:<8}")


@cli.group()
def facilities():
    """设施相关命令"""
    pass


@facilities.command("list")
@click.option("--park", "-p", required=True, help="公园编号")
@click.option("--status", "-s", type=click.Choice(['normal', 'damaged', 'repairing', 'disabled']), help="设施状态")
def list_facilities(park, status):
    """列出公园设施"""
    endpoint = f"/parks/{park}/facilities/"
    if status:
        endpoint += f"?status={status}"
    
    facilities_data = api_get(endpoint)
    
    if facilities_data is None:
        return
    
    if not facilities_data:
        click.echo(click.style("暂无设施数据", fg="yellow"))
        return
    
    status_map = {
        'normal': ('正常', 'green'),
        'damaged': ('损坏', 'red'),
        'repairing': ('维修中', 'yellow'),
        'disabled': ('停用', 'bright_black')
    }
    
    click.echo(click.style(f"\n=== 公园 {park} 设施列表 ===", fg="cyan", bold=True))
    click.echo(f"{'编号':<15} {'名称':<20} {'类型':<12} {'安全相关':<8} {'状态'}")
    click.echo(click.style("-" * 70, fg="bright_black"))
    
    for fac in facilities_data:
        status_text, status_color = status_map.get(fac['status'], ('未知', 'white'))
        safety = click.style('是', fg='red', bold=True) if fac['is_safety_related'] else '否'
        status_display = click.style(status_text, fg=status_color)
        
        click.echo(f"{fac['facility_code']:<15} {fac['name']:<20} {fac['facility_type']:<12} {safety:<8} {status_display}")


@facilities.command("status")
@click.option("--park", "-p", required=True, help="公园编号")
@click.option("--facility", "-f", required=True, help="设施编号")
def facility_status(park, facility):
    """查看设施详细状态"""
    fac = api_get(f"/parks/{park}/facilities/{facility}")
    
    if fac is None:
        return
    
    status_map = {
        'normal': ('正常', 'green'),
        'damaged': ('损坏', 'red'),
        'repairing': ('维修中', 'yellow'),
        'disabled': ('停用', 'bright_black')
    }
    
    click.echo(click.style(f"\n=== 设施详情 ===", fg="cyan", bold=True))
    click.echo(f"设施编号: {fac['facility_code']}")
    click.echo(f"设施名称: {fac['name']}")
    click.echo(f"设施类型: {fac['facility_type']}")
    
    safety = click.style('是 (需24小时内处理)', fg='red', bold=True) if fac['is_safety_related'] else '否'
    click.echo(f"安全相关: {safety}")
    
    status_text, status_color = status_map.get(fac['status'], ('未知', 'white'))
    click.echo(f"当前状态: {click.style(status_text, fg=status_color, bold=True)}")
    click.echo(f"位置描述: {fac.get('location_description', '-')}")


@facilities.command("urgent")
def urgent_reports():
    """查看紧急报告（24小时内需处理）"""
    reports = api_get("/scheduler/reports/urgent")
    
    if reports is None:
        return
    
    if not reports:
        click.echo(click.style("暂无紧急报告", fg="green"))
        return
    
    click.echo(click.style("\n=== 紧急设施报告 ===", fg="red", bold=True))
    click.echo(f"共 {len(reports)} 个紧急报告需要在24小时内处理\n")
    
    for report in reports:
        deadline = report.get('deadline', '')
        if deadline:
            try:
                dt = datetime.fromisoformat(deadline.replace('Z', '+00:00'))
                deadline = dt.strftime('%Y-%m-%d %H:%M')
            except:
                pass
        
        click.echo(f"报告编号: {report['report_code']}")
        click.echo(f"设施ID: {report['facility_id']}")
        click.echo(f"损坏描述: {report['damage_description']}")
        click.echo(f"截止时间: {click.style(deadline, fg='red')}")
        click.echo(click.style("-" * 50, fg="bright_black"))


@cli.group()
def activities():
    """活动相关命令"""
    pass


@activities.command("list")
@click.option("--park", "-p", required=True, help="公园编号")
@click.option("--status", "-s", type=click.Choice(['pending', 'approved', 'rejected', 'completed']), help="活动状态")
def list_activities(park, status):
    """列出公园活动"""
    endpoint = f"/parks/{park}/activities/"
    if status:
        endpoint += f"?status={status}"
    
    acts = api_get(endpoint)
    
    if acts is None:
        return
    
    if not acts:
        click.echo(click.style("暂无活动数据", fg="yellow"))
        return
    
    status_map = {
        'pending': ('待审批', 'yellow'),
        'approved': ('已批准', 'green'),
        'rejected': ('已拒绝', 'red'),
        'completed': ('已完成', 'bright_black')
    }
    
    click.echo(click.style(f"\n=== 公园 {park} 活动列表 ===", fg="cyan", bold=True))
    click.echo(f"{'编号':<15} {'名称':<20} {'面积':<8} {'安保':<6} {'状态'}")
    click.echo(click.style("-" * 65, fg="bright_black"))
    
    for act in acts:
        status_text, status_color = status_map.get(act['status'], ('未知', 'white'))
        security = click.style('是', fg='red', bold=True) if act['requires_security'] else '否'
        status_display = click.style(status_text, fg=status_color)
        
        click.echo(f"{act['activity_code']:<15} {act['name']:<20} {act['area']:<8} {security:<6} {status_display}")


@activities.command("security")
def security_required():
    """查看需要安保的活动"""
    acts = api_get("/scheduler/activities/requires-security")
    
    if acts is None:
        return
    
    if not acts:
        click.echo(click.style("暂无需要安保的活动", fg="green"))
        return
    
    click.echo(click.style("\n=== 需要安保的活动（面积>500㎡） ===", fg="red", bold=True))
    click.echo(f"共 {len(acts)} 个活动需要额外安保\n")
    
    for act in acts:
        start_time = act.get('start_time', '')
        if start_time:
            try:
                dt = datetime.fromisoformat(start_time.replace('Z', '+00:00'))
                start_time = dt.strftime('%Y-%m-%d %H:%M')
            except:
                pass
        
        click.echo(f"活动编号: {act['activity_code']}")
        click.echo(f"活动名称: {act['name']}")
        click.echo(f"使用面积: {click.style(str(act['area']) + ' ㎡', fg='yellow')}")
        click.echo(f"开始时间: {start_time}")
        click.echo(f"状态: {act['status']}")
        click.echo(click.style("-" * 50, fg="bright_black"))


@cli.command()
@click.option("--park", "-p", help="指定公园编号")
def overview(park):
    """系统概览"""
    click.echo(click.style("\n" + "="*60, fg="cyan", bold=True))
    click.echo(click.style("  公园综合管理系统 - 概览", fg="cyan", bold=True))
    click.echo(click.style("="*60 + "\n", fg="cyan", bold=True))
    
    if park:
        facilities_data = api_get(f"/parks/{park}/facilities/")
        activities_data = api_get(f"/parks/{park}/activities/")
        
        if facilities_data:
            damaged = sum(1 for f in facilities_data if f['status'] in ['damaged', 'repairing'])
            click.echo(f"公园设施: {len(facilities_data)} 个 (损坏/维修中: {damaged})")
        
        if activities_data:
            pending = sum(1 for a in activities_data if a['status'] == 'pending')
            click.echo(f"活动申请: {len(activities_data)} 个 (待审批: {pending})")
    else:
        urgent = api_get("/scheduler/reports/urgent")
        overdue = api_get("/scheduler/overdue")
        security = api_get("/scheduler/activities/requires-security")
        
        if urgent is not None:
            click.echo(f"紧急设施报告: {len(urgent)} 个 (需24小时内处理)")
        if overdue is not None:
            click.echo(f"逾期养护任务: {len(overdue)} 个")
        if security is not None:
            click.echo(f"需要安保的活动: {len(security)} 个")
    
    click.echo()


if __name__ == "__main__":
    cli()
