import argparse
import os
import sys
import httpx
from datetime import datetime


def get_server_url():
    port = int(os.environ.get("PORT", 8000))
    return f"http://localhost:{port}"


def get_queue_realtime():
    url = f"{get_server_url()}/queue/realtime"
    try:
        with httpx.Client() as client:
            response = client.get(url, timeout=5)
            response.raise_for_status()
            data = response.json()
            
            print("=" * 50)
            print("实时排队统计")
            print("=" * 50)
            print(f"排队人数: {data['total_queue']} 人")
            print(f"上行轿厢人数: {data['upward_passengers']} 人")
            print(f"有效排队（含上行）: {data['effective_queue']} 人")
            print(f"排队组数: {data['queue_groups']} 组")
            print(f"平均等待: {data['average_wait_seconds']:.1f} 秒")
            if data.get('first_join_time'):
                print(f"最早排队时间: {data['first_join_time']}")
            print("=" * 50)
    except httpx.HTTPError as e:
        print(f"错误: 无法连接到服务器 - {e}")
        sys.exit(1)


def get_queue_estimate():
    url = f"{get_server_url()}/queue/estimate"
    try:
        with httpx.Client() as client:
            response = client.get(url, timeout=5)
            response.raise_for_status()
            data = response.json()
            
            print("=" * 50)
            print("排队预估")
            print("=" * 50)
            print(f"预计等待时间: {data['estimated_wait_minutes']:.1f} 分钟")
            print(f"当前排队规模: {data['queue_size']} 人")
            print(f"当前发车间隔: {data['current_interval_minutes']:.1f} 分钟")
            print(f"距离下次发车: 约 {data['next_dispatch_minutes']:.1f} 分钟")
            print("=" * 50)
    except httpx.HTTPError as e:
        print(f"错误: 无法连接到服务器 - {e}")
        sys.exit(1)


def get_daily_stats(date_str=None):
    url = f"{get_server_url()}/stats/daily"
    params = {}
    if date_str:
        params["target_date"] = date_str
    
    try:
        with httpx.Client() as client:
            response = client.get(url, params=params, timeout=5)
            response.raise_for_status()
            data = response.json()
            
            print("=" * 50)
            print(f"每日运营统计 - {data['date']}")
            print("=" * 50)
            print(f"总发车次数: {data['total_trips']} 次")
            
            passengers = data.get('total_passengers')
            print(f"总乘客数: {passengers if passengers is not None else '-'} 人")
            
            weight = data.get('total_weight')
            print(f"总载重量: {f'{weight:.1f}kg' if weight is not None else '-'}")
            
            avg_weight = data.get('average_weight_per_trip')
            print(f"平均载重: {f'{avg_weight:.1f}kg' if avg_weight is not None else '-'}")
            
            peak = data.get('peak_hour_passengers')
            print(f"高峰小时乘客: {peak if peak is not None else '-'} 人")
            
            print(f"完成维护: {data['maintenance_completed']} 项")
            print(f"超期维护: {data['maintenance_overdue']} 项")
            print("=" * 50)
    except httpx.HTTPError as e:
        print(f"错误: 无法连接到服务器 - {e}")
        sys.exit(1)


def get_shift_stats(date_str=None):
    url = f"{get_server_url()}/stats/shifts"
    params = {}
    if date_str:
        params["target_date"] = date_str
    
    try:
        with httpx.Client() as client:
            response = client.get(url, params=params, timeout=5)
            response.raise_for_status()
            data = response.json()
            
            print("=" * 50)
            print(f"班次统计 - {data['date']}")
            print("=" * 50)
            
            shifts = data.get('shifts', [])
            if not shifts:
                print("暂无班次数据")
            else:
                for shift in shifts:
                    print(f"\n班次 #{shift['shift_number']}")
                    print(f"  开始时间: {shift['start_time']}")
                    print(f"  结束时间: {shift.get('end_time') or '进行中'}")
                    
                    passengers = shift.get('total_passengers')
                    print(f"  乘客数: {passengers if passengers is not None else '-'}")
                    
                    weight = shift.get('total_weight')
                    print(f"  载重量: {f'{weight:.1f}kg' if weight is not None else '-'}")
                    
                    trips = shift.get('total_trips')
                    print(f"  发车次数: {trips if trips is not None else '-'}")
                    
                    revenue = shift.get('total_revenue')
                    print(f"  收入: {f'¥{revenue:.2f}' if revenue is not None else '-'}")
                    
                    wait_time = shift.get('average_wait_time')
                    print(f"  平均等待: {f'{wait_time:.1f}秒' if wait_time is not None else '-'}")
            
            print("\n" + "=" * 50)
    except httpx.HTTPError as e:
        print(f"错误: 无法连接到服务器 - {e}")
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="缆车运营管理系统命令行客户端")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    subparsers.add_parser("queue", help="查看实时排队信息")
    subparsers.add_parser("estimate", help="查看排队预估等待时间")
    
    daily_parser = subparsers.add_parser("daily", help="查看每日运营统计")
    daily_parser.add_argument("--date", help="查询日期 (格式: YYYY-MM-DD)", default=None)
    
    shift_parser = subparsers.add_parser("shift", help="查看班次统计")
    shift_parser.add_argument("--date", help="查询日期 (格式: YYYY-MM-DD)", default=None)
    
    args = parser.parse_args()
    
    if args.command == "queue":
        get_queue_realtime()
    elif args.command == "estimate":
        get_queue_estimate()
    elif args.command == "daily":
        get_daily_stats(args.date)
    elif args.command == "shift":
        get_shift_stats(args.date)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
