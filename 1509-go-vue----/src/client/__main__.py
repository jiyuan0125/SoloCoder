import os
import sys
import json
import argparse
from datetime import datetime

from client.api_client import TraceabilityAPIClient
from client.formatters import (
    format_supply_chain,
    format_inspection,
    format_recall,
    format_todo,
    format_timeline
)


def get_client() -> TraceabilityAPIClient:
    base_url = os.environ.get("SERVER_URL", "http://localhost:8000")
    return TraceabilityAPIClient(base_url)


def cmd_status(args):
    client = get_client()
    result = client.get_status()
    print(f"{result['name']} v{result['version']} - {result['status']}")


def cmd_supply_add(args):
    client = get_client()
    try:
        record = client.add_supply_chain_record(
            batch_number=args.batch,
            stage=args.stage,
            operation_time=args.time or datetime.now(),
            operator=args.operator,
            location=args.location,
            remarks=args.remarks
        )
        print(f"供应链记录已创建: {record['id']}")
        print(format_supply_chain(record))
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_supply_list(args):
    client = get_client()
    records = client.list_supply_chain()
    if not records:
        print("暂无供应链记录")
        return
    for record in records:
        print(format_supply_chain(record))
        print("-" * 40)


def cmd_supply_timeline(args):
    client = get_client()
    records = client.get_supply_chain_timeline(args.batch)
    if not records:
        print(f"批次 {args.batch} 暂无供应链记录")
        return
    print(f"\n批次 {args.batch} 供应链流转时间线")
    print("=" * 60)
    print(format_timeline(records))


def cmd_inspection_add(args):
    client = get_client()
    try:
        record = client.add_inspection_record(
            batch_number=args.batch,
            inspector=args.inspector,
            status=args.status,
            items=args.items.split(",") if args.items else [],
            report=args.report,
            remarks=args.remarks
        )
        print(f"检测记录已创建: {record['id']}")
        print(format_inspection(record))
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_inspection_list(args):
    client = get_client()
    if args.batch:
        records = client.get_inspections_by_batch(args.batch)
        print(f"\n批次 {args.batch} 检测记录:")
    else:
        records = client.list_inspections()
        print("\n所有检测记录:")
    if not records:
        print("  暂无记录")
        return
    print("=" * 60)
    for record in records:
        print(format_inspection(record))
        print("-" * 40)


def cmd_recall_create(args):
    client = get_client()
    try:
        recall = client.create_recall(args.batch, args.reason)
        print(f"召回已创建: {recall['id']}")
        if not recall['can_track']:
            print("警告: 召回范围为空，标记为无法追踪")
        print(format_recall(recall))
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_recall_list(args):
    client = get_client()
    recalls = client.list_recalls(args.batch)
    if not recalls:
        print("暂无召回记录")
        return
    for recall in recalls:
        print(format_recall(recall))
        print("-" * 40)


def cmd_recall_complete(args):
    client = get_client()
    try:
        recall = client.complete_recall(args.id, args.user)
        print(f"召回已完成: {recall['id']}")
        print(format_recall(recall))
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_recall_cancel(args):
    client = get_client()
    try:
        recall = client.cancel_recall(args.id, args.user)
        print(f"召回已取消: {recall['id']}")
        print(format_recall(recall))
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_todo_list(args):
    client = get_client()
    if args.recall_id:
        todos = client.get_recall_todos(args.recall_id)
        print(f"\n召回 {args.recall_id} 的待办事项:")
    else:
        todos = client.list_todos()
        print("\n所有待办事项:")
    if not todos:
        print("  暂无待办")
        return
    print("=" * 60)
    for todo in todos:
        print(format_todo(todo))
        print("-" * 40)


def cmd_todo_confirm(args):
    client = get_client()
    try:
        todo = client.confirm_todo(args.id, args.user, args.remarks)
        print(f"待办已确认: {todo['id']}")
        print(format_todo(todo))
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_export(args):
    client = get_client()
    report = client.export_batch_traceability(args.batch)
    if args.output:
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write(report)
        print(f"报告已保存到: {args.output}")
    else:
        print(report)


