#!/usr/bin/env python3
import argparse
import os
import httpx
from datetime import datetime
import json

BASE_URL = f"http://localhost:{os.getenv('PORT', 9000)}"


def print_table(headers, rows):
    if not rows:
        print("无数据")
        return
    
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            col_widths[i] = max(col_widths[i], len(str(cell)))
    
    separator = "+" + "+".join("-" * (w + 2) for w in col_widths) + "+"
    header_row = "|" + "|".join(f" {h.ljust(w)} " for h, w in zip(headers, col_widths)) + "|"
    
    print(separator)
    print(header_row)
    print(separator)
    
    for row in rows:
        data_row = "|" + "|".join(f" {str(cell).ljust(w)} " for cell, w in zip(row, col_widths)) + "|"
        print(data_row)
        print(separator)


def cmd_list_work_orders(args):
    with httpx.Client() as client:
        params = {}
        if args.section:
            params['section'] = args.section
        if args.status:
            params['status'] = args.status
        
        response = client.get(f"{BASE_URL}/api/work-orders", params=params)
        orders = response.json()
        
        if args.today:
            response = client.get(f"{BASE_URL}/api/work-orders/today")
            orders = response.json()
        
        headers = ["ID", "区段", "类型", "状态", "负责人", "计划时间", "是否逾期"]
        rows = []
        for order in orders:
            plan_time = ""
            if order.get('plan_start_time') and order.get('plan_end_time'):
                start = datetime.fromisoformat(order['plan_start_time'].replace('Z', '+00:00'))
                end = datetime.fromisoformat(order['plan_end_time'].replace('Z', '+00:00'))
                plan_time = f"{start.strftime('%m-%d %H:%M')} ~ {end.strftime('%m-%d %H:%M')}"
            
            overdue = "是" if order.get('is_overdue') else "否"
            status = order['status']
            if order.get('is_overdue'):
                status = f"\033[91m{status}\033[0m"
            
            rows.append([
                order['id'],
                order['section'],
                order['maintenance_type'],
                status,
                order['person_in_charge'],
                plan_time,
                overdue
            ])
        
        print_table(headers, rows)


def cmd_view_work_order(args):
    with httpx.Client() as client:
        response = client.get(f"{BASE_URL}/api/work-orders/{args.id}")
        if response.status_code == 404:
            print("工单不存在")
            return
        order = response.json()
        
        print("\n" + "="*50)
        print(f"工单ID: {order['id']}")
        print(f"区段: {order['section']}")
        print(f"检修类型: {order['maintenance_type']}")
        print(f"状态: {order['status']}")
        print(f"负责人: {order['person_in_charge']}")
        
        if order.get('plan_start_time'):
            start = datetime.fromisoformat(order['plan_start_time'].replace('Z', '+00:00'))
            end = datetime.fromisoformat(order['plan_end_time'].replace('Z', '+00:00'))
            print(f"计划时间: {start.strftime('%Y-%m-%d %H:%M')} ~ {end.strftime('%Y-%m-%d %H:%M')}")
        
        if order.get('actual_start_time'):
            start = datetime.fromisoformat(order['actual_start_time'].replace('Z', '+00:00'))
            print(f"实际开始: {start.strftime('%Y-%m-%d %H:%M')}")
        if order.get('actual_end_time'):
            end = datetime.fromisoformat(order['actual_end_time'].replace('Z', '+00:00'))
            print(f"实际结束: {end.strftime('%Y-%m-%d %H:%M')}")
        
        if order.get('description'):
            print(f"描述: {order['description']}")
        
        print(f"是否逾期: {'是' if order.get('is_overdue') else '否'}")
        print("="*50 + "\n")


def cmd_update_work_order(args):
    with httpx.Client() as client:
        data = {}
        if args.status:
            data['status'] = args.status
        
        response = client.put(f"{BASE_URL}/api/work-orders/{args.id}", json=data)
        if response.status_code == 404:
            print("工单不存在")
            return
        
        order = response.json()
        print(f"工单 {args.id} 状态已更新为: {order['status']}")


