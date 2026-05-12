#!/usr/bin/env python3
import os
import json
import httpx
from datetime import datetime

API_BASE = os.environ.get("PORT_API_URL", "http://localhost:8000/api/v1")

STATUS_MAP = {
    "expected": "预计到港",
    "arrived": "已到港",
    "at_anchor": "锚泊等待",
    "docked": "已靠泊",
    "departed": "已离港"
}

OP_STATUS_MAP = {
    "pending": "待处理",
    "waiting": "等待中",
    "in_progress": "进行中",
    "completed": "已完成"
}


class PortClient:
    def __init__(self, base_url: str = API_BASE):
        self.base_url = base_url
        self.client = httpx.Client(timeout=30.0)
    
    def close(self):
        self.client.close()
    
    def _request(self, method: str, endpoint: str, **kwargs):
        url = f"{self.base_url}{endpoint}"
        try:
            response = self.client.request(method, url, **kwargs)
            response.raise_for_status()
            if response.content:
                return response.json()
            return {}
        except httpx.HTTPStatusError as e:
            error_detail = e.response.json().get("detail", str(e)) if e.response.content else str(e)
            print(f"[错误] HTTP {e.response.status_code}: {error_detail}")
            raise SystemExit(1)
        except Exception as e:
            print(f"[错误] 连接失败: {str(e)}")
            raise SystemExit(1)
    
    def create_berth(self, name, berth_type, capacity, maintenance=False):
        data = {
            "name": name,
            "berth_type": berth_type,
            "capacity": capacity,
            "is_under_maintenance": maintenance
        }
        return self._request("POST", "/berths/", json=data)
    
    def list_berths(self, available_only=None):
        params = {}
        if available_only is not None:
            params["available_only"] = str(available_only).lower()
        return self._request("GET", "/berths/", params=params)
    
    def get_berth(self, berth_id):
        return self._request("GET", f"/berths/{berth_id}")
    
    def update_berth(self, berth_id, **kwargs):
        return self._request("PUT", f"/berths/{berth_id}", json=kwargs)
    
    def delete_berth(self, berth_id):
        return self._request("DELETE", f"/berths/{berth_id}")
    
    def create_ship(self, name, ship_type, draught, expected_arrival, imo_number=None):
        if isinstance(expected_arrival, str):
            expected_arrival = datetime.fromisoformat(expected_arrival)
        data = {
            "name": name,
            "imo_number": imo_number,
            "ship_type": ship_type,
            "draught": draught,
            "expected_arrival": expected_arrival.isoformat()
        }
        return self._request("POST", "/ships/", json=data)
    
    def list_ships(self, status=None):
        params = {}
        if status:
            params["status"] = status
        return self._request("GET", "/ships/", params=params)
    
    def get_ship(self, ship_id):
        return self._request("GET", f"/ships/{ship_id}")
    
    def update_ship(self, ship_id, **kwargs):
        return self._request("PUT", f"/ships/{ship_id}", json=kwargs)
    
    def delete_ship(self, ship_id):
        return self._request("DELETE", f"/ships/{ship_id}")
    
    def get_ship_schedule(self, ship_id):
        return self._request("GET", f"/ships/{ship_id}/schedule")
    
    def create_operation(self, ship_id, cargo_volume, priority=1, berth_id=None, yard_area_id=None, description=None):
        data = {
            "ship_id": ship_id,
            "priority": priority,
            "cargo_volume": cargo_volume,
            "description": description,
            "berth_id": berth_id,
            "yard_area_id": yard_area_id
        }
        return self._request("POST", "/operations/", json=data)
    
    def list_operations(self, status=None, ship_id=None, berth_id=None):
        params = {}
        if status:
            params["status"] = status
        if ship_id:
            params["ship_id"] = ship_id
        if berth_id:
            params["berth_id"] = berth_id
        return self._request("GET", "/operations/", params=params)
    
    def get_operation(self, operation_id):
        return self._request("GET", f"/operations/{operation_id}")
    
    def update_operation(self, operation_id, **kwargs):
        return self._request("PUT", f"/operations/{operation_id}", json=kwargs)
    
    def delete_operation(self, operation_id):
        return self._request("DELETE", f"/operations/{operation_id}")
    
    def complete_operation(self, operation_id):
        return self._request("PUT", f"/operations/{operation_id}", json={"status": "completed"})
    
    def create_yard_area(self, name, total_capacity, available=True):
        data = {
            "name": name,
            "total_capacity": total_capacity,
            "is_available": available
        }
        return self._request("POST", "/yard_areas/", json=data)
    
    def list_yard_areas(self, available_only=None):
        params = {}
        if available_only is not None:
            params["available_only"] = str(available_only).lower()
        return self._request("GET", "/yard_areas/", params=params)
    
    def get_yard_area(self, yard_id):
        return self._request("GET", f"/yard_areas/{yard_id}")
    
    def update_yard_area(self, yard_id, **kwargs):
        return self._request("PUT", f"/yard_areas/{yard_id}", json=kwargs)
    
    def delete_yard_area(self, yard_id):
        return self._request("DELETE", f"/yard_areas/{yard_id}")
    
    def mark_yard_unavailable(self, yard_id):
        return self.update_yard_area(yard_id, is_available=False)
    
    def mark_yard_available(self, yard_id):
        return self.update_yard_area(yard_id, is_available=True)


