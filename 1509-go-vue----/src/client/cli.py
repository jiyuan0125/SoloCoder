import os
import json
import sys
from datetime import datetime
from typing import Optional

import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get("API_URL", "http://localhost:8000")

    def _request(self, method: str, endpoint: str, **kwargs):
        url = f"{self.base_url}{endpoint}"
        try:
            response = requests.request(method, url, **kwargs)
            response.raise_for_status()
            return response
        except requests.ConnectionError:
            print(f"错误: 无法连接到服务端 {self.base_url}")
            sys.exit(1)
        except requests.HTTPError as e:
            if e.response.status_code == 400:
                detail = e.response.json().get("detail", str(e))
                print(f"错误: {detail}")
            elif e.response.status_code == 404:
                print(f"错误: 资源不存在")
            else:
                print(f"错误: HTTP {e.response.status_code}")
            sys.exit(1)

    def health_check(self):
        return self._request("GET", "/").json()

    def create_supply_chain(self, data: dict):
        return self._request("POST", "/api/supply-chain", json=data).json()

    def get_supply_chain(self, batch_number: str):
        return self._request("GET", f"/api/supply-chain/{batch_number}").json()

    def create_inspection(self, data: dict):
        return self._request("POST", "/api/inspection", json=data).json()

    def get_inspection(self, batch_number: str):
        return self._request("GET", f"/api/inspection/{batch_number}").json()

    def initiate_recall(self, data: dict):
        return self._request("POST", "/api/recall", json=data).json()

    def list_recalls(self):
        return self._request("GET", "/api/recall").json()

    def get_recall(self, recall_id: str):
        return self._request("GET", f"/api/recall/{recall_id}").json()

    def get_recall_todos(self, recall_id: str):
        return self._request("GET", f"/api/recall/{recall_id}/todos").json()

    def complete_recall(self, recall_id: str):
        return self._request("POST", f"/api/recall/{recall_id}/complete").json()

    def list_todos(self):
        return self._request("GET", "/api/todos").json()

    def update_todo(self, todo_id: str, data: dict):
        return self._request("PUT", f"/api/todos/{todo_id}", json=data).json()

    def export_traceability(self, batch_number: str):
        return self._request("GET", f"/api/export/{batch_number}").json()

    def export_to_text(self, batch_number: str):
        response = self._request("GET", f"/api/export/{batch_number}/text")
        return response.text


def print_json(data):
    print(json.dumps(data, ensure_ascii=False, indent=2))


def cmd_health(args):
    client = APIClient()
    result = client.health_check()
    print_json(result)


def cmd_supply_chain_add(args):
    client = APIClient()
    data = {
        "batch_number": args.batch_number,
        "stage": args.stage,
        "operation_time": args.operation_time,
        "operator": args.operator,
    }
    if args.location:
        data["location"] = args.location
    if args.notes:
        data["notes"] = args.notes
    result = client.create_supply_chain(data)
    print("供应链记录已创建:")
    print_json(result)


def cmd_supply_chain_list(args):
    client = APIClient()
    result = client.get_supply_chain(args.batch_number)
    if not result:
        print("暂无供应链记录")
        return
    print(f"批次号 {args.batch_number} 的供应链时间线:")
    for i, record in enumerate(result, 1):
        print(f"\n  环节 {i}: {record['stage']}")
        print(f"    操作时间: {record['operation_time']}")
        print(f"    操作人: {record['operator']}")
        if record.get('location'):
            print(f"    地点: {record['location']}")
        if record.get('notes'):
            print(f"    备注: {record['notes']}")


def cmd_inspection_add(args):
    client = APIClient()
    data = {
        "batch_number": args.batch_number,
        "inspector": args.inspector,
        "inspection_time": args.inspection_time,
        "status": args.status,
        "report": args.report,
    }
    if args.item:
        data["item"] = args.item
    result = client.create_inspection(data)
    print("检测记录已创建:")
    print_json(result)


def cmd_inspection_list(args):
    client = APIClient()
    result = client.get_inspection(args.batch_number)
    if not result:
        print("暂无检测记录")
        return
    print(f"批次号 {args.batch_number} 的检测记录:")
    for i, record in enumerate(result, 1):
        print(f"\n  检测 {i}:")
        print(f"    检测时间: {record['inspection_time']}")
        print(f"    检测人: {record['inspector']}")
        print(f"    结果: {record['status']}")
        if record.get('item'):
            print(f"    项目: {record['item']}")
        print(f"    报告: {record['report']}")


