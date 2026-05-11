import argparse
import json
import sys
from typing import Optional

from .api_client import APIClient


def _print_json(data):
    print(json.dumps(data, ensure_ascii=False, indent=2))


def cmd_create_rider(args):
    client = APIClient()
    result = client.create_rider(args.rider_id)
    print(f"已创建骑手: {result['id']}")
    _print_json(result)


def cmd_list_riders(args):
    client = APIClient()
    riders = client.list_riders(args.status)
    print(f"共 {len(riders)} 个骑手:")
    for rider in riders:
        print(f"  {rider['id']}: {rider['status']}")
    if args.detail:
        _print_json(riders)


def cmd_get_rider(args):
    client = APIClient()
    rider = client.get_rider(args.rider_id)
    _print_json(rider)


def cmd_heartbeat(args):
    client = APIClient()
    result = client.heartbeat(args.rider_id)
    print(f"骑手 {args.rider_id} 心跳已更新")
    _print_json(result)


def cmd_set_status(args):
    client = APIClient()
    result = client.set_rider_status(args.rider_id, args.status)
    print(f"骑手 {args.rider_id} 状态已设置为 {args.status}")
    _print_json(result)


def cmd_accept_order(args):
    client = APIClient()
    result = client.accept_order(args.rider_id, args.order_id)
    print(f"骑手 {args.rider_id} 已接单: {args.order_id}")
    _print_json(result)


def cmd_create_order(args):
    client = APIClient()
    result = client.create_order()
    print(f"已创建订单: {result['id']}")
    _print_json(result)


def cmd_list_orders(args):
    client = APIClient()
    orders = client.list_orders(args.status)
    print(f"共 {len(orders)} 个订单:")
    for order in orders:
        rider = order.get('assigned_rider_id', '-')
        print(f"  {order['id']}: {order['status']} (骑手: {rider})")
    if args.detail:
        _print_json(orders)


def cmd_get_order(args):
    client = APIClient()
    order = client.get_order(args.order_id)
    _print_json(order)


def cmd_complete_order(args):
    client = APIClient()
    result = client.complete_order(args.order_id)
    print(f"订单 {args.order_id} 已完成")
    _print_json(result)


def cmd_reassign_order(args):
    client = APIClient()
    result = client.reassign_order(args.order_id, args.rider_id)
    print(f"订单 {args.order_id} 已重新分配给骑手 {args.rider_id}")
    _print_json(result)


def cmd_metrics(args):
    client = APIClient()
    metrics = client.get_metrics()
    print("=" * 50)
    print("  指标看板")
    print("=" * 50)
    print(f"在线骑手数:    {metrics['online_riders']}")
    print(f"待接单数:      {metrics['pending_orders']}")
    print(f"配送中订单数:  {metrics['delivering_orders']}")
    print(f"今日完成数:    {metrics['today_completed']}")
    print(f"平均配送时长:  {metrics['avg_delivery_time_minutes']} 分钟")
    print(f"超时率:        {metrics['timeout_rate'] * 100:.2f}%")
    print("=" * 50)


def main():
    parser = argparse.ArgumentParser(
        prog="client",
        description="外卖平台配送调度系统命令行客户端"
    )
    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    rider_parser = subparsers.add_parser("rider", help="骑手管理")
    rider_sub = rider_parser.add_subparsers(dest="subcommand")
    
    create_rider = rider_sub.add_parser("create", help="创建骑手")
    create_rider.add_argument("rider_id", help="骑手编号")
    create_rider.set_defaults(func=cmd_create_rider)
    
    list_riders = rider_sub.add_parser("list", help="列出所有骑手")
    list_riders.add_argument("--status", help="按状态过滤: idle/delivering/resting/offline")
    list_riders.add_argument("--detail", action="store_true", help="显示详细信息")
    list_riders.set_defaults(func=cmd_list_riders)
    
    get_rider = rider_sub.add_parser("get", help="获取骑手信息")
    get_rider.add_argument("rider_id", help="骑手编号")
    get_rider.set_defaults(func=cmd_get_rider)
    
    heartbeat = rider_sub.add_parser("heartbeat", help="骑手心跳（更新位置时间）")
    heartbeat.add_argument("rider_id", help="骑手编号")
    heartbeat.set_defaults(func=cmd_heartbeat)
    
    set_status = rider_sub.add_parser("set-status", help="设置骑手状态")
    set_status.add_argument("rider_id", help="骑手编号")
    set_status.add_argument("status", choices=["idle", "delivering", "resting"], help="状态")
    set_status.set_defaults(func=cmd_set_status)
    
    accept_order = rider_sub.add_parser("accept-order", help="骑手接单")
    accept_order.add_argument("rider_id", help="骑手编号")
    accept_order.add_argument("order_id", help="订单编号")
    accept_order.set_defaults(func=cmd_accept_order)

    order_parser = subparsers.add_parser("order", help="订单管理")
    order_sub = order_parser.add_subparsers(dest="subcommand")
    
    create_order = order_sub.add_parser("create", help="创建订单")
    create_order.set_defaults(func=cmd_create_order)
    
    list_orders = order_sub.add_parser("list", help="列出所有订单")
    list_orders.add_argument("--status", help="按状态过滤: pending/assigned/delivering/completed/timed_out")
    list_orders.add_argument("--detail", action="store_true", help="显示详细信息")
    list_orders.set_defaults(func=cmd_list_orders)
    
    get_order = order_sub.add_parser("get", help="获取订单信息")
    get_order.add_argument("order_id", help="订单编号")
    get_order.set_defaults(func=cmd_get_order)
    
    complete_order = order_sub.add_parser("complete", help="完成订单")
    complete_order.add_argument("order_id", help="订单编号")
    complete_order.set_defaults(func=cmd_complete_order)
    
    reassign_order = order_sub.add_parser("reassign", help="管理员手动重新分配订单")
    reassign_order.add_argument("order_id", help="订单编号")
    reassign_order.add_argument("rider_id", help="骑手编号")
    reassign_order.set_defaults(func=cmd_reassign_order)

    metrics_parser = subparsers.add_parser("metrics", help="查看指标看板")
    metrics_parser.set_defaults(func=cmd_metrics)

    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    if hasattr(args, 'func'):
        try:
            args.func(args)
        except Exception as e:
            print(f"错误: {e}", file=sys.stderr)
            sys.exit(1)
    else:
        if args.command == "rider":
            rider_parser.print_help()
        elif args.command == "order":
            order_parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()