def format_datetime(dt_str):
    if not dt_str:
        return "-"
    try:
        dt = datetime.fromisoformat(dt_str.replace("Z", "+00:00"))
        return dt.strftime("%Y-%m-%d %H:%M:%S")
    except:
        return dt_str


def print_table(headers, rows):
    if not rows:
        print("(无数据)")
        return
    
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            if i < len(col_widths):
                col_widths[i] = max(col_widths[i], len(str(cell)))
    
    separator = "+" + "+".join("-" * (w + 2) for w in col_widths) + "+"
    header_row = "|" + "|".join(f" {h:<{w}} " for h, w in zip(headers, col_widths)) + "|"
    
    print(separator)
    print(header_row)
    print(separator)
    
    for row in rows:
        print("|" + "|".join(f" {str(cell):<{w}} " for cell, w in zip(row, col_widths)) + "|")
    
    print(separator)


def cmd_berths(args, client):
    if args.action == "list":
        berths = client.list_berths(available_only=args.available)
        rows = []
        for b in berths:
            rows.append([
                b["id"],
                b["name"],
                b["berth_type"],
                f"{b['capacity']}m",
                "是" if b.get("is_available", True) else "否",
                "是" if b.get("is_under_maintenance") else "否"
            ])
        print_table(["ID", "名称", "类型", "靠泊能力", "可用", "维护中"], rows)
    
    elif args.action == "create":
        result = client.create_berth(args.name, args.type, args.capacity, args.maintenance)
        print(f"[成功] 泊位已创建: ID={result['id']} 名称={result['name']}")
    
    elif args.action == "update":
        updates = {}
        if args.name is not None:
            updates["name"] = args.name
        if args.type is not None:
            updates["berth_type"] = args.type
        if args.capacity is not None:
            updates["capacity"] = args.capacity
        if args.maintenance is not None:
            updates["is_under_maintenance"] = args.maintenance
        
        if not updates:
            print("[错误] 请指定要更新的字段")
            return
        
        result = client.update_berth(args.id, **updates)
        print(f"[成功] 泊位已更新: ID={result['id']}")
    
    elif args.action == "delete":
        result = client.delete_berth(args.id)
        print(f"[成功] {result['message']}")
    
    elif args.action == "show":
        b = client.get_berth(args.id)
        print(f"\n泊位详情:")
        print(f"  ID: {b['id']}")
        print(f"  名称: {b['name']}")
        print(f"  类型: {b['berth_type']}")
        print(f"  靠泊能力: {b['capacity']}m")
        print(f"  可用: {'是' if b.get('is_available', True) else '否'}")
        print(f"  维护中: {'是' if b.get('is_under_maintenance') else '否'}")
        print(f"  创建时间: {format_datetime(b['created_at'])}")


