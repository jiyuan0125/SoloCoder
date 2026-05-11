import os
import sys
from datetime import date
from pathlib import Path
from typing import Optional

import click

BASE_DIR = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(BASE_DIR))

from src.client.api_client import APIClient


def get_client():
    return APIClient()


def print_dict(d, indent=0):
    prefix = "  " * indent
    for key, value in d.items():
        if isinstance(value, dict):
            click.echo(f"{prefix}{key}:")
            print_dict(value, indent + 1)
        elif isinstance(value, list):
            click.echo(f"{prefix}{key}:")
            for i, item in enumerate(value):
                if isinstance(item, dict):
                    click.echo(f"{prefix}  [{i}]:")
                    print_dict(item, indent + 2)
                else:
                    click.echo(f"{prefix}  [{i}]: {item}")
        else:
            click.echo(f"{prefix}{key}: {value}")


@click.group()
def cli():
    pass


@click.group()
def pipeline():
    pass


cli.add_command(pipeline)


@pipeline.command("create")
@click.option("--name", required=True, help="管道名称")
@click.option("--code", required=True, help="管道编码")
@click.option("--length", required=True, type=float, help="总长度(米)")
@click.option("--wall-thickness", required=True, type=float, help="设计壁厚(mm)")
@click.option("--operating-pressure", required=True, type=float, help="工作压力(MPa)")
@click.option("--material", help="材质")
@click.option("--diameter", type=float, help="直径(mm)")
@click.option("--risk-level", type=click.Choice(["low", "medium", "high", "critical"]), default="medium", help="风险等级")
@click.option("--description", help="描述")
def create_pipeline(name, code, length, wall_thickness, operating_pressure, material, diameter, risk_level, description):
    data = {
        "name": name,
        "code": code,
        "total_length": length,
        "design_wall_thickness": wall_thickness,
        "operating_pressure": operating_pressure,
        "material": material,
        "diameter": diameter,
        "risk_level": risk_level,
        "description": description
    }
    with get_client() as client:
        result = client.create_pipeline(data)
        click.echo("管道创建成功:")
        print_dict(result)


@pipeline.command("list")
@click.option("--status", help="按状态筛选")
def list_pipelines(status):
    with get_client() as client:
        pipelines = client.list_pipelines(status=status)
        for p in pipelines:
            click.echo(f"ID: {p['id']}, 编码: {p['code']}, 名称: {p['name']}, 状态: {p['status']}, 风险: {p['risk_level']}")


@pipeline.command("get")
@click.argument("pipeline_id", type=int)
def get_pipeline(pipeline_id):
    with get_client() as client:
        pipeline = client.get_pipeline(pipeline_id)
        print_dict(pipeline)


@pipeline.command("delete")
@click.argument("pipeline_id", type=int)
@click.confirmation_option(prompt="确定要删除这条管道吗?")
def delete_pipeline(pipeline_id):
    with get_client() as client:
        result = client.delete_pipeline(pipeline_id)
        click.echo(result["message"])


@click.group()
def segment():
    pass


cli.add_command(segment)


@segment.command("create")
@click.option("--pipeline-id", required=True, type=int, help="所属管道ID")
@click.option("--name", required=True, help="管段名称")
@click.option("--start-pos", required=True, type=float, help="起始位置(m)")
@click.option("--end-pos", required=True, type=float, help="结束位置(m)")
@click.option("--length", required=True, type=float, help="长度(m)")
@click.option("--risk-level", type=click.Choice(["low", "medium", "high", "critical"]), default="medium", help="风险等级")
@click.option("--description", help="描述")
def create_segment(pipeline_id, name, start_pos, end_pos, length, risk_level, description):
    data = {
        "pipeline_id": pipeline_id,
        "name": name,
        "start_position": start_pos,
        "end_position": end_pos,
        "length": length,
        "risk_level": risk_level,
        "status": "operational",
        "description": description
    }
    with get_client() as client:
        result = client.create_segment(data)
        click.echo("管段创建成功:")
        click.echo(f"  巡检频率: 每{result['patrol_frequency_days']}天")
        print_dict(result)


@segment.command("list")
@click.option("--pipeline-id", type=int, help="按管道ID筛选")
def list_segments(pipeline_id):
    with get_client() as client:
        segments = client.list_segments(pipeline_id=pipeline_id)
        for s in segments:
            click.echo(f"ID: {s['id']}, 名称: {s['name']}, 管道ID: {s['pipeline_id']}, 风险: {s['risk_level']}, 巡检频率: {s['patrol_frequency_days']}天")


@click.group()
def pressure():
    pass


cli.add_command(pressure)


@pressure.command("create-point")
@click.option("--pipeline-id", required=True, type=int, help="所属管道ID")
@click.option("--name", required=True, help="监测点名称")
@click.option("--position", required=True, type=float, help="位置(m)")
def create_pressure_point(pipeline_id, name, position):
    data = {
        "pipeline_id": pipeline_id,
        "name": name,
        "position": position
    }
    with get_client() as client:
        result = client.create_pressure_point(data)
        click.echo("压力监测点创建成功:")
        print_dict(result)


