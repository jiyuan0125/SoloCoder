from __future__ import annotations

import argparse
import sys
from typing import Any, Dict, List

from .api_client import ColdChainClient


def format_table(headers: List[str], rows: List[List[str]]) -> str:
    if not rows:
        return "无数据\n"

    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            col_widths[i] = max(col_widths[i], len(str(cell)))

    def format_line(values: List[str]) -> str:
        return "| " + " | ".join(
            f"{str(v):<{w}}" for v, w in zip(values, col_widths)
        ) + " |"

    separator = "+-" + "-+-".join("-" * w for w in col_widths) + "-+"

    output = [separator]
    output.append(format_line(headers))
    output.append(separator)
    for row in rows:
        output.append(format_row(row, col_widths))
    output.append(separator)

    return "\n".join(output) + "\n"


def format_row(values: List[str], col_widths: List[int]) -> str:
    return "| " + " | ".join(
        f"{str(v):<{w}}" for v, w in zip(values, col_widths)
    ) + " |"


def cmd_health(client: ColdChainClient, args: argparse.Namespace) -> None:
    result = client.health()
    print(f"服务状态: {result.get('status', 'unknown')}")


def cmd_locations(client: ColdChainClient, args: argparse.Namespace) -> None:
    locations = client.list_locations()
    headers = ["ID", "名称", "纬度", "经度", "类型"]
    rows = [
        [
            loc["id"],
            loc["name"],
            f"{loc['latitude']:.4f}",
            f"{loc['longitude']:.4f}",
            "产地" if loc["is_origin"] else "配送点",
        ]
        for loc in locations
    ]
    print(format_table(headers, rows))


def cmd_vehicles(client: ColdChainClient, args: argparse.Namespace) -> None:
    vehicles = client.list_vehicles()
    headers = ["ID", "编号", "当前温度(℃)", "载重上限(kg)", "当前载重(kg)", "状态"]
    rows = [
        [
            v["id"],
            v["vehicle_number"],
            f"{v['current_temperature']:.1f}",
            f"{v['max_load']:.0f}",
            f"{v['current_load']:.0f}",
            v["status"],
        ]
        for v in vehicles
    ]
    print(format_table(headers, rows))


def cmd_tasks(client: ColdChainClient, args: argparse.Namespace) -> None:
    tasks = client.list_tasks(status=args.status)
    headers = ["ID", "订单ID", "起点", "终点", "重量(kg)", "温度范围", "状态"]
    rows = []
    for t in tasks:
        temp_range = f"{t['required_min_temp']:.1f}~{t['required_max_temp']:.1f}"
        rows.append([
            t["id"],
            t["order_id"],
            t["origin_id"],
            t["destination_id"],
            f"{t['weight']:.0f}",
            temp_range,
            t["status"],
        ])
    print(format_table(headers, rows))


def cmd_create_task(client: ColdChainClient, args: argparse.Namespace) -> None:
    task = {
        "id": args.id,
        "order_id": args.order_id,
        "origin_id": args.origin,
        "destination_id": args.destination,
        "weight": args.weight,
        "required_min_temp": args.min_temp,
        "required_max_temp": args.max_temp,
    }
    result = client.create_task(task)
    print(f"任务创建成功: {result['id']}")
    print(f"  订单ID: {result['order_id']}")
    print(f"  起点: {result['origin_id']}")
    print(f"  终点: {result['destination_id']}")
    print(f"  重量: {result['weight']} kg")
    print(f"  温度要求: {result['required_min_temp']}~{result['required_max_temp']} ℃")
    print(f"  状态: {result['status']}")