def cmd_recall_initiate(args):
    client = APIClient()
    data = {
        "batch_number": args.batch_number,
        "reason": args.reason,
        "initiator": args.initiator,
    }
    result = client.initiate_recall(data)
    print("召回已发起:")
    print_json(result)


def cmd_recall_list(args):
    client = APIClient()
    result = client.list_recalls()
    if not result:
        print("暂无召回记录")
        return
    print("召回记录列表:")
    for recall in result:
        print(f"\n  召回ID: {recall['id']}")
        print(f"    批次号: {recall['batch_number']}")
        print(f"    状态: {recall['status']}")
        print(f"    原因: {recall['reason']}")
        print(f"    可追踪: {'是' if recall['can_track'] else '否'}")
        print(f"    发起时间: {recall['start_time']}")


def cmd_recall_show(args):
    client = APIClient()
    result = client.get_recall(args.recall_id)
    print_json(result)


def cmd_recall_todos(args):
    client = APIClient()
    result = client.get_recall_todos(args.recall_id)
    if not result:
        print("暂无待办任务")
        return
    print(f"召回 {args.recall_id} 的待办任务:")
    for todo in result:
        status_text = {
            'pending': '待处理',
            'completed': '已完成',
            'overdue': '已逾期',
        }.get(todo['status'], todo['status'])
        print(f"\n  任务ID: {todo['id']}")
        print(f"    环节: {todo['stage']}")
        print(f"    负责人: {todo['handler']}")
        print(f"    状态: {status_text}")
        print(f"    创建时间: {todo['created_at']}")


def cmd_recall_complete(args):
    client = APIClient()
    result = client.complete_recall(args.recall_id)
    print("召回已完成:")
    print_json(result)


def cmd_todo_list(args):
    client = APIClient()
    result = client.list_todos()
    if not result:
        print("暂无待办任务")
        return
    print("所有待办任务:")
    for todo in result:
        status_text = {
            'pending': '待处理',
            'completed': '已完成',
            'overdue': '已逾期',
        }.get(todo['status'], todo['status'])
        print(f"\n  任务ID: {todo['id']}")
        print(f"    批次号: {todo['batch_number']}")
        print(f"    环节: {todo['stage']}")
        print(f"    负责人: {todo['handler']}")
        print(f"    状态: {status_text}")


def cmd_todo_update(args):
    client = APIClient()
    data = {
        "status": args.status,
    }
    if args.processed_by:
        data["processed_by"] = args.processed_by
    if args.notes:
        data["notes"] = args.notes
    result = client.update_todo(args.todo_id, data)
    print("待办任务已更新:")
    print_json(result)


def cmd_export_json(args):
    client = APIClient()
    result = client.export_traceability(args.batch_number)
    print_json(result)


def cmd_export_text(args):
    client = APIClient()
    result = client.export_to_text(args.batch_number)
    if args.output:
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write(result)
        print(f"报告已保存到: {args.output}")
    else:
        print(result)


