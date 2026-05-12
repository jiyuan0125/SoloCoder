#!/usr/bin/env python3
import argparse
import sys
import json
import os

import httpx

from dotenv import load_dotenv

load_dotenv()

APP_PORT = int(os.getenv("APP_PORT", "8000"))
BASE_URL = f"http://localhost:{APP_PORT}"


def get_client():
    return httpx.Client(base_url=BASE_URL, timeout=30)


def format_json(data):
    return json.dumps(data, ensure_ascii=False, indent=2)


def cmd_list_flights(args):
    with get_client() as client:
        response = client.get("/flights")
        if response.status_code == 200:
            flights = response.json()
            if not flights:
                print("暂无航班")
                return
            print("\n=== 航班列表 ===")
            for f in flights:
                print(f"  航班号: {f['flight_number']}")
                print(f"  航线: {f['route']}")
                print(f"  计划起飞: {f['scheduled_departure']}")
                print(f"  计划到达: {f['scheduled_arrival']}")
                print(f"  单价: ¥{f['unit_price']}/kg")
                print(f"  最低配载率: {f['min_load_rate']*100:.0f}%")
                print("-" * 40)
        else:
            print(f"错误: {response.status_code} - {response.text}")


def cmd_get_flight(args):
    with get_client() as client:
        response = client.get(f"/flights/{args.flight_number}")
        if response.status_code == 200:
            print(format_json(response.json()))
        elif response.status_code == 404:
            print("航班不存在")
        else:
            print(f"错误: {response.status_code} - {response.text}")


def cmd_list_cargo(args):
    with get_client() as client:
        response = client.get("/cargo")
        if response.status_code == 200:
            cargoes = response.json()
            if not cargoes:
                print("暂无货运单")
                return
            print("\n=== 货运单列表 ===")
            for c in cargoes:
                print(f"  货运单号: {c['cargo_number']}")
                print(f"  货主: {c['shipper']} -> 收货人: {c['consignee']}")
                print(f"  货物类型: {c['cargo_type']}")
                print(f"  状态: {c['status']}")
                if c['chargeable_weight']:
                    print(f"  计费重量: {c['chargeable_weight']}kg")
                if c['freight_charge']:
                    print(f"  运费: ¥{c['freight_charge']:.2f}")
                if c['storage_charge']:
                    print(f"  保管费: ¥{c['storage_charge']:.2f}")
                print("-" * 40)
        else:
            print(f"错误: {response.status_code} - {response.text}")


def cmd_get_cargo(args):
    with get_client() as client:
        response = client.get(f"/cargo/{args.cargo_number}")
        if response.status_code == 200:
            print(format_json(response.json()))
        elif response.status_code == 404:
            print("货运单不存在")
        else:
            print(f"错误: {response.status_code} - {response.text}")


def cmd_list_compartments(args):
    if args.flight_number:
        with get_client() as client:
            response = client.get(f"/flights/{args.flight_number}/compartments")
            if response.status_code == 200:
                compartments = response.json()
                if not compartments:
                    print("该航班暂无舱位")
                    return
                print(f"\n=== 航班 {args.flight_number} 舱位列表 ===")
                for c in compartments:
                    print(f"  舱位号: {c['compartment_number']}")
                    print(f"  位置: {c['position']}")
                    print(f"  总容量: {c['total_capacity_weight']}kg")
                    print(f"  剩余容量: {c['remaining_capacity_weight']}kg")
                    print(f"  温控舱: {'是' if c['is_temperature_controlled'] else '否'}")
                    print("-" * 40)
            else:
                print(f"错误: {response.status_code} - {response.text}")
    else:
        print("请提供 --flight 参数指定航班号")


def cmd_get_compartment(args):
    with get_client() as client:
        response = client.get(f"/flights/compartments/{args.compartment_number}")
        if response.status_code == 200:
            print(format_json(response.json()))
        elif response.status_code == 404:
            print("舱位不存在")
        else:
            print(f"错误: {response.status_code} - {response.text}")


