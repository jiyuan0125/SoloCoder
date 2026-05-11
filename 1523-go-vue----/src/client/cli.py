import argparse
import json
import os
import sys
from datetime import date

from .api_client import APIClient


def format_json(data):
    return json.dumps(data, indent=2, ensure_ascii=False)


def cmd_equipment(args, client: APIClient):
    if args.action == "create":
        result = client.create_equipment(args.id, args.name, args.type)
        print(format_json(result))
    elif args.action == "list":
        result = client.list_equipment()
        print(format_json(result))
    elif args.action == "get":
        result = client.get_equipment(args.id)
        print(format_json(result))
    elif args.action == "status":
        result = client.update_equipment_status(args.id, args.status)
        print(format_json(result))


def cmd_crushing(args, client: APIClient):
    if args.action == "create":
        result = client.create_crushing_record(
            args.crusher_id,
            args.date,
            args.feed_size,
            args.product_size,
            args.throughput,
        )
        print(format_json(result))
    elif args.action == "list":
        result = client.list_crushing_records(
            args.start_date, args.end_date, args.crusher_id
        )
        print(format_json(result))


def cmd_flotation(args, client: APIClient):
    if args.action == "create":
        result = client.create_flotation_record(
            args.date,
            args.feed_grade,
            args.concentrate_grade,
            args.tailings_grade,
            args.recovery,
        )
        print(format_json(result))
    elif args.action == "list":
        result = client.list_flotation_records(
            args.start_date, args.end_date
        )
        print(format_json(result))


def cmd_calculate(args, client: APIClient):
    result = client.calculate_grade(
        args.feed_grade,
        args.concentrate_grade,
        args.tailings_grade,
        args.feed_throughput,
    )
    print(format_json(result))


def cmd_metrics(args, client: APIClient):
    if args.type == "daily":
        result = client.get_daily_metrics(args.date)
        print(format_json(result))
    elif args.type == "equipment":
        result = client.get_equipment_utilization(args.date)
        print(format_json(result))


def create_parser():
    parser = argparse.ArgumentParser(
        prog="client",
        description="选矿厂工艺管理系统命令行客户端",
    )
    parser.add_argument(
        "--server",
        default=os.getenv("SERVER_URL", "http://localhost:8000"),
        help="服务端地址 (默认: http://localhost:8000)",
    )
    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    equipment_parser = subparsers.add_parser("equipment", help="设备管理")
    equipment_sub = equipment_parser.add_subparsers(dest="action", help="操作")

    equip_create = equipment_sub.add_parser("create", help="创建设备")
    equip_create.add_argument("--id", required=True, help="设备ID")
    equip_create.add_argument("--name", required=True, help="设备名称")
    equip_create.add_argument(
        "--type", required=True, choices=["crusher", "flotation"],
        help="设备类型"
    )

    equipment_sub.add_parser("list", help="列出所有设备")

    equip_get = equipment_sub.add_parser("get", help="获取设备信息")
    equip_get.add_argument("--id", required=True, help="设备ID")

    equip_status = equipment_sub.add_parser("status", help="更新设备状态")
    equip_status.add_argument("--id", required=True, help="设备ID")
    equip_status.add_argument(
        "--status", required=True,
        choices=["running", "stopped", "maintenance"],
        help="设备状态"
    )

    crushing_parser = subparsers.add_parser("crushing", help="破碎记录管理")
    crushing_sub = crushing_parser.add_subparsers(dest="action", help="操作")

    crush_create = crushing_sub.add_parser("create", help="创建破碎记录")
    crush_create.add_argument("--crusher-id", required=True, help="破碎机ID")
    crush_create.add_argument("--date", required=True, help="记录日期 (YYYY-MM-DD)")
    crush_create.add_argument("--feed-size", type=float, required=True, help="破碎前粒度")
    crush_create.add_argument("--product-size", type=float, required=True, help="破碎后粒度")
    crush_create.add_argument("--throughput", type=float, required=True, help="处理量")

    crush_list = crushing_sub.add_parser("list", help="列出破碎记录")
    crush_list.add_argument("--start-date", help="开始日期 (YYYY-MM-DD)")
    crush_list.add_argument("--end-date", help="结束日期 (YYYY-MM-DD)")
    crush_list.add_argument("--crusher-id", help="破碎机ID")

    flotation_parser = subparsers.add_parser("flotation", help="浮选记录管理")
    flotation_sub = flotation_parser.add_subparsers(dest="action", help="操作")

    float_create = flotation_sub.add_parser("create", help="创建浮选记录")
    float_create.add_argument("--date", required=True, help="记录日期 (YYYY-MM-DD)")
    float_create.add_argument("--feed-grade", type=float, required=True, help="原矿品位")
    float_create.add_argument("--concentrate-grade", type=float, required=True, help="精矿品位")
    float_create.add_argument("--tailings-grade", type=float, required=True, help="尾矿品位")
    float_create.add_argument("--recovery", type=float, required=True, help="回收率 (%)")

    float_list = flotation_sub.add_parser("list", help="列出浮选记录")
    float_list.add_argument("--start-date", help="开始日期 (YYYY-MM-DD)")
    float_list.add_argument("--end-date", help="结束日期 (YYYY-MM-DD)")

    calc_parser = subparsers.add_parser("calculate", help="品位计算")
    calc_parser.add_argument("--feed-grade", type=float, required=True, help="原矿品位")
    calc_parser.add_argument("--concentrate-grade", type=float, required=True, help="精矿品位")
    calc_parser.add_argument("--tailings-grade", type=float, required=True, help="尾矿品位")
    calc_parser.add_argument("--feed-throughput", type=float, help="给矿处理量")

    metrics_parser = subparsers.add_parser("metrics", help="指标聚合")
    metrics_parser.add_argument(
        "--type", required=True,
        choices=["daily", "equipment"],
        help="指标类型"
    )
    metrics_parser.add_argument(
        "--date", default=date.today().isoformat(),
        help="日期 (YYYY-MM-DD), 默认今天"
    )

    return parser


def main():
    parser = create_parser()
    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return 1

    client = APIClient(args.server)

    try:
        if args.command == "equipment":
            if not args.action:
                parser.parse_args(["equipment", "--help"])
                return 1
            cmd_equipment(args, client)
        elif args.command == "crushing":
            if not args.action:
                parser.parse_args(["crushing", "--help"])
                return 1
            cmd_crushing(args, client)
        elif args.command == "flotation":
            if not args.action:
                parser.parse_args(["flotation", "--help"])
                return 1
            cmd_flotation(args, client)
        elif args.command == "calculate":
            cmd_calculate(args, client)
        elif args.command == "metrics":
            cmd_metrics(args, client)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