def cmd_ships(args, client):
    if args.action == "list":
        ships = client.list_ships(status=args.status)
        rows = []
        for s in ships:
            rows.append([
                s["id"],
                s["name"],
                s["ship_type"],
                f"{s['draught']}m",
                STATUS_MAP.get(s["status"], s["status"]),
                format_datetime(s["expected_arrival"]),
                s.get("current_berth_id", "-") or "-"
            ])
        print_table(["ID", "名称", "类型", "吃水", "状态", "预计到港", "当前泊位"], rows)
    
    elif args.action == "create":
        result = client.create_ship(
            args.name, args.type, args.draught,
            args.expected_arrival, args.imo
        )
        print(f"[成功] 船舶已创建: ID={result['id']} 名称={result['name']}")
    
    elif args.action == "update":
        updates = {}
        if args.name is not None:
            updates["name"] = args.name
        if args.imo is not None:
            updates["imo_number"] = args.imo
        if args.type is not None:
            updates["ship_type"] = args.type
        if args.draught is not None:
            updates["draught"] = args.draught
        if args.status is not None:
            updates["status"] = args.status
        if args.expected_arrival is not None:
            updates["expected_arrival"] = args.expected_arrival
        
        if not updates:
            print("[错误] 请指定要更新的字段")
            return
        
        result = client.update_ship(args.id, **updates)
        print(f"[成功] 船舶已更新: ID={result['id']} 状态={STATUS_MAP.get(result['status'], result['status'])}")
    
    elif args.action == "delete":
        result = client.delete_ship(args.id)
        print(f"[成功] {result['message']}")
    
    elif args.action == "show":
        s = client.get_ship(args.id)
        print(f"\n船舶详情:")
        print(f"  ID: {s['id']}")
        print(f"  名称: {s['name']}")
        print(f"  IMO编号: {s.get('imo_number') or '-'}")
        print(f"  类型: {s['ship_type']}")
        print(f"  吃水深度: {s['draught']}m")
        print(f"  当前状态: {STATUS_MAP.get(s['status'], s['status'])}")
        print(f"  预计到港: {format_datetime(s['expected_arrival'])}")
        print(f"  当前泊位ID: {s.get('current_berth_id') or '-'}")
    
    elif args.action == "schedule":
        schedule = client.get_ship_schedule(args.id)
        
        print(f"\n=== 调度推荐 ===")
        print(f"\n推荐泊位:")
        if schedule["recommended_berths"]:
            rows = []
            for b in schedule["recommended_berths"]:
                rows.append([
                    b["id"],
                    b["name"],
                    b["berth_type"],
                    f"{b['capacity']}m"
                ])
            print_table(["ID", "名称", "类型", "靠泊能力"], rows)
        else:
            print("  暂无可用泊位，船舶将进入等待队列")
        
        print(f"\n等待队列 (按优先级排序):")
        if schedule["waiting_queue"]:
            rows = []
            for op in schedule["waiting_queue"]:
                rows.append([
                    op["id"],
                    f"船舶#{op['ship_id']}",
                    op["priority"],
                    f"{op['cargo_volume']}",
                    OP_STATUS_MAP.get(op["status"], op["status"]),
                    format_datetime(op["created_at"])
                ])
            print_table(["操作ID", "船舶", "优先级", "货量", "状态", "创建时间"], rows)
        else:
            print("  等待队列为空")


def cmd_operations(args, client):
    if args.action == "list":
        ops = client.list_operations(status=args.status, ship_id=args.ship, berth_id=args.berth)
        rows = []
        for op in ops:
            rows.append([
                op["id"],
                op["ship_id"],
                op.get("berth_id", "-") or "-",
                op.get("yard_area_id", "-") or "-",
                op["priority"],
                f"{op['cargo_volume']}",
                OP_STATUS_MAP.get(op["status"], op["status"]),
                format_datetime(op["started_at"]),
                format_datetime(op["completed_at"])
            ])
        print_table(["ID", "船舶ID", "泊位ID", "堆场ID", "优先级", "货量", "状态", "开始", "完成"], rows)
    
    elif args.action == "create":
        result = client.create_operation(
            args.ship, args.cargo_volume,
            priority=args.priority,
            berth_id=args.berth,
            yard_area_id=args.yard,
            description=args.description
        )
        print(f"[成功] 装卸作业已创建: ID={result['id']} 状态={OP_STATUS_MAP.get(result['status'], result['status'])}")
    
    elif args.action == "complete":
        result = client.complete_operation(args.id)
        print(f"[成功] 装卸作业已完成: ID={result['id']}")
    
    elif args.action == "delete":
        result = client.delete_operation(args.id)
        print(f"[成功] {result['message']}")
    
    elif args.action == "show":
        op = client.get_operation(args.id)
        print(f"\n装卸作业详情:")
        print(f"  ID: {op['id']}")
        print(f"  船舶ID: {op['ship_id']}")
        print(f"  泊位ID: {op.get('berth_id') or '-'}")
        print(f"  堆场ID: {op.get('yard_area_id') or '-'}")
        print(f"  优先级: {op['priority']}")
        print(f"  货量: {op['cargo_volume']}")
        print(f"  状态: {OP_STATUS_MAP.get(op['status'], op['status'])}")
        print(f"  描述: {op.get('description') or '-'}")
        print(f"  开始时间: {format_datetime(op.get('started_at'))}")
        print(f"  完成时间: {format_datetime(op.get('completed_at'))}")