@pressure.command("list-points")
@click.option("--pipeline-id", required=True, type=int, help="管道ID")
def list_pressure_points(pipeline_id):
    with get_client() as client:
        points = client.list_pressure_points(pipeline_id)
        for p in points:
            click.echo(f"ID: {p['id']}, 名称: {p['name']}, 位置: {p['position']}m")


@pressure.command("add-reading")
@click.option("--point-id", required=True, type=int, help="压力监测点ID")
@click.option("--pressure", required=True, type=float, help="压力值(MPa)")
def add_reading(point_id, pressure):
    with get_client() as client:
        result = client.add_pressure_reading(point_id, pressure)
        click.echo("压力读数添加成功:")
        print_dict(result)


@pressure.command("stats")
@click.option("--pipeline-id", required=True, type=int, help="管道ID")
def get_stats(pipeline_id):
    with get_client() as client:
        stats = client.get_pressure_stats(pipeline_id)
        click.echo(f"压力差统计 (基于{stats['count']}条数据):")
        click.echo(f"  均值: {stats['mean']:.4f}")
        click.echo(f"  标准差: {stats['stddev']:.4f}")


@pressure.command("check-leak")
@click.option("--start-point", required=True, type=int, help="起始监测点ID")
@click.option("--end-point", required=True, type=int, help="结束监测点ID")
def check_leak(start_point, end_point):
    with get_client() as client:
        result = client.check_leak(start_point, end_point)
        click.echo("泄漏检测结果:")
        click.echo(f"  疑似泄漏: {'是' if result['is_suspected'] else '否'}")
        click.echo(f"  压力差: {result['pressure_diff']:.4f}")
        click.echo(f"  历史均值: {result['mean']:.4f}")
        click.echo(f"  历史标准差: {result['stddev']:.4f}")
        click.echo(f"  阈值(均值+3σ): {result['threshold']:.4f}")
        click.echo(f"  历史数据量: {result['data_count']}")


@click.group()
def alarm():
    pass


cli.add_command(alarm)


@alarm.command("list")
@click.option("--status", help="按状态筛选")
@click.option("--severity", help="按严重程度筛选")
def list_alarms(status, severity):
    with get_client() as client:
        alarms = client.list_alarms(status=status, severity=severity)
        for a in alarms:
            click.echo(f"ID: {a['id']}, 类型: {a['alarm_type']}, 严重度: {a['severity']}, 状态: {a['status']}, 标题: {a['title']}")


@alarm.command("get")
@click.argument("alarm_id", type=int)
def get_alarm(alarm_id):
    with get_client() as client:
        alarm = client.get_alarm(alarm_id)
        print_dict(alarm)


@alarm.command("update-status")
@click.argument("alarm_id", type=int)
@click.option("--status", required=True, type=click.Choice(["new", "in_process", "urgent", "resolved", "false_alarm"]), help="状态")
@click.option("--handled-by", help="处理人")
@click.option("--false-reason", help="误报原因(误报时必填)")
def update_alarm_status(alarm_id, status, handled_by, false_reason):
    with get_client() as client:
        try:
            result = client.update_alarm_status(
                alarm_id,
                status,
                handled_by=handled_by,
                false_alarm_reason=false_reason
            )
            click.echo("报警状态更新成功:")
            print_dict(result)
        except Exception as e:
            click.echo(f"错误: {e}")


@alarm.command("escalate")
def escalate():
    with get_client() as client:
        result = client.escalate_alarms()
        click.echo(result["message"])


@alarm.command("remind")
def remind():
    with get_client() as client:
        result = client.send_reminders()
        click.echo(result["message"])


@click.group()
def patrol():
    pass


cli.add_command(patrol)


@patrol.command("generate-plan")
def generate_plan():
    with get_client() as client:
        result = client.generate_patrol_plan()
        click.echo(f"{result['message']}")


@patrol.command("list-plans")
@click.option("--date", "plan_date", type=click.DateTime(formats=["%Y-%m-%d"]), help="按日期筛选(YYYY-MM-DD)")
def list_plans(plan_date):
    with get_client() as client:
        plans = client.list_patrol_plans(plan_date=plan_date.date() if plan_date else None)
        for p in plans:
            click.echo(f"ID: {p['id']}, 管道ID: {p['pipeline_id']}, 日期: {p['plan_date']}, 状态: {p['status']}")
            for item in p.get("plan_items", []):
                click.echo(f"  - 管段ID: {item['segment_id']}, 计划时间: {item['scheduled_time']}")


@patrol.command("add-record")
@click.option("--segment-id", required=True, type=int, help="管段ID")
@click.option("--date", "patrol_date", required=True, type=click.DateTime(formats=["%Y-%m-%d"]), help="巡检日期(YYYY-MM-DD)")
@click.option("--inspector", help="巡检员")
@click.option("--findings", help="发现问题")
@click.option("--status", type=click.Choice(["normal", "abnormal"]), default="normal", help="状态")
def add_record(segment_id, patrol_date, inspector, findings, status):
    data = {
        "segment_id": segment_id,
        "patrol_date": patrol_date.date(),
        "inspector": inspector,
        "findings": findings,
        "status": status
    }
    with get_client() as client:
        result = client.add_patrol_record(data)
        click.echo("巡检记录添加成功:")
        print_dict(result)


