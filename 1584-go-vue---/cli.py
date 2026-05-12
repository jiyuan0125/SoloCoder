#!/usr/bin/env python3
import argparse
import sys
from datetime import datetime
from sqlalchemy.orm import Session
from app.database import SessionLocal
from app.models import Task, Line, Staff, TaskStaff, TaskStatus, MaintenanceWindow
from app.services import check_all_conflicts


def format_datetime(dt):
    if not dt:
        return "N/A"
    return dt.strftime("%Y-%m-%d %H:%M:%S")


def list_tasks(status_filter=None):
    db = SessionLocal()
    try:
        query = db.query(Task)
        if status_filter:
            try:
                status_enum = TaskStatus[status_filter.upper()]
                query = query.filter(Task.status == status_enum)
            except KeyError:
                print(f"错误: 无效的状态 '{status_filter}'")
                print(f"有效状态: {', '.join([s.value for s in TaskStatus])}")
                return
        tasks = query.order_by(Task.scheduled_start.desc()).all()
        
        if not tasks:
            print("没有找到作业")
            return
        
        print("\n" + "="*120)
        print(f"{'ID':<5} {'标题':<25} {'状态':<12} {'线路ID':<8} {'区间':<12} {'开始时间':<20} {'结束时间':<20}")
        print("="*120)
        
        for task in tasks:
            line = db.query(Line).filter(Line.id == task.line_id).first()
            line_name = line.name if line else str(task.line_id)
            print(f"{task.id:<5} {task.title[:23]:<25} {task.status.value:<12} {line_name[:6]:<8} "
                  f"KP{task.start_kp}-{task.end_kp:<8} "
                  f"{format_datetime(task.scheduled_start):<20} {format_datetime(task.scheduled_end):<20}")
        
        print("="*120)
        print(f"共 {len(tasks)} 个作业\n")
    finally:
        db.close()


def show_task(task_id):
    db = SessionLocal()
    try:
        task = db.query(Task).filter(Task.id == task_id).first()
        if not task:
            print(f"错误: 作业 {task_id} 不存在")
            return
        
        line = db.query(Line).filter(Line.id == task.line_id).first()
        resp_person = db.query(Staff).filter(Staff.id == task.responsible_person_id).first()
        
        print("\n" + "="*60)
        print(f"作业详情 - ID: {task.id}")
        print("="*60)
        print(f"标题: {task.title}")
        if task.description:
            print(f"描述: {task.description}")
        print(f"线路: {line.name if line else '未知'}")
        print(f"区间: KP {task.start_kp} - KP {task.end_kp}")
        print(f"计划开始: {format_datetime(task.scheduled_start)}")
        print(f"计划结束: {format_datetime(task.scheduled_end)}")
        print(f"状态: {task.status.value}")
        print(f"负责人: {resp_person.name if resp_person else '未知'} (工号: {resp_person.employee_id if resp_person else '未知'})")
        print(f"创建时间: {format_datetime(task.created_at)}")
        print(f"提交时间: {format_datetime(task.submitted_at)}")
        print(f"审批时间: {format_datetime(task.approved_at)}")
        print(f"开始时间: {format_datetime(task.started_at)}")
        print(f"完成时间: {format_datetime(task.completed_at)}")
        print(f"超时时间: {format_datetime(task.timeout_at)}")
        print(f"资质已核验: {'是' if task.qualification_checked else '否'}")
        
        assigned_staff = db.query(TaskStaff).filter(TaskStaff.task_id == task.id).all()
        if assigned_staff:
            print("\n参与人员:")
            for ts in assigned_staff:
                staff = db.query(Staff).filter(Staff.id == ts.staff_id).first()
                if staff:
                    print(f"  - {staff.name} (工号: {staff.employee_id})")
        
        print("="*60 + "\n")
    finally:
        db.close()


def check_conflicts(task_id):
    db = SessionLocal()
    try:
        task = db.query(Task).filter(Task.id == task_id).first()
        if not task:
            print(f"错误: 作业 {task_id} 不存在")
            return
        
        assigned_staff_ids = [
            ts.staff_id for ts in 
            db.query(TaskStaff).filter(TaskStaff.task_id == task.id).all()
        ]
        
        conflicts = check_all_conflicts(db, task, assigned_staff_ids, exclude_task_id=task.id)
        
        print("\n" + "="*60)
        print(f"冲突检测 - 作业 ID: {task_id} ({task.title})")
        print("="*60)
        
        if not conflicts:
            print("✅ 未检测到冲突")
        else:
            print(f"❌ 检测到 {len(conflicts)} 个冲突:")
            for i, conflict in enumerate(conflicts, 1):
                print(f"\n冲突 {i}:")
                print(f"  类型: {'安全冲突' if conflict.conflict_type == 'safety' else '人员冲突'}")
                print(f"  冲突作业: #{conflict.task_id} - {conflict.task_title}")
                print(f"  原因: {conflict.reason}")
        
        print("="*60 + "\n")
    finally:
        db.close()


