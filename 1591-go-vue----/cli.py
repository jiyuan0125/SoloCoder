#!/usr/bin/env python3
import argparse
import requests
import json
import sys

BASE_URL = "http://localhost:8000"


def format_price(cents: int) -> str:
    return f"{cents / 100:.2f} 元"


def print_separator():
    print("=" * 60)


def print_result(result, indent=2):
    print(json.dumps(result, ensure_ascii=False, indent=indent))


def cmd_list_parks(args):
    print_separator()
    print("景区列表")
    print_separator()
    try:
        response = requests.get(f"{BASE_URL}/parks")
        parks = response.json()
        if not parks:
            print("暂无景区")
            return
        for park in parks:
            print(f"ID: {park['id']}")
            print(f"代码: {park['code']}")
            print(f"名称: {park['name']}")
            print(f"最大容量: {park['max_capacity']} 人")
            print(f"当前游客: {park['current_visitors']} 人")
            print("-" * 60)
    except requests.exceptions.RequestException as e:
        print(f"错误: {e}")


def cmd_list_tickets(args):
    print_separator()
    print(f"景区 {args.park_code} 的门票列表")
    print_separator()
    try:
        params = {}
        if args.status:
            params["status"] = args.status
        response = requests.get(f"{BASE_URL}/parks/{args.park_code}/tickets", params=params)
        if response.status_code != 200:
            print(f"错误: {response.json().get('detail', '查询失败')}")
            return
        tickets = response.json()
        if not tickets:
            print("暂无门票记录")
            return
        for ticket in tickets:
            print(f"门票ID: {ticket['id']}")
            print(f"票种: {ticket['ticket_type_name']}")
            print(f"价格: {format_price(ticket['final_price'])}")
            print(f"保险费: {format_price(ticket['insurance_fee'])}")
            print(f"游客: {ticket.get('visitor_name', '未指定')}")
            print(f"状态: {ticket['status']}")
            print(f"购票时间: {ticket['purchase_time']}")
            print("-" * 60)
    except requests.exceptions.RequestException as e:
        print(f"错误: {e}")


def cmd_list_guides(args):
    print_separator()
    print(f"景区 {args.park_code} 的导游列表")
    print_separator()
    try:
        params = {}
        if args.available:
            params["available_only"] = "true"
        response = requests.get(f"{BASE_URL}/parks/{args.park_code}/guides", params=params)
        if response.status_code != 200:
            print(f"错误: {response.json().get('detail', '查询失败')}")
            return
        guides = response.json()
        if not guides:
            print("暂无导游")
            return
        level_names = {"junior": "初级", "intermediate": "中级", "senior": "高级"}
        for guide in guides:
            avg_rating = guide["total_rating"] / guide["rating_count"] if guide["rating_count"] > 0 else 0.0
            print(f"导游ID: {guide['guide_id']}")
            print(f"姓名: {guide['name']}")
            print(f"等级: {level_names.get(guide['level'], guide['level'])}")
            print(f"状态: {'可用' if guide['is_available'] else '忙碌'}")
            print(f"服务次数: {guide['total_assignments']}")
            print(f"平均分: {avg_rating:.1f} ({guide['rating_count']}次评价)")
            print("-" * 60)
    except requests.exceptions.RequestException as e:
        print(f"错误: {e}")


def cmd_query_price(args):
    print_separator()
    print(f"查询 {args.ticket_type} 票价格")
    print_separator()
    try:
        data = {
            "visitor_name": getattr(args, 'visitor_name', None),
            "visitor_age": getattr(args, 'age', None),
            "visitor_height": getattr(args, 'height', None),
            "is_student": getattr(args, 'student', False),
            "is_group": getattr(args, 'group', False),
            "group_size": getattr(args, 'group_size', 1),
            "free_children_count": getattr(args, 'free_children', 0),
            "route_id": getattr(args, 'route_id', None)
        }
        response = requests.post(
            f"{BASE_URL}/parks/{args.park_code}/tickets/{args.ticket_type}/query",
            json=data
        )
        result = response.json()
        print(f"基础票价: {format_price(result['base_price'])}")
        if result['discounts']:
            print("优惠信息:")
            for discount in result['discounts']:
                print(f"  - {discount}")
        print(f"最终票价: {format_price(result['final_price'])}")
        print(f"保险费: {format_price(result['insurance_fee'])}")
        print(f"总计: {format_price(result['total_price'])}")
        print(f"说明: {result['message']}")
    except requests.exceptions.RequestException as e:
        print(f"错误: {e}")


def cmd_purchase_ticket(args):
    print_separator()
    print(f"购买 {args.ticket_type} 票")
    print_separator()
    try:
        data = {
            "visitor_name": getattr(args, 'visitor_name', None),
            "visitor_age": getattr(args, 'age', None),
            "visitor_height": getattr(args, 'height', None),
            "is_student": getattr(args, 'student', False),
            "is_group": getattr(args, 'group', False),
            "group_size": getattr(args, 'group_size', 1),
            "free_children_count": getattr(args, 'free_children', 0),
            "route_id": getattr(args, 'route_id', None)
        }
        response = requests.post(
            f"{BASE_URL}/parks/{args.park_code}/tickets/{args.ticket_type}/purchase",
            json=data
        )
        if response.status_code != 200:
            print(f"错误: {response.json().get('detail', '购票失败')}")
            return
        ticket = response.json()
        print("购票成功!")
        print(f"门票ID: {ticket['id']}")
        print(f"票种: {ticket['ticket_type_name']}")
        print(f"价格: {format_price(ticket['final_price'])}")
        print(f"保险费: {format_price(ticket['insurance_fee'])}")
        print(f"状态: {ticket['status']}")
    except requests.exceptions.RequestException as e:
        print(f"错误: {e}")


