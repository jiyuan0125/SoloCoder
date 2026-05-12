#!/usr/bin/env python3
import argparse
import os
import json
import httpx
from datetime import date, time, datetime

DEFAULT_PORT = int(os.getenv("PORT", 8000))
_DEFAULT_BASE_URL = f"http://localhost:{DEFAULT_PORT}"
_current_base_url = _DEFAULT_BASE_URL


def get_client():
    return httpx.Client(base_url=_current_base_url, timeout=10.0)


def set_base_url(port: int):
    global _current_base_url
    _current_base_url = f"http://localhost:{port}"


def print_table(data, headers=None):
    if not data:
        print("无数据")
        return

    if not headers:
        headers = list(data[0].keys()) if data else []

    col_widths = [len(h) for h in headers]
    for row in data:
        for i, h in enumerate(headers):
            val = str(row.get(h, ""))
            col_widths[i] = max(col_widths[i], len(val))

    separator = "+-" + "-+-".join("-" * w for w in col_widths) + "-+"
    header_row = "| " + " | ".join(f"{h:<{w}}" for h, w in zip(headers, col_widths)) + " |"

    print(separator)
    print(header_row)
    print(separator)

    for row in data:
        row_str = "| " + " | ".join(f"{str(row.get(h, '')):<{w}}" for h, w in zip(headers, col_widths)) + " |"
        print(row_str)

    print(separator)


def cmd_list_collections(args):
    with get_client() as client:
        params = {}
        if args.status:
            params["status"] = args.status
        if args.level:
            params["level"] = args.level
        if args.name:
            params["name"] = args.name

        response = client.get("/api/collections/", params=params)
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return

        collections = response.json()
        if not collections:
            print("未找到藏品")
            return

        table_data = []
        for c in collections:
            table_data.append({
                "ID": c["id"],
                "名称": c["name"],
                "级别": c["level"],
                "状态": c["status"],
                "位置": c.get("location", ""),
            })

        print_table(table_data, ["ID", "名称", "级别", "状态", "位置"])


def cmd_show_collection(args):
    with get_client() as client:
        response = client.get(f"/api/collections/{args.id}")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return

        c = response.json()
        print(f"\n藏品详情:")
        print(f"  ID: {c['id']}")
        print(f"  名称: {c['name']}")
        print(f"  级别: {c['level']}")
        print(f"  状态: {c['status']}")
        print(f"  位置: {c.get('location', '未设置')}")
        print(f"  描述: {c.get('description', '无')}")
        print(f"  创建时间: {c['created_at']}")
        print(f"  更新时间: {c['updated_at']}")

        if c.get("inventory_logs"):
            print(f"\n  出入库记录:")
            for log in c["inventory_logs"]:
                print(f"    - {log['created_at']}: {log['operation_type']} (操作人: {log['operator']})")


def cmd_list_exhibitions(args):
    with get_client() as client:
        params = {}
        if args.status:
            params["status"] = args.status
        if args.name:
            params["name"] = args.name

        response = client.get("/api/exhibitions/", params=params)
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return

        exhibitions = response.json()
        if not exhibitions:
            print("未找到展览")
            return

        table_data = []
        for e in exhibitions:
            table_data.append({
                "ID": e["id"],
                "名称": e["name"],
                "状态": e["status"],
                "开始日期": e.get("start_date", ""),
                "结束日期": e.get("end_date", ""),
                "位置": e.get("location", ""),
            })

        print_table(table_data, ["ID", "名称", "状态", "开始日期", "结束日期", "位置"])


def cmd_show_exhibition(args):
    with get_client() as client:
        response = client.get(f"/api/exhibitions/{args.id}")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return

        e = response.json()
        print(f"\n展览详情:")
        print(f"  ID: {e['id']}")
        print(f"  名称: {e['name']}")
        print(f"  状态: {e['status']}")
        print(f"  位置: {e.get('location', '未设置')}")
        print(f"  开始日期: {e.get('start_date', '未设置')}")
        print(f"  结束日期: {e.get('end_date', '未设置')}")
        print(f"  描述: {e.get('description', '无')}")

        if e.get("items"):
            print(f"\n  展出藏品 ({len(e['items'])} 件):")
            for item in e["items"]:
                print(f"    - {item['collection_id']}: {item['collection_name']}")


def cmd_list_reservations(args):
    with get_client() as client:
        params = {}
        if args.date:
            params["visit_date"] = args.date
        if args.id_number:
            params["id_number"] = args.id_number
        if args.status:
            params["status"] = args.status

        response = client.get("/api/reservations/", params=params)
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return

        reservations = response.json()
        if not reservations:
            print("未找到预约")
            return

        table_data = []
        for r in reservations:
            table_data.append({
                "ID": r["id"],
                "姓名": r["visitor_name"],
                "身份证号": r["id_number"],
                "日期": r["visit_date"],
                "时间段": r["time_slot"],
                "人数": r["visitor_count"],
                "状态": r["status"],
            })

        print_table(table_data, ["ID", "姓名", "身份证号", "日期", "时间段", "人数", "状态"])


