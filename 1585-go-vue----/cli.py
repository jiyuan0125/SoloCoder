#!/usr/bin/env python3
import argparse
import sys
from datetime import datetime
from sqlalchemy.orm import Session

from app.database import SessionLocal, engine
from app.models import Base, VentilationDevice, LightingDevice, DrainageDevice, Alarm


ALARM_COLORS = {
    "low": "\033[94m",
    "medium": "\033[93m",
    "high": "\033[91m",
    "reset": "\033[0m"
}


def get_db():
    db = SessionLocal()
    try:
        return db
    except Exception:
        db.close()
        sys.exit(1)


def format_datetime(dt):
    if dt:
        return dt.strftime("%Y-%m-%d %H:%M:%S")
    return "-"


def show_ventilation_status(args):
    db = get_db()
    try:
        if args.device_id:
            device = db.query(VentilationDevice).filter(VentilationDevice.id == args.device_id).first()
            if not device:
                print(f"设备不存在: {args.device_id}")
                return
            
            print(f"\n{'='*60}")
            print(f"通风设备详情: {device.name} (ID: {device.id})")
            print(f"{'='*60}")
            print(f"状态:     {'运行中' if device.is_running else '已停止'}")
            print(f"模式:     {'自动' if device.mode == 'auto' else '手动'}")
            print(f"故障:     {'是' if device.fault else '否'}")
            print(f"CO浓度:   {device.co_level} ppm")
            print(f"能见度:   {device.visibility} m")
            print(f"启动时间: {format_datetime(device.start_time)}")
            print(f"关延时:   {'激活中' if device.shutdown_delay_active else '未激活'}")
        else:
            devices = db.query(VentilationDevice).all()
            print(f"\n{'='*80}")
            print(f"{'ID':<5}{'名称':<20}{'状态':<10}{'模式':<10}{'故障':<8}{'CO(ppm)':<10}{'能见度(m)':<12}{'启动时间':<20}")
            print(f"{'='*80}")
            for device in devices:
                status = "运行中" if device.is_running else "已停止"
                mode = "自动" if device.mode == "auto" else "手动"
                fault = "是" if device.fault else "否"
                print(f"{device.id:<5}{device.name:<20}{status:<10}{mode:<10}{fault:<8}{device.co_level:<10.1f}{device.visibility:<12.1f}{format_datetime(device.start_time):<20}")
    finally:
        db.close()


def show_lighting_status(args):
    db = get_db()
    try:
        if args.device_id:
            device = db.query(LightingDevice).filter(LightingDevice.id == args.device_id).first()
            if not device:
                print(f"设备不存在: {args.device_id}")
                return
            
            print(f"\n{'='*60}")
            print(f"照明设备详情: {device.name} (ID: {device.id})")
            print(f"{'='*60}")
            print(f"状态:     {'运行中' if device.status == 'running' else '已停止'}")
            print(f"模式:     {'自动' if device.mode == 'auto' else '手动'}")
            print(f"故障:     {'是' if device.fault else '否'}")
            print(f"亮度:     {device.brightness}%")
            print(f"目标亮度: {device.target_brightness}%")
            print(f"分组:     {device.group}")
        else:
            devices = db.query(LightingDevice).all()
            print(f"\n{'='*70}")
            print(f"{'ID':<5}{'名称':<20}{'状态':<10}{'模式':<10}{'故障':<8}{'亮度(%)':<10}{'分组':<15}")
            print(f"{'='*70}")
            for device in devices:
                status = "运行中" if device.status == "running" else "已停止"
                mode = "自动" if device.mode == "auto" else "手动"
                fault = "是" if device.fault else "否"
                print(f"{device.id:<5}{device.name:<20}{status:<10}{mode:<10}{fault:<8}{device.brightness:<10}{device.group:<15}")
    finally:
        db.close()


def show_drainage_status(args):
    db = get_db()
    try:
        if args.device_id:
            device = db.query(DrainageDevice).filter(DrainageDevice.id == args.device_id).first()
            if not device:
                print(f"设备不存在: {args.device_id}")
                return
            
            print(f"\n{'='*60}")
            print(f"排水设备详情: {device.name} (ID: {device.id})")
            print(f"{'='*60}")
            print(f"状态:     {'运行中' if device.is_running else '已停止'}")
            print(f"模式:     {'自动' if device.mode == 'auto' else '手动'}")
            print(f"故障:     {'是' if device.fault else '否'}")
            print(f"型号:     {device.model}")
            print(f"水位:     {device.water_level} m")
            print(f"警戒水位: {device.warning_level} m")
            print(f"安全水位: {device.safe_level} m")
            print(f"启动时间: {format_datetime(device.start_time)}")
            print(f"累计运行: {device.cumulative_runtime:.2f} 小时")
            print(f"保养周期: {device.maintenance_interval_hours} 小时")
        else:
            devices = db.query(DrainageDevice).all()
            print(f"\n{'='*85}")
            print(f"{'ID':<5}{'名称':<20}{'状态':<10}{'模式':<10}{'故障':<8}{'水位(m)':<10}{'累计运行(h)':<15}{'启动时间':<20}")
            print(f"{'='*85}")
            for device in devices:
                status = "运行中" if device.is_running else "已停止"
                mode = "自动" if device.mode == "auto" else "手动"
                fault = "是" if device.fault else "否"
                print(f"{device.id:<5}{device.name:<20}{status:<10}{mode:<10}{fault:<8}{device.water_level:<10.1f}{device.cumulative_runtime:<15.2f}{format_datetime(device.start_time):<20}")
    finally:
        db.close()