def cmd_yards(args, client):
    if args.action == "list":
        yards = client.list_yard_areas(available_only=args.available)
        rows = []
        for y in yards:
            remaining = y.get("remaining_capacity", y["total_capacity"])
            usage = y.get("current_usage", 0)
            rows.append([
                y["id"],
                y["name"],
                f"{y['total_capacity']}",
                f"{usage}",
                f"{remaining}",
                "是" if y["is_available"] else "否"
            ])
        print_table(["ID", "名称", "总容量", "已用", "剩余", "可用"], rows)
    
    elif args.action == "create":
        result = client.create_yard_area(args.name, args.capacity, not args.unavailable)
        print(f"[成功] 堆场已创建: ID={result['id']} 名称={result['name']}")
    
    elif args.action == "update":
        updates = {}
        if args.name is not None:
            updates["name"] = args.name
        if args.capacity is not None:
            updates["total_capacity"] = args.capacity
        if args.available is not None:
            updates["is_available"] = args.available
        
        if not updates:
            print("[错误] 请指定要更新的字段")
            return
        
        result = client.update_yard_area(args.id, **updates)
        print(f"[成功] 堆场已更新: ID={result['id']}")
    
    elif args.action == "delete":
        result = client.delete_yard_area(args.id)
        print(f"[成功] {result['message']}")
    
    elif args.action == "show":
        y = client.get_yard_area(args.id)
        remaining = y.get("remaining_capacity", y["total_capacity"])
        usage = y.get("current_usage", 0)
        print(f"\n堆场详情:")
        print(f"  ID: {y['id']}")
        print(f"  名称: {y['name']}")
        print(f"  总容量: {y['total_capacity']}")
        print(f"  已用容量: {usage}")
        print(f"  剩余容量: {remaining}")
        print(f"  可用: {'是' if y['is_available'] else '否'}")


def cmd_help(args, client):
    help_text = """
港口泊位与装卸调度系统 - 命令行客户端

使用方法:
  python client.py <命令> <子命令> [选项]

命令:
  berths       泊位管理
  ships        船舶管理
  operations   装卸作业管理
  yards        堆场管理

子命令示例:
  berths list                    列出所有泊位
  berths create --name B01 --type container --capacity 12.5
  ships list                     列出所有船舶
  ships create --name OCEAN --type container --draught 10.0 --expected 2026-05-15T08:00:00
  operations list                列出所有作业
  operations create --ship 1 --cargo 500 --priority 2
  yards list                     列出所有堆场
  yards create --name YARD-A --capacity 10000

环境变量:
  PORT_API_URL    API 地址 (默认: http://localhost:8000/api/v1)

详细帮助:
  python client.py berths --help
  python client.py ships --help
  python client.py operations --help
  python client.py yards --help
"""
    print(help_text)