def cmd_overdue_borrowings(args):
    with get_client() as client:
        response = client.get("/api/borrowings/overdue/today")
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return

        data = response.json()
        print(f"今天: {data['today']}")
        print(f"逾期外借数量: {data['overdue_count']}")

        if data["overdue_items"]:
            table_data = []
            for item in data["overdue_items"]:
                table_data.append({
                    "ID": item["id"],
                    "藏品ID": item["collection_id"],
                    "藏品名称": item.get("collection_name", ""),
                    "借方": item["borrower"],
                    "预计归还": item["expected_return_date"],
                    "逾期天数": item["overdue_days"],
                })
            print_table(table_data, ["ID", "藏品ID", "藏品名称", "借方", "预计归还", "逾期天数"])


def cmd_todos(args):
    with get_client() as client:
        params = {}
        if args.completed is not None:
            params["is_completed"] = args.completed

        response = client.get("/api/todos/", params=params)
        if response.status_code != 200:
            print(f"错误: {response.status_code} - {response.text}")
            return

        todos = response.json()
        if not todos:
            print("无待办事项")
            return

        table_data = []
        for t in todos:
            table_data.append({
                "ID": t["id"],
                "借方": t.get("borrower", ""),
                "藏品": t.get("collection_name", ""),
                "逾期天数": t["overdue_days"],
                "已完成": "是" if t["is_completed"] else "否",
                "创建时间": t["created_at"][:19],
            })

        print_table(table_data, ["ID", "借方", "藏品", "逾期天数", "已完成", "创建时间"])


def cmd_generate_reminders(args):
    from app.database import SessionLocal
    from app.models import Borrowing, ReminderTodo, Collection
    from datetime import date as date_obj

    db = SessionLocal()
    try:
        today = date_obj.today()
        overdue = db.query(Borrowing).filter(
            Borrowing.status == "外借中",
            Borrowing.expected_return_date < today
        ).all()

        count = 0
        for b in overdue:
            existing = db.query(ReminderTodo).filter(
                ReminderTodo.borrowing_id == b.id,
                ReminderTodo.is_completed == False
            ).first()

            if existing:
                continue

            collection = db.query(Collection).filter(Collection.id == b.collection_id).first()
            overdue_days = (today - b.expected_return_date).days

            todo = ReminderTodo(
                borrowing_id=b.id,
                overdue_days=overdue_days,
                message=f"藏品【{collection.name if collection else '未知'}】已逾期 {overdue_days} 天，借方: {b.borrower}"
            )
            db.add(todo)
            count += 1

        db.commit()
        print(f"已生成 {count} 条催还待办")
    finally:
        db.close()


def main():
    parser = argparse.ArgumentParser(description="博物馆数字化管理系统 CLI")
    parser.add_argument("--port", type=int, default=DEFAULT_PORT, help=f"API 端口 (默认: {DEFAULT_PORT})")

    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    list_coll = subparsers.add_parser("list-collections", help="列出藏品")
    list_coll.add_argument("--status", help="按状态筛选 (在库/展出/外借/修复中/已注销)")
    list_coll.add_argument("--level", help="按级别筛选 (一级/二级/三级/一般)")
    list_coll.add_argument("--name", help="按名称模糊搜索")

    show_coll = subparsers.add_parser("show-collection", help="显示藏品详情")
    show_coll.add_argument("id", type=int, help="藏品 ID")

    list_exh = subparsers.add_parser("list-exhibitions", help="列出展览")
    list_exh.add_argument("--status", help="按状态筛选 (策划/布展/展出/撤展/已完成)")
    list_exh.add_argument("--name", help="按名称模糊搜索")

    show_exh = subparsers.add_parser("show-exhibition", help="显示展览详情")
    show_exh.add_argument("id", type=int, help="展览 ID")

    list_res = subparsers.add_parser("list-reservations", help="列出参观预约")
    list_res.add_argument("--date", help="按日期筛选 (YYYY-MM-DD)")
    list_res.add_argument("--id-number", help="按身份证号筛选")
    list_res.add_argument("--status", help="按状态筛选")

    subparsers.add_parser("overdue-borrowings", help="查看逾期外借")

    list_todos = subparsers.add_parser("list-todos", help="列出待办事项")
    list_todos.add_argument("--completed", type=bool, help="是否已完成 (true/false)")

    subparsers.add_parser("generate-reminders", help="生成催还待办 (无需 API 运行)")

    args = parser.parse_args()

    if args.port and args.port != DEFAULT_PORT:
        set_base_url(args.port)

    if args.command == "list-collections":
        cmd_list_collections(args)
    elif args.command == "show-collection":
        cmd_show_collection(args)
    elif args.command == "list-exhibitions":
        cmd_list_exhibitions(args)
    elif args.command == "show-exhibition":
        cmd_show_exhibition(args)
    elif args.command == "list-reservations":
        cmd_list_reservations(args)
    elif args.command == "overdue-borrowings":
        cmd_overdue_borrowings(args)
    elif args.command == "list-todos":
        cmd_todos(args)
    elif args.command == "generate-reminders":
        cmd_generate_reminders(args)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
