#!/usr/bin/env python3
import argparse
import sys
import httpx
from datetime import datetime

BASE_URL = "http://localhost:8000"

def format_datetime(dt_str):
    if not dt_str:
        return "N/A"
    try:
        dt = datetime.fromisoformat(dt_str.replace('Z', '+00:00'))
        return dt.strftime("%Y-%m-%d %H:%M:%S")
    except:
        return dt_str

def list_vehicles():
    with httpx.Client(base_url=BASE_URL) as client:
        response = client.get("/vehicles/")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return
        vehicles = response.json()
        print("\n" + "="*100)
        print(f"{'ID':<5} {'车牌号':<12} {'状态':<10} {'电量(%)':<10} {'当前站点':<12} {'当前线路':<12} {'最后上报':<20}")
        print("-"*100)
        for v in vehicles:
            print(f"{v['id']:<5} {v['plate_number']:<12} {v['status']:<10} {v['current_battery']:<10.1f} "
                  f"{str(v['current_station_id'] or 'N/A'):<12} "
                  f"{str(v['current_route_id'] or 'N/A'):<12} "
                  f"{format_datetime(v['last_report_time']):<20}")
        print("="*100 + "\n")

def get_vehicle_status(vehicle_id):
    with httpx.Client(base_url=BASE_URL) as client:
        response = client.get(f"/vehicles/{vehicle_id}/status")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return
        status = response.json()
        print("\n" + "="*80)
        print(f"车辆详细状态 (ID: {status['vehicle_id']})")
        print("="*80)
        print(f"车牌号: {status['plate_number']}")
        print(f"状态: {status['status']}")
        print(f"当前电量: {status['current_battery']:.1f}%")
        print(f"当前站点ID: {status['current_station_id'] or 'N/A'}")
        print(f"当前线路ID: {status['current_route_id'] or 'N/A'}")
        print(f"最后上报时间: {format_datetime(status['last_report_time'])}")
        
        if status['charging_status']:
            cs = status['charging_status']
            print("\n" + "-"*80)
            print("充电状态:")
            print(f"  充电会话ID: {cs['session_id']}")
            print(f"  充电桩名称: {cs['charger_name']}")
            print(f"  开始时间: {format_datetime(cs['start_time'])}")
            print(f"  预计结束: {format_datetime(cs['estimated_end'])}")
            print(f"  当前电量: {cs['current_battery']:.1f}%")
            print(f"  目标电量: {cs['target_battery']:.1f}%")
            
            if cs['start_time'] and cs['estimated_end'] and cs['current_battery'] and cs['target_battery']:
                progress = (cs['current_battery'] - (cs['target_battery'] - 60)) / 60 * 100
                progress = max(0, min(100, progress))
                print(f"  充电进度: {progress:.1f}%")
        print("="*80 + "\n")

def list_chargers():
    with httpx.Client(base_url=BASE_URL) as client:
        response = client.get("/chargers/")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return
        chargers = response.json()
        print("\n" + "="*80)
        print(f"{'ID':<5} {'名称':<15} {'站点ID':<10} {'状态':<12} {'功率(kW)':<10} {'充电次数':<10}")
        print("-"*80)
        for c in chargers:
            print(f"{c['id']:<5} {c['name']:<15} {str(c['station_id'] or 'N/A'):<10} "
                  f"{c['status']:<12} {c['power_kw']:<10.1f} {c['total_charging_count']:<10}")
        print("="*80 + "\n")

def list_active_charging():
    with httpx.Client(base_url=BASE_URL) as client:
        response = client.get("/charging-sessions/active")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return
        sessions = response.json()
        if not sessions:
            print("\n当前没有进行中的充电会话\n")
            return
        print("\n" + "="*100)
        print(f"{'会话ID':<8} {'车辆ID':<8} {'充电桩ID':<10} {'开始时间':<20} {'预计结束':<20} {'进度':<10}")
        print("-"*100)
        for s in sessions:
            print(f"{s['id']:<8} {s['vehicle_id']:<8} {s['charger_id']:<10} "
                  f"{format_datetime(s['start_time']):<20} "
                  f"{format_datetime(s['estimated_end_time']):<20} "
                  f"{f'{s['current_battery']:.1f}%' if s['current_battery'] else 'N/A':<10}")
        print("="*100 + "\n")

def list_alerts():
    with httpx.Client(base_url=BASE_URL) as client:
        response = client.get("/alerts/active")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return
        alerts = response.json()
        if not alerts:
            print("\n当前没有活动告警\n")
            return
        print("\n" + "="*100)
        print(f"{'ID':<5} {'类型':<15} {'状态':<12} {'通知方式':<12} {'创建时间':<20}")
        print("-"*100)
        for a in alerts:
            print(f"{a['id']:<5} {a['type']:<15} {a['status']:<12} "
                  f"{a['notification_type']:<12} {format_datetime(a['created_at']):<20}")
            print(f"  消息: {a['message']}")
        print("="*100 + "\n")