def cmd_refund_ticket(args):
    print_separator()
    print(f"退票 (门票ID: {args.ticket_id})")
    print_separator()
    try:
        params = {"ticket_id": args.ticket_id}
        response = requests.post(
            f"{BASE_URL}/parks/{args.park_code}/tickets/{args.ticket_type}/refund",
            params=params
        )
        result = response.json()
        if response.status_code != 200:
            print(f"错误: {result.get('detail', '退票失败')}")
            return
        print(result.get('message', '退票成功'))
    except requests.exceptions.RequestException as e:
        print(f"错误: {e}")


def cmd_list_routes(args):
    print_separator()
    print(f"景区 {args.park_code} 的游线列表")
    print_separator()
    try:
        params = {}
        if args.active:
            params["active_only"] = "true"
        response = requests.get(f"{BASE_URL}/parks/{args.park_code}/routes", params=params)
        if response.status_code != 200:
            print(f"错误: {response.json().get('detail', '查询失败')}")
            return
        routes = response.json()
        if not routes:
            print("暂无游线")
            return
        difficulty_names = {"easy": "简单", "normal": "普通", "challenging": "挑战"}
        for route in routes:
            print(f"游线ID: {route['route_id']}")
            print(f"名称: {route['name']}")
            print(f"难度: {difficulty_names.get(route['difficulty'], route['difficulty'])}")
            print(f"容量: {route['current_visitors']}/{route['capacity']}")
            print(f"预计时长: {route['duration_hours']} 小时")
            print(f"基础保险费: {format_price(route['base_insurance_fee'])}")
            print(f"状态: {'激活' if route['is_active'] else '停用'}")
            print("-" * 60)
    except requests.exceptions.RequestException as e:
        print(f"错误: {e}")


def main():
    global BASE_URL
    original_base_url = BASE_URL

    parser = argparse.ArgumentParser(description="景区综合管理系统客户端")
    parser.add_argument("--url", default=original_base_url, help="API服务器地址")

    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    list_parks_parser = subparsers.add_parser("list-parks", help="列出所有景区")
    list_parks_parser.set_defaults(func=cmd_list_parks)

    list_tickets_parser = subparsers.add_parser("list-tickets", help="列出门票")
    list_tickets_parser.add_argument("park_code", help="景区代码")
    list_tickets_parser.add_argument("--status", choices=["valid", "refunded"], help="按状态筛选")
    list_tickets_parser.set_defaults(func=cmd_list_tickets)

    list_guides_parser = subparsers.add_parser("list-guides", help="列出导游")
    list_guides_parser.add_argument("park_code", help="景区代码")
    list_guides_parser.add_argument("--available", action="store_true", help="仅显示可用导游")
    list_guides_parser.set_defaults(func=cmd_list_guides)

    list_routes_parser = subparsers.add_parser("list-routes", help="列出游线")
    list_routes_parser.add_argument("park_code", help="景区代码")
    list_routes_parser.add_argument("--active", action="store_true", help="仅显示激活的游线")
    list_routes_parser.set_defaults(func=cmd_list_routes)

    query_parser = subparsers.add_parser("query-price", help="查询门票价格")
    query_parser.add_argument("park_code", help="景区代码")
    query_parser.add_argument("ticket_type", help="票种名称")
    query_parser.add_argument("--age", type=int, help="游客年龄")
    query_parser.add_argument("--height", type=float, help="游客身高(米)")
    query_parser.add_argument("--student", action="store_true", help="是否为学生")
    query_parser.add_argument("--group", action="store_true", help="是否为团队")
    query_parser.add_argument("--group-size", type=int, default=1, help="团队人数")
    query_parser.add_argument("--free-children", type=int, default=0, help="免票儿童数")
    query_parser.add_argument("--route-id", type=int, help="游线ID")
    query_parser.set_defaults(func=cmd_query_price)

    purchase_parser = subparsers.add_parser("purchase", help="购买门票")
    purchase_parser.add_argument("park_code", help="景区代码")
    purchase_parser.add_argument("ticket_type", help="票种名称")
    purchase_parser.add_argument("--visitor-name", help="游客姓名")
    purchase_parser.add_argument("--age", type=int, help="游客年龄")
    purchase_parser.add_argument("--height", type=float, help="游客身高(米)")
    purchase_parser.add_argument("--student", action="store_true", help="是否为学生")
    purchase_parser.add_argument("--group", action="store_true", help="是否为团队")
    purchase_parser.add_argument("--group-size", type=int, default=1, help="团队人数")
    purchase_parser.add_argument("--free-children", type=int, default=0, help="免票儿童数")
    purchase_parser.add_argument("--route-id", type=int, help="游线ID")
    purchase_parser.set_defaults(func=cmd_purchase_ticket)

    refund_parser = subparsers.add_parser("refund", help="退票")
    refund_parser.add_argument("park_code", help="景区代码")
    refund_parser.add_argument("ticket_type", help="票种名称")
    refund_parser.add_argument("ticket_id", type=int, help="门票ID")
    refund_parser.set_defaults(func=cmd_refund_ticket)

    args = parser.parse_args()

    BASE_URL = args.url

    if not args.command:
        parser.print_help()
        sys.exit(1)

    args.func(args)


if __name__ == "__main__":
    main()