def main():
    parser = argparse.ArgumentParser(
        prog="tracecli",
        description="食品安全溯源管理系统命令行客户端"
    )
    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    subparsers.add_parser("status", help="查看服务状态")

    supply_parser = subparsers.add_parser("supply", help="供应链管理")
    supply_sub = supply_parser.add_subparsers(dest="subcommand")

    supply_add = supply_sub.add_parser("add", help="添加供应链记录")
    supply_add.add_argument("--batch", required=True, help="批次号")
    supply_add.add_argument("--stage", required=True, 
                           choices=["raw_material", "production", "processing", 
                                   "packaging", "warehouse", "distribution", "retail"],
                           help="环节类型")
    supply_add.add_argument("--time", type=datetime.fromisoformat, 
                           help="操作时间 (ISO格式)")
    supply_add.add_argument("--operator", required=True, help="操作人")
    supply_add.add_argument("--location", help="地点")
    supply_add.add_argument("--remarks", help="备注")

    supply_sub.add_parser("list", help="列出所有供应链记录")

    supply_timeline = supply_sub.add_parser("timeline", help="查看批次时间线")
    supply_timeline.add_argument("--batch", required=True, help="批次号")

    inspection_parser = subparsers.add_parser("inspection", help="检测管理")
    inspection_sub = inspection_parser.add_subparsers(dest="subcommand")

    inspection_add = inspection_sub.add_parser("add", help="添加检测记录")
    inspection_add.add_argument("--batch", required=True, help="批次号")
    inspection_add.add_argument("--inspector", required=True, help="检测人")
    inspection_add.add_argument("--status", required=True, 
                               choices=["pending", "passed", "failed"],
                               help="检测结果")
    inspection_add.add_argument("--items", help="检测项目 (逗号分隔)")
    inspection_add.add_argument("--report", help="检测报告")
    inspection_add.add_argument("--remarks", help="备注")

    inspection_list = inspection_sub.add_parser("list", help="列出检测记录")
    inspection_list.add_argument("--batch", help="按批次筛选")

    recall_parser = subparsers.add_parser("recall", help="召回管理")
    recall_sub = recall_parser.add_subparsers(dest="subcommand")

    recall_create = recall_sub.add_parser("create", help="发起召回")
    recall_create.add_argument("--batch", required=True, help="批次号")
    recall_create.add_argument("--reason", required=True, help="召回原因")

    recall_list = recall_sub.add_parser("list", help="列出召回记录")
    recall_list.add_argument("--batch", help="按批次筛选")

    recall_complete = recall_sub.add_parser("complete", help="完成召回")
    recall_complete.add_argument("--id", required=True, help="召回ID")
    recall_complete.add_argument("--user", required=True, help="操作人")

    recall_cancel = recall_sub.add_parser("cancel", help="取消召回")
    recall_cancel.add_argument("--id", required=True, help="召回ID")
    recall_cancel.add_argument("--user", required=True, help="操作人")

    todo_parser = subparsers.add_parser("todo", help="待办管理")
    todo_sub = todo_parser.add_subparsers(dest="subcommand")

    todo_list = todo_sub.add_parser("list", help="列出待办事项")
    todo_list.add_argument("--recall-id", help="按召回ID筛选")

    todo_confirm = todo_sub.add_parser("confirm", help="确认待办事项")
    todo_confirm.add_argument("--id", required=True, help="待办ID")
    todo_confirm.add_argument("--user", required=True, help="操作人")
    todo_confirm.add_argument("--remarks", help="处理说明")

    export_parser = subparsers.add_parser("export", help="导出溯源报告")
    export_parser.add_argument("--batch", required=True, help="批次号")
    export_parser.add_argument("--output", "-o", help="输出文件路径")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return

    if args.command == "status":
        cmd_status(args)
    elif args.command == "supply":
        if args.subcommand == "add":
            cmd_supply_add(args)
        elif args.subcommand == "list":
            cmd_supply_list(args)
        elif args.subcommand == "timeline":
            cmd_supply_timeline(args)
        else:
            supply_parser.print_help()
    elif args.command == "inspection":
        if args.subcommand == "add":
            cmd_inspection_add(args)
        elif args.subcommand == "list":
            cmd_inspection_list(args)
        else:
            inspection_parser.print_help()
    elif args.command == "recall":
        if args.subcommand == "create":
            cmd_recall_create(args)
        elif args.subcommand == "list":
            cmd_recall_list(args)
        elif args.subcommand == "complete":
            cmd_recall_complete(args)
        elif args.subcommand == "cancel":
            cmd_recall_cancel(args)
        else:
            recall_parser.print_help()
    elif args.command == "todo":
        if args.subcommand == "list":
            cmd_todo_list(args)
        elif args.subcommand == "confirm":
            cmd_todo_confirm(args)
        else:
            todo_parser.print_help()
    elif args.command == "export":
        cmd_export(args)


if __name__ == "__main__":
    main()