def list_dispatch_logs(limit=20):
    with httpx.Client(base_url=BASE_URL) as client:
        response = client.get("/dispatch-logs/")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return
        logs = response.json()
        logs = logs[-limit:]
        print("\n" + "="*100)
        print(f"{'ID':<5} {'级别':<8} {'时间':<20} {'车辆ID':<10} {'充电桩ID':<10} {'消息'}")
        print("-"*100)
        for log in logs:
            print(f"{log['id']:<5} {log['level']:<8} {format_datetime(log['created_at']):<20} "
                  f"{str(log['related_vehicle_id'] or 'N/A'):<10} "
                  f"{str(log['related_charger_id'] or 'N/A'):<10} {log['message']}")
        print("="*100 + "\n")

def run_monitor():
    with httpx.Client(base_url=BASE_URL) as client:
        response = client.post("/monitor/run")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return
        result = response.json()
        print("\n" + "="*80)
        print("调度监控结果")
        print("="*80)
        print(f"总车辆数: {result['total_vehicles']}")
        print(f"处理事件数: {result['processed_events']}")
        if result['events']:
            print("\n事件详情:")
            for event in result['events']:
                print(f"  - 车辆ID: {event['vehicle_id']}, 类型: {event['type']}")
                if 'data' in event:
                    if 'message' in event['data']:
                        print(f"    消息: {event['data']['message']}")
        print("="*80 + "\n")

def get_vehicle_range(vehicle_id):
    with httpx.Client(base_url=BASE_URL) as client:
        response = client.get(f"/monitor/vehicle-range/{vehicle_id}")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return
        data = response.json()
        print("\n" + "="*80)
        print(f"车辆续航信息 (ID: {data['vehicle_id']})")
        print("="*80)
        print(f"车牌号: {data['plate_number']}")
        print(f"当前电量: {data['current_battery']:.1f}%")
        print(f"预估剩余里程: {data['remaining_range_km']:.2f} km")
        print(f"能否完成当前交路: {'是' if data['can_complete_current_assignment'] else '否'}")
        print(f"当前线路ID: {data['current_route_id'] or 'N/A'}")
        print("="*80 + "\n")

def main():
    parser = argparse.ArgumentParser(
        description="有轨电车运营管理系统 - 命令行客户端",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例用法:
  python cli.py vehicles              # 列出所有车辆
  python cli.py vehicle 1             # 查看车辆1的详细状态
  python cli.py chargers              # 列出所有充电桩
  python cli.py charging              # 查看进行中的充电会话
  python cli.py alerts                # 查看活动告警
  python cli.py logs                  # 查看调度日志
  python cli.py monitor               # 运行调度监控
  python cli.py range 1               # 查看车辆1的续航信息
        """
    )
    
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    subparsers.add_parser("vehicles", help="列出所有车辆")
    
    vehicle_parser = subparsers.add_parser("vehicle", help="查看车辆详细状态")
    vehicle_parser.add_argument("vehicle_id", type=int, help="车辆ID")
    
    subparsers.add_parser("chargers", help="列出所有充电桩")
    
    subparsers.add_parser("charging", help="查看进行中的充电会话")
    
    subparsers.add_parser("alerts", help="查看活动告警")
    
    logs_parser = subparsers.add_parser("logs", help="查看调度日志")
    logs_parser.add_argument("--limit", type=int, default=20, help="显示最近N条日志")
    
    subparsers.add_parser("monitor", help="运行调度监控")
    
    range_parser = subparsers.add_parser("range", help="查看车辆续航信息")
    range_parser.add_argument("vehicle_id", type=int, help="车辆ID")
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    try:
        if args.command == "vehicles":
            list_vehicles()
        elif args.command == "vehicle":
            get_vehicle_status(args.vehicle_id)
        elif args.command == "chargers":
            list_chargers()
        elif args.command == "charging":
            list_active_charging()
        elif args.command == "alerts":
            list_alerts()
        elif args.command == "logs":
            list_dispatch_logs(args.limit)
        elif args.command == "monitor":
            run_monitor()
        elif args.command == "range":
            get_vehicle_range(args.vehicle_id)
    except httpx.ConnectError:
        print("错误: 无法连接到服务器，请确保服务已启动 (uvicorn app.main:app --reload)")
        sys.exit(1)
    except Exception as e:
        print(f"错误: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