def list_lines():
    db = SessionLocal()
    try:
        lines = db.query(Line).all()
        if not lines:
            print("没有找到线路")
            return
        
        print("\n" + "="*80)
        print(f"{'ID':<5} {'名称':<20} {'描述':<50}")
        print("="*80)
        for line in lines:
            desc = (line.description[:47] + '...') if line.description and len(line.description) > 50 else (line.description or '')
            print(f"{line.id:<5} {line.name:<20} {desc:<50}")
        print("="*80)
        print(f"共 {len(lines)} 条线路\n")
    finally:
        db.close()


def list_windows(line_id):
    db = SessionLocal()
    try:
        line = db.query(Line).filter(Line.id == line_id).first()
        if not line:
            print(f"错误: 线路 {line_id} 不存在")
            return
        
        windows = db.query(MaintenanceWindow).filter(
            MaintenanceWindow.line_id == line_id
        ).all()
        
        if not windows:
            print(f"线路 {line.name} 没有配置天窗时间")
            return
        
        print("\n" + "="*80)
        print(f"线路 {line.name} 的天窗时间配置")
        print("="*80)
        print(f"{'ID':<5} {'类型':<10} {'日期':<12} {'开始时间':<10} {'结束时间':<10}")
        print("="*80)
        
        for window in windows:
            wtype = "通用" if window.is_general else "特定日期"
            date_str = str(window.date) if window.date else "N/A"
            print(f"{window.id:<5} {wtype:<10} {date_str:<12} {str(window.start_time):<10} {str(window.end_time):<10}")
        
        print("="*80 + "\n")
    finally:
        db.close()


def list_staff():
    db = SessionLocal()
    try:
        staff_list = db.query(Staff).all()
        if not staff_list:
            print("没有找到人员")
            return
        
        print("\n" + "="*80)
        print(f"{'ID':<5} {'姓名':<15} {'工号':<15} {'入职日期':<12} {'部门':<20}")
        print("="*80)
        for staff in staff_list:
            join_date = str(staff.join_date) if staff.join_date else "N/A"
            dept = staff.department or "N/A"
            print(f"{staff.id:<5} {staff.name:<15} {staff.employee_id:<15} {join_date:<12} {dept:<20}")
        print("="*80)
        print(f"共 {len(staff_list)} 人\n")
    finally:
        db.close()


def main():
    parser = argparse.ArgumentParser(description='天窗作业调度管理系统 CLI')
    subparsers = parser.add_subparsers(dest='command', help='可用命令')
    
    list_parser = subparsers.add_parser('list', help='列出资源')
    list_parser.add_argument('resource', choices=['tasks', 'lines', 'staff', 'windows'], help='资源类型')
    list_parser.add_argument('--status', help='过滤作业状态')
    list_parser.add_argument('--line-id', type=int, help='线路ID（用于查询天窗时间）')
    
    show_parser = subparsers.add_parser('show', help='显示详情')
    show_parser.add_argument('resource', choices=['task'], help='资源类型')
    show_parser.add_argument('id', type=int, help='资源ID')
    
    conflict_parser = subparsers.add_parser('check-conflicts', help='检测作业冲突')
    conflict_parser.add_argument('task_id', type=int, help='作业ID')
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(0)
    
    if args.command == 'list':
        if args.resource == 'tasks':
            list_tasks(args.status)
        elif args.resource == 'lines':
            list_lines()
        elif args.resource == 'staff':
            list_staff()
        elif args.resource == 'windows':
            if not args.line_id:
                print("错误: 查询天窗时间需要指定 --line-id")
                sys.exit(1)
            list_windows(args.line_id)
    
    elif args.command == 'show':
        if args.resource == 'task':
            show_task(args.id)
    
    elif args.command == 'check-conflicts':
        check_conflicts(args.task_id)


if __name__ == '__main__':
    main()
