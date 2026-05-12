#!/usr/bin/env python3
import click
import requests
import os
from datetime import date, datetime
from typing import Optional

API_HOST = os.getenv("API_HOST", "localhost")
API_PORT = int(os.getenv("API_PORT", "8000"))
BASE_URL = f"http://{API_HOST}:{API_PORT}"


def get_endpoint(path: str) -> str:
    return f"{BASE_URL}{path}"


@click.group()
def cli():
    """动物园综合管理系统命令行客户端"""
    pass


@cli.group()
def animals():
    """动物信息管理"""
    pass


@animals.command("list")
@click.option("--zone-id", type=int, help="按区域ID筛选")
@click.option("--species-id", type=int, help="按物种ID筛选")
@click.option("--status", help="按健康状态筛选 (healthy/sick/isolation/dead/pregnant)")
def list_animals(zone_id, species_id, status):
    """列出所有动物"""
    params = {}
    if zone_id:
        params["zone_id"] = zone_id
    if species_id:
        params["species_id"] = species_id
    if status:
        params["health_status"] = status
    
    try:
        response = requests.get(get_endpoint("/api/animals"), params=params)
        response.raise_for_status()
        animals = response.json()
        
        if not animals:
            click.echo("没有找到动物。")
            return
        
        click.echo("=" * 80)
        click.echo(f"{'ID':<5} {'名称':<15} {'物种ID':<10} {'区域ID':<10} {'体重(kg)':<10} {'状态':<15}")
        click.echo("=" * 80)
        
        for animal in animals:
            click.echo(
                f"{animal['id']:<5} "
                f"{animal['name']:<15} "
                f"{animal['species_id']:<10} "
                f"{animal['zone_id']:<10} "
                f"{animal['weight']:<10} "
                f"{animal['health_status']:<15}"
            )
        click.echo("=" * 80)
        click.echo(f"总计: {len(animals)} 只动物")
        
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@animals.command("show")
@click.argument("animal_id", type=int)
def show_animal(animal_id):
    """显示动物详细信息"""
    try:
        response = requests.get(get_endpoint(f"/api/animals/{animal_id}"))
        response.raise_for_status()
        animal = response.json()
        
        click.echo("=" * 60)
        click.echo(f"动物 ID: {animal['id']}")
        click.echo(f"名称: {animal['name']}")
        click.echo(f"物种 ID: {animal['species_id']}")
        click.echo(f"区域 ID: {animal['zone_id']}")
        click.echo(f"芯片 ID: {animal.get('chip_id', 'N/A')}")
        click.echo(f"性别: {animal.get('gender', 'N/A')}")
        click.echo(f"出生日期: {animal.get('birth_date', 'N/A')}")
        click.echo(f"体重: {animal['weight']} kg")
        click.echo(f"健康状态: {animal['health_status']}")
        click.echo(f"隔离中: {'是' if animal.get('is_isolation') else '否'}")
        if animal.get('isolation_start_date'):
            click.echo(f"隔离开始日期: {animal['isolation_start_date']}")
        if animal.get('notes'):
            click.echo(f"备注: {animal['notes']}")
        click.echo(f"创建时间: {animal['created_at']}")
        click.echo(f"更新时间: {animal['updated_at']}")
        click.echo("=" * 60)
        
    except requests.exceptions.HTTPError as e:
        if e.response.status_code == 404:
            click.echo(f"错误: 未找到 ID 为 {animal_id} 的动物", err=True)
        else:
            click.echo(f"错误: {e}", err=True)
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@cli.group()
def feeding():
    """饲养记录管理"""
    pass