def main():
    import argparse

    parser = argparse.ArgumentParser(
        prog="traceability-cli",
        description="食品安全溯源管理系统命令行客户端",
    )
    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    subparsers.add_parser("health", help="检查服务端健康状态")

    sc_parser = subparsers.add_parser("supply-chain", help="供应链管理")
    sc_sub = sc_parser.add_subparsers(dest="subcommand")

    sc_add = sc_sub.add_parser("add", help="添加供应链记录")
    sc_add.add_argument("--batch-number", required=True, help="批次号")
    sc_add.add_argument("--stage", required=True, 
                       choices=["raw_material", "processing", "packaging", 
                                "distribution", "retail", "on_shelf"],
                       help="环节类型")
    sc_add.add_argument("--operation-time", required=True, 
                       help="操作时间 (ISO格式, 如: 2024-01-15T10:30:00)")
    sc_add.add_argument("--operator", required=True, help="操作人")
    sc_add.add_argument("--location", help="地点")
    sc_add.add_argument("--notes", help="备注")

    sc_list = sc_sub.add_parser("list", help="查询供应链时间线")
    sc_list.add_argument("--batch-number", required=True, help="批次号")

    in_parser = subparsers.add_parser("inspection", help="检测管理")
    in_sub = in_parser.add_subparsers(dest="subcommand")

    in_add = in_sub.add_parser("add", help="添加检测记录")
    in_add.add_argument("--batch-number", required=True, help="批次号")
    in_add.add_argument("--inspector", required=True, help="检测人")
    in_add.add_argument("--inspection-time", required=True, 
                       help="检测时间 (ISO格式)")
    in_add.add_argument("--status", required=True, 
                       choices=["pending", "passed", "failed"],
                       help="检测结果")
    in_add.add_argument("--report", required=True, help="检测报告")
    in_add.add_argument("--item", help="检测项目")

    in_list = in_sub.add_parser("list", help="查询检测记录")
    in_list.add_argument("--batch-number", required=True, help="批次号")

    rc_parser = subparsers.add_parser("recall", help="召回管理")
    rc_sub = rc_parser.add_subparsers(dest="subcommand")

    rc_init = rc_sub.add_parser("initiate", help="发起召回")
    rc_init.add_argument("--batch-number", required=True, help="批次号")
    rc_init.add_argument("--reason", required=True, help="召回原因")
    rc_init.add_argument("--initiator", required=True, help="发起人")

    rc_sub.add_parser("list", help="列出所有召回")

    rc_show = rc_sub.add_parser("show", help="查看召回详情")
    rc_show.add_argument("--recall-id", required=True, help="召回ID")

    rc_todos = rc_sub.add_parser("todos", help="查看召回待办")
    rc_todos.add_argument("--recall-id", required=True, help="召回ID")

    rc_complete = rc_sub.add_parser("complete", help="完成召回")
    rc_complete.add_argument("--recall-id", required=True, help="召回ID")

    td_parser = subparsers.add_parser("todo", help="待办管理")
    td_sub = td_parser.add_subparsers(dest="subcommand")

    td_sub.add_parser("list", help="列出所有待办")

    td_update = td_sub.add_parser("update", help="更新待办状态")
    td_update.add_argument("--todo-id", required=True, help="待办ID")
    td_update.add_argument("--status", required=True,
                          choices=["pending", "completed", "overdue"],
                          help="状态")
    td_update.add_argument("--processed-by", help="处理人")
    td_update.add_argument("--notes", help="处理备注")

    ex_parser = subparsers.add_parser("export", help="数据导出")
    ex_sub = ex_parser.add_subparsers(dest="subcommand")

    ex_json = ex_sub.add_parser("json", help="导出为JSON")
    ex_json.add_argument("--batch-number", required=True, help="批次号")

    ex_text = ex_sub.add_parser("text", help="导出为文本")
    ex_text.add_argument("--batch-number", required=True, help="批次号")
    ex_text.add_argument("--output", "-o", help="输出文件路径")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return

    if args.command == "health":
        cmd_health(args)
    elif args.command == "supply-chain":
        if args.subcommand == "add":
            cmd_supply_chain_add(args)
        elif args.subcommand == "list":
            cmd_supply_chain_list(args)
        else:
            sc_parser.print_help()
    elif args.command == "inspection":
        if args.subcommand == "add":
            cmd_inspection_add(args)
        elif args.subcommand == "list":
            cmd_inspection_list(args)
        else:
            in_parser.print_help()
    elif args.command == "recall":
        if args.subcommand == "initiate":
            cmd_recall_initiate(args)
        elif args.subcommand == "list":
            cmd_recall_list(args)
        elif args.subcommand == "show":
            cmd_recall_show(args)
        elif args.subcommand == "todos":
            cmd_recall_todos(args)
        elif args.subcommand == "complete":
            cmd_recall_complete(args)
        else:
            rc_parser.print_help()
    elif args.command == "todo":
        if args.subcommand == "list":
            cmd_todo_list(args)
        elif args.subcommand == "update":
            cmd_todo_update(args)
        else:
            td_parser.print_help()
    elif args.command == "export":
        if args.subcommand == "json":
            cmd_export_json(args)
        elif args.subcommand == "text":
            cmd_export_text(args)
        else:
            ex_parser.print_help()


if __name__ == "__main__":
    main()