def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="港口泊位与装卸调度系统客户端")
    subparsers = parser.add_subparsers(dest="command", help="命令")
    
    berth_parser = subparsers.add_parser("berths", help="泊位管理")
    berth_sub = berth_parser.add_subparsers(dest="action", help="操作")
    
    blist = berth_sub.add_parser("list", help="列出泊位")
    blist.add_argument("--available", type=lambda x: x.lower() == "true", help="仅显示可用/不可用泊位")
    
    bcreate = berth_sub.add_parser("create", help="创建泊位")
    bcreate.add_argument("--name", required=True, help="泊位名称")
    bcreate.add_argument("--type", required=True, help="泊位类型 (如 container, bulk)")
    bcreate.add_argument("--capacity", type=float, required=True, help="靠泊能力(米)")
    bcreate.add_argument("--maintenance", action="store_true", help="是否处于维护中")
    
    bupdate = berth_sub.add_parser("update", help="更新泊位")
    bupdate.add_argument("--id", type=int, required=True, help="泊位ID")
    bupdate.add_argument("--name", help="新名称")
    bupdate.add_argument("--type", help="新类型")
    bupdate.add_argument("--capacity", type=float, help="新靠泊能力")
    bupdate.add_argument("--maintenance", type=lambda x: x.lower() == "true", help="是否维护中")
    
    bdelete = berth_sub.add_parser("delete", help="删除泊位")
    bdelete.add_argument("--id", type=int, required=True, help="泊位ID")
    
    bshow = berth_sub.add_parser("show", help="查看泊位详情")
    bshow.add_argument("--id", type=int, required=True, help="泊位ID")
    
    ship_parser = subparsers.add_parser("ships", help="船舶管理")
    ship_sub = ship_parser.add_subparsers(dest="action", help="操作")
    
    slist = ship_sub.add_parser("list", help="列出船舶")
    slist.add_argument("--status", help="按状态过滤")
    
    screate = ship_sub.add_parser("create", help="创建船舶")
    screate.add_argument("--name", required=True, help="船舶名称")
    screate.add_argument("--imo", help="IMO编号")
    screate.add_argument("--type", required=True, help="船舶类型")
    screate.add_argument("--draught", type=float, required=True, help="吃水深度(米)")
    screate.add_argument("--expected", required=True, help="预计到港时间 (ISO格式)")
    
    supdate = ship_sub.add_parser("update", help="更新船舶")
    supdate.add_argument("--id", type=int, required=True, help="船舶ID")
    supdate.add_argument("--name", help="新名称")
    supdate.add_argument("--imo", help="新IMO编号")
    supdate.add_argument("--type", help="新类型")
    supdate.add_argument("--draught", type=float, help="新吃水深度")
    supdate.add_argument("--status", help="新状态 (expected/arrived/at_anchor/docked/departed)")
    supdate.add_argument("--expected", help="新预计到港时间")
    
    sdelete = ship_sub.add_parser("delete", help="删除船舶")
    sdelete.add_argument("--id", type=int, required=True, help="船舶ID")
    
    sshow = ship_sub.add_parser("show", help="查看船舶详情")
    sshow.add_argument("--id", type=int, required=True, help="船舶ID")
    
    sched = ship_sub.add_parser("schedule", help="查看调度推荐")
    sched.add_argument("--id", type=int, required=True, help="船舶ID")
    
    op_parser = subparsers.add_parser("operations", help="装卸作业管理")
    op_sub = op_parser.add_subparsers(dest="action", help="操作")
    
    olist = op_sub.add_parser("list", help="列出作业")
    olist.add_argument("--status", help="按状态过滤")
    olist.add_argument("--ship", type=int, help="按船舶ID过滤")
    olist.add_argument("--berth", type=int, help="按泊位ID过滤")
    
    ocreate = op_sub.add_parser("create", help="创建装卸作业")
    ocreate.add_argument("--ship", type=int, required=True, help="船舶ID")
    ocreate.add_argument("--cargo", type=float, required=True, help="货物量")
    ocreate.add_argument("--priority", type=int, default=1, help="优先级(越高越优先)")
    ocreate.add_argument("--berth", type=int, help="指定泊位ID")
    ocreate.add_argument("--yard", type=int, help="指定堆场ID")
    ocreate.add_argument("--description", help="作业描述")
    
    ocomplete = op_sub.add_parser("complete", help="完成作业")
    ocomplete.add_argument("--id", type=int, required=True, help="作业ID")
    
    odelete = op_sub.add_parser("delete", help="删除作业")
    odelete.add_argument("--id", type=int, required=True, help="作业ID")
    
    oshow = op_sub.add_parser("show", help="查看作业详情")
    oshow.add_argument("--id", type=int, required=True, help="作业ID")
    
    yard_parser = subparsers.add_parser("yards", help="堆场管理")
    yard_sub = yard_parser.add_subparsers(dest="action", help="操作")
    
    ylist = yard_sub.add_parser("list", help="列出堆场")
    ylist.add_argument("--available", type=lambda x: x.lower() == "true", help="仅显示可用/不可用堆场")
    
    ycreate = yard_sub.add_parser("create", help="创建堆场")
    ycreate.add_argument("--name", required=True, help="堆场名称")
    ycreate.add_argument("--capacity", type=float, required=True, help="总容量")
    ycreate.add_argument("--unavailable", action="store_true", help="创建时标记为不可用")
    
    yupdate = yard_sub.add_parser("update", help="更新堆场")
    yupdate.add_argument("--id", type=int, required=True, help="堆场ID")
    yupdate.add_argument("--name", help="新名称")
    yupdate.add_argument("--capacity", type=float, help="新容量")
    yupdate.add_argument("--available", type=lambda x: x.lower() == "true", help="是否可用")
    
    ydelete = yard_sub.add_parser("delete", help="删除堆场")
    ydelete.add_argument("--id", type=int, required=True, help="堆场ID")
    
    yshow = yard_sub.add_parser("show", help="查看堆场详情")
    yshow.add_argument("--id", type=int, required=True, help="堆场ID")
    
    help_parser = subparsers.add_parser("help", help="显示帮助")
    
    args = parser.parse_args()
    
    if args.command is None or args.command == "help":
        cmd_help(args, None)
        return
    
    if not hasattr(args, "action") or args.action is None:
        parser.print_help()
        return
    
    client = PortClient()
    try:
        if args.command == "berths":
            cmd_berths(args, client)
        elif args.command == "ships":
            cmd_ships(args, client)
        elif args.command == "operations":
            cmd_operations(args, client)
        elif args.command == "yards":
            cmd_yards(args, client)
    finally:
        client.close()


if __name__ == "__main__":
    main()