@feeding.command("records")
@click.option("--animal-id", type=int, help="按动物ID筛选")
@click.option("--start-date", help="开始日期 (YYYY-MM-DD)")
@click.option("--end-date", help="结束日期 (YYYY-MM-DD)")
@click.option("--status", help="按进食状态筛选 (normal/refused/partial)")
def list_feeding_records(animal_id, start_date, end_date, status):
    """列出投喂记录"""
    params = {}
    if animal_id:
        params["animal_id"] = animal_id
    if start_date:
        params["start_date"] = start_date
    if end_date:
        params["end_date"] = end_date
    if status:
        params["status"] = status
    
    try:
        response = requests.get(get_endpoint("/api/feeding/records"), params=params)
        response.raise_for_status()
        records = response.json()
        
        if not records:
            click.echo("没有找到投喂记录。")
            return
        
        click.echo("=" * 100)
        click.echo(f"{'ID':<5} {'动物ID':<10} {'饲料ID':<10} {'日期':<12} {'班次':<10} {'计划量':<10} {'实际量':<10} {'状态':<10}")
        click.echo("=" * 100)
        
        for record in records:
            click.echo(
                f"{record['id']:<5} "
                f"{record['animal_id']:<10} "
                f"{record['feed_id']:<10} "
                f"{record['feeding_date']:<12} "
                f"{record['shift']:<10} "
                f"{record['planned_amount']:<10} "
                f"{record['actual_amount']:<10} "
                f"{record['status']:<10}"
            )
        click.echo("=" * 100)
        click.echo(f"总计: {len(records)} 条记录")
        
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@feeding.command("shifts")
@click.option("--date", "shift_date", help="查询日期 (YYYY-MM-DD)，默认今天")
def list_shifts(shift_date):
    """列出班次安排"""
    if not shift_date:
        shift_date = date.today().isoformat()
    
    try:
        response = requests.get(get_endpoint(f"/api/feeding/shifts/coverage"), params={"check_date": shift_date})
        response.raise_for_status()
        coverage = response.json()
        
        click.echo(f"日期: {coverage['date']}")
        click.echo("=" * 80)
        
        if coverage["has_issues"]:
            click.echo("警告: 存在未覆盖的班次!")
            for issue in coverage["issues"]:
                click.echo(f"  - {issue['message']}")
        
        click.echo("\n班次覆盖详情:")
        for zone_id, zone_data in coverage["coverage"].items():
            click.echo(f"\n区域: {zone_data['name']} (ID: {zone_id})")
            for shift, shift_data in zone_data["shifts"].items():
                if shift_data["covered"]:
                    click.echo(f"  {shift:<10}: {shift_data['employee']}")
                else:
                    click.echo(f"  {shift:<10}: 未安排人员")
        
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@feeding.command("feeds")
@click.option("--name", help="按饲料名称筛选")
def list_feeds(name):
    """列出饲料库存"""
    params = {}
    if name:
        params["name"] = name
    
    try:
        response = requests.get(get_endpoint("/api/feeding/feeds"), params=params)
        response.raise_for_status()
        feeds = response.json()
        
        if not feeds:
            click.echo("没有找到饲料。")
            return
        
        click.echo("=" * 100)
        click.echo(f"{'ID':<5} {'名称':<20} {'单位':<10} {'当前库存':<15} {'安全库存':<15} {'状态':<10}")
        click.echo("=" * 100)
        
        for feed in feeds:
            status = "正常" if feed["current_stock"] >= feed["safety_stock"] else "库存不足"
            click.echo(
                f"{feed['id']:<5} "
                f"{feed['name']:<20} "
                f"{feed['unit']:<10} "
                f"{feed['current_stock']:<15} "
                f"{feed['safety_stock']:<15} "
                f"{status:<10}"
            )
        click.echo("=" * 100)
        click.echo(f"总计: {len(feeds)} 种饲料")
        
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@cli.group()
def todos():
    """待办事项管理"""
    pass