@patrol.command("list-records")
@click.option("--segment-id", type=int, help="按管段筛选")
def list_records(segment_id):
    with get_client() as client:
        records = client.list_patrol_records(segment_id=segment_id)
        for r in records:
            click.echo(f"ID: {r['id']}, 管段ID: {r['segment_id']}, 日期: {r['patrol_date']}, 状态: {r['status']}, 巡检员: {r['inspector']}")


@click.group()
def integrity():
    pass


cli.add_command(integrity)


@integrity.command("add-record")
@click.option("--pipeline-id", required=True, type=int, help="管道ID")
@click.option("--segment-id", type=int, help="管段ID")
@click.option("--inspection-date", required=True, type=click.DateTime(formats=["%Y-%m-%d"]), help="检测日期(YYYY-MM-DD)")
@click.option("--wall-thickness", required=True, type=float, help="壁厚(mm)")
@click.option("--location", help="检测位置")
@click.option("--notes", help="备注")
def add_integrity_record(pipeline_id, segment_id, inspection_date, wall_thickness, location, notes):
    data = {
        "pipeline_id": pipeline_id,
        "segment_id": segment_id,
        "inspection_date": inspection_date.date(),
        "wall_thickness": wall_thickness,
        "location": location,
        "notes": notes
    }
    with get_client() as client:
        result = client.add_integrity_record(data)
        click.echo("完整性记录添加成功:")
        print_dict(result)


@integrity.command("list-records")
@click.option("--pipeline-id", type=int, help="按管道筛选")
def list_integrity_records(pipeline_id):
    with get_client() as client:
        records = client.list_integrity_records(pipeline_id=pipeline_id)
        for r in records:
            click.echo(f"ID: {r['id']}, 管道ID: {r['pipeline_id']}, 检测日期: {r['inspection_date']}, 壁厚: {r['wall_thickness']}mm")


@integrity.command("check-wall")
@click.argument("record_id", type=int)
def check_wall(record_id):
    with get_client() as client:
        result = client.check_wall_thickness(record_id)
        click.echo("壁厚检查结果:")
        click.echo(f"  需要维修: {'是' if result['needs_maintenance'] else '否'}")
        if result.get("design_thickness"):
            click.echo(f"  设计壁厚: {result['design_thickness']:.2f}mm")
            click.echo(f"  阈值(80%): {result['threshold']:.2f}mm")
            click.echo(f"  当前壁厚: {result['current_thickness']:.2f}mm")
            click.echo(f"  比例: {result['ratio']:.1%}")


@integrity.command("list-maintenance")
@click.option("--status", help="按状态筛选")
def list_maintenance(status):
    with get_client() as client:
        tasks = client.list_maintenance_tasks(status=status)
        for t in tasks:
            click.echo(f"ID: {t['id']}, 类型: {t['task_type']}, 优先级: {t['priority']}, 状态: {t['status']}, 标题: {t['title']}")


@integrity.command("update-maintenance")
@click.argument("task_id", type=int)
@click.option("--status", type=click.Choice(["pending", "in_progress", "completed"]), help="状态")
@click.option("--priority", type=click.Choice(["low", "medium", "high"]), help="优先级")
def update_maintenance(task_id, status, priority):
    data = {}
    if status:
        data["status"] = status
    if priority:
        data["priority"] = priority
    with get_client() as client:
        result = client.update_maintenance_task(task_id, data)
        click.echo("维修任务更新成功:")
        print_dict(result)


@click.group()
def report():
    pass


cli.add_command(report)


@report.command("generate")
@click.option("--date", "report_date", type=click.DateTime(formats=["%Y-%m-%d"]), help="报告日期(YYYY-MM-DD)，默认今天")
def generate_report(report_date):
    with get_client() as client:
        result = client.generate_report(report_date=report_date.date() if report_date else None)
        click.echo(f"{result['message']}, 日期: {result['date']}")


@report.command("list")
def list_reports():
    with get_client() as client:
        reports = client.list_reports()
        for r in reports:
            click.echo(f"日期: {r['report_date']}, 管道数: {r['total_pipelines']}, 报警数: {r['total_alarms']}")


@report.command("view")
@click.option("--date", "report_date", required=True, type=click.DateTime(formats=["%Y-%m-%d"]), help="报告日期(YYYY-MM-DD)")
def view_report(report_date):
    with get_client() as client:
        report = client.get_report(report_date.date())
        click.echo("=" * 60)
        click.echo(report["content"])
        click.echo("=" * 60)


@cli.command("health")
def health():
    with get_client() as client:
        result = client.health()
        click.echo(f"服务状态: {result['status']}")


if __name__ == "__main__":
    cli()
