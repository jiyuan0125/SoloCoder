#!/usr/bin/env python3
import argparse
import sys
import os
from dotenv import load_dotenv
from datetime import datetime, timedelta
from sqlalchemy.orm import Session

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from app.database import SessionLocal
from app.models import (
    Schedule, ScheduleStatus, ArrivalRecord, Route,
    Direction, Vehicle, Station, RouteStation,
    OperationParameter
)
from app.scheduler import get_punctuality_report


def print_table(headers, rows):
    if not rows:
        print("暂无数据")
        return
    
    col_widths = []
    for header in headers:
        col_widths.append(len(str(header)))
    
    for row in rows:
        for i, cell in enumerate(row):
            if len(str(cell)) > col_widths[i]:
                col_widths[i] = len(str(cell))
    
    total_width = sum(col_widths) + 3 * len(col_widths) + 1
    print("-" * total_width)
    
    header_str = "|"
    for i, header in enumerate(headers):
        header_str += f" {str(header):^{col_widths[i]}} |"
    print(header_str)
    print("-" * total_width)
    
    for row in rows:
        row_str = "|"
        for i, cell in enumerate(row):
            row_str += f" {str(cell):{col_widths[i]}} |"
        print(row_str)
        print("-" * total_width)


def cmd_list_schedules(args):
    db = SessionLocal()
    try:
        query = db.query(Schedule)
        
        if args.route_id:
            query = query.filter(Schedule.route_id == args.route_id)
        
        if args.date:
            date_obj = datetime.strptime(args.date, "%Y-%m-%d")
            date_start = date_obj.replace(hour=0, minute=0, second=0, microsecond=0)
            date_end = date_start + timedelta(days=1)
            query = query.filter(
                Schedule.start_time >= date_start,
                Schedule.start_time < date_end
            )
        
        schedules = query.order_by(Schedule.start_time).all()
        
        if not schedules:
            print("未找到排班记录")
            return
        
        rows = []
        for s in schedules:
            route = db.query(Route).filter(Route.id == s.route_id).first()
            vehicle = db.query(Vehicle).filter(Vehicle.id == s.vehicle_id).first()
            rows.append([
                s.id,
                route.name if route else "未知线路",
                vehicle.vehicle_code if vehicle else "未知车辆",
                s.direction.value,
                s.start_time.strftime("%H:%M"),
                s.end_time.strftime("%H:%M"),
                s.status.value
            ])
        
        headers = ["ID", "线路", "车辆", "方向", "发车时间", "结束时间", "状态"]
        print_table(headers, rows)
        
    finally:
        db.close()