def cmd_assign_task(client: ColdChainClient, args: argparse.Namespace) -> None:
    try:
        result = client.assign_task(args.task_id)
        print(f"任务分配成功:")
        print(f"  任务ID: {result['task_id']}")
        print(f"  车辆ID: {result['vehicle_id']}")
        print(f"  车辆编号: {result['vehicle_number']}")
        print(f"  状态: {result['status']}")
    except Exception as e:
        print(f"分配失败: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_start_task(client: ColdChainClient, args: argparse.Namespace) -> None:
    try:
        result = client.start_task(args.task_id)
        print(f"任务已开始: {result['id']}")
        print(f"  状态: {result['status']}")
    except Exception as e:
        print(f"启动失败: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_complete_task(client: ColdChainClient, args: argparse.Namespace) -> None:
    try:
        result = client.complete_task(args.task_id)
        print(f"任务已完成: {result['id']}")
        print(f"  最终状态: {result['status']}")
        print(f"  完成时间: {result['completed_at']}")
    except Exception as e:
        print(f"完成失败: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_report_temp(client: ColdChainClient, args: argparse.Namespace) -> None:
    try:
        result = client.report_temperature(
            args.task_id, args.vehicle_id, args.temperature
        )
        status = "正常" if result["is_normal"] else "异常"
        print(f"温度上报:")
        print(f"  温度: {result['temperature']:.1f} ℃")
        print(f"  状态: {status}")
        print(f"  上报时间: {result['reported_at']}")
    except Exception as e:
        print(f"上报失败: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_report_loc(client: ColdChainClient, args: argparse.Namespace) -> None:
    try:
        result = client.report_location(
            args.task_id, args.vehicle_id, args.latitude, args.longitude
        )
        print(f"位置上报:")
        print(f"  纬度: {result['latitude']}")
        print(f"  经度: {result['longitude']}")
        print(f"  上报时间: {result['reported_at']}")
    except Exception as e:
        print(f"上报失败: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_check_comm(client: ColdChainClient, args: argparse.Namespace) -> None:
    result = client.check_communication()
    if result:
        print(f"发现 {len(result)} 个通讯中断的任务:")
        headers = ["任务ID", "订单ID", "车辆ID", "状态"]
        rows = [
            [t["task_id"], t["order_id"], t["vehicle_id"], t["status"]]
            for t in result
        ]
        print(format_table(headers, rows))
    else:
        print("所有任务通讯正常。")


def cmd_export(client: ColdChainClient, args: argparse.Namespace) -> None:
    result = client.export_delivery_records(args.start_date, args.end_date)
    print(result)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="coldchain",
        description="冷链物流系统命令行客户端",
    )
    parser.add_argument(
        "--url",
        default=None,
        help="服务端地址 (默认: http://localhost:8000 或 COLD_CHAIN_BASE_URL 环境变量)",
    )

    subparsers = parser.add_subparsers(dest="command", required=True)

    subparsers.add_parser("health", help="检查服务端健康状态")
    subparsers.add_parser("locations", help="列出所有位置")
    subparsers.add_parser("vehicles", help="列出所有车辆")

    tasks_parser = subparsers.add_parser("tasks", help="列出配送任务")
    tasks_parser.add_argument(
        "--status",
        choices=["pending", "assigned", "in_transit", "completed", "temperature_abnormal", "communication_lost"],
        default=None,
        help="按状态筛选",
    )

    create_parser = subparsers.add_parser("create-task", help="创建配送任务")
    create_parser.add_argument("--id", required=True, help="任务ID")
    create_parser.add_argument("--order-id", required=True, help="订单ID")
    create_parser.add_argument("--origin", required=True, help="起点位置ID")
    create_parser.add_argument("--destination", required=True, help="终点位置ID")
    create_parser.add_argument("--weight", type=float, required=True, help="重量(kg)")
    create_parser.add_argument("--min-temp", type=float, required=True, help="最低温度要求(℃)")
    create_parser.add_argument("--max-temp", type=float, required=True, help="最高温度要求(℃)")

    assign_parser = subparsers.add_parser("assign-task", help="分配任务到车辆")
    assign_parser.add_argument("task_id", help="任务ID")

    start_parser = subparsers.add_parser("start-task", help="开始配送")
    start_parser.add_argument("task_id", help="任务ID")

    complete_parser = subparsers.add_parser("complete-task", help="完成配送")
    complete_parser.add_argument("task_id", help="任务ID")

    temp_parser = subparsers.add_parser("report-temp", help="上报温度")
    temp_parser.add_argument("task_id", help="任务ID")
    temp_parser.add_argument("vehicle_id", help="车辆ID")
    temp_parser.add_argument("temperature", type=float, help="车厢温度(℃)")

    loc_parser = subparsers.add_parser("report-loc", help="上报位置")
    loc_parser.add_argument("task_id", help="任务ID")
    loc_parser.add_argument("vehicle_id", help="车辆ID")
    loc_parser.add_argument("latitude", type=float, help="纬度")
    loc_parser.add_argument("longitude", type=float, help="经度")

    subparsers.add_parser("check-comm", help="检查通讯状态(24小时无上报标记为中断)")

    export_parser = subparsers.add_parser("export", help="导出配送记录(纯文本表格)")
    export_parser.add_argument("--start", required=True, help="开始日期 (YYYY-MM-DD)")
    export_parser.add_argument("--end", required=True, help="结束日期 (YYYY-MM-DD)")

    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()

    commands = {
        "health": cmd_health,
        "locations": cmd_locations,
        "vehicles": cmd_vehicles,
        "tasks": cmd_tasks,
        "create-task": cmd_create_task,
        "assign-task": cmd_assign_task,
        "start-task": cmd_start_task,
        "complete-task": cmd_complete_task,
        "report-temp": cmd_report_temp,
        "report-loc": cmd_report_loc,
        "check-comm": cmd_check_comm,
        "export": cmd_export,
    }

    with ColdChainClient(base_url=args.url) as client:
        if args.command == "export":
            result = client.export_delivery_records(args.start, args.end)
            print(result)
        else:
            commands[args.command](client, args)


if __name__ == "__main__":
    main()