@todos.command("list")
@click.option("--pending", is_flag=True, help="只显示待处理的")
@click.option("--type", "todo_type", help="按类型筛选 (purchase/veterinary/report_death/upgrade_treatment/vaccination)")
def list_todos(pending, todo_type):
    """列出待办事项"""
    if pending:
        endpoint = "/api/todos/pending"
        params = {}
    else:
        endpoint = "/api/todos"
        params = {}
        if todo_type:
            params["todo_type"] = todo_type
    
    try:
        response = requests.get(get_endpoint(endpoint), params=params)
        response.raise_for_status()
        todos_list = response.json()
        
        if not todos_list:
            click.echo("没有找到待办事项。")
            return
        
        click.echo("=" * 120)
        click.echo(f"{'ID':<5} {'类型':<20} {'状态':<10} {'截止日期':<12} {'标题':<50}")
        click.echo("=" * 120)
        
        for todo in todos_list:
            click.echo(
                f"{todo['id']:<5} "
                f"{todo['todo_type']:<20} "
                f"{todo['status']:<10} "
                f"{(todo.get('due_date') or 'N/A'):<12} "
                f"{todo['title']:<50}"
            )
        click.echo("=" * 120)
        click.echo(f"总计: {len(todos_list)} 条待办事项")
        
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@todos.command("stats")
def todo_stats():
    """显示待办事项统计"""
    try:
        response = requests.get(get_endpoint("/api/todos/statistics"))
        response.raise_for_status()
        stats = response.json()
        
        click.echo("=" * 60)
        click.echo("待办事项统计")
        click.echo("=" * 60)
        click.echo(f"待处理总数: {stats['total_pending']}")
        click.echo(f"已完成总数: {stats['total_completed']}")
        click.echo("\n按类型统计:")
        for type_name, type_stats in stats["by_type"].items():
            click.echo(f"  {type_name:<25}: {type_stats['pending']} 待处理")
        click.echo("=" * 60)
        
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@todos.command("complete")
@click.argument("todo_id", type=int)
@click.option("--by", default="cli", help="完成者")
def complete_todo(todo_id, by):
    """完成待办事项"""
    try:
        response = requests.post(
            get_endpoint(f"/api/todos/{todo_id}/complete"),
            params={"completed_by": by}
        )
        response.raise_for_status()
        todo = response.json()
        
        click.echo(f"待办事项已完成:")
        click.echo(f"  ID: {todo['id']}")
        click.echo(f"  标题: {todo['title']}")
        click.echo(f"  完成时间: {todo['completed_at']}")
        click.echo(f"  完成者: {todo['completed_by']}")
        
    except requests.exceptions.HTTPError as e:
        if e.response.status_code == 404:
            click.echo(f"错误: 未找到 ID 为 {todo_id} 的待办事项", err=True)
        else:
            click.echo(f"错误: {e}", err=True)
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@cli.command("dashboard")
def dashboard():
    """显示系统概览"""
    try:
        response = requests.get(get_endpoint("/dashboard"))
        response.raise_for_status()
        data = response.json()
        
        click.echo("=" * 60)
        click.echo("动物园综合管理系统 - 系统概览")
        click.echo("=" * 60)
        click.echo(f"动物总数:     {data['total_animals']}")
        click.echo(f"患病动物:     {data['sick_animals']}")
        click.echo(f"隔离中动物:   {data['isolation_animals']}")
        click.echo(f"低库存饲料:   {data['low_stock_feeds']}")
        click.echo(f"今日投喂记录: {data['today_feeding_records']}")
        click.echo(f"待办事项:     {data['pending_todos']}")
        click.echo("=" * 60)
        
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