def cmd_schedule_detail(args):
    db = SessionLocal()
    try:
        schedule = db.query(Schedule).filter(Schedule.id == args.id).first()
        if not schedule:
            print(f"排班 ID {args.id} 不存在")
            return
        
        route = db.query(Route).filter(Route.id == schedule.route_id).first()
        vehicle = db.query(Vehicle).filter(Vehicle.id == schedule.vehicle_id).first()
        
        print(f"\n=== 排班详情 ===")
        print(f"ID: {schedule.id}")
        print(f"线路: {route.name if route else '未知'}")
        print(f"车辆: {vehicle.vehicle_code if vehicle else '未知'} ({vehicle.plate_number if vehicle else ''})")
        print(f"方向: {'上行' if schedule.direction == Direction.UP else '下行'}")
        print(f"发车时间: {schedule.start_time.strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"结束时间: {schedule.end_time.strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"状态: {schedule.status.value}")
        
        arrival_records = db.query(ArrivalRecord).filter(
            ArrivalRecord.schedule_id == schedule.id
        ).order_by(ArrivalRecord.planned_arrival_time).all()
        
        print(f"\n=== 站点时刻 ===")
        rows = []
        for ar in arrival_records:
            station = db.query(Station).filter(Station.id == ar.station_id).first()
            actual_time = ar.actual_arrival_time.strftime("%H:%M") if ar.actual_arrival_time else "-"
            on_time = "准点" if ar.is_on_time else ("晚点" if ar.is_on_time is False else "未到")
            
            rows.append([
                station.name if station else "未知",
                ar.planned_arrival_time.strftime("%H:%M"),
                actual_time,
                on_time
            ])
        
        headers = ["站点", "计划到站", "实际到站", "状态"]
        print_table(headers, rows)
        
    finally:
        db.close()


def cmd_punctuality(args):
    db = SessionLocal()
    try:
        start_dt = datetime.strptime(args.start, "%Y-%m-%d") if args.start else None
        end_dt = datetime.strptime(args.end, "%Y-%m-%d") if args.end else None
        
        reports = get_punctuality_report(db, args.route_id, start_dt, end_dt)
        
        if not reports:
            print("暂无准点率数据")
            return
        
        rows = []
        for r in reports:
            rows.append([
                r["route_id"],
                r["route_name"],
                r["total_schedules"],
                r["on_time_count"],
                f"{r['punctuality_rate']:.2f}%"
            ])
        
        headers = ["线路ID", "线路名称", "总班次", "准点班次", "准点率"]
        print_table(headers, rows)
        
    finally:
        db.close()


def cmd_list_routes(args):
    db = SessionLocal()
    try:
        routes = db.query(Route).all()
        if not routes:
            print("暂无线路")
            return
        
        rows = []
        for r in routes:
            rows.append([r.id, r.name, r.code, r.start_station, r.end_station])
        
        headers = ["ID", "名称", "编码", "起点", "终点"]
        print_table(headers, rows)
    finally:
        db.close()


def cmd_route_stations(args):
    db = SessionLocal()
    try:
        query = db.query(RouteStation).filter(RouteStation.route_id == args.route_id)
        if args.direction:
            direction = Direction.UP if args.direction.lower() == "up" else Direction.DOWN
            query = query.filter(RouteStation.direction == direction)
        
        route_stations = query.order_by(RouteStation.direction, RouteStation.sequence).all()
        
        if not route_stations:
            print("该线路暂无站点配置")
            return
        
        rows = []
        for rs in route_stations:
            station = db.query(Station).filter(Station.id == rs.station_id).first()
            direction_str = "上行" if rs.direction == Direction.UP else "下行"
            rows.append([
                rs.id,
                direction_str,
                rs.sequence,
                station.name if station else "未知",
                station.code if station else "-",
                f"{rs.travel_time_from_prev}分钟",
                f"{rs.stop_time}分钟"
            ])
        
        headers = ["ID", "方向", "顺序", "站点名称", "站点编码", "到站耗时", "停靠时间"]
        print_table(headers, rows)
    finally:
        db.close()


def cmd_list_vehicles(args):
    db = SessionLocal()
    try:
        vehicles = db.query(Vehicle).all()
        if not vehicles:
            print("暂无车辆")
            return
        
        rows = []
        for v in vehicles:
            load_ratio = v.current_load / v.capacity if v.capacity > 0 else 0
            crowd_status = "拥挤" if load_ratio > 0.8 else "正常"
            route = db.query(Route).filter(Route.id == v.current_route_id).first()
            
            rows.append([
                v.id,
                v.vehicle_code,
                v.plate_number,
                v.capacity,
                v.current_load,
                f"{load_ratio*100:.1f}%",
                crowd_status,
                route.name if route else "-",
                v.status
            ])
        
        headers = ["ID", "车辆编号", "车牌号", "容量", "当前载客", "负载率", "状态", "当前线路", "运营状态"]
        print_table(headers, rows)
    finally:
        db.close()


def cmd_generate_schedule(args):
    from app.scheduler import generate_schedule_for_route
    
    db = SessionLocal()
    try:
        date = datetime.strptime(args.date, "%Y-%m-%d")
        schedules = generate_schedule_for_route(db, args.route_id, date)
        print(f"成功生成 {len(schedules)} 个排班")
        print(f"排班ID列表: {[s.id for s in schedules]}")
    except ValueError as e:
        print(f"错误: {e}")
    finally:
        db.close()


def main():
    parser = argparse.ArgumentParser(description="BRT运营调度管理系统 - 命令行客户端")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    parser_schedules = subparsers.add_parser("schedules", help="查看排班列表")
    parser_schedules.add_argument("--route-id", type=int, help="按线路ID筛选")
    parser_schedules.add_argument("--date", help="按日期筛选 (YYYY-MM-DD)")
    parser_schedules.set_defaults(func=cmd_list_schedules)
    
    parser_schedule_detail = subparsers.add_parser("schedule-detail", help="查看排班详情")
    parser_schedule_detail.add_argument("--id", type=int, required=True, help="排班ID")
    parser_schedule_detail.set_defaults(func=cmd_schedule_detail)
    
    parser_punctuality = subparsers.add_parser("punctuality", help="查看准点率")
    parser_punctuality.add_argument("--route-id", type=int, help="按线路ID筛选")
    parser_punctuality.add_argument("--start", help="开始日期 (YYYY-MM-DD)")
    parser_punctuality.add_argument("--end", help="结束日期 (YYYY-MM-DD)")
    parser_punctuality.set_defaults(func=cmd_punctuality)
    
    parser_routes = subparsers.add_parser("routes", help="查看线路列表")
    parser_routes.set_defaults(func=cmd_list_routes)
    
    parser_route_stations = subparsers.add_parser("route-stations", help="查看线路站点")
    parser_route_stations.add_argument("--route-id", type=int, required=True, help="线路ID")
    parser_route_stations.add_argument("--direction", help="方向: up 或 down")
    parser_route_stations.set_defaults(func=cmd_route_stations)
    
    parser_vehicles = subparsers.add_parser("vehicles", help="查看车辆列表")
    parser_vehicles.set_defaults(func=cmd_list_vehicles)
    
    parser_generate = subparsers.add_parser("generate-schedule", help="生成排班")
    parser_generate.add_argument("--route-id", type=int, required=True, help="线路ID")
    parser_generate.add_argument("--date", required=True, help="日期 (YYYY-MM-DD)")
    parser_generate.set_defaults(func=cmd_generate_schedule)
    
    args = parser.parse_args()
    
    if args.command is None:
        parser.print_help()
        return
    
    args.func(args)


if __name__ == "__main__":
    main()