def cmd_power_status(args):
    with httpx.Client() as client:
        if args.section:
            response = client.get(f"{BASE_URL}/api/power-supply/data/{args.section}/latest")
            if response.status_code == 404:
                print("未找到该区段的供电数据")
                return
            data = response.json()
            
            status = "正常"
            if data['is_power_outage']:
                status = "\033[91m停电\033[0m"
            elif data['is_alarm']:
                status = "\033[93m告警\033[0m"
            
            print("\n" + "="*50)
            print(f"区段: {data['section']}")
            print(f"状态: {status}")
            print(f"电流: {data['current']} A")
            print(f"电压: {data['voltage']} V")
            print(f"额定电流: {data['rated_current']} A")
            recorded = datetime.fromisoformat(data['recorded_at'].replace('Z', '+00:00'))
            print(f"记录时间: {recorded.strftime('%Y-%m-%d %H:%M:%S')}")
            print("="*50 + "\n")
        else:
            print("请指定区段名称: --section <区段名>")


def cmd_section_stats(args):
    with httpx.Client() as client:
        response = client.get(f"{BASE_URL}/api/power-supply/statistics/{args.section}")
        stats = response.json()
        
        print("\n" + "="*50)
        print(f"区段: {stats['section']}")
        print(f"累计巡检次数: {stats['total_inspection_count']}")
        print(f"累计停电时长: {stats['total_power_outage_duration_minutes']} 分钟")
        print(f"累计告警次数: {stats['total_alarm_count']}")
        updated = datetime.fromisoformat(stats['last_updated_at'].replace('Z', '+00:00'))
        print(f"最后更新: {updated.strftime('%Y-%m-%d %H:%M:%S')}")
        print("="*50 + "\n")


def cmd_list_alarms(args):
    with httpx.Client() as client:
        params = {}
        if args.status:
            params['status'] = args.status
        
        response = client.get(f"{BASE_URL}/api/alarms", params=params)
        alarms = response.json()
        
        headers = ["ID", "区段", "严重程度", "状态", "关联工单", "创建时间"]
        rows = []
        for alarm in alarms:
            created = datetime.fromisoformat(alarm['created_at'].replace('Z', '+00:00'))
            severity = alarm['severity']
            if severity == "高":
                severity = f"\033[91m{severity}\033[0m"
            elif severity == "中":
                severity = f"\033[93m{severity}\033[0m"
            
            rows.append([
                alarm['id'],
                alarm['section'],
                severity,
                alarm['status'],
                alarm.get('original_work_order_id', '-'),
                created.strftime('%m-%d %H:%M')
            ])
        
        print_table(headers, rows)


def main():
    parser = argparse.ArgumentParser(description="接触网检修管理系统 - 命令行客户端")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    list_parser = subparsers.add_parser("list", help="列出工单")
    list_parser.add_argument("--section", help="按区段筛选")
    list_parser.add_argument("--status", help="按状态筛选")
    list_parser.add_argument("--today", action="store_true", help="仅显示今天的工单")
    list_parser.set_defaults(func=cmd_list_work_orders)
    
    view_parser = subparsers.add_parser("view", help="查看工单详情")
    view_parser.add_argument("id", type=int, help="工单ID")
    view_parser.set_defaults(func=cmd_view_work_order)
    
    update_parser = subparsers.add_parser("update", help="更新工单状态")
    update_parser.add_argument("id", type=int, help="工单ID")
    update_parser.add_argument("--status", required=True, 
                               choices=["创建", "待执行", "执行中", "已完成"],
                               help="新状态")
    update_parser.set_defaults(func=cmd_update_work_order)
    
    power_parser = subparsers.add_parser("power", help="查看供电状态")
    power_parser.add_argument("--section", help="区段名称")
    power_parser.set_defaults(func=cmd_power_status)
    
    stats_parser = subparsers.add_parser("stats", help="查看区段统计")
    stats_parser.add_argument("section", help="区段名称")
    stats_parser.set_defaults(func=cmd_section_stats)
    
    alarms_parser = subparsers.add_parser("alarms", help="列出告警工单")
    alarms_parser.add_argument("--status", help="按状态筛选")
    alarms_parser.set_defaults(func=cmd_list_alarms)
    
    args = parser.parse_args()
    if args.command:
        args.func(args)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