def cmd_list_assignments(args):
    with get_client() as client:
        if args.cargo_number:
            response = client.get(f"/load/cargo/{args.cargo_number}")
            if response.status_code == 200:
                assignments = response.json()
                if not assignments:
                    print("该货运单暂无配载记录")
                    return
                print(f"\n=== 货运单 {args.cargo_number} 配载记录 ===")
                for a in assignments:
                    print(f"  舱位: {a['compartment_number']}")
                    print(f"  分配重量: {a['assigned_weight']}kg")
                    print(f"  状态: {'已确认' if a['is_confirmed'] else '待确认'}")
                    if a['confirmed_at']:
                        print(f"  确认时间: {a['confirmed_at']}")
                    print("-" * 40)
            else:
                print(f"错误: {response.status_code} - {response.text}")
        elif args.compartment_number:
            response = client.get(f"/load/{args.compartment_number}/assignments")
            if response.status_code == 200:
                assignments = response.json()
                if not assignments:
                    print("该舱位暂无配载记录")
                    return
                print(f"\n=== 舱位 {args.compartment_number} 配载记录 ===")
                for a in assignments:
                    print(f"  货运单: {a['cargo_number']}")
                    print(f"  分配重量: {a['assigned_weight']}kg")
                    print(f"  状态: {'已确认' if a['is_confirmed'] else '待确认'}")
                    if a['confirmed_at']:
                        print(f"  确认时间: {a['confirmed_at']}")
                    print("-" * 40)
            else:
                print(f"错误: {response.status_code} - {response.text}")
        else:
            print("请提供 --cargo 或 --compartment 参数")


def main():
    parser = argparse.ArgumentParser(description="航空货运管理系统 CLI")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    flights_parser = subparsers.add_parser("flights", help="航班相关操作")
    flights_sub = flights_parser.add_subparsers(dest="subcommand")
    
    flights_list = flights_sub.add_parser("list", help="列出所有航班")
    flights_list.set_defaults(func=cmd_list_flights)
    
    flights_get = flights_sub.add_parser("get", help="获取航班详情")
    flights_get.add_argument("flight_number", help="航班号")
    flights_get.set_defaults(func=cmd_get_flight)
    
    cargo_parser = subparsers.add_parser("cargo", help="货运单相关操作")
    cargo_sub = cargo_parser.add_subparsers(dest="subcommand")
    
    cargo_list = cargo_sub.add_parser("list", help="列出所有货运单")
    cargo_list.set_defaults(func=cmd_list_cargo)
    
    cargo_get = cargo_sub.add_parser("get", help="获取货运单详情")
    cargo_get.add_argument("cargo_number", help="货运单号")
    cargo_get.set_defaults(func=cmd_get_cargo)
    
    compartments_parser = subparsers.add_parser("compartments", help="舱位相关操作")
    compartments_sub = compartments_parser.add_subparsers(dest="subcommand")
    
    compartments_list = compartments_sub.add_parser("list", help="列出舱位")
    compartments_list.add_argument("--flight", dest="flight_number", help="航班号")
    compartments_list.set_defaults(func=cmd_list_compartments)
    
    compartments_get = compartments_sub.add_parser("get", help="获取舱位详情")
    compartments_get.add_argument("compartment_number", help="舱位号")
    compartments_get.set_defaults(func=cmd_get_compartment)
    
    load_parser = subparsers.add_parser("load", help="配载相关操作")
    load_sub = load_parser.add_subparsers(dest="subcommand")
    
    load_list = load_sub.add_parser("list", help="列出配载记录")
    load_list.add_argument("--cargo", dest="cargo_number", help="货运单号")
    load_list.add_argument("--compartment", dest="compartment_number", help="舱位号")
    load_list.set_defaults(func=cmd_list_assignments)
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    if "func" not in args:
        subparser_map = {
            "flights": flights_parser,
            "cargo": cargo_parser,
            "compartments": compartments_parser,
            "load": load_parser
        }
        if args.command in subparser_map:
            subparser_map[args.command].print_help()
        sys.exit(1)
    
    args.func(args)


if __name__ == "__main__":
    main()