@cli.command("init-data")
def init_sample_data():
    """初始化示例数据"""
    try:
        employee1 = requests.post(get_endpoint("/api/employees"), json={
            "name": "张饲养员",
            "position": "饲养员",
            "phone": "13800138001"
        }).json()
        
        employee2 = requests.post(get_endpoint("/api/employees"), json={
            "name": "李兽医",
            "position": "兽医",
            "phone": "13800138002"
        }).json()
        
        zone1 = requests.post(get_endpoint("/api/zones"), json={
            "name": "猛兽区",
            "description": "大型猫科动物区域"
        }).json()
        
        zone2 = requests.post(get_endpoint("/api/zones"), json={
            "name": "食草区",
            "description": "食草动物区域"
        }).json()
        
        species = requests.post(get_endpoint("/api/species"), json={
            "name": "东北虎",
            "scientific_name": "Panthera tigris altaica",
            "conservation_level": "endangered",
            "description": "西伯利亚虎"
        }).json()
        
        species2 = requests.post(get_endpoint("/api/species"), json={
            "name": "大熊猫",
            "scientific_name": "Ailuropoda melanoleuca",
            "conservation_level": "vulnerable",
            "description": "中国国宝"
        }).json()
        
        requests.post(get_endpoint("/api/feeding-standards"), json={
            "species_id": species["id"],
            "min_weight": 150,
            "max_weight": 250,
            "feed_type": "牛肉",
            "daily_amount": 8,
            "frequency": 2,
            "is_special": False
        })
        
        requests.post(get_endpoint("/api/feeding-standards"), json={
            "species_id": species2["id"],
            "min_weight": 80,
            "max_weight": 150,
            "feed_type": "新鲜竹子",
            "daily_amount": 12,
            "frequency": 3,
            "is_special": False
        })
        
        requests.post(get_endpoint("/api/feeding-standards"), json={
            "species_id": species["id"],
            "min_weight": 150,
            "max_weight": 250,
            "feed_type": "易消化肉类",
            "daily_amount": 6,
            "frequency": 3,
            "is_special": True,
            "for_health_status": "sick"
        })
        
        feed1 = requests.post(get_endpoint("/api/feeding/feeds"), json={
            "name": "牛肉",
            "unit": "kg",
            "current_stock": 50,
            "safety_stock": 20,
            "supplier": "优质肉类供应商"
        }).json()
        
        feed2 = requests.post(get_endpoint("/api/feeding/feeds"), json={
            "name": "新鲜竹子",
            "unit": "kg",
            "current_stock": 100,
            "safety_stock": 30,
            "supplier": "四川竹子基地"
        }).json()
        
        requests.post(get_endpoint("/api/feeding/feeds"), json={
            "name": "易消化肉类",
            "unit": "kg",
            "current_stock": 15,
            "safety_stock": 10,
            "supplier": "专用饲料供应商"
        })
        
        animal1 = requests.post(get_endpoint("/api/animals"), json={
            "name": "虎王",
            "species_id": species["id"],
            "zone_id": zone1["id"],
            "chip_id": "CHIP001",
            "gender": "male",
            "weight": 200,
            "health_status": "healthy"
        }).json()
        
        animal2 = requests.post(get_endpoint("/api/animals"), json={
            "name": "圆圆",
            "species_id": species2["id"],
            "zone_id": zone2["id"],
            "chip_id": "CHIP002",
            "gender": "female",
            "weight": 110,
            "health_status": "healthy"
        }).json()
        
        from datetime import date, timedelta
        today = date.today().isoformat()
        
        requests.post(get_endpoint("/api/feeding/shifts"), json={
            "zone_id": zone1["id"],
            "employee_id": employee1["id"],
            "shift_date": today,
            "shift": "morning"
        })
        
        click.echo("示例数据初始化完成!")
        click.echo(f"创建的员工: {employee1['name']}, {employee2['name']}")
        click.echo(f"创建的区域: {zone1['name']}, {zone2['name']}")
        click.echo(f"创建的物种: {species['name']}, {species2['name']}")
        click.echo(f"创建的动物: {animal1['name']}, {animal2['name']}")
        click.echo(f"创建的饲料: {feed1['name']}, {feed2['name']}")
        
    except requests.exceptions.RequestException as e:
        click.echo(f"错误: {e}", err=True)


if __name__ == "__main__":
    cli()
