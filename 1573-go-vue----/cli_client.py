#!/usr/bin/env python3
import argparse
import os
import httpx
import json

BASE_URL = os.getenv("API_URL", "http://localhost:8000")


def get_client():
    return httpx.Client(base_url=BASE_URL)


def list_devices(device_type):
    with get_client() as client:
        response = client.get(f"/{device_type}/")
        if response.status_code == 200:
            devices = response.json()
            print(f"\n=== {device_type.upper()} 设备列表 ===")
            if not devices:
                print("无设备")
            for device in devices:
                print(json.dumps(device, indent=2, ensure_ascii=False))
        else:
            print(f"错误: {response.status_code} - {response.text}")


def get_device(device_type, device_id):
    with get_client() as client:
        response = client.get(f"/{device_type}/{device_id}")
        if response.status_code == 200:
            print(f"\n=== {device_type.upper()} 设备详情 ===")
            print(json.dumps(response.json(), indent=2, ensure_ascii=False))
        else:
            print(f"错误: {response.status_code} - {response.text}")


def get_alerts():
    with get_client() as client:
        response = client.get("/stats/alerts")
        if response.status_code == 200:
            alerts = response.json()
            print("\n=== 告警列表 ===")
            if not alerts:
                print("无告警")
            for alert in alerts:
                print(json.dumps(alert, indent=2, ensure_ascii=False))
        else:
            print(f"错误: {response.status_code} - {response.text}")


def get_stats(period):
    with get_client() as client:
        response = client.get(f"/stats/{period}")
        if response.status_code == 200:
            stats = response.json()
            print(f"\n=== {period.upper()} 统计 ===")
            print(json.dumps(stats, indent=2, ensure_ascii=False))
        else:
            print(f"错误: {response.status_code} - {response.text}")


def operate_switch(device_id, action, position=None, locked=None):
    with get_client() as client:
        if action == "switch":
            response = client.post(
                f"/signals/{device_id}/switch",
                json={"position": position}
            )
        elif action == "lock":
            response = client.post(
                f"/signals/{device_id}/lock",
                json={"locked": locked}
            )
        
        if response.status_code == 200:
            print(f"操作成功: {response.json()}")
        else:
            print(f"错误: {response.status_code} - {response.text}")


def operate_block(device_id, action):
    with get_client() as client:
        response = client.post(f"/block/{device_id}/{action}")
        if response.status_code == 200:
            print(f"操作成功: {response.json()}")
        else:
            print(f"错误: {response.status_code} - {response.text}")


def operate_semaphore(device_id, action):
    with get_client() as client:
        response = client.post(f"/semaphore/{device_id}/{action}")
        if response.status_code == 200:
            print(f"操作成功: {response.json()}")
        else:
            print(f"错误: {response.status_code} - {response.text}")


def operate_interlocking(device_id, action, route_data=None):
    with get_client() as client:
        if action == "set-route":
            response = client.post(
                f"/interlocking/{device_id}/set-route",
                json=route_data
            )
        elif action == "cancel-route":
            response = client.post(f"/interlocking/{device_id}/cancel-route")
        
        if response.status_code == 200:
            print(f"操作成功: {response.json()}")
        else:
            print(f"错误: {response.status_code} - {response.text}")


def main():
    parser = argparse.ArgumentParser(description="铁路信号监控系统命令行客户端")
    subparsers = parser.add_subparsers(dest="command", help="命令")
    
    list_parser = subparsers.add_parser("list", help="列出设备")
    list_parser.add_argument("device_type", choices=["signals", "interlocking", "block", "semaphore"], help="设备类型")
    
    get_parser = subparsers.add_parser("get", help="获取设备详情")
    get_parser.add_argument("device_type", choices=["signals", "interlocking", "block", "semaphore"], help="设备类型")
    get_parser.add_argument("device_id", help="设备ID")
    
    subparsers.add_parser("alerts", help="查看告警")
    
    stats_parser = subparsers.add_parser("stats", help="查看统计")
    stats_parser.add_argument("period", choices=["hourly", "daily"], help="统计周期")
    
    switch_parser = subparsers.add_parser("switch", help="操作道岔")
    switch_parser.add_argument("device_id", help="道岔ID")
    switch_parser.add_argument("action", choices=["switch", "lock"], help="操作类型")
    switch_parser.add_argument("--position", help="道岔位置 (switch操作时)")
    switch_parser.add_argument("--locked", type=bool, help="是否锁闭 (lock操作时)")
    
    block_parser = subparsers.add_parser("block", help="操作闭塞分区")
    block_parser.add_argument("device_id", help="闭塞分区ID")
    block_parser.add_argument("action", choices=["occupy", "release"], help="操作类型")
    
    semaphore_parser = subparsers.add_parser("semaphore", help="操作信号机")
    semaphore_parser.add_argument("device_id", help="信号机ID")
    semaphore_parser.add_argument("action", choices=["open", "close"], help="操作类型")
    
    interlock_parser = subparsers.add_parser("interlock", help="操作联锁")
    interlock_parser.add_argument("device_id", help="联锁设备ID")
    interlock_parser.add_argument("action", choices=["set-route", "cancel-route"], help="操作类型")
    interlock_parser.add_argument("--route-id", help="进路ID (set-route时)")
    interlock_parser.add_argument("--switches", nargs="+", help="道岔列表 (set-route时)")
    interlock_parser.add_argument("--blocks", nargs="+", help="闭塞分区列表 (set-route时)")
    interlock_parser.add_argument("--semaphores", nargs="+", help="信号机列表 (set-route时)")
    
    args = parser.parse_args()
    
    if args.command == "list":
        list_devices(args.device_type)
    elif args.command == "get":
        get_device(args.device_type, args.device_id)
    elif args.command == "alerts":
        get_alerts()
    elif args.command == "stats":
        get_stats(args.period)
    elif args.command == "switch":
        operate_switch(args.device_id, args.action, args.position, args.locked)
    elif args.command == "block":
        operate_block(args.device_id, args.action)
    elif args.command == "semaphore":
        operate_semaphore(args.device_id, args.action)
    elif args.command == "interlock":
        route_data = None
        if args.action == "set-route":
            route_data = {
                "route_id": args.route_id,
                "switches": args.switches or [],
                "blocks": args.blocks or [],
                "semaphores": args.semaphores or [],
                "conflicting_routes": []
            }
        operate_interlocking(args.device_id, args.action, route_data)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