def show_alarms(args):
    db = get_db()
    try:
        query = db.query(Alarm)
        
        if args.active:
            query = query.filter(Alarm.status.in_(['active', 'acknowledged']))
        if args.subsystem:
            query = query.filter(Alarm.subsystem == args.subsystem)
        if args.level:
            query = query.filter(Alarm.level == args.level)
        
        alarms = query.order_by(Alarm.created_at.desc()).limit(args.limit).all()
        
        print(f"\n{'='*100}")
        print(f"{'ID':<5}{'子系统':<12}{'设备':<15}{'级别':<10}{'状态':<15}{'创建时间':<20}{'消息':<30}")
        print(f"{'='*100}")
        
        for alarm in alarms:
            level_color = ALARM_COLORS.get(alarm.level, ALARM_COLORS['reset'])
            level_display = f"{level_color}{alarm.level}{ALARM_COLORS['reset']}"
            status_display = {
                'active': '活动',
                'acknowledged': '已确认',
                'resolved': '已解决'
            }.get(alarm.status, alarm.status)
            
            message_truncated = alarm.message[:30]
            if len(alarm.message) > 30:
                message_truncated += "..."
            
            print(f"{alarm.id:<5}{alarm.subsystem:<12}{alarm.device_name:<15}{level_display:<10}{status_display:<15}{format_datetime(alarm.created_at):<20}{message_truncated:<30}")
        
        print(f"\n级别颜色说明:")
        print(f"  {ALARM_COLORS['low']}low (蓝色){ALARM_COLORS['reset']}: 低级别告警")
        print(f"  {ALARM_COLORS['medium']}medium (黄色){ALARM_COLORS['reset']}: 中级别告警")
        print(f"  {ALARM_COLORS['high']}high (红色){ALARM_COLORS['reset']}: 高级别告警")
    finally:
        db.close()


def show_all_status(args):
    print("\n" + "#"*60)
    print("# 通风系统")
    print("#"*60)
    show_ventilation_status(args)
    
    print("\n" + "#"*60)
    print("# 照明系统")
    print("#"*60)
    show_lighting_status(args)
    
    print("\n" + "#"*60)
    print("# 排水系统")
    print("#"*60)
    show_drainage_status(args)
    
    print("\n" + "#"*60)
    print("# 活动告警")
    print("#"*60)
    args.active = True
    show_alarms(args)


def main():
    parser = argparse.ArgumentParser(description='隧道综合监控系统 - 命令行客户端')
    subparsers = parser.add_subparsers(dest='command', help='可用命令')
    
    vent_parser = subparsers.add_parser('ventilation', help='查看通风系统状态')
    vent_parser.add_argument('--device-id', type=int, help='指定设备ID查看详情')
    vent_parser.set_defaults(func=show_ventilation_status)
    
    light_parser = subparsers.add_parser('lighting', help='查看照明系统状态')
    light_parser.add_argument('--device-id', type=int, help='指定设备ID查看详情')
    light_parser.set_defaults(func=show_lighting_status)
    
    drain_parser = subparsers.add_parser('drainage', help='查看排水系统状态')
    drain_parser.add_argument('--device-id', type=int, help='指定设备ID查看详情')
    drain_parser.set_defaults(func=show_drainage_status)
    
    alarm_parser = subparsers.add_parser('alarms', help='查看告警')
    alarm_parser.add_argument('--active', action='store_true', help='只显示活动告警')
    alarm_parser.add_argument('--subsystem', choices=['ventilation', 'lighting', 'drainage'], help='按子系统过滤')
    alarm_parser.add_argument('--level', choices=['low', 'medium', 'high'], help='按级别过滤')
    alarm_parser.add_argument('--limit', type=int, default=20, help='显示数量限制')
    alarm_parser.set_defaults(func=show_alarms)
    
    all_parser = subparsers.add_parser('all', help='查看所有系统状态')
    all_parser.add_argument('--device-id', type=int, help='指定设备ID（对通风、照明、排水有效）')
    all_parser.set_defaults(func=show_all_status)
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    Base.metadata.create_all(bind=engine)
    args.func(args)


if __name__ == '__main__':
    main()
